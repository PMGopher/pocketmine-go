package block

import (
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
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

// getAdjacentBlocksExceptDown is a port of Lava::getAdjacentBlocksExceptDown.
func (l *Lava) getAdjacentBlocksExceptDown() []Behavior {
	geo := l.self.(blockGeometry)
	var result []Behavior
	for _, face := range math.AllFacing {
		if face == math.Down {
			continue
		}
		result = append(result, geo.GetSide(face, 1))
	}
	return result
}

// checkForHarden is a port of Lava::checkForHarden. BLUE_ICE isn't ported as its own singleton
// yet, so the soul-soil/Basalt branch checks the raw type ID (BLUE_ICE below), matching the PHP
// original's getTypeId() === BlockTypeIds::BLUE_ICE comparison exactly.
func (l *Lava) checkForHarden() bool {
	if l.Falling {
		return false
	}
	for _, colliding := range l.getAdjacentBlocksExceptDown() {
		if _, ok := colliding.(*Water); ok {
			if l.Decay == 0 {
				l.liquidCollide(colliding, VanillaObsidian())
				return true
			} else if l.Decay <= 4 {
				l.liquidCollide(colliding, VanillaCobblestone())
				return true
			}
		}
	}

	if l.self.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == SOUL_SOIL {
		for _, colliding := range l.getAdjacentBlocksExceptDown() {
			if colliding.GetTypeId() == BLUE_ICE {
				l.liquidCollide(colliding, VanillaBasalt())
				return true
			}
		}
	}

	return false
}

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
