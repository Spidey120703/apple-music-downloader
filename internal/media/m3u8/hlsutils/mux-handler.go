package hlsutils

import (
	"downloader/internal/media/mp4/mp4utils"
	"downloader/pkg/utils"
	"errors"
	"os"
	"path"
	"slices"
)

type MuxHandler struct {
	*Context
}

func (ctx *MuxHandler) initializeMux() (err error) {
	for _, entry := range ctx.MediaPlaylistEntries {
		if entry.Decryptor != nil {
			entry.Muxer = mp4utils.OfDecryptor(entry.Decryptor)
		} else {
			if err = func() (err error) {
				var inputs []*os.File
				if inputs, err = loadCaches(ctx.TempDir, entry); err != nil {
					return
				}
				defer utils.CloseQuietlyAll(inputs)

				if len(inputs) == 0 {
					return errors.New("empty media playlist")
				}

				entry.Muxer = mp4utils.NewMuxContext()
				if err = entry.Muxer.Initialize(inputs[0]); err != nil {
					return
				}

				for _, input := range inputs[1:] {
					if _, err = entry.Muxer.MergeSegment(input); err != nil {
						return
					}
				}

				return
			}(); err != nil {
				return err
			}
		}
	}
	// sort segments shortest first (ascending duration)
	slices.SortFunc(ctx.MediaPlaylistEntries, func(a *MediaPlaylistEntry, b *MediaPlaylistEntry) int {
		var maxDuration = func(e *MediaPlaylistEntry) (d int) {
			for _, samples := range e.Muxer.Samples {
				var t int
				for _, sample := range samples {
					t += int(sample.SampleDuration)
				}
				d = max(d, t)
			}
			return
		}
		return maxDuration(b) - maxDuration(a)
	})
	return
}

func (ctx *MuxHandler) applyMetadata() (err error) {
	if ctx.MetaData != nil {
		if err = ctx.MetaData.Attach(ctx.MediaPlaylistEntries[0].Muxer.Root); err != nil {
			return
		}
	}
	return
}

func (ctx *MuxHandler) mergeSegments() (err error) {
	for _, entry := range ctx.MediaPlaylistEntries {
		if err = entry.Muxer.Desegmentize(); err != nil {
			return
		}
	}
	return
}

func (ctx *MuxHandler) muxTracks() (err error) {
	ctx.Muxer = ctx.MediaPlaylistEntries[0].Muxer

	for _, entry := range ctx.MediaPlaylistEntries[1:] {
		if err = ctx.Muxer.MuxTrack(entry.Muxer); err != nil {
			return
		}
	}

	return
}

func (ctx *MuxHandler) finalizeMux() (err error) {
	if len(ctx.TargetPath) == 0 {
		return errors.New("target path is empty")
	}
	if err = os.MkdirAll(path.Dir(ctx.TargetPath), os.ModePerm); err != nil {
		return
	}

	var output *os.File
	if output, err = os.Create(ctx.TargetPath); err != nil {
		return
	}
	return ctx.Muxer.Finalize(output)
}

func (ctx *MuxHandler) Execute() (err error) {
	if err = ctx.initializeMux(); err != nil {
		return
	}
	if err = ctx.applyMetadata(); err != nil {
		return
	}
	if err = ctx.mergeSegments(); err != nil {
		return
	}
	if err = ctx.muxTracks(); err != nil {
		return
	}
	return ctx.finalizeMux()
}
