package item

import (
	"fmt"
	stdmath "math"
	"unicode/utf8"
)

// WritableBookPage limits, WritableBookPage::PAGE_LENGTH_HARD_LIMIT_BYTES and
// PHOTO_NAME_LENGTH_HARD_LIMIT_BYTES (Limits::INT16_MAX).
const (
	WritableBookPageLengthHardLimitBytes      = stdmath.MaxInt16
	WritableBookPhotoNameLengthHardLimitBytes = stdmath.MaxInt16
)

// WritableBookPage is a port of pocketmine\item\WritableBookPage. The constructor's
// InvalidArgumentException (text or photo name too long, text not valid UTF-8) is a panic, like
// this port's other constructor argument checks; item NBT deserialization recovers it like PHP
// catches it.
type WritableBookPage struct {
	Text      string
	PhotoName string
}

// checkWritableBookPageLength is a port of WritableBookPage::checkLength.
func checkWritableBookPageLength(s, name string, maxLength int) {
	if len(s) > maxLength {
		panic(fmt.Sprintf("%s must be at most %d bytes, but have %d bytes", name, maxLength, len(s)))
	}
}

// NewWritableBookPage is a port of WritableBookPage::__construct with no photo.
func NewWritableBookPage(text string) WritableBookPage {
	return NewWritableBookPageWithPhoto(text, "")
}

// NewWritableBookPageWithPhoto is a port of WritableBookPage::__construct.
func NewWritableBookPageWithPhoto(text, photoName string) WritableBookPage {
	checkWritableBookPageLength(text, "Text", WritableBookPageLengthHardLimitBytes)
	checkWritableBookPageLength(photoName, "Photo name", WritableBookPhotoNameLengthHardLimitBytes)
	if !utf8.ValidString(text) { // Utils::checkUTF8
		panic("Text must be valid UTF-8")
	}
	return WritableBookPage{Text: text, PhotoName: photoName}
}

func (p WritableBookPage) GetText() string { return p.Text }

func (p WritableBookPage) GetPhotoName() string { return p.PhotoName }
