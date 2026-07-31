package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// TextFormat is a port of pocketmine\utils\TextFormat: handling of Minecraft chat
// formatting codes, and conversion to other formats like HTML.
const (
	Escape = "\xc2\xa7" // §
	EOL    = "\n"

	Black             = Escape + "0"
	DarkBlue          = Escape + "1"
	DarkGreen         = Escape + "2"
	DarkAqua          = Escape + "3"
	DarkRed           = Escape + "4"
	DarkPurple        = Escape + "5"
	Gold              = Escape + "6"
	Gray              = Escape + "7"
	DarkGray          = Escape + "8"
	Blue              = Escape + "9"
	Green             = Escape + "a"
	Aqua              = Escape + "b"
	Red               = Escape + "c"
	LightPurple       = Escape + "d"
	Yellow            = Escape + "e"
	White             = Escape + "f"
	MinecoinGold      = Escape + "g"
	MaterialQuartz    = Escape + "h"
	MaterialIron      = Escape + "i"
	MaterialNetherite = Escape + "j"
	MaterialRedstone  = Escape + "m"
	MaterialCopper    = Escape + "n"
	MaterialGold      = Escape + "p"
	MaterialEmerald   = Escape + "q"
	MaterialDiamond   = Escape + "s"
	MaterialLapis     = Escape + "t"
	MaterialAmethyst  = Escape + "u"
	MaterialResin     = Escape + "v"

	Obfuscated = Escape + "k"
	Bold       = Escape + "l"
	Italic     = Escape + "o"

	Reset = Escape + "r"
)

// Colors is the set of all valid Minecraft color codes (as a lookup set).
var Colors = map[string]bool{
	Black: true, DarkBlue: true, DarkGreen: true, DarkAqua: true, DarkRed: true,
	DarkPurple: true, Gold: true, Gray: true, DarkGray: true, Blue: true,
	Green: true, Aqua: true, Red: true, LightPurple: true, Yellow: true, White: true,
	MinecoinGold: true, MaterialQuartz: true, MaterialIron: true, MaterialNetherite: true,
	MaterialRedstone: true, MaterialCopper: true, MaterialGold: true, MaterialEmerald: true,
	MaterialDiamond: true, MaterialLapis: true, MaterialAmethyst: true, MaterialResin: true,
}

// Formats is the set of all valid Minecraft formatting (non-color) codes.
var Formats = map[string]bool{
	Obfuscated: true, Bold: true, Italic: true,
}

var (
	formatTokenPattern    = regexp.MustCompile("(" + Escape + "[0-9a-v])")
	privateUseAreaPattern = regexp.MustCompile(`[\x{E000}-\x{F8FF}]`)
	formatCodePattern     = regexp.MustCompile(Escape + "[0-9a-v]")
	ansiEscapePattern     = regexp.MustCompile("\x1b[\\(\\][0-9;\\[\\(]+[Bm]")
)

// Tokenize splits the string by format tokens, keeping the tokens themselves in the result.
func Tokenize(s string) []string {
	matches := formatTokenPattern.FindAllStringIndex(s, -1)
	result := make([]string, 0, len(matches)*2+1)
	last := 0
	for _, m := range matches {
		if m[0] > last {
			result = append(result, s[last:m[0]])
		}
		result = append(result, s[m[0]:m[1]])
		last = m[1]
	}
	if last < len(s) {
		result = append(result, s[last:])
	}
	return result
}

// Clean cleans the string from Minecraft codes, ANSI escape codes and invalid UTF-8 characters.
//
// mb_scrub()'s default substitute is approximated here with U+FFFD, matching Go's own
// invalid-UTF-8 replacement convention; PHP's exact substitute character depends on
// mb_substitute_character() and may differ byte-for-byte in edge cases.
func Clean(s string, removeFormat bool) string {
	s = strings.ToValidUTF8(s, "�")
	s = privateUseAreaPattern.ReplaceAllString(s, "")
	if removeFormat {
		s = strings.ReplaceAll(formatCodePattern.ReplaceAllString(s, ""), Escape, "")
	}
	return strings.ReplaceAll(ansiEscapePattern.ReplaceAllString(s, ""), "\x1b", "")
}

// Colorize replaces placeholders of § with the correct character. Only valid codes are converted.
func Colorize(s string, placeholder string) string {
	pattern := regexp.MustCompile(regexp.QuoteMeta(placeholder) + "([0-9a-v])")
	return pattern.ReplaceAllString(s, Escape+"$1")
}

// AddBase adds base formatting to the string. The given format codes are inserted directly
// after any RESET (§r) codes, so that resets return to the base format rather than the
// terminal's default colour.
func AddBase(baseFormat string, s string) string {
	for _, part := range Tokenize(baseFormat) {
		if !Formats[part] && !Colors[part] {
			panic(fmt.Sprintf("Unexpected base format token %q, expected only color and format tokens", part))
		}
	}
	baseFormat = Reset + baseFormat
	return baseFormat + strings.ReplaceAll(s, Reset, baseFormat)
}

// JavaToBedrock converts any Java formatting codes in the given string to Bedrock.
//
// As of 1.21.50, strikethrough (§m) and underline (§n) are not supported by Bedrock, and
// these symbols instead represent additional colours in Bedrock. To avoid unintended
// formatting, this strips those codes rather than translating them.
func JavaToBedrock(s string) string {
	return strings.NewReplacer(Escape+"m", "", Escape+"n", "").Replace(s)
}

var htmlColors = map[string]string{
	Black: "color:#000", DarkBlue: "color:#00A", DarkGreen: "color:#0A0", DarkAqua: "color:#0AA",
	DarkRed: "color:#A00", DarkPurple: "color:#A0A", Gold: "color:#FA0", Gray: "color:#AAA",
	DarkGray: "color:#555", Blue: "color:#55F", Green: "color:#5F5", Aqua: "color:#5FF",
	Red: "color:#F55", LightPurple: "color:#F5F", Yellow: "color:#FF5", White: "color:#FFF",
	MinecoinGold: "color:#dd0", MaterialQuartz: "color:#e2d3d1", MaterialIron: "color:#cec9c9",
	MaterialNetherite: "color:#44393a", MaterialRedstone: "color:#961506", MaterialCopper: "color:#b4684d",
	MaterialGold: "color:#deb02c", MaterialEmerald: "color:#119f36", MaterialDiamond: "color:#2cb9a8",
	MaterialLapis: "color:#20487a", MaterialAmethyst: "color:#9a5cc5", MaterialResin: "color:#fc7812",
	Bold: "font-weight:bold", Italic: "font-style:italic",
}

// ToHTML returns an HTML-formatted string with colors/markup.
func ToHTML(s string) string {
	var b strings.Builder
	tokens := 0
	for _, token := range Tokenize(s) {
		if formatString, ok := htmlColors[token]; ok {
			b.WriteString(fmt.Sprintf(`<span style="%s">`, formatString))
			tokens++
		} else if token == Reset {
			b.WriteString(strings.Repeat("</span>", tokens))
			tokens = 0
		} else {
			b.WriteString(token)
		}
	}
	b.WriteString(strings.Repeat("</span>", tokens))
	return b.String()
}
