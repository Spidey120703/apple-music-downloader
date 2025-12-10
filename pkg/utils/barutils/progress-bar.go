package barutils

import (
	"downloader/pkg/ansi"
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func NewProgressBar(max int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stdout, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    "=" + ansi.CSIFg256(237),
			SaucerPadding: "-",
			BarStart:      ansi.CSIFgRGB(114, 156, 31),
			BarEnd:        ansi.CSIReset,
		}),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)
}

func NewProgressBarBytes(max int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stdout, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    "=" + ansi.CSIFg256(237),
			SaucerPadding: "-",
			BarStart:      ansi.CSIFgRGB(114, 156, 31),
			BarEnd:        ansi.CSIReset,
		}),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)
}
