package ansi

import (
	"strconv"
	"strings"
)

// CSI (Control Sequence Introducer)

func CSI(sgr ...int) string {
	if len(sgr) == 0 {
		return "\033[0m"
	}
	var str strings.Builder
	str.WriteString("\033[")
	for i, c := range sgr {
		str.WriteString(strconv.Itoa(c & 0xff))
		if i != len(sgr)-1 {
			str.WriteByte(';')
		}
	}
	str.WriteByte('m')
	return str.String()
}

const (
	CSIReset         = "\033[0m"
	CSIBold          = "\033[1m"
	CSIFaint         = "\033[2m"
	CSIItalic        = "\033[3m"
	CSIUnderline     = "\033[4m"
	CSISlowBlink     = "\033[5m"
	CSIRapidBlink    = "\033[6m"
	CSIReverseVideo  = "\033[7m"
	CSIConceal       = "\033[8m"
	CSIStrikethrough = "\033[9m"

	CSIFgBlack      = "\033[30m"
	CSIFgRed        = "\033[31m"
	CSIFgGreen      = "\033[32m"
	CSIFgYellow     = "\033[33m"
	CSIFgBlue       = "\033[34m"
	CSIFgMagenta    = "\033[35m"
	CSIFgCyan       = "\033[36m"
	CSIFgWhite      = "\033[37m"
	CSIFgResetColor = "\033[39m"

	CSIBgBlack      = "\033[40m"
	CSIBgRed        = "\033[41m"
	CSIBgGreen      = "\033[42m"
	CSIBgYellow     = "\033[43m"
	CSIBgBlue       = "\033[44m"
	CSIBgMagenta    = "\033[45m"
	CSIBgCyan       = "\033[46m"
	CSIBgWhite      = "\033[47m"
	CSIBgResetColor = "\033[49m"

	CSIFgBrightBlack   = "\033[90m"
	CSIFgBrightRed     = "\033[91m"
	CSIFgBrightGreen   = "\033[92m"
	CSIFgBrightYellow  = "\033[93m"
	CSIFgBrightBlue    = "\033[94m"
	CSIFgBrightMagenta = "\033[95m"
	CSIFgBrightCyan    = "\033[96m"
	CSIFgBrightWhite   = "\033[97m"

	CSIBgBrightBlack   = "\033[100m"
	CSIBgBrightRed     = "\033[101m"
	CSIBgBrightGreen   = "\033[102m"
	CSIBgBrightYellow  = "\033[103m"
	CSIBgBrightBlue    = "\033[104m"
	CSIBgBrightMagenta = "\033[105m"
	CSIBgBrightCyan    = "\033[106m"
	CSIBgBrightWhite   = "\033[107m"
)

func CSIFg256(c int) string {
	return "\033[38;5;" + strconv.Itoa(c&0xff) + "m"
}

func CSIFgRGB(r, g, b int) string {
	return "\033[38;2;" + strconv.Itoa(r&0xff) + ";" + strconv.Itoa(g&0xff) + ";" + strconv.Itoa(b&0xff) + "m"
}

func CSIBg256(c int) string {
	return "\033[48;5;" + strconv.Itoa(c&0xff) + "m"
}

func CSIBgRGB(r, g, b int) string {
	return "\033[48;2;" + strconv.Itoa(r&0xff) + ";" + strconv.Itoa(g&0xff) + ";" + strconv.Itoa(b&0xff) + "m"
}
