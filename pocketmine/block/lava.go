package block

import (
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/world/sound"
)

// Lava is a port of pocketmine\block\Lava.
type Lava struct {
	Liquid
}

func NewLava(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Lava {
	l := &Lava{Liquid{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}}
	l.Init(l)
	return l
}

func (l *Lava) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Lava) GetLightLevel() int { return 15 }

func (l *Lava) GetBucketFillSound() sound.Sound { return sound.BucketFillLavaSound{} }

func (l *Lava) GetBucketEmptySound() sound.Sound { return sound.BucketEmptyLavaSound{} }

func (l *Lava) TickRate() int { return 30 }

func (l *Lava) GetFlowDecayPerBlock() int { return 2 } // TODO: this is 1 in the nether

// checkForHarden is a port of Lava::checkForHarden. The PHP original turns adjacent Water into
// Obsidian/Cobblestone (or, on Soul Soil, adjacent Blue Ice into Basalt) via liquidCollide, which
// needs BlockEventHelper.Form and the block registry, neither ported yet, so this is a no-op stub
// for now (same category of gap as everywhere else BlockEventHelper is needed).
func (l *Lava) checkForHarden() bool { return false }

// OnEntityInside is a port of Lava::onEntityInside.
func (l *Lava) OnEntityInside(e Entity) bool {
	dmgEv := entity.NewEntityDamageByBlockEvent(l.self, e, entity.EntityDamageCauseLava, 4, nil)
	e.Attack(dmgEv)

	combustEv := entity.NewEntityCombustByBlockEvent(l.self, e, 8)
	combustEv.Call()
	if !combustEv.IsCancelled() {
		e.SetOnFire(combustEv.GetDuration())
	}

	e.ResetFallDistance()
	return true
}
