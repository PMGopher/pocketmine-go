package block

import (
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/world/sound"
)

// Water is a port of pocketmine\block\Water.
type Water struct {
	Liquid
}

func NewWater(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Water {
	w := &Water{Liquid{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}}
	w.Init(w)
	return w
}

func (w *Water) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *Water) GetLightFilter() int { return 2 }

func (w *Water) GetBucketFillSound() sound.Sound { return sound.BucketFillWaterSound{} }

func (w *Water) GetBucketEmptySound() sound.Sound { return sound.BucketEmptyWaterSound{} }

func (w *Water) TickRate() int { return 5 }

func (w *Water) checkForHarden() bool { return false }

func (w *Water) GetMinAdjacentSourcesToFormSource() (int, bool) { return 2, true }

func (w *Water) OnEntityInside(e Entity) bool {
	e.ResetFallDistance()
	if e.IsOnFire() {
		e.ExtinguishWithCause(entity.EntityExtinguishCauseWater)
	}
	return true
}
