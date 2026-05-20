package config

import "time"

const (
	NumThreads     = 5
	BarWidth       = 60
	BarRefreshRate = 100 * time.Millisecond

	DefaultTargetPath = "Downloads/"
	DefaultTempPath   = "Temp/"

	DefaultFairPlayServerAddr = "127.0.0.1:10020"
	DefaultUserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
	DefaultOrigin             = "https://beta.music.apple.com"
	DefaultReferer            = "https://beta.music.apple.com/"
	DefaultNumThreads         = 5
	DefaultStorefront         = "cn"
	DefaultAMLanguage         = "zh-Hans-CN"
)
