package mp4utils

import (
	"downloader/mp4/cmaf"
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

func testDecrypt(readers []io.ReadSeekCloser, writer io.WriteSeeker, key [][]byte) (muxer *MuxContext) {
	var err error

	ctx := DecryptContext{}
	if err = ctx.Initialize(readers[0]); err != nil {
		panic(err)
	}

	if err = ctx.DecryptHeader(key); err != nil {
		panic(err)
	}

	for _, reader := range readers[1:] {
		var seg *cmaf.Segment
		if seg, err = ctx.Context.MergeSegment(reader); err != nil {
			panic(err)
		}
		if err = ctx.DecryptSegment(seg, key); err != nil {
			panic(err)
		}
	}

	muxer = NewMuxContext()
	muxer.Context = ctx.Context
	muxer.Samples = ctx.Samples
	if err = muxer.Desegmentize(); err != nil {
		panic(err)
	}
	return
}

func Main() {
	//input, err := os.Open("temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-0.mp4")
	/*
		input, err := os.Open("temp/P918331953_A1770791065_audio_en_gr2304_alac_m.mp4")
		if err != nil {
			panic(err)
		}

		ctx := DecryptContext{
			Parameters: FairPlayParameters{
				AdamID: "1770791065",
			},
			DecryptSample: ADecryptSample,
		}
		if err = ctx.Initialize(input); err != nil {
			panic(err)
		}

		if err = ctx.DecryptHeader(
			[][]byte{
				[]byte("skd://itunes.apple.com/P000000000/s1/e1"),
				[]byte("skd://itunes.apple.com/p918331953/c23"),
			}); err != nil {
			panic(err)
		}

		muxer := NewMuxContext()
		muxer.Context = ctx.Context
		muxer.Samples = ctx.Samples
		if err = muxer.Desegmentize(); err != nil {
			panic(err)
		}

		mp4Output, err := os.Create("alac.mp4")
		if err != nil {
			panic(err)
		}
		defer utils.CloseQuietly(mp4Output)
		if err = ctx.Finalize(mp4Output); err != nil {
			panic(err)
		}
	*/

	/*
		if true {
			// Video
			var videoFiles []io.ReadSeekCloser
			for _, path := range []string{
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--0.mp4",
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--1.m4s",
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--2.m4s",
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--3.m4s",
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--4.m4s",
				"Temp/P915444077_A1770791066_MV_video_gr290_sdr_1488x1080_cbcs_--5.m4s",
			} {
				input, err := os.Open(path)
				if err != nil {
					panic(err)
				}
				videoFiles = append(videoFiles, input)
			}
			defer utils.CloseQuietlyAll(videoFiles)
			videoOutput, err := os.Create("v.mp4")
			if err != nil {
				panic(err)
			}
			defer utils.CloseQuietly(videoOutput)

			videoKey := [][]byte{{0x52, 0xc1, 0xd0, 0x69, 0x85, 0x9d, 0x50, 0xdc, 0xa1, 0x1c, 0x1d, 0x36, 0xbd, 0x27, 0xc7, 0x0e}}
			println(hex.EncodeToString(videoKey[0]))
			videoMuxer := decrypt(videoFiles, videoOutput, videoKey)

			// Audio
			var audioFiles []io.ReadSeekCloser
			for _, path := range []string{
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-0.mp4",
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-1.m4s",
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-2.m4s",
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-3.m4s",
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-4.m4s",
				"Temp/P915444077_A1770791066_audio_en_gr256_mp4a-40-2-5.m4s",
			} {
				input, err := os.Open(path)
				if err != nil {
					panic(err)
				}
				audioFiles = append(audioFiles, input)
			}
			defer utils.CloseQuietlyAll(audioFiles)
			audioOutput, err := os.Create("a.mp4")
			if err != nil {
				panic(err)
			}
			defer utils.CloseQuietly(audioOutput)

			audioKey := [][]byte{{0xec, 0x7a, 0xd3, 0x42, 0x2d, 0xaf, 0xd6, 0xeb, 0xd1, 0xce, 0x8f, 0x3a, 0xba, 0xa0, 0xe2, 0x0d}}
			println(hex.EncodeToString(audioKey[0]))
			audioMuxer := decrypt(audioFiles, audioOutput, audioKey)

			title := "123"
			md := metadata.Metadata{
				Title: &title,
			}
			err = md.Attach(audioMuxer.Root)
			if err != nil {
				panic(err)
			}

			err = audioMuxer.MuxTrack(videoMuxer)
			if err != nil {
				panic(err)
			}
			mp4Output, err := os.Create("mv.mp4")
			if err != nil {
				panic(err)
			}
			defer utils.CloseQuietly(mp4Output)

			println(audioMuxer.Header.Node.Stringify())
			if err = audioMuxer.Finalize(mp4Output); err != nil {
				panic(err)
			}
		}
	*/

	//println(string(Header.Ftyp.MajorBrand[:]))
	//println(len(Header.Moov.Trak[0].Mdia.Minf.Stbl.Sgpd))
	//println(Header.Moov.Trak[0].Mdia.Minf.Stbl.Sgpd[0].DefaultSampleDescriptionIndex)
	//println(Header.Moov.Trak[0].Mdia.Minf.Stbl.Sgpd[0].DefaultLength)
	//println(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].AudioSampleEntry.GetSampleRateInt())
	//println(string(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].Sinf.Frma.DataFormat[:]))
	//println(string(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].Sinf.Schm.SchemeType[:]))
	//println(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].Sinf.Schm.SchemeUri)
	//println(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].Sinf.Schm.SchemeVersion)
	//println(hex.Dump(Header.Moov.Trak[0].Mdia.Minf.Stbl.Stsd.Entries[0].Sinf.Schi.Tenc.DefaultKID[:]))
	//println(Header.Moof[0].Traf.Saio[0].AuxInfoTypeParameter)
	//println(Header.Moof[0].Traf.Senc.SampleEntries[0].SubsampleCount)
}
