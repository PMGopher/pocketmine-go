package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// WritableBookBase is a forward-compatible marker for pocketmine\item\WritableBookBase - same
// pattern as the Axe/Dye/Durable/Shovel markers elsewhere. PageExists is structurally identical
// to the real item.WritableBookBase.PageExists (both take a page index and return a bool), so any
// real writable/written book satisfies this automatically despite the self-referential-return-type
// limitation documented on ItemFrame - PageExists never needs to return an Item.
type WritableBookBase interface {
	Item
	PageExists(pageID int) bool
}

// Lectern is a port of pocketmine\block\Lectern, minus cloning the book on every get/set (block.Item
// has no Clone() method - same deviation as ItemFrame.GetFramedItem/SetFramedItem) and actually
// dropping the ejected book into the world on OnAttack (World.DropItem isn't in the ported World
// interface - see SweetBerryBush's doc comment for the same gap). Turning pages, placement-
// facing, and the producing-signal state machine are all fully real.
type Lectern struct {
	Transparent
	HorizontalFacingComponent

	ViewedPage      int
	Book            WritableBookBase
	ProducingSignal bool
}

func NewLectern(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Lectern {
	l := &Lectern{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	l.Init(l)
	return l
}

func (l *Lectern) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Lectern) DescribeBlockOnlyState(w runtime.DataDescriber) {
	l.DescribeHorizontalFacing(w)
	w.Bool(&l.ProducingSignal)
}

// ReadStateFromWorld is a port of Lectern::readStateFromWorld.
func (l *Lectern) ReadStateFromWorld() Behavior {
	l.Transparent.ReadStateFromWorld()
	world, err := l.position.GetWorld()
	if err != nil {
		return l.self
	}
	t, ok := world.GetTile(l.position)
	if !ok {
		return l.self
	}
	tileLectern, ok := t.(*tile.Lectern)
	if !ok {
		return l.self
	}
	l.ViewedPage = tileLectern.GetViewedPage()
	l.Book = nil
	if tileBook, has := tileLectern.GetBook(); has {
		if wb, ok := tileBook.(WritableBookBase); ok {
			l.Book = wb
		}
	}
	return l.self
}

func (l *Lectern) GetFlammability() int { return 30 }

func (l *Lectern) GetDrops(item Item) []Item {
	drops := l.Block.GetDrops(item)
	if l.Book != nil {
		drops = append(drops, l.Book)
	}
	return drops
}

func (l *Lectern) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 0.1)}
}

func (l *Lectern) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (l *Lectern) IsProducingSignal() bool { return l.ProducingSignal }

func (l *Lectern) SetProducingSignal(producingSignal bool) { l.ProducingSignal = producingSignal }

func (l *Lectern) GetViewedPage() int { return l.ViewedPage }

func (l *Lectern) SetViewedPage(viewedPage int) { l.ViewedPage = viewedPage }

// GetBook returns the book directly (no defensive clone - see type doc comment).
func (l *Lectern) GetBook() WritableBookBase { return l.Book }

// SetBook is a port of Lectern::setBook, minus cloning/re-counting the incoming book (see type
// doc comment).
func (l *Lectern) SetBook(book WritableBookBase) {
	if book == nil || book.IsNull() {
		l.Book = nil
	} else {
		l.Book = book
	}
	l.ViewedPage = 0
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (l *Lectern) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		l.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of Lectern::onInteract.
func (l *Lectern) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if l.Book == nil {
		if wb, ok := item.(WritableBookBase); ok {
			l.SetBook(wb)
			l.setSelf()
			l.addSound(sound.LecternPlaceBookSound{})
			item.Pop()
		}
	}
	return true
}

// OnAttack is a port of Lectern::onAttack, minus actually dropping the ejected book into the
// world (see type doc comment).
func (l *Lectern) OnAttack(item Item, face math.Facing, player Player) bool {
	if l.Book != nil {
		l.SetBook(nil)
		l.setSelf()
	}
	return false
}

// OnPageTurn is a port of Lectern::onPageTurn.
func (l *Lectern) OnPageTurn(newPage int) bool {
	if newPage == l.ViewedPage {
		return true
	}
	if l.Book == nil || !l.Book.PageExists(newPage) {
		return false
	}
	l.ViewedPage = newPage
	if !l.ProducingSignal {
		l.ProducingSignal = true
		world, err := l.position.GetWorld()
		if err == nil {
			world.ScheduleDelayedBlockUpdate(l.position.Vector3, 1)
		}
	}
	l.setSelf()
	return true
}

// OnScheduledUpdate is a port of Lectern::onScheduledUpdate.
func (l *Lectern) OnScheduledUpdate() {
	if l.ProducingSignal {
		l.ProducingSignal = false
		l.setSelf()
	}
}

func (l *Lectern) addSound(s sound.Sound) {
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(l.position.Vector3, s)
}

func (l *Lectern) setSelf() {
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(l.position, l.self)
}
