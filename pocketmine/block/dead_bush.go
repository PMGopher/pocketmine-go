package block

import "pocketmine-go/pocketmine/math"

// DeadBush is a port of pocketmine\block\DeadBush.
type DeadBush struct {
	Flowable
}

func NewDeadBush(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DeadBush {
	d := &DeadBush{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	d.Init(d)
	return d
}

func (d *DeadBush) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DeadBush) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	if geo.HasTypeTag(BlockTypeTagsSand) || geo.HasTypeTag(BlockTypeTagsMud) {
		return true
	}
	switch support.GetTypeId() {
	//can't use DIRT tag here because it includes farmland
	case PODZOL, MYCELIUM, DIRT, GRASS, HARDENED_CLAY, STAINED_CLAY:
		return true
	//TODO: moss block
	default:
		return false
	}
}

func (d *DeadBush) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return d.canBeSupportedAt(blockReplace) && d.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (d *DeadBush) OnNearbyBlockChange() {
	if !d.canBeSupportedAt(d.self) {
		if world, err := d.position.GetWorld(); err == nil {
			world.UseBreakOn(d.position.AsVector3())
		}
	} else {
		d.Flowable.OnNearbyBlockChange()
	}
}

// GetDropsForIncompatibleTool should return [VanillaItems.Stick().SetCount(rand.Intn(3))] — needs
// the unported item package for real Item construction (see Block.GetDropsForCompatibleTool's
// doc comment), so this returns nil for now.
func (d *DeadBush) GetDropsForIncompatibleTool(item Item) []Item { return nil }

func (d *DeadBush) IsAffectedBySilkTouch() bool { return true }

func (d *DeadBush) GetFlameEncouragement() int { return 60 }

func (d *DeadBush) GetFlammability() int { return 100 }
