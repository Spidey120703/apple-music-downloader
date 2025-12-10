package boxtree

import (
	"errors"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Spidey120703/go-mp4"
)

type BoxNode struct {
	Info     *mp4.BoxInfo
	Box      mp4.IBox
	Path     mp4.BoxPath
	Parent   *BoxNode
	Children []*BoxNode
	Cache    map[mp4.BoxType][]*BoxNode
}

func (n *BoxNode) IsLeaf() bool {
	return len(n.Children) == 0
}

func (n *BoxNode) IsRoot() bool {
	return n.Parent == nil
}

func (n *BoxNode) Caching() (err error) {
	if n.IsLeaf() {
		return
	}
	n.Cache = make(map[mp4.BoxType][]*BoxNode)
	for i := range n.Children {
		n.Children[i].Parent = n
		if n.Children[i].Info == nil {
			return errors.New("child info is nil")
		}
		boxType := n.Children[i].Info.Type
		_, found := n.Cache[boxType]
		if !found {
			n.Cache[boxType] = make([]*BoxNode, 0)
		}
		n.Cache[boxType] = append(n.Cache[boxType], n.Children[i])
	}
	return
}

func (n *BoxNode) P(path string) (forest []*BoxNode, err error) {
	node := n
	parts := strings.Split(strings.Trim(path, ". "), ".")
	for _, p := range parts {
		var idx = -1
		if strings.Contains(p, "[") && strings.Contains(p, "]") {
			buf := strings.Split(p, "[")
			p = buf[0]
			idx, err = strconv.Atoi(buf[1][:len(buf[1])-1])
			if err != nil {
				return
			}
		}
		nodes, found := node.Cache[mp4.StrToBoxType(p)]
		if !found || len(nodes) == 0 {
			return nil, errors.New("not found " + p)
		}
		if idx == -1 {
			forest = nodes
			node = nodes[0]
		} else {
			if idx < 0 || idx >= len(nodes) {
				return nil, errors.New("index out of range for " + p)
			}
			forest = []*BoxNode{nodes[idx]}
			node = nodes[idx]
		}
	}
	return
}

func (n *BoxNode) Stringify() (str string) {
	if len(n.Path) > 0 {
		var boxStr string
		var err error
		tabs := strings.Repeat("  ", len(n.Path)-1)
		if n.Info.Type == mp4.BoxTypeMdat() {
			boxStr = "Data=[...]"
		} else {
			boxStr, err = mp4.Stringify(n.Box, n.Info.Context)
		}
		if err != nil {
			str += tabs + "[" + n.Info.Type.String() + "]\n"
		} else {
			str += tabs + "[" + n.Info.Type.String() + "] " + boxStr + "\n"
		}
	}
	for _, child := range n.Children {
		str += child.Stringify()
	}
	return str
}

func (n *BoxNode) Remove(boxType mp4.BoxType) (size uint64, err error) {
	for _, child := range n.Children {
		if child.Info.Type == boxType {
			size += child.Info.Size
		}
	}
	n.Children = slices.DeleteFunc(n.Children, func(node *BoxNode) bool {
		return node.Info.Type == boxType
	})
	err = n.Caching()
	return
}

func (n *BoxNode) Append(boxType mp4.BoxType, box mp4.IBox) (err error) {
	n.Children = append(n.Children, &BoxNode{
		Info: &mp4.BoxInfo{Type: boxType},
		Box:  box,
		Path: ToAppendedPath(n.Path, boxType),
	})
	err = n.Caching()
	return
}

func (n *BoxNode) Insert(idx int, boxType mp4.BoxType, box mp4.IBox) (err error) {
	n.Children = slices.Insert(n.Children, idx, &BoxNode{
		Info: &mp4.BoxInfo{Type: boxType},
		Box:  box,
		Path: ToAppendedPath(n.Path, boxType),
	})
	err = n.Caching()
	return
}

func ToAppendedPath(path mp4.BoxPath, boxType ...mp4.BoxType) (target mp4.BoxPath) {
	for _, p := range path {
		target = append(target, p)
	}
	for _, bt := range boxType {
		target = append(target, bt)
	}
	return target
}

func Unmarshal(reader io.ReadSeeker) (*BoxNode, error) {
	var convert = func(any []interface{}) []*BoxNode {
		if len(any) == 0 {
			return nil
		}
		nodes := make([]*BoxNode, len(any))
		for i := range any {
			nodes[i] = any[i].(*BoxNode)
		}
		return nodes
	}
	var handler = func(handle *mp4.ReadHandle) (interface{}, error) {
		node := &BoxNode{Info: &handle.BoxInfo, Path: handle.Path}
		if payload, _, err := handle.ReadPayload(); err != nil {
			return nil, err
		} else {
			node.Box = payload
		}
		if expand, err := handle.Expand(); err != nil {
			return nil, err
		} else {
			node.Children = convert(expand)
		}
		if err := node.Caching(); err != nil {
			// println(handle.Path[len(handle.Path)-1].String())
			return nil, err
		}
		return node, nil
	}
	if vals, err := mp4.ReadBoxStructure(reader, handler); err != nil {
		return nil, err
	} else {
		children := convert(vals)
		node := &BoxNode{Children: children, Path: mp4.BoxPath{}}
		err = node.Caching()
		return node, err
	}
}

func Marshal(writer io.WriteSeeker, root *BoxNode) (n uint64, err error) {
	w := mp4.NewWriter(writer)

	var handler func(*BoxNode) (uint64, error)
	handler = func(node *BoxNode) (n uint64, err error) {
		var b uint64
		var boxInfo *mp4.BoxInfo
		if !node.IsRoot() {
			if boxInfo, err = w.StartBox(node.Info); err != nil {
				return
			}

			if b, err = mp4.Marshal(w, node.Box, node.Info.Context); err != nil {
				return
			}
			n += boxInfo.HeaderSize + b
		}

		for _, child := range node.Children {
			if b, err = handler(child); err != nil {
				return
			}
			n += b
		}

		if !node.IsRoot() {
			if boxInfo, err = w.EndBox(); err != nil {
				return
			}
			node.Info.Offset = boxInfo.Offset
			node.Info.Size = boxInfo.Size
			node.Info.HeaderSize = boxInfo.HeaderSize
			node.Info.Type = boxInfo.Type
			node.Info.ExtendToEOF = boxInfo.ExtendToEOF
		}

		return
	}

	return handler(root)
}

func Main() {
	//f, err := os.Open("temp/P918331953_A1770791065_audio_en_gr2304_alac_m.mp4")
	//f, err := os.Open("temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-0.mp4")
	f, err := os.Open("temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--0.mp4")
	//f, err := os.Open("Downloads/Coldplay/2014-05-19 - Ghost Stories [825646299133]/Disc 1/1. Always In My Head.m4a")
	if err != nil {
		panic(err)
	}
	w, err := os.Create("a.m4a")
	if err != nil {
		panic(err)
	}
	defer w.Close()
	defer f.Close()
	//isobmff := ISOBaseMediaFileFormat{}
	//err = Unmarshal(f, &isobmff)
	//if err != nil {
	//	panic(err)
	//}
	//println(isobmff.Ftyp.Info.Type.String())
	//println(string(isobmff.Ftyp.Payload.MajorBrand[:]))
	//println(isobmff.Moov.Info.Type.String())
	//println(isobmff.Moov.Mvhd.Info.Type.String())
	//println(isobmff.Moov.Trak[0].Mdia.Hdlr.Payload.Name)
	//println(isobmff.Moov.Trak[0].Mdia.Minf.Stbl.Stco.Payload.EntryCount)

	root, err := Unmarshal(f)
	if err != nil {
		panic(err)
	}
	println(root.Stringify())
	n, err := Marshal(w, root)
	if err != nil {
		panic(err)
	}
	println(n)

	fst, err := root.P("moov.trak.mdia.minf.stbl.stsd")
	println(fst[0].Box.(*mp4.Stsd).EntryCount)
	for _, node := range fst[0].Children {
		println(node.Box.GetType().String())
		enca := node.Box.(*mp4.AudioSampleEntry)
		fst, _ := node.P("sinf.schi.tenc")
		tenc := fst[0].Box.(*mp4.Tenc)
		println(string(tenc.DefaultConstantIV[:]))
		println(enca.SampleSize)
		println()
	}

	println(ValidateISOBMFF(root))
	fst, err = root.P("moof[3]")
	println(fst)

	/*
		fst, err = root.P("moov.trak.udta.ludt")
		println("LoudnessBases: []LoudnessBase{")
		for _, node := range fst[0].Children {
			println(node.Info.Type.String())
			loudness := node.Box.(*mp4.LoudnessEntry)
			println(loudness.Version)
			println(loudness.Flags[2])
			println(loudness.LoudnessBaseCount)
			for _, base := range loudness.LoudnessBases {
				println("EQSetID:", base.EQSetID, ",")
				println("DownmixID:", base.DownmixID, ",")
				println("DRCSetID:", base.DRCSetID, ",")
				println("BsSamplePeakLevel:", base.BsSamplePeakLevel, ",")
				println("BsTruePeakLevel:", base.BsTruePeakLevel, ",")
				println("MeasurementSystemForTP:", base.MeasurementSystemForTP, ",")
				println("ReliabilityForTP:", base.ReliabilityForTP, ",")
				println("MeasurementCount:", base.MeasurementCount, ",")
				println("Measurements: []LoudnessMeasurement{")

				for _, measure := range base.Measurements {
					println("{")
					println("    MethodDefinition:", measure.MethodDefinition, ",")
					println("    MethodValue:", measure.MethodValue, ",")
					println("    MeasurementSystem:", measure.MeasurementSystem, ",")
					println("    Reliability:", measure.Reliability, ",")
					println("},")
				}

				println("},")
			}
			println("},")
		}
	*/

	//fst, err = root.P("moof.traf.senc")
	//senc := fst[0].Box.(*box_types.Senc)
	//for _, sample := range senc.SampleEntries {
	//	for _, subsample := range sample.SubsampleEntries {
	//		println(subsample.BytesOfProtectedData, subsample.BytesOfClearData)
	//	}
	//}
	//fst, err := forest[1].Path("udta.meta.ilst.sonm.data")
	//if err != nil {
	//	panic(err)
	//}
	//println(string(fst[0].Box.(*mp4.Data).Data))
}
