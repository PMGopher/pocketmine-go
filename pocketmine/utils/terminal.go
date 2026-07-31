package utils

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Terminal is a port of pocketmine\utils\Terminal.
//
// PHP's original chooses between `tput`-queried escape codes (Unix, when $TERM is set) and a
// hardcoded fallback table (Windows/Android, or as a general fallback). Shelling out to `tput`
// spawns a process per code and is fragile across differing terminfo databases, so this port
// always uses the fixed 256-colour ANSI codes (PHP's own fallback table) — they render
// identically on any terminal that advertises 256-colour support, which is effectively
// universal today.
var (
	FormatBold          string
	FormatObfuscated    string
	FormatItalic        string
	FormatUnderline     string
	FormatStrikethrough string
	FormatReset         string

	ColorBlack, ColorDarkBlue, ColorDarkGreen, ColorDarkAqua, ColorDarkRed, ColorPurple,
	ColorGold, ColorGray, ColorDarkGray, ColorBlue, ColorGreen, ColorAqua, ColorRed,
	ColorLightPurple, ColorYellow, ColorWhite, ColorMinecoinGold, ColorMaterialQuartz,
	ColorMaterialIron, ColorMaterialNetherite, ColorMaterialRedstone, ColorMaterialCopper,
	ColorMaterialGold, ColorMaterialEmerald, ColorMaterialDiamond, ColorMaterialLapis,
	ColorMaterialAmethyst, ColorMaterialResin string
)

var terminalFormattingEnabled *bool

func HasFormattingCodes() bool {
	if terminalFormattingEnabled == nil {
		panic("Formatting codes have not been initialized")
	}
	return *terminalFormattingEnabled
}

func detectFormattingCodesSupport() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func ansi256(code int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", code)
}

func loadFallbackEscapeCodes() {
	FormatBold = "\x1b[1m"
	FormatObfuscated = ""
	FormatItalic = "\x1b[3m"
	FormatUnderline = "\x1b[4m"
	FormatStrikethrough = "\x1b[9m"
	FormatReset = "\x1b[m"

	ColorBlack = ansi256(16)
	ColorDarkBlue = ansi256(19)
	ColorDarkGreen = ansi256(34)
	ColorDarkAqua = ansi256(37)
	ColorDarkRed = ansi256(124)
	ColorPurple = ansi256(127)
	ColorGold = ansi256(214)
	ColorGray = ansi256(145)
	ColorDarkGray = ansi256(59)
	ColorBlue = ansi256(63)
	ColorGreen = ansi256(83)
	ColorAqua = ansi256(87)
	ColorRed = ansi256(203)
	ColorLightPurple = ansi256(207)
	ColorYellow = ansi256(227)
	ColorWhite = ansi256(231)
	ColorMinecoinGold = ansi256(184)
	ColorMaterialQuartz = ansi256(188)
	ColorMaterialIron = ansi256(251)
	ColorMaterialNetherite = ansi256(237)
	ColorMaterialRedstone = ansi256(88)
	ColorMaterialCopper = ansi256(131)
	ColorMaterialGold = ansi256(178)
	ColorMaterialEmerald = ansi256(35)
	ColorMaterialDiamond = ansi256(37)
	ColorMaterialLapis = ansi256(24)
	ColorMaterialAmethyst = ansi256(98)
	ColorMaterialResin = ansi256(208)
}

// InitTerminal detects (or is told whether) the terminal supports formatting, and loads the
// escape code table if so. Pass nil to auto-detect, matching Terminal::init(null).
func InitTerminal(enableFormatting *bool) {
	enabled := detectFormattingCodesSupport()
	if enableFormatting != nil {
		enabled = *enableFormatting
	}
	terminalFormattingEnabled = &enabled
	if !enabled {
		return
	}
	loadFallbackEscapeCodes()
}

func IsTerminalInit() bool {
	return terminalFormattingEnabled != nil
}

// ToANSI returns a string with colorized ANSI escape codes for the current terminal.
func ToANSI(s string) string {
	var b strings.Builder
	for _, token := range Tokenize(s) {
		switch token {
		case Bold:
			b.WriteString(FormatBold)
		case Obfuscated:
			b.WriteString(FormatObfuscated)
		case Italic:
			b.WriteString(FormatItalic)
		case Reset:
			b.WriteString(FormatReset)
		case Black:
			b.WriteString(ColorBlack)
		case DarkBlue:
			b.WriteString(ColorDarkBlue)
		case DarkGreen:
			b.WriteString(ColorDarkGreen)
		case DarkAqua:
			b.WriteString(ColorDarkAqua)
		case DarkRed:
			b.WriteString(ColorDarkRed)
		case DarkPurple:
			b.WriteString(ColorPurple)
		case Gold:
			b.WriteString(ColorGold)
		case Gray:
			b.WriteString(ColorGray)
		case DarkGray:
			b.WriteString(ColorDarkGray)
		case Blue:
			b.WriteString(ColorBlue)
		case Green:
			b.WriteString(ColorGreen)
		case Aqua:
			b.WriteString(ColorAqua)
		case Red:
			b.WriteString(ColorRed)
		case LightPurple:
			b.WriteString(ColorLightPurple)
		case Yellow:
			b.WriteString(ColorYellow)
		case White:
			b.WriteString(ColorWhite)
		case MinecoinGold:
			b.WriteString(ColorMinecoinGold)
		case MaterialQuartz:
			b.WriteString(ColorMaterialQuartz)
		case MaterialIron:
			b.WriteString(ColorMaterialIron)
		case MaterialNetherite:
			b.WriteString(ColorMaterialNetherite)
		case MaterialRedstone:
			b.WriteString(ColorMaterialRedstone)
		case MaterialCopper:
			b.WriteString(ColorMaterialCopper)
		case MaterialGold:
			b.WriteString(ColorMaterialGold)
		case MaterialEmerald:
			b.WriteString(ColorMaterialEmerald)
		case MaterialDiamond:
			b.WriteString(ColorMaterialDiamond)
		case MaterialLapis:
			b.WriteString(ColorMaterialLapis)
		case MaterialAmethyst:
			b.WriteString(ColorMaterialAmethyst)
		case MaterialResin:
			b.WriteString(ColorMaterialResin)
		default:
			b.WriteString(token)
		}
	}
	return b.String()
}

// WriteTerminal emits a string containing Minecraft colour codes to the console, formatted
// with native colours.
func WriteTerminal(line string) {
	fmt.Print(ToANSI(line))
}

// WriteTerminalLine is WriteTerminal followed by a reset and a newline.
func WriteTerminalLine(line string) {
	fmt.Println(ToANSI(line) + FormatReset)
}
