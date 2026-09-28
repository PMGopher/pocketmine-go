package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Bucket is a port of pocketmine\item\Bucket (the empty bucket).
type Bucket struct {
	ItemBase
}

func NewBucket(identifier ItemIdentifier, name string) *Bucket {
	b := &Bucket{}
	b.Init(b, identifier, name)
	return b
}

func (b *Bucket) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bucket) GetMaxStackSize() int { return 16 }

// liquidBlock is the part of block.Liquid Bucket::onInteractBlock needs (`instanceof Liquid`).
type liquidBlock interface {
	block.Behavior
	IsSource() bool
	GetBucketFillSound() sound.Sound
	GetBucketEmptySound() sound.Sound
	GetFlowingForm() block.Behavior
}

// OnInteractBlock is a port of Bucket::onInteractBlock: filling the bucket from a liquid source.
func (b *Bucket) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	//TODO: move this to generic placement logic
	liquid, ok := blockClicked.(liquidBlock)
	if !ok || !liquid.IsSource() {
		return ItemUseResultNone
	}

	var resultItem Item
	switch blockClicked.GetTypeId() {
	case block.LAVA:
		resultItem = VanillaItem("lava_bucket")
	case block.WATER:
		resultItem = VanillaItem("water_bucket")
	default:
		return ItemUseResultFail
	}

	p, ok := player.(playerevent.Player)
	if !ok {
		return ItemUseResultNone
	}
	ev := playerevent.NewPlayerBucketFillEvent(p, blockReplace, int(face), b, resultItem)
	event.Call(ev)
	if ev.IsCancelled() {
		return ItemUseResultFail
	}
	pos := blockClicked.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return ItemUseResultFail
	}
	_ = world.SetBlock(pos, block.VanillaAir())
	world.AddSound(pos.Add(0.5, 0.5, 0.5), liquid.GetBucketFillSound())

	b.Pop()
	*returnedItems = append(*returnedItems, ev.GetItem().(Item))
	return ItemUseResultSuccess
}
