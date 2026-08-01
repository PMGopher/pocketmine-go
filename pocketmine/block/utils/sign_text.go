package blockutils

import (
	"strings"

	"pocketmine-go/pocketmine/color"
)

const SignTextLineCount = 4

// SignText is a port of pocketmine\block\utils\SignText.
type SignText struct {
	lines     [SignTextLineCount]string
	baseColor color.Color
	glowing   bool
}

// NewSignText panics if more than SignTextLineCount lines are given, mirroring the PHP original's
// InvalidArgumentException (a programmer error at the call site). Unlike the PHP constructor,
// UTF-8/newline validation is not enforced here (see Utils.CheckUTF8's doc comment elsewhere in
// this port for the general "not enforcing PHP's string validation helpers" convention) - pass
// nil for baseColor to default to black.
func NewSignText(lines []string, baseColor *color.Color, glowing bool) SignText {
	if len(lines) > SignTextLineCount {
		panic("Expected at most 4 lines")
	}
	t := SignText{glowing: glowing}
	if baseColor != nil {
		t.baseColor = *baseColor
	} else {
		t.baseColor = color.NewColor(0, 0, 0)
	}
	copy(t.lines[:], lines)
	return t
}

// SignTextFromBlob is a port of SignText::fromBlob.
func SignTextFromBlob(blob string, baseColor *color.Color, glowing bool) SignText {
	lines := strings.SplitN(blob, "\n", SignTextLineCount+1)
	if len(lines) > SignTextLineCount {
		lines = lines[:SignTextLineCount]
	}
	return NewSignText(lines, baseColor, glowing)
}

func (t SignText) GetLines() [SignTextLineCount]string { return t.lines }

// GetLine panics if index is out of bounds, mirroring the PHP original's InvalidArgumentException.
func (t SignText) GetLine(index int) string {
	if index < 0 || index >= SignTextLineCount {
		panic("Line index is out of bounds")
	}
	return t.lines[index]
}

func (t SignText) GetBaseColor() color.Color { return t.baseColor }

func (t SignText) IsGlowing() bool { return t.glowing }
