package item

import (
	"fmt"
	stdmath "math"
	"unicode/utf8"

	"pocketmine-go/pocketmine/nbt"
)

const (
	WrittenBookGenerationOriginal   = 0
	WrittenBookGenerationCopy       = 1
	WrittenBookGenerationCopyOfCopy = 2
	WrittenBookGenerationTattered   = 3
)

const (
	writtenBookTagGeneration = "generation"
	writtenBookTagAuthor     = "author"
	writtenBookTagTitle      = "title"
)

// WrittenBook is a port of pocketmine\item\WrittenBook.
type WrittenBook struct {
	WritableBookBase

	Generation int
	Author     string
	Title      string
}

func NewWrittenBook(identifier ItemIdentifier, name string) *WrittenBook {
	w := &WrittenBook{Generation: WrittenBookGenerationOriginal}
	w.Init(w, identifier, name)
	return w
}

func (w *WrittenBook) Clone() Item {
	c := *w
	c.Pages = append([]WritableBookPage(nil), w.Pages...)
	c.rebind(&c)
	return &c
}

func (w *WrittenBook) GetMaxStackSize() int { return 16 }

func (w *WrittenBook) GetGeneration() int { return w.Generation }

// SetGeneration panics if generation is out of range, mirroring the PHP original's
// InvalidArgumentException (a programmer error at the call site).
func (w *WrittenBook) SetGeneration(generation int) {
	if generation < 0 || generation > 3 {
		panic("Generation is out of range")
	}
	w.Generation = generation
}

func (w *WrittenBook) GetAuthor() string { return w.Author }

// checkWrittenBookString is WrittenBook::setAuthor/setTitle's checks: at most Limits::INT16_MAX
// bytes and valid UTF-8 (InvalidArgumentException, a panic here).
func checkWrittenBookString(s, name string) {
	if len(s) > stdmath.MaxInt16 {
		panic(fmt.Sprintf("%s must be at most %d bytes, but have %d bytes", name, stdmath.MaxInt16, len(s)))
	}
	if !utf8.ValidString(s) {
		panic(name + " must be valid UTF-8")
	}
}

// SetAuthor is a port of WrittenBook::setAuthor.
func (w *WrittenBook) SetAuthor(author string) {
	checkWrittenBookString(author, "Author")
	w.Author = author
}

func (w *WrittenBook) GetTitle() string { return w.Title }

// SetTitle is a port of WrittenBook::setTitle.
func (w *WrittenBook) SetTitle(title string) {
	checkWrittenBookString(title, "Title")
	w.Title = title
}

// deserializeCompoundTag/serializeCompoundTag extend WritableBookBase's own pair, the same
// self-dispatch participation described on Durable's.
func (w *WrittenBook) deserializeCompoundTag(tag *nbt.CompoundTag) {
	w.WritableBookBase.deserializeCompoundTag(tag)
	w.Generation = int(tag.GetIntOr(writtenBookTagGeneration, nbt.IntTag(w.Generation)))
	w.Author = string(tag.GetStringOr(writtenBookTagAuthor, nbt.StringTag(w.Author)))
	w.Title = string(tag.GetStringOr(writtenBookTagTitle, nbt.StringTag(w.Title)))
}

func (w *WrittenBook) serializeCompoundTag(tag *nbt.CompoundTag) {
	w.WritableBookBase.serializeCompoundTag(tag)
	tag.SetInt(writtenBookTagGeneration, nbt.IntTag(w.Generation))
	tag.SetString(writtenBookTagAuthor, nbt.StringTag(w.Author))
	tag.SetString(writtenBookTagTitle, nbt.StringTag(w.Title))
}
