package mp4utils

import (
	"downloader/internal/media/mp4/boxtree"
	"downloader/internal/media/mp4/cmaf"
	"downloader/pkg/utils"
	"errors"
	"io"
	"slices"

	"github.com/Spidey120703/go-mp4"
)

const DefaultChunkSize uint32 = 22

type MuxContext struct {
	cmaf.Context
	Samples   map[uint32][]Sample
	ChunkSize uint32
}

func NewMuxContext() *MuxContext {
	return &MuxContext{
		ChunkSize: DefaultChunkSize,
		Samples:   make(map[uint32][]Sample),
	}
}

func OfDecryptor(decryptor IDecryptor) *MuxContext {
	muxer := NewMuxContext()
	muxer.Root = decryptor.GetRoot()
	muxer.Samples = decryptor.GetSamples()
	return muxer
}

func (ctx *MuxContext) GetChunkSize() uint32 {
	if ctx.ChunkSize == 0 {
		return DefaultChunkSize
	}
	return ctx.ChunkSize
}

func (ctx *MuxContext) getTrackExtendsBox(trackID uint32) (trex *mp4.Trex, err error) {
	for _, trex = range ctx.Header.Moov.Mvex.Trex {
		if trex.TrackID == trackID {
			break
		}
	}
	if trex == nil || trex.TrackID != trackID {
		err = errors.New("traf track id does not match with any trex box")
		return
	}
	return
}

func (ctx *MuxContext) Initialize(input io.ReadSeeker) (err error) {
	if ctx.ChunkSize == 0 {
		ctx.ChunkSize = DefaultChunkSize
	}
	if ctx.Samples == nil {
		ctx.Samples = make(map[uint32][]Sample)
	}
	err = ctx.Context.Initialize(input)
	if err != nil {
		return
	}
	for idx := range ctx.Header.Moof {
		var trex *mp4.Trex
		var samples []Sample
		for _, traf := range ctx.Header.Moof[idx].Traf {
			if trex, err = ctx.getTrackExtendsBox(traf.Tfhd.TrackID); err != nil {
				return
			}
			samples = GetFullSamples(traf, ctx.Header.Mdat[idx], trex)
			ctx.Samples[traf.Tfhd.TrackID] = append(ctx.Samples[traf.Tfhd.TrackID], samples...)
		}
	}
	return
}

func (ctx *MuxContext) MergeSegment(input io.ReadSeeker) (seg *cmaf.Segment, err error) {
	seg, err = ctx.Context.MergeSegment(input)
	if err != nil {
		return
	}
	for idx := range seg.Moof {
		var trex *mp4.Trex
		var samples []Sample
		for _, traf := range seg.Moof[idx].Traf {
			if trex, err = ctx.getTrackExtendsBox(traf.Tfhd.TrackID); err != nil {
				return
			}
			samples = GetFullSamples(traf, seg.Mdat[idx], trex)
			ctx.Samples[traf.Tfhd.TrackID] = append(ctx.Samples[traf.Tfhd.TrackID], samples...)
		}
	}
	return
}

func (ctx *MuxContext) Desegmentize() (err error) {

	if ctx.Header, err = cmaf.InitializeHeader(ctx.Root); err != nil {
		return
	}

	{ // moov.mvhd
		var maxDuration uint64
		for _, samples := range ctx.Samples {
			var duration uint64
			for _, sample := range samples {
				duration += uint64(sample.SampleDuration)
			}
			maxDuration = max(maxDuration, duration)
		}
		switch ctx.Header.Moov.Mvhd.GetVersion() {
		case 0:
			ctx.Header.Moov.Mvhd.DurationV0 = uint32(maxDuration)
		case 1:
			ctx.Header.Moov.Mvhd.DurationV1 = maxDuration
		default:
			ctx.Header.Moov.Mvhd.DurationV0 = uint32(maxDuration)
			ctx.Header.Moov.Mvhd.DurationV1 = maxDuration
		}
	}

	for _, trak := range ctx.Header.Moov.Trak {
		samples := ctx.Samples[trak.Tkhd.TrackID]

		var duration uint64
		for _, sample := range samples {
			duration += uint64(sample.SampleDuration)
		}

		{ // moov.trak.tkhd
			trak.Tkhd.SetFlags(0x3)
			switch trak.Tkhd.GetVersion() {
			case 0:
				trak.Tkhd.DurationV0 = uint32(duration)
			case 1:
				trak.Tkhd.DurationV1 = duration
			default:
				trak.Tkhd.DurationV0 = uint32(duration)
				trak.Tkhd.DurationV1 = duration
			}
		}

		{ // moov.trak.mdia.mdhd
			switch trak.Mdia.Mdhd.GetVersion() {
			case 0:
				trak.Mdia.Mdhd.DurationV0 = uint32(duration)
			case 1:
				trak.Mdia.Mdhd.DurationV1 = duration
			default:
				trak.Mdia.Mdhd.DurationV0 = uint32(duration)
				trak.Mdia.Mdhd.DurationV1 = duration
			}
		}

		{ // moov.trak.mdia.minf.stbl.stts
			trak.Mdia.Minf.Stbl.Stts.Entries = []mp4.SttsEntry{}
			for _, sample := range samples {
				if len(trak.Mdia.Minf.Stbl.Stts.Entries) > 0 {
					last := &trak.Mdia.Minf.Stbl.Stts.Entries[len(trak.Mdia.Minf.Stbl.Stts.Entries)-1]
					if last.SampleDelta == sample.SampleDuration {
						last.SampleCount++
						continue
					}
				}
				trak.Mdia.Minf.Stbl.Stts.Entries = append(trak.Mdia.Minf.Stbl.Stts.Entries, mp4.SttsEntry{
					SampleCount: 1,
					SampleDelta: sample.SampleDuration,
				})
			}
			trak.Mdia.Minf.Stbl.Stts.EntryCount = uint32(len(trak.Mdia.Minf.Stbl.Stts.Entries))
		}

		sampleCount := uint32(len(samples))
		{ // moov.trak.mdia.minf.stbl.stsc
			trak.Mdia.Minf.Stbl.Stsc.Entries = []mp4.StscEntry{
				{
					FirstChunk:             1,
					SamplesPerChunk:        ctx.GetChunkSize(),
					SampleDescriptionIndex: 1,
				},
			}
			if sampleCount%ctx.GetChunkSize() != 0 {
				trak.Mdia.Minf.Stbl.Stsc.Entries = append(trak.Mdia.Minf.Stbl.Stsc.Entries, mp4.StscEntry{
					FirstChunk:             sampleCount/ctx.GetChunkSize() + 1,
					SamplesPerChunk:        sampleCount % ctx.GetChunkSize(),
					SampleDescriptionIndex: 1,
				})
			}
			trak.Mdia.Minf.Stbl.Stsc.EntryCount = uint32(len(trak.Mdia.Minf.Stbl.Stsc.Entries))
		}

		{ // moov.trak.mdia.minf.stbl.stsz
			trak.Mdia.Minf.Stbl.Stsz.SampleCount = sampleCount
			trak.Mdia.Minf.Stbl.Stsz.EntrySize = []uint32{}
			for _, sample := range samples {
				trak.Mdia.Minf.Stbl.Stsz.EntrySize = append(trak.Mdia.Minf.Stbl.Stsz.EntrySize, sample.SampleSize)
			}
		}

		{ // moov.trak.mdia.minf.stbl.stco
			chunkCount := (sampleCount + ctx.GetChunkSize() - 1) / ctx.GetChunkSize()
			trak.Mdia.Minf.Stbl.Stco.EntryCount = chunkCount
			trak.Mdia.Minf.Stbl.Stco.ChunkOffset = make([]uint32, chunkCount)
		}

		var flags uint32 = 0
		for _, moof := range ctx.Header.Moof {
			for _, traf := range moof.Traf {
				if traf.Tfhd.TrackID == trak.Tkhd.TrackID {
					for _, trun := range traf.Trun {
						flags |= trun.GetFlags()
					}
				}
			}
		}

		if flags&0x800 != 0 { // moov.trak.mdia.minf.stbl.ctts
			ctts := &mp4.Ctts{}

			ctts.SetVersion(samples[0].Version)
			ctts.Entries = []mp4.CttsEntry{}
			for _, sample := range samples {
				if len(ctts.Entries) > 0 {
					last := &ctts.Entries[len(ctts.Entries)-1]
					if func() bool {
						if sample.Version == 1 {
							return last.SampleOffsetV1 == sample.SampleCompositionTimeOffsetV1
						} else {
							return last.SampleOffsetV0 == sample.SampleCompositionTimeOffsetV0
						}
					}() {
						last.SampleCount++
						continue
					}
				}
				entry := mp4.CttsEntry{SampleCount: 1}
				if sample.Version == 1 {
					entry.SampleOffsetV1 = sample.SampleCompositionTimeOffsetV1
				} else {
					entry.SampleOffsetV0 = sample.SampleCompositionTimeOffsetV0
				}
				ctts.Entries = append(ctts.Entries, entry)
			}
			ctts.EntryCount = uint32(len(ctts.Entries))

			if err = trak.Mdia.Minf.Stbl.Node.Insert(2, mp4.BoxTypeCtts(), ctts); err != nil {
				return
			}
		}

		if flags&0x400 != 0 { // moov.trak.mdia.minf.stbl.sdtp
			sdtp := &mp4.Sdtp{}

			for _, sample := range samples {
				var sampleFlags SampleFlags
				if sampleFlags, err = UnmarshalSampleFlags(sample.SampleFlags); err != nil {
					return
				}

				sdtp.Samples = append(sdtp.Samples, mp4.SdtpSampleElem{
					IsLeading:           sampleFlags.IsLeading,
					SampleDependsOn:     sampleFlags.SampleDependsOn,
					SampleIsDependedOn:  sampleFlags.SampleIsDependedOn,
					SampleHasRedundancy: sampleFlags.SampleHasRedundancy,
				})
			}

			if err = trak.Mdia.Minf.Stbl.Node.Insert(2, mp4.BoxTypeSdtp(), sdtp); err != nil {
				return
			}
		}

		if flags&0x400 != 0 { // moov.trak.mdia.minf.stbl.stss
			stss := &mp4.Stss{}

			for idx, sample := range samples {
				var sampleFlags SampleFlags
				if sampleFlags, err = UnmarshalSampleFlags(sample.SampleFlags); err != nil {
					return
				}

				if sampleFlags.SampleIsNonSyncSample == 0 {
					stss.SampleNumber = append(stss.SampleNumber, uint32(idx+1))
				}
			}
			stss.EntryCount = uint32(len(stss.SampleNumber))

			if err = trak.Mdia.Minf.Stbl.Node.Insert(2, mp4.BoxTypeStss(), stss); err != nil {
				return
			}
		}

		if trak.Mdia.Minf.Stbl.Sgpd != nil {
			sgpd, found := trak.Mdia.Minf.Stbl.Node.Cache[mp4.BoxTypeSgpd()]
			if !found {
				goto end
			}
			if _, err = trak.Mdia.Minf.Stbl.Node.Remove(mp4.BoxTypeSgpd()); err != nil {
				return
			}
			trak.Mdia.Minf.Stbl.Node.Children = append(
				trak.Mdia.Minf.Stbl.Node.Children,
				sgpd[0],
			)

			sbgp := &mp4.Sbgp{}
			sbgp.Entries = append(sbgp.Entries, mp4.SbgpEntry{
				SampleCount:           sampleCount,
				GroupDescriptionIndex: 1,
			})
			sbgp.EntryCount = uint32(len(sbgp.Entries))

			if err = trak.Mdia.Minf.Stbl.Node.Append(mp4.BoxTypeSbgp(), sbgp); err != nil {
				return
			}
		}

	end:
		err = trak.Mdia.Minf.Stbl.Node.Caching()
		if err != nil {
			return
		}
	}

	if _, err = ctx.Header.Moov.Node.Remove(mp4.BoxTypeMvex()); err != nil {
		return
	}

	if _, err = ctx.Root.Remove(mp4.BoxTypeMoof()); err != nil {
		return
	}
	if _, err = ctx.Root.Remove(mp4.BoxTypeMdat()); err != nil {
		return
	}

	var offset uint64
	counter := utils.NewNullWriter()
	if offset, err = boxtree.Marshal(counter, ctx.Root); err != nil {
		return
	}

	{ // mdat
		offset += 8
		mdat := &mp4.Mdat{}
		for _, trak := range ctx.Header.Moov.Trak {
			samples := ctx.Samples[trak.Tkhd.TrackID]

			var i int
			for idx, sample := range samples {
				if uint32(idx)%ctx.ChunkSize == 0 {
					trak.Mdia.Minf.Stbl.Stco.ChunkOffset[i] = uint32(offset)
					i += 1
				}
				mdat.Data = append(mdat.Data, sample.Data...)
				offset += uint64(len(sample.Data))
			}
		}
		if err = ctx.Root.Append(mp4.BoxTypeMdat(), mdat); err != nil {
			return
		}
	}

	if ctx.Header, err = cmaf.InitializeHeader(ctx.Root); err != nil {
		return
	}

	return
}

func (ctx *MuxContext) MuxTrack(other *MuxContext) (err error) {
	if ctx.Header, err = cmaf.InitializeHeader(ctx.Root); err != nil {
		return
	}
	if other.Header, err = cmaf.InitializeHeader(other.Root); err != nil {
		return
	}
	if len(ctx.Header.Moof) != 0 || len(other.Header.Moof) != 0 {
		return errors.New("fMP4 is not compatible")
	}
	if len(ctx.Header.Mdat) != 1 || len(other.Header.Mdat) != 1 {
		return errors.New("multi-mdat box is not compatible")
	}

	var traks []*boxtree.BoxNode
	if traks, err = other.Root.P("moov.trak"); err != nil {
		return
	}
	trackID := ctx.Header.Moov.Mvhd.NextTrackID
	for idx, trak := range traks {
		tkhd := trak.Cache[mp4.BoxTypeTkhd()][0]
		tkhd.Box.(*mp4.Tkhd).TrackID = trackID
		ctx.Header.Moov.Node.Children = slices.Insert(ctx.Header.Moov.Node.Children, 2+idx, trak)
		trackID += 1
	}
	ctx.Header.Moov.Mvhd.NextTrackID = trackID

	if ctx.Header, err = cmaf.InitializeHeader(ctx.Root); err != nil {
		return
	}

	var offset uint64
	counter := utils.NewNullWriter()
	{ // other's moov.trak.mdia.minf.stbl.stco
		if offset, err = boxtree.Marshal(counter, ctx.Root); err != nil {
			return
		}
		for _, trak := range other.Header.Moov.Trak {
			headerSize := trak.Mdia.Minf.Stbl.Stco.ChunkOffset[0]
			for idx := range trak.Mdia.Minf.Stbl.Stco.ChunkOffset {
				trak.Mdia.Minf.Stbl.Stco.ChunkOffset[idx] += uint32(offset) - headerSize
			}
		}
	}

	mdatNode := ctx.Root.Cache[mp4.BoxTypeMdat()][0]

	if _, err = ctx.Root.Remove(mp4.BoxTypeMdat()); err != nil {
		return
	}

	mdat := mdatNode.Box.(*mp4.Mdat)
	mdat.Data = append(mdat.Data, other.Header.Mdat[0].Data...)

	{ // ctx's moov.trak.mdia.minf.stbl.stco
		if offset, err = boxtree.Marshal(counter, ctx.Root); err != nil {
			return
		}
		offset += 8
		for _, trak := range ctx.Header.Moov.Trak {
			oldHeaderSize := trak.Mdia.Minf.Stbl.Stco.ChunkOffset[0]
			for idx := range trak.Mdia.Minf.Stbl.Stco.ChunkOffset {
				trak.Mdia.Minf.Stbl.Stco.ChunkOffset[idx] += uint32(offset) - oldHeaderSize
			}
		}
	}

	ctx.Root.Children = append(ctx.Root.Children, mdatNode)
	return ctx.Root.Caching()
}
