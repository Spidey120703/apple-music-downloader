package config

import "time"

const TargetPath = "Downloads/"
const TempPath = "Temp/"

const FairPlayServerAddr = "127.0.0.1:10020"

const BarWidth = 60
const NumThreads = 5
const BarRefreshRate = 100 * time.Millisecond

const ChromeUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
const AMPLibraryUserAgent = "AMPLibraryAgent/1.6 (Windows 10.0.26120 x64; x64) Chromium/128.0.2739.63 build/112 (dt:2)"
const AppleMusicUserAgent = "Music/1.6 (Windows 10.0.26120 x64; x64) Chromium/128.0.2739.63 build/112 (dt:2)"

const UserAgent = ChromeUserAgent
const Origin = "https://beta.music.apple.com"
const Referer = "https://beta.music.apple.com/"

const Storefront = "cn"
const MediaUserToken = "Ai00hPjqDQcpdKvILsHxXWeJLNt2miOjjBe7cgSI0uIpZu0U90Fu7DQsovYaMHU+p+gJyOHUKfgA2vbGN19XbGy40oWwO3u+46cEucIzORDAuTaPQsrBvMZidhP2krg5QhPW3jYXuFgK2xUaFWrZ45jrun0MX4KeD3G/Lck8cwACZ+5BHeh4V65fpcTjLa6Sm8Uy7Na+R6bse+iBiuvgnVkirt1FmQdVK22RfyXAX7uJYpaAgw=="
const Language = "zh-Hans-CN"

const UseOriginalExt = true
