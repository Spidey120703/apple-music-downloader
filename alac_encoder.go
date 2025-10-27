package main

import (
	"downloader/alac"
	"downloader/applemusic"
	"downloader/itunes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"strconv"

	"github.com/abema/go-mp4"
	"github.com/schollz/progressbar/v3"
)

type AlacInfo struct {
	reader io.ReadSeeker
	params *alac.Alac
	chunks []Chunk
}

func (s *AlacInfo) Duration() (ret uint64) {
	for _, chunk := range s.chunks {
		for _, sample := range chunk.samples {
			ret += uint64(sample.SampleDuration)
		}
	}
	return
}

func (s *AlacInfo) Samples() (ret []Sample) {
	for _, chunk := range s.chunks {
		for _, sample := range chunk.samples {
			ret = append(ret, sample)
		}
	}
	return
}

func extractAlac(input io.ReadSeeker) (*AlacInfo, error) {
	alacInfo := &AlacInfo{reader: input}

	// note: extracting alac atom
	{
		// Box Path: moov/trak/mdia/minf/stbl
		stbl, err := mp4.ExtractBox(input, nil, mp4.BoxPath{
			mp4.BoxTypeMoov(),
			mp4.BoxTypeTrak(),
			mp4.BoxTypeMdia(),
			mp4.BoxTypeMinf(),
			mp4.BoxTypeStbl(),
		})
		if err != nil || len(stbl) != 1 {
			return nil, err
		}

		// Box Path: moov/trak/mdia/minf/stbl/stsd/enca
		enca, err := mp4.ExtractBoxWithPayload(input, stbl[0], mp4.BoxPath{
			mp4.BoxTypeStsd(),
			mp4.BoxTypeEnca(),
		})
		if err != nil || len(enca) == 0 {
			return nil, err
		}

		box, err := mp4.ExtractBoxWithPayload(input, &enca[0].Info, mp4.BoxPath{alac.BoxTypeAlac()})
		if err != nil || len(box) != 1 {
			return nil, err
		}
		alacInfo.params = box[0].Payload.(*alac.Alac)
	}

	// note: extracting samples
	{
		// Box Path: moov/mvex/trex
		trex, err := mp4.ExtractBoxWithPayload(input, nil, mp4.BoxPath{
			mp4.BoxTypeMoov(),
			mp4.BoxTypeMvex(),
			mp4.BoxTypeTrex(),
		})
		if err != nil || len(trex) != 1 {
			return nil, err
		}
		trexPayload := trex[0].Payload.(*mp4.Trex)

		// Box Path: moof[]
		moofs, err := mp4.ExtractBox(input, nil, mp4.BoxPath{mp4.BoxTypeMoof()})
		if err != nil || len(moofs) <= 0 {
			return nil, err
		}
		Info.Printf("Found %d '%s' Boxes", len(moofs), mp4.BoxTypeMoof())

		// Box Path: mdat[]
		mdats, err := mp4.ExtractBoxWithPayload(input, nil, mp4.BoxPath{mp4.BoxTypeMdat()})
		if err != nil || len(mdats) != len(moofs) {
			return nil, err
		}
		Info.Printf("Found %d '%s' Boxes", len(mdats), mp4.BoxTypeMdat())

		for i, moof := range moofs {

			// Box Path: moof[]/traf/tfhd
			tfhd, err := mp4.ExtractBoxWithPayload(input, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTfhd(),
			})
			if err != nil || len(tfhd) != 1 {
				return nil, err
			}
			tfhdPayload := tfhd[0].Payload.(*mp4.Tfhd)

			sampleDescriptionIndex := tfhdPayload.SampleDescriptionIndex
			if sampleDescriptionIndex != 0 {
				sampleDescriptionIndex--
			}

			// Box Path: moof[]/traf/trun
			truns, err := mp4.ExtractBoxWithPayload(input, moof, mp4.BoxPath{
				mp4.BoxTypeTraf(),
				mp4.BoxTypeTrun(),
			})
			if err != nil || len(truns) <= 0 {
				return nil, err
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
				alacInfo.chunks = append(alacInfo.chunks, chunk)
			}
			if len(mdatPayloadData) != 0 {
				return nil, errors.New("size mismatch")
			}
		}
	}
	return alacInfo, nil
}

func decryptSample(samples []Sample, song *applemusic.Songs, keys []string) (decryptedSamples [][]byte, err error) {
	conn, err := net.Dial("tcp", "127.0.0.1:10020")
	if err != nil {
		return
	}
	defer CloseQuietly(conn)

	var lastIndex uint32 = math.MaxUint32

	bar := progressbar.Default(int64(len(samples)), "Decrypting")

	for _, sample := range samples {
		if lastIndex != sample.SampleDescriptionIndex {
			if lastIndex != uint32(math.MaxUint32) {
				_, err = conn.Write([]byte{0, 0, 0, 0})
				if err != nil {
					return
				}
			}
			keyUri := keys[sample.SampleDescriptionIndex]
			id := *song.ID
			if keyUri == PrefetchKeyUri {
				id = DefaultId
			}

			_, err = conn.Write([]byte{byte(len(id))})
			if err != nil {
				return
			}
			_, err = io.WriteString(conn, id)
			if err != nil {
				return
			}

			_, err = conn.Write([]byte{byte(len(keyUri))})
			if err != nil {
				return
			}
			_, err = io.WriteString(conn, keyUri)
			if err != nil {
				return
			}
		}
		lastIndex = sample.SampleDescriptionIndex

		err = binary.Write(conn, binary.LittleEndian, uint32(len(sample.Data)))
		if err != nil {
			return
		}
		_, err = conn.Write(sample.Data)
		if err != nil {
			return
		}

		decrypted := make([]byte, len(sample.Data))
		_, err = io.ReadFull(conn, decrypted)
		if err != nil {
			return
		}

		// println(hex.Dump(decrypted))

		decryptedSamples = append(decryptedSamples, decrypted)

		err = bar.Add(1)
		if err != nil {
			return
		}
	}
	_, _ = conn.Write([]byte{0, 0, 0, 0, 0})

	//println(bytes.Equal((func() (data []byte) {
	//	for _, sample := range samples {
	//		data = append(data, sample.Data...)
	//	}
	//	return
	//})(), (func() (data []byte) {
	//	for _, sample := range decryptedSamples {
	//		data = append(data, sample...)
	//	}
	//	return
	//})()))

	return
}

func writeM4A(
	writer *mp4.Writer,
	alacInfo *AlacInfo,
	itunesSongInfo *itunes.Song,
	song *applemusic.Songs,
	album *applemusic.Albums,
	data []byte,
	coverData []byte,
	lyrics string) error {
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
	duration := alacInfo.Duration()
	numSamples := uint32(len(alacInfo.Samples()))
	var stcoBoxInfo *mp4.BoxInfo

	{ // `moov` - Movie Box
		_, err := writer.StartBox(&mp4.BoxInfo{
			Type: mp4.BoxTypeMoov(),
		})
		if err != nil {
			return err
		}

		// Box Path: moov
		box, err := mp4.ExtractBox(alacInfo.reader, nil, mp4.BoxPath{mp4.BoxTypeMoov()})
		if err != nil || len(box) != 1 {
			return err
		}
		originalMoov := box[0]

		{ // `mvhd` - Movie Header Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMvhd()})
			if err != nil {
				return err
			}

			// Box Path: moov/mvhd
			box, err := mp4.ExtractBoxWithPayload(alacInfo.reader, originalMoov, mp4.BoxPath{mp4.BoxTypeMvhd()})
			if err != nil || len(box) != 1 {
				return err
			}
			mvhdPayload := box[0].Payload.(*mp4.Mvhd)

			// ISO/IEC 14496-12 Section 6.2.2
			// Box Version:
			// 	- 0x000000 - 32-bit sizes fields
			// 	- 0x000001 - 64-bit sizes fields
			if mvhdPayload.Version == 1 {
				mvhdPayload.DurationV1 = duration
			} else { // tkhd.Version == 0
				mvhdPayload.DurationV0 = uint32(duration)
			}

			_, err = mp4.Marshal(writer, mvhdPayload, box[0].Info.Context)
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
			box, err := mp4.ExtractBox(alacInfo.reader, originalMoov, mp4.BoxPath{mp4.BoxTypeTrak()})
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
				box, err := mp4.ExtractBoxWithPayload(alacInfo.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeTkhd()})
				if err != nil || len(box) != 1 {
					return err
				}
				tkhdPayload := box[0].Payload.(*mp4.Tkhd)

				// ISO/IEC 14496-12 Section 6.2.2
				if tkhdPayload.Version == 1 {
					tkhdPayload.DurationV1 = duration
				} else { // tkhd.Version == 0
					tkhdPayload.DurationV0 = uint32(duration)
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
				box, err := mp4.ExtractBox(alacInfo.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeMdia()})
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
					box, err := mp4.ExtractBoxWithPayload(alacInfo.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeMdhd()})
					if err != nil || len(box) != 1 {
						return err
					}
					mdhdPayload := box[0].Payload.(*mp4.Mdhd)

					// ISO/IEC 14496-12 Section 6.2.2
					if mdhdPayload.Version == 1 {
						mdhdPayload.DurationV1 = duration
					} else { // tkhd.Version == 0
						mdhdPayload.DurationV0 = uint32(duration)
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
					box, err = mp4.ExtractBox(alacInfo.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeHdlr()})
					if err != nil {
						return err
					}

					err = writer.CopyBox(alacInfo.reader, box[0])
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
					box, err := mp4.ExtractBox(alacInfo.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeMinf()})
					if err != nil || len(box) != 1 {
						return err
					}
					originalMinf := box[0]

					{ // `smhd` - Sound Media Header Box, `dinf` - Data Information Box

						// Box Path: moov/trak/mdia/minf/smhd
						// 			 moov/trak/mdia/minf/dinf
						boxes, err := mp4.ExtractBoxes(alacInfo.reader, originalMinf, []mp4.BoxPath{
							{mp4.BoxTypeSmhd()},
							{mp4.BoxTypeDinf()},
						})
						if err != nil {
							return err
						}

						for _, box := range boxes {
							err = writer.CopyBox(alacInfo.reader, box)
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

							{ // `alac` - Apple Lossless Audio Codec Encapsulation

								context := mp4.Context{UnderStsd: true}

								_, err := writer.StartBox(&mp4.BoxInfo{Type: alac.BoxTypeAlac()})
								if err != nil {
									return err
								}

								_, err = mp4.Marshal(writer, &mp4.AudioSampleEntry{
									SampleEntry: mp4.SampleEntry{
										AnyTypeBox: mp4.AnyTypeBox{
											Type: alac.BoxTypeAlac(),
										},
										DataReferenceIndex: 1,
									},
									ChannelCount: uint16(alacInfo.params.NumChannels),
									SampleSize:   uint16(alacInfo.params.BitDepth),
									SampleRate:   alacInfo.params.SampleRate,
								}, context)
								if err != nil {
									return err
								}

								{ // `alac` - Apple Lossless Audio Codec

									box, err := writer.StartBox(&mp4.BoxInfo{Type: alac.BoxTypeAlac()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, alacInfo.params, box.Context)
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
							for _, sample := range alacInfo.Samples() {
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

							if numSamples%chunkSize != 0 {
								stsc.Entries = append(stsc.Entries, mp4.StscEntry{
									FirstChunk:             numSamples/chunkSize + 1,
									SamplesPerChunk:        numSamples % chunkSize,
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

							stsz := mp4.Stsz{SampleCount: numSamples}
							for _, sample := range alacInfo.Samples() {
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

							numChunks := (numSamples + chunkSize - 1) / chunkSize

							_, err = mp4.Marshal(writer, &mp4.Stco{
								EntryCount:  numChunks,
								ChunkOffset: make([]uint32, numChunks),
							}, box.Context)
							if err != nil {
								return err
							}

							stcoBoxInfo, err = writer.EndBox()
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

			{ // `udta` - User Data Box

				// Box Path: moov/trak/udta
				box, err := mp4.ExtractBox(alacInfo.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeUdta()})
				if err != nil {
					return err
				}

				err = writer.CopyBox(alacInfo.reader, box[0])
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

			{ // `meta` - Meta Box
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
								return errors.New("unknown data type")
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
						/*
							AddMetaItunes := func(name string, value any) error {
								_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxType{'-', '-', '-', '-'}, Context: context})
								if err != nil {
									return err
								}

								{
									_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxType{'m', 'e', 'a', 'n'}, Context: context})
									if err != nil {
										return err
									}

									_, err = writer.Write([]byte{0, 0, 0, 0})
									if err != nil {
										return err
									}
									_, err = io.WriteString(writer, "com.apple.iTunes")
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								{
									_, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxType{'n', 'a', 'm', 'e'}, Context: context})
									if err != nil {
										return err
									}

									_, err = writer.Write([]byte{0, 0, 0, 0})
									if err != nil {
										return err
									}
									_, err = io.WriteString(writer, name)
									if err != nil {
										return err
									}

									_, err = writer.EndBox()
									if err != nil {
										return err
									}
								}

								err = MarshalData(value)
								if err != nil {
									return err
								}

								_, err = writer.EndBox()
								return err
							}
						*/

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

							/*
								genreID := GuessID3v1GenreID(song.Attributes.GenreNames[0])
								if genreID >= 0 {
									err = AddMeta(mp4.BoxType{'g', 'n', 'r', 'e'}, uint32(genreID))
									if err != nil {
										return err
									}
								}

							*/
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

						if *song.Attributes.HasLyrics {
							err = AddMeta(mp4.BoxType{'\xa9', 'l', 'y', 'r'}, lyrics)
							if err != nil {
								return err
							}
						}

						/*
							err = AddMeta(mp4.BoxType{'t', 'm', 'p', 'o'}, uint8(0))
							if err != nil {
								return err
							}
						*/
						err = AddMeta(mp4.BoxType{'r', 't', 'n', 'g'}, uint8(0))
						if err != nil {
							return err
						}
						err = AddMeta(mp4.BoxType{'s', 't', 'i', 'k'}, uint8(1))
						if err != nil {
							return err
						}

						if len(song.Attributes.GenreNames) > 0 {
							//var genreID uint32
							//if song.Relationships.Genres != nil {
							//	for _, genre := range song.Relationships.Genres.Data {
							//		if *genre.Attributes.Name == song.Attributes.GenreNames[0] {
							//			id, err := strconv.ParseUint(*genre.ID, 10, 32)
							//			if err != nil {
							//				return err
							//			}
							//			genreID = uint32(id)
							//		}
							//	}
							//}
							//if genreID == 0 {
							//	genreID = uint32(GetQTGenreID(song.Attributes.GenreNames))
							//}
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
						/*
							err = AddMeta(mp4.BoxType{'x', 'i', 'd', ' '}, ":isrc:"+*song.Attributes.Isrc)
							if err != nil {
								return err
							}
						*/

						/*
							err = AddMeta(mp4.BoxType{'s', 'o', 'a', 'a'}, *album.Attributes.ArtistName)
							if err != nil {
								return err
							}
							err = AddMeta(mp4.BoxType{'\xa9', 'c', 'o', 'm'}, composerName)
							if err != nil {
								return err
							}
							err = AddMeta(mp4.BoxType{'\xa9', 'c', 'p', 'y'}, *album.Attributes.Copyright)
							if err != nil {
								return err
							}
							err = AddMeta(mp4.BoxType{'\xa9', 'p', 'r', 'f'}, *song.Attributes.ArtistName)
							if err != nil {
								return err
							}
							err = AddMeta(mp4.BoxType{'\xa9', 'p', 'u', 'b'}, *album.Attributes.RecordLabel)
							if err != nil {
								return err
							}

							// "iTunes 12.10.11.2"
							// "iTunes 12.12.8.2"
							err = AddMeta(mp4.BoxType{'\xa9', 't', 'o', 'o'}, "iTunes 12.13.8.3")
							if err != nil {
								return err
							}

							err = AddMetaItunes("ITUNESALBUMID", *album.ID)
							if err != nil {
								return err
							}
							err = AddMetaItunes("RELEASETIME", *song.Attributes.ReleaseDate)
							if err != nil {
								return err
							}
							err = AddMetaItunes("PERFORMER", *song.Attributes.ArtistName)
							if err != nil {
								return err
							}
							err = AddMetaItunes("ISRC", *song.Attributes.Isrc)
							if err != nil {
								return err
							}
							err = AddMetaItunes("LABEL", *album.Attributes.RecordLabel)
							if err != nil {
								return err
							}
							err = AddMetaItunes("UPC", *album.Attributes.Upc)
							if err != nil {
								return err
							}
						*/

						/*
							err = AddMetaItunes("Encoding Params", []byte{
								'v', 'e', 'r', 's', 0, 0, 0, 1,
								'a', 'c', 'd', 'f', 0, 0, 0, 3,
								'v', 'b', 'r', 'q', 0, 0, 0, 0,
							})
							if err != nil {
								return err
							}
							err = AddMetaItunes("iTunNORM",
								" 00000000 00000000 00000000 00000000 00000000"+
									" 00000000 00000000 00000000 00000000 00000000")
							if err != nil {
								return err
							}
							err = AddMetaItunes("iTunes_CDDB_IDs",
								strconv.Itoa(*song.Relationships.Albums.Data[0].Attributes.TrackCount)+"+"+
									""+"+"+
									"")
							if err != nil {
								return err
							}
							err = AddMetaItunes("UFIDhttp://www.cddb.com/id3/taginfo1.html", []byte{
								0xE4, 0x8C, 0xB3, 0xE3, 0x8D, 0x84, 0xE3, 0x00,
								0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
								0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
								0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
								0x00, 0x00, 0x00, 0x00,
							})
							if err != nil {
								return err
							}
						*/
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

		_, err = mp4.Marshal(writer, &mp4.Mdat{Data: data}, box.Context)
		if err != nil {
			return err
		}

		mdatBoxInfo, err := writer.EndBox()
		if err != nil {
			return err
		}

		var stco mp4.Stco

		offset := mdatBoxInfo.Offset + mdatBoxInfo.HeaderSize
		for i, sample := range alacInfo.Samples() {
			if uint32(i)%chunkSize == 0 {
				stco.EntryCount++
				stco.ChunkOffset = append(stco.ChunkOffset, uint32(offset))
			}
			offset += uint64(len(sample.Data))
		}

		_, err = stcoBoxInfo.SeekToPayload(writer)
		if err != nil {
			return err
		}

		_, err = mp4.Marshal(writer, &stco, box.Context)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeFMP4(writer *mp4.Writer, alacContext *AlacInfo, data []byte) error {
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
			MajorBrand:   mp4.BrandISO5(),
			MinorVersion: 0,
			CompatibleBrands: []mp4.CompatibleBrandElem{
				{mp4.BrandISOM()},
				{mp4.BrandISO5()},
				{[4]byte{'h', 'l', 's', 'f'}},
				{[4]byte{'m', 'p', '4', '2'}},
				{[4]byte{'a', 'l', 'a', 'c'}},
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

	//duration := alacContext.Duration()
	var err error

	{ // `moov` - Movie Box
		_, err = writer.StartBox(&mp4.BoxInfo{
			Type: mp4.BoxTypeMoov(),
		})
		if err != nil {
			return err
		}

		// Box Path: moov
		box, err := mp4.ExtractBox(alacContext.reader, nil, mp4.BoxPath{mp4.BoxTypeMoov()})
		if err != nil || len(box) != 1 {
			return err
		}
		originalMoov := box[0]

		{ // `mvhd` - Movie Header Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMvhd()})
			if err != nil {
				return err
			}

			// Box Path: moov/mvhd
			box, err := mp4.ExtractBoxWithPayload(alacContext.reader, originalMoov, mp4.BoxPath{mp4.BoxTypeMvhd()})
			if err != nil || len(box) != 1 {
				return err
			}
			mvhdPayload := box[0].Payload.(*mp4.Mvhd)

			// ISO/IEC 14496-12 Section 6.2.2
			// Box Version:
			// 	- 0x000000 - 32-bit sizes fields
			// 	- 0x000001 - 64-bit sizes fields
			//if mvhdPayload.Version == 1 {
			//	mvhdPayload.DurationV1 = duration
			//} else { // tkhd.Version == 0
			//	mvhdPayload.DurationV0 = uint32(duration)
			//}

			_, err = mp4.Marshal(writer, mvhdPayload, box[0].Info.Context)
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
			box, err := mp4.ExtractBox(alacContext.reader, originalMoov, mp4.BoxPath{mp4.BoxTypeTrak()})
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
				box, err := mp4.ExtractBoxWithPayload(alacContext.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeTkhd()})
				if err != nil || len(box) != 1 {
					return err
				}
				tkhdPayload := box[0].Payload.(*mp4.Tkhd)

				// ISO/IEC 14496-12 Section 6.2.2
				//if tkhdPayload.Version == 1 {
				//	tkhdPayload.DurationV1 = duration
				//} else { // tkhd.Version == 0
				//	tkhdPayload.DurationV0 = uint32(duration)
				//}
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
				box, err := mp4.ExtractBox(alacContext.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeMdia()})
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
					box, err := mp4.ExtractBoxWithPayload(alacContext.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeMdhd()})
					if err != nil || len(box) != 1 {
						return err
					}
					mdhdPayload := box[0].Payload.(*mp4.Mdhd)

					// ISO/IEC 14496-12 Section 6.2.2
					//if mdhdPayload.Version == 1 {
					//	mdhdPayload.DurationV1 = duration
					//} else { // tkhd.Version == 0
					//	mdhdPayload.DurationV0 = uint32(duration)
					//}

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
					box, err := mp4.ExtractBox(alacContext.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeHdlr()})
					if err != nil {
						return err
					}

					err = writer.CopyBox(alacContext.reader, box[0])
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
					box, err := mp4.ExtractBox(alacContext.reader, originalMdia, mp4.BoxPath{mp4.BoxTypeMinf()})
					if err != nil || len(box) != 1 {
						return err
					}
					originalMinf := box[0]

					{ // `smhd` - Sound Media Header Box, `dinf` - Data Information Box

						// Box Path: moov/trak/mdia/minf/smhd
						// 			 moov/trak/mdia/minf/dinf
						boxes, err := mp4.ExtractBoxes(alacContext.reader, originalMinf, []mp4.BoxPath{
							{mp4.BoxTypeSmhd()},
							{mp4.BoxTypeDinf()},
						})
						if err != nil {
							return err
						}

						for _, box := range boxes {
							err = writer.CopyBox(alacContext.reader, box)
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

							{ // `alac` - Apple Lossless Audio Codec Encapsulation
								_, err = writer.StartBox(&mp4.BoxInfo{Type: alac.BoxTypeAlac()})
								if err != nil {
									return err
								}

								// ISO/IEC 14496-12 Section 12.2.3.1
								// type AudioSampleEntry struct {
								//		Reserved 			[3]uint16
								err = binary.Write(writer, binary.BigEndian, [3]uint16{})
								if err != nil {
									return err
								}
								//		DataReferenceIndex  uint16
								err = binary.Write(writer, binary.BigEndian, uint16(1))
								if err != nil {
									return err
								}
								//		EntryVersion 		[2]uint32
								err = binary.Write(writer, binary.BigEndian, uint16(0))
								if err != nil {
									return err
								}
								//		Reserved 			[3]uint16
								err = binary.Write(writer, binary.BigEndian, [3]uint16{})
								if err != nil {
									return err
								}
								// 		ChannelCount 		uint16
								err = binary.Write(writer, binary.BigEndian, uint16(alacContext.params.NumChannels))
								if err != nil {
									return err
								}
								// 	 	SampleSize 			uint16
								err = binary.Write(writer, binary.BigEndian, uint16(alacContext.params.BitDepth))
								if err != nil {
									return err
								}
								//		PreDefined			uint16
								err = binary.Write(writer, binary.BigEndian, uint16(0))
								if err != nil {
									return err
								}
								//		Reserved			uint16
								err = binary.Write(writer, binary.BigEndian, uint16(0))
								if err != nil {
									return err
								}
								// 	 	SampleRate 			uint32
								err = binary.Write(writer, binary.BigEndian, alacContext.params.SampleRate)
								if err != nil {
									return err
								}
								// }

								// 		QuickTimeData		[]byte
								{ // `alac` - Apple Lossless Audio Codec
									box, err := writer.StartBox(&mp4.BoxInfo{Type: alac.BoxTypeAlac()})
									if err != nil {
										return err
									}

									_, err = mp4.Marshal(writer, alacContext.params, box.Context)
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

							_, err = mp4.Marshal(writer, &mp4.Stts{}, box.Context)
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

							_, err = mp4.Marshal(writer, &mp4.Stsc{}, box.Context)
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

							_, err = mp4.Marshal(writer, &mp4.Stsz{}, box.Context)
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

							_, err = mp4.Marshal(writer, &mp4.Stco{}, box.Context)
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

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}

			{ // `udta` - User Data Box

				// Box Path: moov/trak/udta
				box, err := mp4.ExtractBox(alacContext.reader, originalTrak, mp4.BoxPath{mp4.BoxTypeUdta()})
				if err != nil || len(box) != 1 {
					return err
				}

				err = writer.CopyBox(alacContext.reader, box[0])
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

			// Box Path: moov/udta
			box, err := mp4.ExtractBox(alacContext.reader, originalMoov, mp4.BoxPath{mp4.BoxTypeUdta()})
			if err != nil || len(box) != 1 {
				return err
			}

			err = writer.CopyBox(alacContext.reader, box[0])
			if err != nil {
				return err
			}
		}

		{ // `mvex` - Movie Extends Box
			_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMvex()})
			if err != nil {
				return err
			}

			{ // `trex` - Track Extends Box

				// Box Path: moov/mvex/trex
				box, err := mp4.ExtractBox(alacContext.reader, originalMoov, mp4.BoxPath{
					mp4.BoxTypeMvex(),
					mp4.BoxTypeTrex(),
				})
				if err != nil || len(box) != 1 {
					return err
				}

				err = writer.CopyBox(alacContext.reader, box[0])
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

	{ // `moof`, `mdat`
		moofBoxes, err := mp4.ExtractBox(alacContext.reader, nil, mp4.BoxPath{mp4.BoxTypeMoof()})
		if err != nil || len(moofBoxes) == 0 {
			return err
		}

		mdatBoxes, err := mp4.ExtractBoxWithPayload(alacContext.reader, nil, mp4.BoxPath{mp4.BoxTypeMdat()})
		if err != nil || len(mdatBoxes) != len(moofBoxes) {
			return err
		}

		offset := uint64(0)
		for i, originalMoof := range moofBoxes {
			var dataOffsets []uint32
			var trunBoxInfos []*mp4.BoxInfo
			var truns []mp4.Trun

			{ // `moof` - Movie Fragment Box
				_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMoof()})
				if err != nil {
					return err
				}

				{ // `mfhd` - Movie Fragment Header Box
					box, err := mp4.ExtractBox(alacContext.reader, originalMoof, mp4.BoxPath{mp4.BoxTypeMfhd()})
					if err != nil || len(box) != 1 {
						return err
					}

					err = writer.CopyBox(alacContext.reader, box[0])
					if err != nil {
						return err
					}
				}

				{ // `traf` - Track Fragment Box
					_, err = writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTraf()})
					if err != nil {
						return err
					}

					{ // `tfhd` - Track Fragment Header Box
						box, err := mp4.ExtractBox(alacContext.reader, originalMoof, mp4.BoxPath{
							mp4.BoxTypeTraf(),
							mp4.BoxTypeTfhd(),
						})
						if err != nil || len(box) != 1 {
							return err
						}

						err = writer.CopyBox(alacContext.reader, box[0])
						if err != nil {
							return err
						}
					}

					{ // `tfdt` - Track Fragment Header Box
						box, err := mp4.ExtractBox(alacContext.reader, originalMoof, mp4.BoxPath{
							mp4.BoxTypeTraf(),
							mp4.BoxTypeTfdt(),
						})
						if err != nil || len(box) != 1 {
							return err
						}

						err = writer.CopyBox(alacContext.reader, box[0])
						if err != nil {
							return err
						}
					}

					{ // `trun` - Trunk Run Box
						box, err := mp4.ExtractBoxWithPayload(alacContext.reader, originalMoof, mp4.BoxPath{
							mp4.BoxTypeTraf(),
							mp4.BoxTypeTrun(),
						})
						if err != nil || len(box) == 0 {
							return err
						}

						dataOffset := uint32(0)

						for _, originalTrun := range box {
							box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeTrun()})
							if err != nil {
								return err
							}
							originalTrunPayload := originalTrun.Payload.(*mp4.Trun)

							trun := mp4.Trun{
								FullBox: mp4.FullBox{
									Flags: originalTrunPayload.Flags,
								},
								SampleCount:      originalTrunPayload.SampleCount,
								DataOffset:       int32(dataOffset),
								FirstSampleFlags: originalTrunPayload.FirstSampleFlags,
								Entries:          originalTrunPayload.Entries,
							}

							truns = append(truns, trun)

							_, err = mp4.Marshal(writer, &trun, box.Context)
							if err != nil {
								return err
							}

							dataOffsets = append(dataOffsets, dataOffset)

							for _, entry := range originalTrunPayload.Entries {
								dataOffset += entry.SampleSize
							}

							trunBoxInfo, err := writer.EndBox()
							if err != nil {
								return err
							}
							trunBoxInfos = append(trunBoxInfos, trunBoxInfo)
						}
					}

					_, err = writer.EndBox()
					if err != nil {
						return err
					}
				}

				moofBoxInfo, err := writer.EndBox()
				if err != nil {
					return err
				}

				for i, trunBoxInfo := range trunBoxInfos {
					_, err = trunBoxInfo.SeekToPayload(writer)
					if err != nil {
						return err
					}

					trun := truns[i]
					trun.DataOffset = int32(uint32(moofBoxInfo.Size)+dataOffsets[i]) + 8

					_, err = mp4.Marshal(writer, &trun, trunBoxInfo.Context)
					if err != nil {
						return err
					}
				}

				_, err = writer.Seek(0, io.SeekEnd)
				if err != nil {
					return err
				}
			}

			originalMdat := mdatBoxes[i]
			{ // `mdat` - Media Data Box
				box, err := writer.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeMdat()})
				if err != nil {
					return err
				}

				size := originalMdat.Info.Size - 8

				_, err = mp4.Marshal(writer, &mp4.Mdat{Data: data[offset : offset+size]}, box.Context)
				if err != nil {
					return err
				}

				offset += size

				_, err = writer.EndBox()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
