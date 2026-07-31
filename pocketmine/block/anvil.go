package block

import (
	stdmath "math"
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	AnvilUndamaged       = 0
	AnvilSlightlyDamaged = 1
	AnvilVeryDamaged     = 2
)

// Anvil is a port of pocketmine\block\Anvil.
type Anvil struct {
	Transparent
	FallableComponent
	HorizontalFacingComponent

	Damage int
}

func NewAnvil(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Anvil {
	a := &Anvil{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	a.Init(a)
	return a
}

func (a *Anvil) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *Anvil) DescribeBlockItemState(w runtime.DataDescriber) {
	w.BoundedIntAuto(AnvilUndamaged, AnvilVeryDamaged, &a.Damage)
}

func (a *Anvil) DescribeBlockOnlyState(w runtime.DataDescriber) { a.DescribeHorizontalFacing(w) }

func (a *Anvil) GetDamage() int { return a.Damage }

// SetDamage panics if damage is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (a *Anvil) SetDamage(damage int) {
	if damage < AnvilUndamaged || damage > AnvilVeryDamaged {
		panic("Damage must be in range 0 ... 2")
	}
	a.Damage = damage
}

func (a *Anvil) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	axis := math.FacingAxis(math.RotateY(a.Facing, false))
	return []math.AxisAlignedBB{math.OneAABB().SquashedCopy(axis, 1.0/8)}
}

func (a *Anvil) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// OnInteract should open an AnvilInventory for the interacting player — needs the unported
// block/inventory package, so this is a no-op for now; it still returns true, matching the PHP
// original's unconditional `return true;`.
func (a *Anvil) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return true
}

func (a *Anvil) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		a.Facing = math.RotateY(player.GetHorizontalFacing(), false)
	}
	return a.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (a *Anvil) OnHitGround(blockEntity FallingBlockEntity) bool {
	threshold := 0.05 + (stdmath.Round(blockEntity.GetFallDistance())-1)*0.05
	if rand.Float64() < threshold {
		if a.Damage != AnvilVeryDamaged {
			a.Damage++
		} else {
			return false
		}
	}
	return true
}

func (a *Anvil) GetFallDamagePerBlock() float64 { return 2.0 }

func (a *Anvil) GetMaxFallDamage() float64 { return 40.0 }

func (a *Anvil) GetLandSound() (sound.Sound, bool) { return sound.AnvilFallSound{}, true }
