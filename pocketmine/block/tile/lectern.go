package tile

import (
	"fmt"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	lecternTagHasBook    = "hasBook"
	lecternTagPage       = "page"
	lecternTagTotalPages = "totalPages"
	lecternTagBook       = "book"
)

// Lectern is a port of pocketmine\block\tile\Lectern. The book is this package's minimal Item
// (a WritableBookBase), nil for none.
type Lectern struct {
	SpawnableBase

	viewedPage int
	book       Item
}

func NewLectern(world World, pos math.Vector3) *Lectern {
	l := &Lectern{}
	l.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	l.Init(l)
	return l
}

func (l *Lectern) SaveID() string { return "Lectern" }

func (l *Lectern) GetViewedPage() int { return l.viewedPage }

func (l *Lectern) SetViewedPage(viewedPage int) { l.viewedPage = viewedPage }

func (l *Lectern) GetBook() (Item, bool) { return l.book, l.book != nil }

// SetBook is a port of pocketmine\block\tile\Lectern::setBook (the block passes a copy).
func (l *Lectern) SetBook(book Item) {
	if isNullItem(book) {
		l.book = nil
	} else {
		l.book = book
	}
}

// isWritableBook is `$book instanceof WritableBookBase`.
func isWritableBook(it Item) bool {
	_, ok := it.(interface{ PageExists(pageID int) bool })
	return ok
}

// ReadSaveData is a port of Lectern::readSaveData.
func (l *Lectern) ReadSaveData(tag *nbt.CompoundTag) error {
	l.viewedPage = int(tag.GetIntOr(lecternTagPage, 0))
	itemTag, ok, err := tag.GetCompoundTag(lecternTagBook)
	if err != nil {
		return err
	}
	if ok {
		book := loadItem(itemTag, fmt.Sprintf("Lectern (%v) book", l.GetPosition().Vector3))
		if isWritableBook(book) && !isNullItem(book) {
			l.book = book
		}
	}
	return nil
}

// WriteSaveData is a port of Lectern::writeSaveData.
func (l *Lectern) WriteSaveData(tag *nbt.CompoundTag) {
	hasBook := nbt.ByteTag(0)
	if l.book != nil {
		hasBook = 1
	}
	tag.SetByte(lecternTagHasBook, hasBook)
	tag.SetInt(lecternTagPage, nbt.IntTag(l.viewedPage))
	if l.book != nil {
		if bookTag := saveItem(l.book, -1); bookTag != nil {
			tag.SetTag(lecternTagBook, bookTag)
		}
		tag.SetInt(lecternTagTotalPages, nbt.IntTag(bookPageCount(l.book)))
	}
}

// AddAdditionalSpawnData is a port of Lectern::addAdditionalSpawnData.
func (l *Lectern) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	hasBook := nbt.ByteTag(0)
	if l.book != nil {
		hasBook = 1
	}
	tag.SetByte(lecternTagHasBook, hasBook)
	tag.SetInt(lecternTagPage, nbt.IntTag(l.viewedPage))
	if l.book != nil {
		if bookTag := networkItemNbt(l.book); bookTag != nil {
			tag.SetTag(lecternTagBook, bookTag)
		}
		tag.SetInt(lecternTagTotalPages, nbt.IntTag(bookPageCount(l.book)))
	}
}
