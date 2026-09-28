package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/math"
)

// LiquidBucket is a port of pocketmine\item\LiquidBucket.
type LiquidBucket struct {
	ItemBase

	Liquid block.Behavior
}

func NewLiquidBucket(identifier ItemIdentifier, name string, liquid block.Behavior) *LiquidBucket {
	l := &LiquidBucket{Liquid: liquid}
	l.Init(l, identifier, name)
	return l
}

func (l *LiquidBucket) Clone() Item {
	c := *l
	c.Liquid = l.Liquid.Clone()
	c.rebind(&c)
	return &c
}

func (l *LiquidBucket) GetMaxStackSize() int { return 1 }

// GetFuelTime is a port of LiquidBucket::getFuelTime.
func (l *LiquidBucket) GetFuelTime() int {
	if _, ok := l.Liquid.(*block.Lava); ok {
		return 20000
	}
	return 0
}

func (l *LiquidBucket) GetLiquid() block.Behavior { return l.Liquid }

// GetFuelResidue is a port of LiquidBucket::getFuelResidue.
func (l *LiquidBucket) GetFuelResidue() Item { return VanillaBucket() }

// OnInteractBlock is a port of LiquidBucket::onInteractBlock: emptying the bucket onto a
// replaceable block.
func (l *LiquidBucket) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if !blockReplace.CanBeReplaced() {
		return ItemUseResultNone
	}

	//TODO: move this to generic placement logic
	resultBlock, ok := l.Liquid.Clone().(liquidBlock)
	if !ok {
		return ItemUseResultNone
	}

	p, ok := player.(playerevent.Player)
	if !ok {
		return ItemUseResultNone
	}
	ev := playerevent.NewPlayerBucketEmptyEvent(p, blockReplace, int(face), l, VanillaBucket())
	event.Call(ev)
	if ev.IsCancelled() {
		return ItemUseResultFail
	}
	pos := blockReplace.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return ItemUseResultFail
	}
	_ = world.SetBlock(pos, resultBlock.GetFlowingForm())
	world.AddSound(pos.Add(0.5, 0.5, 0.5), resultBlock.GetBucketEmptySound())

	l.Pop()
	*returnedItems = append(*returnedItems, ev.GetItem().(Item))
	return ItemUseResultSuccess
}
