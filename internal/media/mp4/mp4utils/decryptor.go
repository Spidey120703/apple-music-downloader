package mp4utils

import (
	"downloader/internal/media/mp4/cmaf"
	"errors"
	"fmt"
	"io"

	"github.com/Spidey120703/go-mp4"
)

type IDecryptor interface {
	cmaf.IContext
	GetSamples() map[uint32][]Sample
	DecryptHeader([][]byte) error
	DecryptSegment(*cmaf.Segment, [][]byte) error
	DecryptFragment(cmaf.MovieFragmentBox, *mp4.Mdat, [][]byte) error
	DecryptSample(string, []Sample, *mp4.Tenc, *mp4.Senc, [][]byte) error
}

type DecryptContext struct {
	cmaf.Context
	Sub     IDecryptor
	Samples map[uint32][]Sample
}

func (ctx *DecryptContext) Initialize(input io.ReadSeeker) (err error) {
	if ctx.Samples == nil {
		ctx.Samples = make(map[uint32][]Sample)
	}
	return ctx.Context.Initialize(input)
}

func (ctx *DecryptContext) DecryptHeader(keys [][]byte) (err error) {

	{ // Ftyp
		switch ctx.Header.GetMediaType() {
		case cmaf.MediaTypeAudio:
			ctx.Header.Ftyp.MajorBrand = mp4.BrandM4A()
		case cmaf.MediaTypeVideo:
			ctx.Header.Ftyp.MajorBrand = mp4.BrandM4V()
		default:
		}
		ctx.Header.Ftyp.MinorVersion = 0
		ctx.Header.Ftyp.CompatibleBrands = []mp4.CompatibleBrandElem{
			{ctx.Header.Ftyp.MajorBrand},
			{mp4.BrandMP42()},
			{mp4.BrandISOM()},
			{[4]byte{}},
		}
	}

	{ // SampleEntry
		for _, trak := range ctx.Header.Moov.Trak {
			for _, entry := range trak.Mdia.Minf.Stbl.Stsd.Entries {
				if entry.Sinf == nil {
					continue
				}
				entry.Node.Info.Type = entry.Sinf.Frma.DataFormat
				switch ctx.Header.GetMediaType() {
				case cmaf.MediaTypeAudio:
					entry.AudioSampleEntry.AnyTypeBox.Type = entry.Sinf.Frma.DataFormat
				case cmaf.MediaTypeVideo:
					entry.VisualSampleEntry.AnyTypeBox.Type = entry.Sinf.Frma.DataFormat
				default:
					entry.SampleEntry.AnyTypeBox.Type = entry.Sinf.Frma.DataFormat
				}
				if _, err = entry.Node.Remove(mp4.BoxTypeSinf()); err != nil {
					return
				}
			}
		}
	}

	{ // Pssh
		if _, err = ctx.Header.Moov.Node.Remove(mp4.BoxTypePssh()); err != nil {
			return
		}
	}

	if len(ctx.Header.Moof) != len(ctx.Header.Mdat) {
		return cmaf.ErrFragmentationMismatch
	}
	for idx := range ctx.Header.Moof {
		if err = ctx.DecryptFragment(ctx.Header.Moof[idx], ctx.Header.Mdat[idx], keys); err != nil {
			return
		}
	}
	return
}

func (ctx *DecryptContext) DecryptSegment(seg *cmaf.Segment, keys [][]byte) (err error) {

	if len(seg.Moof) != len(seg.Mdat) {
		return cmaf.ErrFragmentationMismatch
	}
	for idx := range seg.Moof {
		if err = ctx.DecryptFragment(seg.Moof[idx], seg.Mdat[idx], keys); err != nil {
			return
		}
	}

	return
}

func (ctx *DecryptContext) Finalize(output io.WriteSeeker) (err error) {
	return ctx.Context.Finalize(output)
}

type TrackCryptoInfo struct {
	Sinf *cmaf.ProtectionSchemeInformationBox
	Trex *mp4.Trex
	Pssh []*mp4.Pssh
}

func (t *TrackCryptoInfo) IsEncrypted() bool {
	return t.Sinf != nil
}

func (ctx *DecryptContext) GetTrackCryptoInfo(trackID uint32) (info TrackCryptoInfo, err error) {
	for _, trak := range ctx.Header.Moov.Trak {
		if trak.Tkhd.TrackID == trackID {
			for _, entry := range trak.Mdia.Minf.Stbl.Stsd.Entries {
				info.Sinf = entry.Sinf
			}
		}
	}
	for _, trex := range ctx.Header.Moov.Mvex.Trex {
		if trex.TrackID == trackID {
			info.Trex = trex
		}
	}
	info.Pssh = ctx.Header.Moov.Pssh
	return
}

func (ctx *DecryptContext) DecryptFragment(moof cmaf.MovieFragmentBox, mdat *mp4.Mdat, keys [][]byte) (err error) {
	var totalDeleted, size uint64
	if size, err = moof.Node.Remove(mp4.BoxTypePssh()); err != nil {
		return
	} else {
		totalDeleted += size
	}
	var samples []Sample
	for _, traf := range moof.Traf {
		var cryptoInfo TrackCryptoInfo
		if cryptoInfo, err = ctx.GetTrackCryptoInfo(traf.Tfhd.TrackID); err != nil {
			return
		}
		samples = GetFullSamples(traf, mdat, cryptoInfo.Trex)
		if !cryptoInfo.IsEncrypted() {
			ctx.Samples[traf.Tfhd.TrackID] = append(ctx.Samples[traf.Tfhd.TrackID], samples...)
			continue
		}
		schemeType := string(cryptoInfo.Sinf.Schm.SchemeType[:])
		if schemeType != "cbcs" {
			return fmt.Errorf("unsupported scheme type: %s", schemeType)
		}
		if err = ctx.DecryptSample(schemeType, samples, cryptoInfo.Sinf.Schi.Tenc, traf.Senc, keys); err != nil {
			return
		}
		ctx.Samples[traf.Tfhd.TrackID] = append(ctx.Samples[traf.Tfhd.TrackID], samples...)
		for _, boxType := range []mp4.BoxType{
			mp4.BoxTypeSaiz(),
			mp4.BoxTypeSaio(),
			mp4.BoxTypeSenc(),
			mp4.BoxTypeSbgp(),
			mp4.BoxTypeSgpd(),
		} {
			if size, err = traf.Node.Remove(boxType); err != nil {
				return
			} else {
				totalDeleted += size
			}
		}
	}
	for _, traf := range moof.Traf {
		for _, trun := range traf.Trun {
			trun.DataOffset -= int32(totalDeleted)
		}
	}
	return
}

func (ctx *DecryptContext) DecryptSample(schemeType string, samples []Sample, tenc *mp4.Tenc, senc *mp4.Senc, keys [][]byte) error {
	if ctx.Sub == nil {
		panic(errors.New("DecryptSample not implemented"))
	}
	return ctx.Sub.DecryptSample(schemeType, samples, tenc, senc, keys)
}

func (ctx *DecryptContext) GetSamples() map[uint32][]Sample {
	return ctx.Samples
}
