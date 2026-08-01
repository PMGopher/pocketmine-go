package item

import "testing"

var (
	_ Item = (*Book)(nil)
	_ Item = (*EnchantedBook)(nil)
	_ Item = (*WritableBook)(nil)
	_ Item = (*WrittenBook)(nil)
)

func TestEnchantedBookMaxStackSize(t *testing.T) {
	e := NewEnchantedBook(NewItemIdentifier(ENCHANTED_BOOK), "Enchanted Book")
	if e.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", e.GetMaxStackSize())
	}
}

func TestWritableBookAddAndGetPageText(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "Hello")
	w.SetPageText(2, "World")

	if !w.PageExists(0) || !w.PageExists(1) || !w.PageExists(2) {
		t.Fatal("expected SetPageText(2, ...) to backfill pages 0-2")
	}
	if w.GetPageText(0) != "Hello" {
		t.Errorf("GetPageText(0) = %q, want %q", w.GetPageText(0), "Hello")
	}
	if w.GetPageText(1) != "" {
		t.Errorf("GetPageText(1) = %q, want empty (backfilled)", w.GetPageText(1))
	}
	if w.GetPageText(2) != "World" {
		t.Errorf("GetPageText(2) = %q, want %q", w.GetPageText(2), "World")
	}
}

func TestWritableBookGetPageTextPanicsWhenMissing(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	defer func() {
		if recover() == nil {
			t.Error("expected GetPageText to panic for a nonexistent page")
		}
	}()
	w.GetPageText(0)
}

func TestWritableBookDeletePage(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "A")
	w.SetPageText(1, "B")
	w.SetPageText(2, "C")

	w.DeletePage(1)

	if len(w.GetPages()) != 2 {
		t.Fatalf("len(GetPages()) = %d, want 2", len(w.GetPages()))
	}
	if w.GetPageText(0) != "A" || w.GetPageText(1) != "C" {
		t.Errorf("pages after delete = [%q, %q], want [A, C]", w.GetPageText(0), w.GetPageText(1))
	}
}

func TestWritableBookInsertPage(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "A")
	w.SetPageText(1, "C")

	w.InsertPage(1, "B")

	if w.GetPageText(0) != "A" || w.GetPageText(1) != "B" || w.GetPageText(2) != "C" {
		t.Errorf("pages after insert = [%q, %q, %q], want [A, B, C]", w.GetPageText(0), w.GetPageText(1), w.GetPageText(2))
	}
}

func TestWritableBookSwapPages(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "A")
	w.SetPageText(1, "B")

	if !w.SwapPages(0, 1) {
		t.Fatal("expected SwapPages to succeed")
	}
	if w.GetPageText(0) != "B" || w.GetPageText(1) != "A" {
		t.Errorf("pages after swap = [%q, %q], want [B, A]", w.GetPageText(0), w.GetPageText(1))
	}
}

func TestWritableBookPagesRoundTripThroughNBT(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "Page one")
	w.SetPageText(1, "Page two")

	decoded := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	decoded.SetNamedTag(w.GetNamedTag())

	pages := decoded.GetPages()
	if len(pages) != 2 {
		t.Fatalf("len(GetPages()) = %d, want 2", len(pages))
	}
	if pages[0].GetText() != "Page one" || pages[1].GetText() != "Page two" {
		t.Errorf("pages = %+v, want [Page one, Page two]", pages)
	}
}

func TestWritableBookCloneIsIndependent(t *testing.T) {
	w := NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Writable Book")
	w.SetPageText(0, "Original")

	clone := w.Clone().(*WritableBook)
	clone.SetPageText(0, "Changed")

	if w.GetPageText(0) != "Original" {
		t.Error("expected cloning not to affect the original book's pages")
	}
}

func TestWrittenBookDefaultsAndSetters(t *testing.T) {
	wb := NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	if wb.GetGeneration() != WrittenBookGenerationOriginal {
		t.Errorf("GetGeneration() = %d, want Original", wb.GetGeneration())
	}
	if wb.GetMaxStackSize() != 16 {
		t.Errorf("GetMaxStackSize() = %d, want 16", wb.GetMaxStackSize())
	}

	wb.SetAuthor("Steve")
	wb.SetTitle("My Book")
	wb.SetGeneration(WrittenBookGenerationCopy)

	if wb.GetAuthor() != "Steve" || wb.GetTitle() != "My Book" || wb.GetGeneration() != WrittenBookGenerationCopy {
		t.Errorf("got author=%q title=%q generation=%d", wb.GetAuthor(), wb.GetTitle(), wb.GetGeneration())
	}
}

func TestWrittenBookSetGenerationRejectsOutOfRange(t *testing.T) {
	wb := NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	defer func() {
		if recover() == nil {
			t.Error("expected SetGeneration to panic for an out-of-range value")
		}
	}()
	wb.SetGeneration(4)
}

func TestWrittenBookRoundTripsThroughNBTIncludingPages(t *testing.T) {
	wb := NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	wb.SetAuthor("Steve")
	wb.SetTitle("My Book")
	wb.SetGeneration(WrittenBookGenerationCopyOfCopy)
	wb.SetPageText(0, "Once upon a time")

	decoded := NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	decoded.SetNamedTag(wb.GetNamedTag())

	if decoded.GetAuthor() != "Steve" || decoded.GetTitle() != "My Book" {
		t.Errorf("got author=%q title=%q", decoded.GetAuthor(), decoded.GetTitle())
	}
	if decoded.GetGeneration() != WrittenBookGenerationCopyOfCopy {
		t.Errorf("GetGeneration() = %d, want CopyOfCopy", decoded.GetGeneration())
	}
	if len(decoded.GetPages()) != 1 || decoded.GetPageText(0) != "Once upon a time" {
		t.Errorf("pages = %+v, want [Once upon a time]", decoded.GetPages())
	}
}
