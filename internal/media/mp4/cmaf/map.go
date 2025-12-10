package cmaf

import (
	"downloader/internal/media/mp4/boxtree"
	"errors"
	"slices"

	"github.com/Spidey120703/go-mp4"
)

type ProtectionSchemeInformationBox struct {
	Frma *mp4.Frma
	Schm *mp4.Schm
	Schi struct {
		Tenc *mp4.Tenc
	}
}

type ProtectedSampleEntry struct {
	SampleEntry                      *mp4.SampleEntry
	AudioSampleEntry                 *mp4.AudioSampleEntry
	VisualSampleEntry                *mp4.VisualSampleEntry
	ClosedCaptionSubtitleSampleEntry *mp4.ClosedCaptionSubtitleSampleEntry
	Sinf                             *ProtectionSchemeInformationBox
	Node                             *boxtree.BoxNode
}

type TrackBox struct {
	Tkhd *mp4.Tkhd
	Mdia struct {
		Mdhd *mp4.Mdhd
		Hdlr *mp4.Hdlr
		Minf struct {
			Vmhd *mp4.Vmhd
			Smhd *mp4.Smhd
			Nmhd *mp4.Nmhd
			Dinf struct {
				Dref *mp4.Dref
			}
			Stbl struct {
				Stsd struct {
					Box     *mp4.Stsd
					Entries []ProtectedSampleEntry
				}
				Stts *mp4.Stts
				Stsc *mp4.Stsc
				Stsz *mp4.Stsz
				Stz2 *mp4.Stz2
				Stco *mp4.Stco
				Co64 *mp4.Co64
				Sbgp []*mp4.Sbgp
				Sgpd []*mp4.Sgpd
				Saiz []*mp4.Saiz
				Saio []*mp4.Saio
				Stss *mp4.Stss
				Sdtp *mp4.Sdtp
				Ctts *mp4.Ctts
				Node *boxtree.BoxNode
			}
		}
	}
	Udta struct {
		Ludt struct {
			Tlou *mp4.LoudnessEntry
			Alou *mp4.LoudnessEntry
		}
	}
	Node *boxtree.BoxNode
}

type TrackFragmentBox struct {
	Tfhd *mp4.Tfhd
	Tfdt *mp4.Tfdt
	Trun []*mp4.Trun
	Senc *mp4.Senc
	Saiz []*mp4.Saiz
	Saio []*mp4.Saio
	Sbgp []*mp4.Sbgp
	Sgpd []*mp4.Sgpd
	Node *boxtree.BoxNode
}

type MovieFragmentBox struct {
	Mfhd *mp4.Mfhd
	Pssh []*mp4.Pssh
	Traf []TrackFragmentBox
	Node *boxtree.BoxNode
}

type Header struct {
	Ftyp *mp4.Ftyp
	Moov struct {
		Mvhd *mp4.Mvhd
		Trak []TrackBox
		Udta struct {
			Swre *mp4.Swre
			Node *boxtree.BoxNode
		}
		Mvex struct {
			Trex []*mp4.Trex
		}
		Pssh []*mp4.Pssh
		Node *boxtree.BoxNode
	}
	Moof []MovieFragmentBox
	Mdat []*mp4.Mdat
	Node *boxtree.BoxNode
}

type Segment struct {
	Styp *mp4.Styp
	Sidx []*mp4.Sidx
	Moof []MovieFragmentBox
	Mdat []*mp4.Mdat
	Node *boxtree.BoxNode
}

type MediaType int

const (
	MediaTypeOther MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
)

func (isom *Header) GetMediaType() MediaType {
	if slices.ContainsFunc(isom.Moov.Trak, func(box TrackBox) bool { return box.Mdia.Minf.Vmhd != nil }) {
		return MediaTypeVideo
	} else if !slices.ContainsFunc(isom.Moov.Trak, func(box TrackBox) bool { return box.Mdia.Minf.Smhd == nil }) {
		return MediaTypeAudio
	}
	return MediaTypeOther
}

var ErrIllegalISOBMFF = errors.New("this file is not an ISO Base Media File")
var ErrFragmentationMismatch = errors.New("number of moof and mdat boxes do not match")

func assignOnce[T mp4.IBox](root *boxtree.BoxNode, target *T, path string) {
	forest, err := root.P(path)
	if err != nil || len(forest) != 1 {
		// println("debug: not found", path)
		return
	}
	*target = forest[0].Box.(T)
}

func assignMore[T mp4.IBox](root *boxtree.BoxNode, target *[]T, path string) {
	forest, err := root.P(path)
	if err != nil {
		// println("debug: not found", path)
		return
	}
	for _, box := range forest {
		*target = append(*target, box.Box.(T))
	}
}

func checkMediaHeader(trak *TrackBox) bool {
	cnt := 0
	if trak.Mdia.Minf.Smhd != nil {
		cnt += 1
	}
	if trak.Mdia.Minf.Vmhd != nil {
		cnt += 1
	}
	if trak.Mdia.Minf.Nmhd != nil {
		cnt += 1
	}
	return cnt == 1
}

func InitializeHeader(root *boxtree.BoxNode) (hdr *Header, err error) {
	if !boxtree.ValidateISOBMFF(root) {
		err = ErrIllegalISOBMFF
		return
	}
	hdr = &Header{Node: root}

	assignOnce(root, &hdr.Ftyp, "ftyp")
	assignOnce(root, &hdr.Moov.Mvhd, "moov.mvhd")

	trakNodes, _ := root.P("moov.trak")
	for _, trackNode := range trakNodes {
		trak := TrackBox{Node: trackNode}
		assignOnce(trackNode, &trak.Tkhd, "tkhd")
		assignOnce(trackNode, &trak.Mdia.Mdhd, "mdia.mdhd")
		assignOnce(trackNode, &trak.Mdia.Hdlr, "mdia.hdlr")
		assignOnce(trackNode, &trak.Mdia.Minf.Smhd, "mdia.minf.smhd")
		assignOnce(trackNode, &trak.Mdia.Minf.Vmhd, "mdia.minf.vmhd")
		assignOnce(trackNode, &trak.Mdia.Minf.Nmhd, "mdia.minf.nmhd")
		assignOnce(trackNode, &trak.Mdia.Minf.Dinf.Dref, "mdia.minf.dinf.dref")

		stbl, _ := trackNode.P("mdia.minf.stbl")
		trak.Mdia.Minf.Stbl.Node = stbl[0]

		{ // mdia.minf.stbl.stsd
			if !checkMediaHeader(&trak) {
				return nil, errors.New("invalid track info")
			}
			stsd, _ := trackNode.P("mdia.minf.stbl.stsd")
			trak.Mdia.Minf.Stbl.Stsd.Box = stsd[0].Box.(*mp4.Stsd)
			for _, entryInfo := range stsd[0].Children {
				entry := ProtectedSampleEntry{}
				if trak.Mdia.Minf.Smhd != nil {
					entry.AudioSampleEntry = entryInfo.Box.(*mp4.AudioSampleEntry)
				} else if trak.Mdia.Minf.Vmhd != nil {
					entry.VisualSampleEntry = entryInfo.Box.(*mp4.VisualSampleEntry)
				} else if trak.Mdia.Minf.Nmhd != nil {
					entry.ClosedCaptionSubtitleSampleEntry = entryInfo.Box.(*mp4.ClosedCaptionSubtitleSampleEntry)
				} else {
					entry.SampleEntry = entryInfo.Box.(*mp4.SampleEntry)
				}
				sinfNodes, found := entryInfo.Cache[mp4.BoxTypeSinf()]
				if found && len(sinfNodes) > 0 {
					entry.Sinf = &ProtectionSchemeInformationBox{}
					assignOnce(sinfNodes[0], &entry.Sinf.Frma, "frma")
					assignOnce(sinfNodes[0], &entry.Sinf.Schm, "schm")
					assignOnce(sinfNodes[0], &entry.Sinf.Schi.Tenc, "schi.tenc")
				}
				entry.Node = entryInfo
				trak.Mdia.Minf.Stbl.Stsd.Entries = append(trak.Mdia.Minf.Stbl.Stsd.Entries, entry)
			}
		}
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stts, "mdia.minf.stbl.stts")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stsc, "mdia.minf.stbl.stsc")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stsz, "mdia.minf.stbl.stsz")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stz2, "mdia.minf.stbl.stz2")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stco, "mdia.minf.stbl.stco")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Co64, "mdia.minf.stbl.co64")
		assignMore(trackNode, &trak.Mdia.Minf.Stbl.Sbgp, "mdia.minf.stbl.sbgp")
		assignMore(trackNode, &trak.Mdia.Minf.Stbl.Sgpd, "mdia.minf.stbl.sgpd")
		assignMore(trackNode, &trak.Mdia.Minf.Stbl.Saiz, "mdia.minf.stbl.saiz")
		assignMore(trackNode, &trak.Mdia.Minf.Stbl.Saio, "mdia.minf.stbl.saio")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Stss, "mdia.minf.stbl.stss")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Sdtp, "mdia.minf.stbl.sdtp")
		assignOnce(trackNode, &trak.Mdia.Minf.Stbl.Ctts, "mdia.minf.stbl.ctts")
		assignOnce(trackNode, &trak.Udta.Ludt.Tlou, "udta.ludt.tlou")
		assignOnce(trackNode, &trak.Udta.Ludt.Alou, "udta.ludt.alou")
		hdr.Moov.Trak = append(hdr.Moov.Trak, trak)
	}

	udta, _ := root.P("moov.udta")
	hdr.Moov.Udta.Node = udta[0]

	assignOnce(root, &hdr.Moov.Udta.Swre, "moov.udta.swre")
	assignMore(root, &hdr.Moov.Mvex.Trex, "moov.mvex.trex")

	assignMore(root, &hdr.Moov.Pssh, "moov.pssh")
	hdr.Moov.Node = root.Cache[mp4.BoxTypeMoov()][0]

	if hdr.Moof, err = initializeMovieFragmentBox(root); err != nil {
		return
	}

	assignMore(root, &hdr.Mdat, "mdat")

	if len(hdr.Moof) > 0 && len(hdr.Moof) != len(hdr.Mdat) {
		err = ErrFragmentationMismatch
	}

	return
}

func InitializeSegment(root *boxtree.BoxNode) (seg *Segment, err error) {
	seg = &Segment{Node: root}

	assignOnce(root, &seg.Styp, "styp")
	assignMore(root, &seg.Sidx, "sidx")

	if seg.Moof, err = initializeMovieFragmentBox(root); err != nil {
		return
	}
	assignMore(root, &seg.Mdat, "mdat")

	if len(seg.Moof) != len(seg.Mdat) {
		err = ErrFragmentationMismatch
	}

	return
}

func initializeMovieFragmentBox(root *boxtree.BoxNode) (moofs []MovieFragmentBox, err error) {
	moofNodes, _ := root.P("moof")
	for _, moofNode := range moofNodes {
		moof := MovieFragmentBox{}
		assignOnce(moofNode, &moof.Mfhd, "mfhd")
		trafNodes, _ := moofNode.P("traf")
		for _, trafNode := range trafNodes {
			traf := TrackFragmentBox{}
			assignOnce(trafNode, &traf.Tfhd, "tfhd")
			assignOnce(trafNode, &traf.Tfdt, "tfdt")
			assignOnce(trafNode, &traf.Senc, "senc")
			assignMore(trafNode, &traf.Trun, "trun")
			assignMore(trafNode, &traf.Saiz, "saiz")
			assignMore(trafNode, &traf.Saio, "saio")
			assignMore(trafNode, &traf.Sbgp, "sbgp")
			assignMore(trafNode, &traf.Sgpd, "sgpd")
			traf.Node = trafNode
			moof.Traf = append(moof.Traf, traf)
		}
		moof.Node = moofNode
		moofs = append(moofs, moof)
	}
	return
}
