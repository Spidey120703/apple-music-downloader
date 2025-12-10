package ansi

// SGR (Select Graphic Rendition)
// https://ecma-international.org/wp-content/uploads/ECMA-48_5th_edition_june_1991.pdf (8.3.117)

const (
	TextReset = iota
	TextBold
	TextFaint
	TextItalic
	TextUnderline
	TextSlowBlink
	TextRapidBlink
	TextReverseVideo
	TextConceal
	TextStrikethrough

	TextDoublyUnderline = iota + 11
	TextNormalIntensity
	TextNoItalic
	TextNoUnderline
	TextNoBlink
	_
	TextNoReverseVideo
	TextNoConceal
	TextNoStrikethrough

	ForegroundBlack
	ForegroundRed
	ForegroundGreen
	ForegroundYellow
	ForegroundBlue
	ForegroundMagenta
	ForegroundCyan
	ForegroundWhite
	ForegroundSetColor
	ForegroundResetColor

	BackgroundBlack
	BackgroundRed
	BackgroundGreen
	BackgroundYellow
	BackgroundBlue
	BackgroundMagenta
	BackgroundCyan
	BackgroundWhite
	BackgroundSetColor
	BackgroundResetColor

	_

	TextFrame
	TextEncircle
	TextOverline
	TextNoFrameOrEncircle
	TextNoOverline

	ForegroundBrightBlack = iota + 45
	ForegroundBrightRed
	ForegroundBrightGreen
	ForegroundBrightYellow
	ForegroundBrightBlue
	ForegroundBrightMagenta
	ForegroundBrightCyan
	ForegroundBrightWhite

	BackgroundBrightBlack = iota + 47
	BackgroundBrightRed
	BackgroundBrightGreen
	BackgroundBrightYellow
	BackgroundBrightBlue
	BackgroundBrightMagenta
	BackgroundBrightCyan
	BackgroundBrightWhite
)

const (
	ColorMode256 = 5
	ColorModeRGB = 2
)
