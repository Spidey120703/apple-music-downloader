package main

import (
	"downloader/applemusic"
	"downloader/avc"
	"downloader/itunes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"

	"github.com/abema/go-mp4"
)

type VideoInfo struct {
	readers struct {
		video []io.ReadSeeker
		audio []io.ReadSeeker
	}
	params struct {
		avc1 struct {
			encv *mp4.VisualSampleEntry
			avcC *mp4.AVCDecoderConfiguration
			colr *mp4.Colr
			fiel *mp4.Fiel
			chrm *avc.Chrm
			pasp *mp4.PixelAspectRatioBox
		}
		mp4a struct {
			enca *mp4.AudioSampleEntry
			esds *mp4.Esds
		}
	}
	chunk struct {
		video []Chunk
		audio []Chunk
	}
}

func (s *VideoInfo) VideoDuration() (ret uint64) {
	for _, chunk := range s.chunk.video {
		for _, sample := range chunk.samples {
			ret += uint64(sample.SampleDuration)
		}
	}
	return
}

func (s *VideoInfo) AudioDuration() (ret uint64) {
	for _, chunk := range s.chunk.audio {
		for _, sample := range chunk.samples {
			ret += uint64(sample.SampleDuration)
		}
	}
	return
}

func (s *VideoInfo) VideoSamples() (ret []Sample) {
	for _, chunk := range s.chunk.video {
		for _, sample := range chunk.samples {
			ret = append(ret, sample)
		}
	}
	return
}

func (s *VideoInfo) AudioSamples() (ret []Sample) {
	for _, chunk := range s.chunk.audio {
		for _, sample := range chunk.samples {
			ret = append(ret, sample)
		}
	}
	return
}

func extractAvc[T io.ReadSeeker](input []T, videoInfo *VideoInfo) error {

	videoInfo.readers.video = append(videoInfo.readers.video, input[0])

	// note: extracting avc atom
	{
		// Box Path: moov/trak/mdia/minf/stbl
		stbl, err := mp4.ExtractBox(input[0], nil, mp4.BoxPath{
			mp4.BoxTypeMoov(),
			mp4.BoxTypeTrak(),
			mp4.BoxTypeMdia(),
			mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(),
		})
		if err != nil || len(stbl) != 1 {
			return err
		}

		// Box Path: moov/trak/mdia/minf/stbl/stsd/encv
		encv, err := mp4.ExtractBoxWithPayload(input[0], stbl[0], mp4.BoxPath{
			mp4.BoxTypeStsd(),
			mp4.BoxTypeEncv(),
		})
		if err != nil || len(encv) == 0 {
			return err
		}
		videoInfo.params.avc1.encv = encv[0].Payload.(*mp4.VisualSampleEntry)

		avcC, err := mp4.ExtractBoxWithPayload(input[0], &encv[0].Info, mp4.BoxPath{mp4.BoxTypeAvcC()})
		if err != nil || len(avcC) != 1 {
			return err
		}
		videoInfo.params.avc1.avcC = avcC[0].Payload.(*mp4.AVCDecoderConfiguration)

		colr, err := mp4.ExtractBoxWithPayload(input[0], &encv[0].Info, mp4.BoxPath{mp4.BoxTypeColr()})
		if err != nil || len(colr) != 1 {
			return err
		}
		videoInfo.params.avc1.colr = colr[0].Payload.(*mp4.Colr)

		fiel, err := mp4.ExtractBoxWithPayload(input[0], &encv[0].Info, mp4.BoxPath{mp4.BoxTypeFiel()})
		if err != nil || len(colr) != 1 {
			return err
		}
		videoInfo.params.avc1.fiel = fiel[0].Payload.(*mp4.Fiel)

		chrm, err := mp4.ExtractBoxWithPayload(input[0], &encv[0].Info, mp4.BoxPath{avc.BoxTypeChrm()})
		if err != nil || len(colr) != 1 {
			return err
		}
		videoInfo.params.avc1.chrm = chrm[0].Payload.(*avc.Chrm)

		pasp, err := mp4.ExtractBoxWithPayload(input[0], &encv[0].Info, mp4.BoxPath{mp4.BoxTypePasp()})
		if err != nil || len(colr) != 1 {
			return err
		}
		videoInfo.params.avc1.pasp = pasp[0].Payload.(*mp4.PixelAspectRatioBox)
	}

	// Box Path: moov/mvex/trex
	trex, err := mp4.ExtractBoxWithPayload(input[0], nil, mp4.BoxPath{
		mp4.BoxTypeMoov(),
		mp4.BoxTypeMvex(),
		mp4.BoxTypeTrex(),
	})
	if err != nil || len(trex) != 1 {
		return err
	}
	trexPayload := trex[0].Payload.(*mp4.Trex)

	// note: extracting samples
	for _, segment := range input[1:] {
		videoInfo.readers.video = append(videoInfo.readers.video, segment)

		// Box Path: moof[]
		moofs, err := mp4.ExtractBox(segment, nil, mp4.BoxPath{mp4.BoxTypeMoof()})
		if err != nil || len(moofs) <= 0 {
			return err
		}
		Info.Printf("Found %d '%s' Boxes", len(moofs), mp4.BoxTypeMoof())

		// Box Path: mdat[]
		mdats, err := mp4.ExtractBoxWithPayload(segment, nil, mp4.BoxPath{mp4.BoxTypeMdat()})
		if err != nil || len(mdats) != len(moofs) {
			return err
		}
		Info.Printf("Found %d '%s' Boxes", len(mdats), mp4.BoxTypeMdat())

		for i, moof := range moofs {

			// Box Path: moof[]/traf/tfhd
			tfhd, err := mp4.ExtractBoxWithPayload(segment, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTfhd(),
			})
			if err != nil || len(tfhd) != 1 {
				return err
			}
			tfhdPayload := tfhd[0].Payload.(*mp4.Tfhd)

			sampleDescriptionIndex := tfhdPayload.SampleDescriptionIndex
			if sampleDescriptionIndex != 0 {
				sampleDescriptionIndex--
			}

			// Box Path: moof[]/traf/trun
			truns, err := mp4.ExtractBoxWithPayload(segment, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTrun(),
			})
			if err != nil || len(truns) <= 0 {
				return err
			}

			mdatPayloadData := mdats[i].Payload.(*mp4.Mdat).Data

			for _, trun := range truns {
				var chunk Chunk
				for _, entry := range trun.Payload.(*mp4.Trun).Entries {
					// Priority: `trun` > `tfhd` > `trex`
					// ISO/IEC 14496-12:
					// 	- `trun`: Section 8.8.8.1
					// 	- `tfhd`: Section 8.8.7.1

					sample := Sample{SampleDescriptionIndex: sampleDescriptionIndex}

					sampleSize := func() uint32 {
						if trun.Payload.CheckFlag(0x200) { // `trun` : sample‐size‐present
							return entry.SampleSize
						} else if tfhdPayload.CheckFlag(0x10) { // `tfhd` : default‐sample‐size‐present
							return tfhdPayload.DefaultSampleSize
						} else {
							return trexPayload.DefaultSampleSize
						}
					}()
					sample.Data = mdatPayloadData[:sampleSize]
					mdatPayloadData = mdatPayloadData[sampleSize:]

					sample.SampleDuration = func() uint32 {
						if trun.Payload.CheckFlag(0x100) { // `trun` : sample‐SampleDuration‐present
							return entry.SampleDuration
						} else if tfhdPayload.CheckFlag(0x8) { // `tfhd` : default‐sample‐SampleDuration‐present
							return tfhdPayload.DefaultSampleDuration
						} else {
							return trexPayload.DefaultSampleDuration
						}
					}()

					chunk.samples = append(chunk.samples, sample)
				}
				videoInfo.chunk.video = append(videoInfo.chunk.video, chunk)
			}
			if len(mdatPayloadData) != 0 {
				return errors.New("size mismatch")
			}
		}
	}
	return err
}

func extractMp4a[T io.ReadSeeker](input []T, videoInfo *VideoInfo) error {
	videoInfo.readers.audio = append(videoInfo.readers.audio, input[0])

	// note: extracting alac atom
	{
		// Box Path: moov/trak/mdia/minf/stbl
		stbl, err := mp4.ExtractBox(input[0], nil, mp4.BoxPath{
			mp4.BoxTypeMoov(),
			mp4.BoxTypeTrak(),
			mp4.BoxTypeMdia(),
			mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(),
		})
		if err != nil || len(stbl) != 1 {
			return err
		}

		// Box Path: moov/trak/mdia/minf/stbl/stsd/encv
		enca, err := mp4.ExtractBoxWithPayload(input[0], stbl[0], mp4.BoxPath{
			mp4.BoxTypeStsd(),
			mp4.BoxTypeEnca(),
		})
		if err != nil || len(enca) == 0 {
			return err
		}
		videoInfo.params.mp4a.enca = enca[0].Payload.(*mp4.AudioSampleEntry)

		esds, err := mp4.ExtractBoxWithPayload(input[0], &enca[0].Info, mp4.BoxPath{mp4.BoxTypeEsds()})
		if err != nil || len(esds) != 1 {
			return err
		}
		videoInfo.params.mp4a.esds = esds[0].Payload.(*mp4.Esds)
	}

	// Box Path: moov/mvex/trex
	trex, err := mp4.ExtractBoxWithPayload(input[0], nil, mp4.BoxPath{
		mp4.BoxTypeMoov(),
		mp4.BoxTypeMvex(),
		mp4.BoxTypeTrex(),
	})
	if err != nil || len(trex) != 1 {
		return err
	}
	trexPayload := trex[0].Payload.(*mp4.Trex)

	// note: extracting samples
	for _, segment := range input[1:] {
		videoInfo.readers.audio = append(videoInfo.readers.audio, segment)

		// Box Path: moof[]
		moofs, err := mp4.ExtractBox(segment, nil, mp4.BoxPath{mp4.BoxTypeMoof()})
		if err != nil || len(moofs) <= 0 {
			return err
		}
		Info.Printf("Found %d '%s' Boxes", len(moofs), mp4.BoxTypeMoof())

		// Box Path: mdat[]
		mdats, err := mp4.ExtractBoxWithPayload(segment, nil, mp4.BoxPath{mp4.BoxTypeMdat()})
		if err != nil || len(mdats) != len(moofs) {
			return err
		}
		Info.Printf("Found %d '%s' Boxes", len(mdats), mp4.BoxTypeMdat())

		for i, moof := range moofs {

			// Box Path: moof[]/traf/tfhd
			tfhd, err := mp4.ExtractBoxWithPayload(segment, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTfhd(),
			})
			if err != nil || len(tfhd) != 1 {
				return err
			}
			tfhdPayload := tfhd[0].Payload.(*mp4.Tfhd)

			sampleDescriptionIndex := tfhdPayload.SampleDescriptionIndex
			if sampleDescriptionIndex != 0 {
				sampleDescriptionIndex--
			}

			// Box Path: moof[]/traf/trun
			truns, err := mp4.ExtractBoxWithPayload(segment, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTrun(),
			})
			if err != nil || len(truns) <= 0 {
				return err
			}

			mdatPayloadData := mdats[i].Payload.(*mp4.Mdat).Data

			for _, trun := range truns {
				var chunk Chunk
				for _, entry := range trun.Payload.(*mp4.Trun).Entries {
					// Priority: `trun` > `tfhd` > `trex`
					// ISO/IEC 14496-12:
					// 	- `trun`: Section 8.8.8.1
					// 	- `tfhd`: Section 8.8.7.1

					sample := Sample{SampleDescriptionIndex: sampleDescriptionIndex}

					sampleSize := func() uint32 {
						if trun.Payload.CheckFlag(0x200) { // `trun` : sample‐size‐present
							return entry.SampleSize
						} else if tfhdPayload.CheckFlag(0x10) { // `tfhd` : default‐sample‐size‐present
							return tfhdPayload.DefaultSampleSize
						} else {
							return trexPayload.DefaultSampleSize
						}
					}()
					sample.Data = mdatPayloadData[:sampleSize]
					mdatPayloadData = mdatPayloadData[sampleSize:]

					sample.SampleDuration = func() uint32 {
						if trun.Payload.CheckFlag(0x100) { // `trun` : sample‐SampleDuration‐present
							return entry.SampleDuration
						} else if tfhdPayload.CheckFlag(0x8) { // `tfhd` : default‐sample‐SampleDuration‐present
							return tfhdPayload.DefaultSampleDuration
						} else {
							return trexPayload.DefaultSampleDuration
						}
					}()

					chunk.samples = append(chunk.samples, sample)
				}
				videoInfo.chunk.audio = append(videoInfo.chunk.audio, chunk)
			}
			if len(mdatPayloadData) != 0 {
				return errors.New("size mismatch")
			}
		}
	}
	return err
}

func writeMP4(
	writer *mp4.Writer,
	videoInfo *VideoInfo,
	itunesSongInfo *itunes.MusicVideo,
	song *applemusic.Songs,
	album *applemusic.Albums,
	videoData []byte,
	audioData []byte,
	coverData []byte) error {
	/*
		Mandatory Boxes of MPEG ISO Base Media File Format (ISO/IEC 14496-12):
			- ftyp 					- File Type Box
			+ moov					- Movie Box
				- mvhd				- Movie Header Box
				+ trak				- Track Box
					- tkhd			- Track Header Box
					+ mdia			- Media Box
						- mdhd		- Media Header Box
						- hdlr		- Handler Box
						+ minf		- Media Information Box
							- vmhd 	- Video Media Header Box (video track only)
							- smhd 	- Sound Media Header Box (sound track only)
							- hmhd 	- Hint Media Header Box (hint track only)
							- sthd 	- Subtitle Media Header Box (subtitle track only)
							- nmhd 	- Null Media Header Box (some tracks only)
							- dinf	- Data Information Box
							- dref	- Data Reference Box
						+ stbl		- Sample Table Box
							+ stsd	- Sample Description Box
							- stts	- Time to Sample Box
							- stsc	- Sample to Chunk Box
							- stsz	- Sample Sizes Box (framing only)
							- stco	- Chunk Offset Box
	*/
	{ // `ftyp` - File Type Box
		box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeFtyp()})
		if err != nil {
			return err
		}

		_, err = mp4.Marshal(writer, &mp4.Ftyp{
			MajorBrand:   [4]byte{'M', '4', 'A', ' '},
			MinorVersion: 0,
			CompatibleBrands: []mp4.CompatibleBrandElem{
				{[4]byte{'M', '4', 'A', ' '}},
				{[4]byte{'m', 'p', '4', '2'}},
				{mp4.BrandISOM()},
				{[4]byte{0, 0, 0, 0}},
			},
		}, box.Context)
		if err != nil {
			return err
		}

		_, err = writer.EndBox()
		if err != nil {
			return err
		}
	}

	const chunkSize uint32 = 5
	videoDuration := videoInfo.VideoDuration()
	audioDuration := videoInfo.AudioDuration()
	videoNumSamples := uint32(len(videoInfo.VideoSamples()))
	audioNumSamples := uint32(len(videoInfo.AudioSamples()))
	var videoStcoBoxInfo *mp4.BoxInfo
	var audioStcoBoxInfo *mp4.BoxInfo

	{ // `moov` - Movie Box
		_, err := writer.StartBox(&mp4.BoxInfo{
			Type: mp4.BoxTypeMoov(),
		})
		if err != nil {
			return err
		}

		// Box Path: moov
		box, err := mp4.ExtractBox(videoInfo.readers.video[0], nil, mp4.BoxPath{mp4.BoxTypeMoov()})
		if err != nil || len(box) != 1 {
			return err
		}
		originalVideoMoov := box[0]

		// Box Path: moov
		box, err = mp4.ExtractBox(videoInfo.readers.audio[0], nil, mp4.BoxPath{mp4.BoxTypeMoov()})
		if err != nil || len(box) != 1 {
			return err
		}
		originalAudioMoov := box[0]

		{ // `mvhd` - Movie Header Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMvhd()})
			if err != nil {
				return err
			}

			// Box Path: moov/mvhd
			box, err := mp4.ExtractBoxWithPayload(videoInfo.readers.video[0], originalVideoMoov, mp4.BoxPath{mp4.BoxTypeMvhd()})
			if err != nil || len(box) != 1 {
				return err
			}
			videoMvhdPayload := box[0].Payload.(*mp4.Mvhd)

			// Box Path: moov/mvhd
			box, err = mp4.ExtractBoxWithPayload(videoInfo.readers.audio[0], originalAudioMoov, mp4.BoxPath{mp4.BoxTypeMvhd()})
			if err != nil || len(box) != 1 {
				return err
			}
			audioMvhdPayload := box[0].Payload.(*mp4.Mvhd)

			mvhdPayload := mp4.Mvhd{}

			// ISO/IEC 14496-12 Section 6.2.2
			// Box Version:
			// 	- 0x000000 - 32-bit sizes fields
			// 	- 0x000001 - 64-bit sizes fields
			if videoMvhdPayload.Version == 1 && audioMvhdPayload.Version == 1 {
				mvhdPayload.CreationTimeV1 = videoMvhdPayload.CreationTimeV1
				mvhdPayload.ModificationTimeV1 = videoMvhdPayload.ModificationTimeV1
				mvhdPayload.DurationV1 = max(videoDuration, audioDuration)
			} else if videoMvhdPayload.Version == 0 && audioMvhdPayload.Version == 0 {
				mvhdPayload.CreationTimeV0 = videoMvhdPayload.CreationTimeV0
				mvhdPayload.ModificationTimeV0 = videoMvhdPayload.ModificationTimeV0
				mvhdPayload.DurationV0 = uint32(max(videoDuration, audioDuration))
			}

			mvhdPayload.Timescale = max(videoMvhdPayload.Timescale, audioMvhdPayload.Timescale)
			mvhdPayload.Rate = videoMvhdPayload.Rate
			mvhdPayload.Volume = videoMvhdPayload.Volume
			mvhdPayload.Matrix = videoMvhdPayload.Matrix
			mvhdPayload.NextTrackID = 3

			_, err = mp4.Marshal(writer, &mvhdPayload, box[0].Info.Context)
			if err != nil {
				return err
			}

			_, err = writer.EndBox()
			if err != nil {
				return err
			}
		}

		{ // `trak` - Track Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTrak()})
			if err != nil {
				return err
			}

			// Box Path: moov/trak
			box, err := mp4.ExtractBox(videoInfo.readers.video[0], originalVideoMoov, mp4.BoxPath{mp4.BoxTypeTrak()})
			if err != nil || len(box) != 1 {
				return err
			}
			originalTrak := box[0]

			{ // `tkhd` - Track Header Box
				_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTkhd()})
				if err != nil {
					return err
				}

				// Box Path: moov/trak/tkhd
				box, err := mp4.ExtractBoxWithPayload(videoInfo.readers.video[0], originalTrak, mp4.BoxPath{mp4.BoxTypeTkhd()})
				if err != nil || len(box) != 1 {
					return err
				}
				tkhdPayload := box[0].Payload.(*mp4.Tkhd)

				// ISO/IEC 14496-12 Section 6.2.2
				if tkhdPayload.Version == 1 {
					tkhdPayload.DurationV1 = videoDuration
				} else { // tkhd.Version == 0
					tkhdPayload.DurationV0 = uint32(videoDuration)
				}
				// ISO/IEC 14496-12 Section 8.3.2.3
				// track_enabled | track_in_movie | track_in_preview
				// 	- 0x000001 - track_enabled
				// 	- 0x000010 - track_in_movie
				// 	- 0x000100 - track_in_preview
				tkhdPayload.SetFlags(0x7)

				_, err = mp4.Marshal(writer, tkhdPayload, box[0].Info.Context)
				if err != nil {
					return err
				}

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			{ // `mdia` - Media Box
				_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdia()})
				if err != nil {
					return err
				}

				// Box Path: moov/trak/mdia
				box, err := mp4.ExtractBox(videoInfo.readers.video[0], originalTrak, mp4.BoxPath{mp4.BoxTypeMdia()})
				if err != nil || len(box) != 1 {
					return err
				}
				originalMdia := box[0]

				{ // `mdhd` - Media Header Box
					_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdhd()})
					if err != nil {
						return err
					}

					// Box Path: moov/trak/mdia/mdhd
					box, err := mp4.ExtractBoxWithPayload(videoInfo.readers.video[0], originalMdia, mp4.BoxPath{mp4.BoxTypeMdhd()})
					if err != nil || len(box) != 1 {
						return err
					}
					mdhdPayload := box[0].Payload.(*mp4.Mdhd)

					// ISO/IEC 14496-12 Section 6.2.2
					if mdhdPayload.Version == 1 {
						mdhdPayload.DurationV1 = videoDuration
					} else { // tkhd.Version == 0
						mdhdPayload.DurationV0 = uint32(videoDuration)
					}

					_, err = mp4.Marshal(writer, mdhdPayload, box[0].Info.Context)
					if err != nil {
						return err
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				{ // `hdlr` - Handler Box

					// Box Path: moov/trak/mdia/hdlr
					box, err = mp4.ExtractBox(videoInfo.readers.video[0], originalMdia, mp4.BoxPath{mp4.BoxTypeHdlr()})
					if err != nil {
						return err
					}

					err = writer.CopyBox(videoInfo.readers.video[0], box[0])
					if err != nil {
						return err
					}
				}

				{ // `minf` - Media Information Box
					_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMinf()})
					if err != nil {
						return err
					}

					// Box Path: moov/trak/mdia/minf
					box, err := mp4.ExtractBox(videoInfo.readers.video[0], originalMdia, mp4.BoxPath{mp4.BoxTypeMinf()})
					if err != nil || len(box) != 1 {
						return err
					}
					originalMinf := box[0]

					{ // `smhd` - Sound Media Header Box, `dinf` - Data Information Box

						// Box Path: moov/trak/mdia/minf/vmhd
						// 			 moov/trak/mdia/minf/dinf
						boxes, err := mp4.ExtractBoxes(videoInfo.readers.video[0], originalMinf, []mp4.BoxPath{
							{mp4.BoxTypeVmhd()},
							{mp4.BoxTypeDinf()},
						})
						if err != nil {
							return err
						}

						for _, box := range boxes {
							err = writer.CopyBox(videoInfo.readers.video[0], box)
							if err != nil {
								return err
							}
						}
					}

					{ // `stbl` - Sample Table Box
						_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStbl()})
						if err != nil {
							return err
						}

						{ // `stsd` - Sample Description Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsd()})
							if err != nil {
								return err
							}

							_, err = mp4.Marshal(writer, &mp4.Stsd{EntryCount: 1}, box.Context)
							if err != nil {
								return err
							}

							{ // Visual Sample Entry

								box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeAvc1()})
								if err != nil {
									return err
								}

								_, err = mp4.Marshal(writer, &mp4.VisualSampleEntry{
									SampleEntry: mp4.SampleEntry{
										AnyTypeBox: mp4.AnyTypeBox{
											Type: mp4.BoxTypeAvc1(),
										},
										DataReferenceIndex: 1,
									},
									Width:           videoInfo.params.avc1.encv.Width,
									Height:          videoInfo.params.avc1.encv.Height,
									Horizresolution: videoInfo.params.avc1.encv.Horizresolution,
									Vertresolution:  videoInfo.params.avc1.encv.Vertresolution,
									FrameCount:      videoInfo.params.avc1.encv.FrameCount,
									Compressorname:  videoInfo.params.avc1.encv.Compressorname,
									Depth:           videoInfo.params.avc1.encv.Depth,
								}, box.Context)
								if err != nil {
									return err
								}

								{ // `avcC`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeAvcC()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.avc1.avcC, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								{ // `colr`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeColr()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.avc1.colr, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								{ // `fiel`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeFiel()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.avc1.fiel, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								{ // `chrm`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: avc.BoxTypeChrm()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.avc1.chrm, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								{ // `pasp`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypePasp()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.avc1.pasp, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								_, err = writer.EndBox()
								if err != nil {
									return err
								}
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stts` - Time to Sample Box

							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStts()})
							if err != nil {
								return err
							}

							var stts mp4.Stts

							// Handle variable sample durations
							// 	- some samples may differ in duration (e.g., Variable Frame Rate content)
							for _, sample := range videoInfo.VideoSamples() {
								if len(stts.Entries) != 0 {
									last := &stts.Entries[len(stts.Entries)-1]
									if last.SampleDelta == sample.SampleDuration {
										last.SampleCount++
										continue
									}
								}
								stts.Entries = append(stts.Entries, mp4.SttsEntry{
									SampleCount: 1,
									SampleDelta: sample.SampleDuration,
								})
							}
							stts.EntryCount = uint32(len(stts.Entries))

							_, err = mp4.Marshal(writer, &stts, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stsc` - Sample to Chunk Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsc()})
							if err != nil {
								return err
							}

							stsc := mp4.Stsc{
								EntryCount: 1,
								Entries: []mp4.StscEntry{
									{
										FirstChunk:             1,
										SamplesPerChunk:        chunkSize,
										SampleDescriptionIndex: 1,
									},
								},
							}

							if videoNumSamples%chunkSize != 0 {
								stsc.Entries = append(stsc.Entries, mp4.StscEntry{
									FirstChunk:             videoNumSamples/chunkSize + 1,
									SamplesPerChunk:        videoNumSamples % chunkSize,
									SampleDescriptionIndex: 1,
								})
								stsc.EntryCount++
							}

							_, err = mp4.Marshal(writer, &stsc, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stsz` - Sample Sizes Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsz()})
							if err != nil {
								return err
							}

							stsz := mp4.Stsz{SampleCount: videoNumSamples}
							for _, sample := range videoInfo.VideoSamples() {
								stsz.EntrySize = append(stsz.EntrySize, uint32(len(sample.Data)))
							}

							_, err = mp4.Marshal(writer, &stsz, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stco` - Chunk Offset Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStco()})
							if err != nil {
								return err
							}

							numChunks := (videoNumSamples + chunkSize - 1) / chunkSize

							_, err = mp4.Marshal(writer, &mp4.Stco{
								EntryCount:  numChunks,
								ChunkOffset: make([]uint32, numChunks),
							}, box.Context)
							if err != nil {
								return err
							}

							videoStcoBoxInfo, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						_, err := writer.EndBox()
						if err != nil {
							return err
						}
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			_, err = writer.EndBox()
			if err != nil {
				return err
			}
		}

		{ // `trak` - Track Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTrak()})
			if err != nil {
				return err
			}

			// Box Path: moov/trak
			box, err := mp4.ExtractBox(videoInfo.readers.audio[0], originalAudioMoov, mp4.BoxPath{mp4.BoxTypeTrak()})
			if err != nil || len(box) != 1 {
				return err
			}
			originalTrak := box[0]

			{ // `tkhd` - Track Header Box
				_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTkhd()})
				if err != nil {
					return err
				}

				// Box Path: moov/trak/tkhd
				box, err := mp4.ExtractBoxWithPayload(videoInfo.readers.audio[0], originalTrak, mp4.BoxPath{mp4.BoxTypeTkhd()})
				if err != nil || len(box) != 1 {
					return err
				}
				tkhdPayload := box[0].Payload.(*mp4.Tkhd)

				// ISO/IEC 14496-12 Section 6.2.2
				if tkhdPayload.Version == 1 {
					tkhdPayload.DurationV1 = audioDuration
				} else { // tkhd.Version == 0
					tkhdPayload.DurationV0 = uint32(audioDuration)
				}
				// ISO/IEC 14496-12 Section 8.3.2.3
				// track_enabled | track_in_movie | track_in_preview
				// 	- 0x000001 - track_enabled
				// 	- 0x000010 - track_in_movie
				// 	- 0x000100 - track_in_preview
				tkhdPayload.SetFlags(0x7)

				tkhdPayload.TrackID = 2

				_, err = mp4.Marshal(writer, tkhdPayload, box[0].Info.Context)
				if err != nil {
					return err
				}

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			{ // `mdia` - Media Box
				_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdia()})
				if err != nil {
					return err
				}

				// Box Path: moov/trak/mdia
				box, err := mp4.ExtractBox(videoInfo.readers.audio[0], originalTrak, mp4.BoxPath{mp4.BoxTypeMdia()})
				if err != nil || len(box) != 1 {
					return err
				}
				originalMdia := box[0]

				{ // `mdhd` - Media Header Box
					_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdhd()})
					if err != nil {
						return err
					}

					// Box Path: moov/trak/mdia/mdhd
					box, err := mp4.ExtractBoxWithPayload(videoInfo.readers.audio[0], originalMdia, mp4.BoxPath{mp4.BoxTypeMdhd()})
					if err != nil || len(box) != 1 {
						return err
					}
					mdhdPayload := box[0].Payload.(*mp4.Mdhd)

					// ISO/IEC 14496-12 Section 6.2.2
					if mdhdPayload.Version == 1 {
						mdhdPayload.DurationV1 = audioDuration
					} else { // tkhd.Version == 0
						mdhdPayload.DurationV0 = uint32(audioDuration)
					}

					_, err = mp4.Marshal(writer, mdhdPayload, box[0].Info.Context)
					if err != nil {
						return err
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				{ // `hdlr` - Handler Box

					// Box Path: moov/trak/mdia/hdlr
					box, err = mp4.ExtractBox(videoInfo.readers.audio[0], originalMdia, mp4.BoxPath{mp4.BoxTypeHdlr()})
					if err != nil {
						return err
					}

					err = writer.CopyBox(videoInfo.readers.audio[0], box[0])
					if err != nil {
						return err
					}
				}

				{ // `minf` - Media Information Box
					_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMinf()})
					if err != nil {
						return err
					}

					// Box Path: moov/trak/mdia/minf
					box, err := mp4.ExtractBox(videoInfo.readers.audio[0], originalMdia, mp4.BoxPath{mp4.BoxTypeMinf()})
					if err != nil || len(box) != 1 {
						return err
					}
					originalMinf := box[0]

					{ // `smhd` - Sound Media Header Box, `dinf` - Data Information Box

						// Box Path: moov/trak/mdia/minf/smhd
						// 			 moov/trak/mdia/minf/dinf
						boxes, err := mp4.ExtractBoxes(videoInfo.readers.audio[0], originalMinf, []mp4.BoxPath{
							{mp4.BoxTypeSmhd()},
							{mp4.BoxTypeDinf()},
						})
						if err != nil {
							return err
						}

						for _, box := range boxes {
							err = writer.CopyBox(videoInfo.readers.audio[0], box)
							if err != nil {
								return err
							}
						}
					}

					{ // `stbl` - Sample Table Box
						_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStbl()})
						if err != nil {
							return err
						}

						{ // `stsd` - Sample Description Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsd()})
							if err != nil {
								return err
							}

							_, err = mp4.Marshal(writer, &mp4.Stsd{EntryCount: 1}, box.Context)
							if err != nil {
								return err
							}

							{ // Audio Sample Entry
								box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMp4a()})
								if err != nil {
									return err
								}

								_, err = mp4.Marshal(writer, &mp4.AudioSampleEntry{
									SampleEntry: mp4.SampleEntry{
										AnyTypeBox: mp4.AnyTypeBox{
											Type: mp4.BoxTypeMp4a(),
										},
										DataReferenceIndex: 1,
									},
									EntryVersion: videoInfo.params.mp4a.enca.EntryVersion,
									ChannelCount: videoInfo.params.mp4a.enca.ChannelCount,
									SampleSize:   videoInfo.params.mp4a.enca.SampleSize,
									PreDefined:   videoInfo.params.mp4a.enca.PreDefined,
									Reserved2:    videoInfo.params.mp4a.enca.Reserved2,
									SampleRate:   videoInfo.params.mp4a.enca.SampleRate,
								}, box.Context)
								if err != nil {
									return err
								}

								{ // `esds`
									box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeEsds()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, videoInfo.params.mp4a.esds, box.Context)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								_, err = writer.EndBox()
								if err != nil {
									return err
								}
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stts` - Time to Sample Box

							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStts()})
							if err != nil {
								return err
							}

							var stts mp4.Stts

							// Handle variable sample durations
							// 	- some samples may differ in duration (e.g., Variable Frame Rate content)
							for _, sample := range videoInfo.AudioSamples() {
								if len(stts.Entries) != 0 {
									last := &stts.Entries[len(stts.Entries)-1]
									if last.SampleDelta == sample.SampleDuration {
										last.SampleCount++
										continue
									}
								}
								stts.Entries = append(stts.Entries, mp4.SttsEntry{
									SampleCount: 1,
									SampleDelta: sample.SampleDuration,
								})
							}
							stts.EntryCount = uint32(len(stts.Entries))

							_, err = mp4.Marshal(writer, &stts, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stsc` - Sample to Chunk Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsc()})
							if err != nil {
								return err
							}

							stsc := mp4.Stsc{
								EntryCount: 1,
								Entries: []mp4.StscEntry{
									{
										FirstChunk:             1,
										SamplesPerChunk:        chunkSize,
										SampleDescriptionIndex: 1,
									},
								},
							}

							if audioNumSamples%chunkSize != 0 {
								stsc.Entries = append(stsc.Entries, mp4.StscEntry{
									FirstChunk:             audioNumSamples/chunkSize + 1,
									SamplesPerChunk:        audioNumSamples % chunkSize,
									SampleDescriptionIndex: 1,
								})
								stsc.EntryCount++
							}

							_, err = mp4.Marshal(writer, &stsc, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stsz` - Sample Sizes Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStsz()})
							if err != nil {
								return err
							}

							stsz := mp4.Stsz{SampleCount: audioNumSamples}
							for _, sample := range videoInfo.AudioSamples() {
								stsz.EntrySize = append(stsz.EntrySize, uint32(len(sample.Data)))
							}

							_, err = mp4.Marshal(writer, &stsz, box.Context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						{ // `stco` - Chunk Offset Box
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeStco()})
							if err != nil {
								return err
							}

							numChunks := (audioNumSamples + chunkSize - 1) / chunkSize

							_, err = mp4.Marshal(writer, &mp4.Stco{
								EntryCount:  numChunks,
								ChunkOffset: make([]uint32, numChunks),
							}, box.Context)
							if err != nil {
								return err
							}

							audioStcoBoxInfo, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						_, err := writer.EndBox()
						if err != nil {
							return err
						}
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			_, err = writer.EndBox()
			if err != nil {
				return err
			}
		}

		{ // `udta` - User Data Box
			context := mp4.Context{UnderUdta: true}

			_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeUdta(), Context: context})
			if err != nil {
				return err
			}

			if &context == nil { // `meta` - Meta Box
				context.UnderIlstMeta = true

				_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMeta(), Context: context})
				if err != nil {
					return err
				}

				_, err = mp4.Marshal(writer, &mp4.Meta{}, context)
				if err != nil {
					return err
				}

				{ // `hldr` - Handler Box
					_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeHdlr(), Context: context})
					if err != nil {
						return err
					}

					_, err = mp4.Marshal(writer, &mp4.Hdlr{
						HandlerType: [4]byte{'m', 'd', 'i', 'r'},
						Reserved:    [3]uint32{0x6170706c},
						Name:        "\x00",
					}, context)
					if err != nil {
						return err
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				{ // `ilst` - Ilst Box
					context.UnderIlst = true

					_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeIlst(), Context: context})
					if err != nil {
						return err
					}

					{
						MarshalData := func(value any) error {
							_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeData(), Context: context})
							if err != nil {
								return err
							}

							data := mp4.Data{}
							switch val := value.(type) {
							case string:
								data.DataType = mp4.DataTypeStringUTF8
								data.Data = []byte(val)
							case uint8:
								data.DataType = mp4.DataTypeSignedIntBigEndian
								data.Data = []byte{val}
							case uint32:
								data.DataType = mp4.DataTypeSignedIntBigEndian
								data.Data = make([]byte, 4)
								binary.BigEndian.PutUint32(data.Data, val)
							case []byte:
								data.DataType = mp4.DataTypeBinary
								data.Data = val
							default:
								return errors.New("unknown videoData type")
							}

							_, err = mp4.Marshal(writer, &data, context)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							return err
						}
						AddMeta := func(boxType mp4.BoxType, value any) error {
							_, err := writer.StartBox(&mp4.BoxInfo{Type: boxType, Context: context})
							if err != nil {
								return err
							}

							err = MarshalData(value)
							if err != nil {
								return err
							}

							_, err = writer.EndBox()
							return err
						}

						err = AddMeta(mp4.BoxType{'\xa9', 'n', 'a', 'm'}, *song.Attributes.Name)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'\xa9', 'A', 'R', 'T'}, *song.Attributes.ArtistName)
						if err != nil {
							return err
						}
						composerName := ""
						if song.Attributes.ComposerName != nil {
							composerName = *song.Attributes.ComposerName
						}
						err = AddMeta(mp4.BoxType{'\xa9', 'w', 'r', 't'}, composerName)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'\xa9', 'a', 'l', 'b'}, *song.Attributes.AlbumName)
						if err != nil {
							return err
						}

						if len(song.Attributes.GenreNames) > 0 {
							err = AddMeta(mp4.BoxType{'\xa9', 'g', 'e', 'n'}, itunesSongInfo.PrimaryGenreName)
							if err != nil {
								return err
							}
						}

						trkn := make([]byte, 8)
						disk := make([]byte, 8)
						binary.BigEndian.PutUint32(trkn[0:4], uint32(itunesSongInfo.TrackNumber))
						binary.BigEndian.PutUint16(trkn[4:6], uint16(itunesSongInfo.TrackCount))
						binary.BigEndian.PutUint32(disk[0:4], uint32(itunesSongInfo.DiscNumber))
						binary.BigEndian.PutUint16(disk[4:6], uint16(itunesSongInfo.DiscCount))
						err = AddMeta(mp4.BoxType{'t', 'r', 'k', 'n'}, trkn)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'d', 'i', 's', 'k'}, disk)
						if err != nil {
							return err
						}

						err = AddMeta(mp4.BoxType{'\xa9', 'd', 'a', 'y'}, *song.Attributes.ReleaseDate)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'c', 'p', 'i', 'l'}, uint8(0))
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'p', 'g', 'a', 'p'}, uint8(0))
						if err != nil {
							return err
						}

						{ // Cover
							_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxType{'c', 'o', 'v', 'r'}, Context: context})
							if err != nil {
								return err
							}

							{
								_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxType{'d', 'a', 't', 'a'}, Context: context})
								if err != nil {
									return err
								}

								data := mp4.Data{DataType: mp4.DataTypeStringJPEG, Data: coverData}

								_, err = mp4.Marshal(writer, &data, context)
								if err != nil {
									return err
								}

								_, err = writer.EndBox()
								if err != nil {
									return err
								}
							}

							_, err = writer.EndBox()
							if err != nil {
								return err
							}
						}

						err = AddMeta(mp4.BoxType{'r', 't', 'n', 'g'}, uint8(0))
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 't', 'i', 'k'}, uint8(1))
						if err != nil {
							return err
						}

						if len(song.Attributes.GenreNames) > 0 {
							err = AddMeta(mp4.BoxType{'g', 'e', 'I', 'D'}, uint32(GenreMapped[itunesSongInfo.PrimaryGenreName]))
							if err != nil {
								return err
							}
						}
						err = AddMeta(mp4.BoxType{'s', 'f', 'I', 'D'}, uint32(GetStoreFrontID(StoreFront)))
						if err != nil {
							return err
						}

						artistID, err := strconv.ParseUint(*song.Relationships.Artists.Data[0].ID, 10, 32)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'a', 't', 'I', 'D'}, uint32(artistID))
						if err != nil {
							return err
						}

						if len(song.Relationships.Composers.Data) > 0 {
							composerID, err := strconv.ParseUint(*song.Relationships.Composers.Data[0].ID, 10, 32)
							if err != nil {
								return err
							}
							err = AddMeta(mp4.BoxType{'c', 'm', 'I', 'D'}, uint32(composerID))
							if err != nil {
								return err
							}
						}

						playlistID, err := strconv.ParseUint(*album.ID, 10, 32)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'p', 'l', 'I', 'D'}, uint32(playlistID))
						if err != nil {
							return err
						}

						catalogID, err := strconv.ParseUint(*song.ID, 10, 32)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'c', 'n', 'I', 'D'}, uint32(catalogID))
						if err != nil {
							return err
						}

						err = AddMeta(mp4.BoxType{'a', 'A', 'R', 'T'}, *album.Attributes.ArtistName)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 'o', 'a', 'r'}, *song.Attributes.ArtistName)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 'o', 'n', 'm'}, *song.Attributes.Name)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 'o', 'a', 'l'}, *album.Attributes.Name)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 'o', 'c', 'o'}, composerName)
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'c', 'p', 'r', 't'}, *album.Attributes.Copyright)
						if err != nil {
							return err
						}
					}

					context.UnderIlst = false

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				{ // `free` - Free Box
					_, err := writer.StartBox(&mp4.BoxInfo{
						Type: mp4.BoxTypeFree(),
					})
					if err != nil {
						return err
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				context.UnderIlstMeta = false

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			context.UnderUdta = false

			_, err = writer.EndBox()
			if err != nil {
				return err
			}
		}

		_, err = writer.EndBox()
		if err != nil {
			return err
		}
	}

	{ // `free` - Free Box
		_, err := writer.StartBox(&mp4.BoxInfo{
			Type: mp4.BoxTypeFree(),
		})
		if err != nil {
			return err
		}

		_, err = writer.EndBox()
		if err != nil {
			return err
		}
	}

	{ // `mdat` - Media Data Box
		box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdat()})
		if err != nil {
			return err
		}

		_, err = mp4.Marshal(writer, &mp4.Mdat{Data: videoData}, box.Context)
		if err != nil {
			return err
		}

		_, err = mp4.Marshal(writer, &mp4.Mdat{Data: videoData}, box.Context)
		if err != nil {
			return err
		}

		_, err = mp4.Marshal(writer, &mp4.Mdat{Data: audioData}, box.Context)
		if err != nil {
			return err
		}

		mdatBoxInfo, err := writer.EndBox()
		if err != nil {
			return err
		}

		offset := mdatBoxInfo.Offset + mdatBoxInfo.HeaderSize
		{
			var stco mp4.Stco

			for i, sample := range videoInfo.VideoSamples() {
				if uint32(i)%chunkSize == 0 {
					stco.EntryCount++
					stco.ChunkOffset = append(stco.ChunkOffset, uint32(offset))
				}
				offset += uint64(len(sample.Data))
			}

			_, err = videoStcoBoxInfo.SeekToPayload(writer)
			if err != nil {
				return err
			}

			_, err = mp4.Marshal(writer, &stco, box.Context)
			if err != nil {
				return err
			}
		}
		{
			var stco mp4.Stco

			for i, sample := range videoInfo.AudioSamples() {
				if uint32(i)%chunkSize == 0 {
					stco.EntryCount++
					stco.ChunkOffset = append(stco.ChunkOffset, uint32(offset))
				}
				offset += uint64(len(sample.Data))
			}

			_, err = audioStcoBoxInfo.SeekToPayload(writer)
			if err != nil {
				return err
			}

			_, err = mp4.Marshal(writer, &stco, box.Context)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
