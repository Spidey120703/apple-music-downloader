package barutils

import (
	"downloader/pkg/ansi"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func NewProgress(wg *sync.WaitGroup, width int, d time.Duration) *mpb.Progress {
	return mpb.New(
		mpb.WithWaitGroup(wg),
		mpb.WithWidth(width),
		mpb.WithRefreshRate(d),
	)
}

func NewBar(p *mpb.Progress, total int64, title string) *mpb.Bar {
	return p.New(total,
		mpb.BarStyle().Lbound(ansi.CSIFgRGB(249, 38, 114)).Filler("=").Padding("-").Tip("").TipMeta(func(a string) string { return ansi.CSIFg256(237) }).Rbound(ansi.CSIReset),
		mpb.BarFillerOnComplete(ansi.CSIFgRGB(114, 156, 31)+strings.Repeat("=", 40)),
		mpb.PrependDecorators(
			decor.Name(title),
			decor.NewPercentage("%4d", decor.WC{C: decor.DextraSpace, W: 5}),
		),
		mpb.AppendDecorators(
			decor.Name(ansi.CSIFgGreen),
			decor.Counters(decor.SizeB1024(0), "%.1f/%.1f"),
			decor.Name(ansi.CSIReset),
			decor.Name(" "),
			decor.Name(ansi.CSIFgRed),
			decor.EwmaSpeed(decor.SizeB1024(0), "%.1f", 30),
			decor.Name(ansi.CSIReset),
			decor.Name(" eta "),
			decor.Name(ansi.CSIFgBlue),
			decor.EwmaETA(decor.ET_STYLE_HHMMSS, 30, decor.WC{}),
			decor.Name(ansi.CSIReset),
		),
	)
}
