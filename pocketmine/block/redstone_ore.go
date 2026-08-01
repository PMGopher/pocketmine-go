package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// RedstoneOre is a port of pocketmine\block\RedstoneOre.
type RedstoneOre struct {
	Opaque
	LightableComponent
}

func NewRedstoneOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneOre {
	r := &RedstoneOre{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	r.Init(r)
	return r
}

func (r *RedstoneOre) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneOre) DescribeBlockOnlyState(w runtime.DataDescriber) { r.DescribeLit(w) }

func (r *RedstoneOre) GetLightLevel() int {
	if r.Lit {
		return 9
	}
	return 0
}

// OnInteract deliberately always returns false (matching the PHP original's comment: lighting up
// shouldn't prevent block placement), even though it may still light the ore up as a side effect.
func (r *RedstoneOre) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if !r.Lit {
		r.Lit = true
		if world, err := r.position.GetWorld(); err == nil {
			if err := world.SetBlock(r.position, r.self); err != nil {
				panic(err)
			}
		}
	}
	return false
}

func (r *RedstoneOre) OnNearbyBlockChange() {
	if !r.Lit {
		r.Lit = true
		if world, err := r.position.GetWorld(); err == nil {
			if err := world.SetBlock(r.position, r.self); err != nil {
				panic(err)
			}
		}
	}
}

func (r *RedstoneOre) TicksRandomly() bool { return r.Lit }

func (r *RedstoneOre) OnRandomTick() {
	if r.Lit {
		r.Lit = false
		if world, err := r.position.GetWorld(); err == nil {
			if err := world.SetBlock(r.position, r.self); err != nil {
				panic(err)
			}
		}
	}
}

// GetDropsForCompatibleTool should return redstone dust scaled via FortuneDropHelper — needs real
// Item construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (r *RedstoneOre) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (r *RedstoneOre) IsAffectedBySilkTouch() bool { return true }

func (r *RedstoneOre) GetXpDropAmount() int { return rand.Intn(5) + 1 }
