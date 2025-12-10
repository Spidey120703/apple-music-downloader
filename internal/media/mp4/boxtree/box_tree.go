package boxtree

import (
	"errors"
	"io"
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
