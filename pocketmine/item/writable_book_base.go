package item

import (
	"fmt"

	"pocketmine-go/pocketmine/nbt"
)

const (
	writableBookTagPages     = "pages"
	writableBookTagPageText  = "text"
	writableBookTagPagePhoto = "photoname"
)

// WritableBookBase is a port of pocketmine\item\WritableBookBase.
type WritableBookBase struct {
	ItemBase

	Pages []WritableBookPage
}

func (w *WritableBookBase) PageExists(pageID int) bool {
	return pageID >= 0 && pageID < len(w.Pages)
}

// GetPageText panics if the page doesn't exist, mirroring the PHP original's OutOfRangeException
// (a programmer error at the call site) - same convention as e.g. block.PitcherCrop.SetAge.
func (w *WritableBookBase) GetPageText(pageID int) string {
	if !w.PageExists(pageID) {
		panic(fmt.Sprintf("Page %d does not exist", pageID))
	}
	return w.Pages[pageID].Text
}

// SetPageText is a port of WritableBookBase::setPageText.
func (w *WritableBookBase) SetPageText(pageID int, text string) {
	if !w.PageExists(pageID) {
		w.AddPage(pageID)
	}
	w.Pages[pageID] = NewWritableBookPage(text)
}

// AddPage is a port of WritableBookBase::addPage.
func (w *WritableBookBase) AddPage(pageID int) {
	if pageID < 0 {
		panic(fmt.Sprintf("Page number %d is out of range", pageID))
	}
	for current := len(w.Pages); current <= pageID; current++ {
		w.Pages = append(w.Pages, NewWritableBookPage(""))
	}
}

// DeletePage is a port of WritableBookBase::deletePage.
func (w *WritableBookBase) DeletePage(pageID int) {
	if !w.PageExists(pageID) {
		return
	}
	w.Pages = append(w.Pages[:pageID], w.Pages[pageID+1:]...)
}

// InsertPage is a port of WritableBookBase::insertPage.
func (w *WritableBookBase) InsertPage(pageID int, text string) {
	if pageID < 0 || pageID > len(w.Pages) {
		panic("Page ID must not be negative")
	}
	newPages := make([]WritableBookPage, 0, len(w.Pages)+1)
	newPages = append(newPages, w.Pages[:pageID]...)
	newPages = append(newPages, NewWritableBookPage(text))
	newPages = append(newPages, w.Pages[pageID:]...)
	w.Pages = newPages
}

// SwapPages is a port of WritableBookBase::swapPages.
func (w *WritableBookBase) SwapPages(pageID1, pageID2 int) bool {
	text1 := w.GetPageText(pageID1)
	text2 := w.GetPageText(pageID2)
	w.SetPageText(pageID1, text2)
	w.SetPageText(pageID2, text1)
	return true
}

func (w *WritableBookBase) GetMaxStackSize() int { return 1 }

func (w *WritableBookBase) GetPages() []WritableBookPage { return w.Pages }

func (w *WritableBookBase) SetPages(pages []WritableBookPage) { w.Pages = pages }

// deserializeCompoundTag/serializeCompoundTag are WritableBookBase's participation in the
// compoundTagCodec self-dispatch chain (see ItemBase.GetNamedTag's doc comment). Only the PE
// (compound-tag) page format is handled on write, matching the PHP original's serializeCompoundTag;
// on read, both the PE format and the legacy PC (plain string list) format are recognised,
// matching deserializeCompoundTag's dual-format cast() checks.
func (w *WritableBookBase) deserializeCompoundTag(tag *nbt.CompoundTag) {
	w.ItemBase.deserializeCompoundTag(tag)
	w.Pages = nil

	pages, ok, _ := tag.GetListTag(writableBookTagPages)
	if !ok {
		return
	}
	switch pages.GetTagType() {
	case nbt.TagCompound:
		for _, v := range pages.Values() {
			page, ok := v.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			text := string(page.GetStringOr(writableBookTagPageText, ""))
			photo := string(page.GetStringOr(writableBookTagPagePhoto, ""))
			w.Pages = append(w.Pages, NewWritableBookPageWithPhoto(text, photo))
		}
	case nbt.TagString:
		for _, v := range pages.Values() {
			if s, ok := v.(nbt.StringTag); ok {
				w.Pages = append(w.Pages, NewWritableBookPage(string(s)))
			}
		}
	}
}

func (w *WritableBookBase) serializeCompoundTag(tag *nbt.CompoundTag) {
	w.ItemBase.serializeCompoundTag(tag)
	if len(w.Pages) == 0 {
		tag.RemoveTag(writableBookTagPages)
		return
	}
	values := make([]nbt.Tag, len(w.Pages))
	for i, page := range w.Pages {
		values[i] = nbt.NewCompoundTag().
			SetString(writableBookTagPageText, nbt.StringTag(page.Text)).
			SetString(writableBookTagPagePhoto, nbt.StringTag(page.PhotoName))
	}
	pages, err := nbt.NewListTag(values, nbt.TagCompound)
	if err != nil {
		panic(err)
	}
	tag.SetTag(writableBookTagPages, pages)
}
