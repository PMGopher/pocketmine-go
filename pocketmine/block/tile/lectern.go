package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	lecternTagHasBook    = "hasBook"
	lecternTagPage       = "page"
	lecternTagTotalPages = "totalPages"
)

// Lectern is a port of pocketmine\block\tile\Lectern, minus the book's NBT round-trip
// (Item::safeNbtDeserialize/nbtSerialize aren't ported - see Jukebox's doc comment for the same
// gap) and the network spawn-data translation (TypeConverter isn't ported either). The book is
// instead held as this package's own minimal Item interface, with nil standing in for "no book" -
// same shape as ItemFrame's framed item (see its doc comment for why there's no Air-like sentinel
// here). ViewedPage fully round-trips through NBT.
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

// SetBook is a port of pocketmine\block\tile\Lectern::setBook, minus cloning the incoming book
// (the null-check is the caller's responsibility too - see ItemFrame.SetItem's doc comment for
// the same shape).
func (l *Lectern) SetBook(book Item) { l.book = book }

func (l *Lectern) ReadSaveData(tag *nbt.CompoundTag) error {
	l.viewedPage = int(tag.GetIntOr(lecternTagPage, 0))
	return nil
}

func (l *Lectern) WriteSaveData(tag *nbt.CompoundTag) {
	hasBook := nbt.ByteTag(0)
	if l.book != nil {
		hasBook = 1
	}
	tag.SetByte(lecternTagHasBook, hasBook)
	tag.SetInt(lecternTagPage, nbt.IntTag(l.viewedPage))
}

// AddAdditionalSpawnData is a port of Lectern::addAdditionalSpawnData, minus the item-to-network-
// NBT translation (TypeConverter isn't ported) and the total-page-count tag (would need the
// unported item's page list) - with no real book NBT to read from anyway, this only reports
// hasBook/page, same reduced shape as WriteSaveData above.
func (l *Lectern) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	hasBook := nbt.ByteTag(0)
	if l.book != nil {
		hasBook = 1
	}
	tag.SetByte(lecternTagHasBook, hasBook)
	tag.SetInt(lecternTagPage, nbt.IntTag(l.viewedPage))
}
