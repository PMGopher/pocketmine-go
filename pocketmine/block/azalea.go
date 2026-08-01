package block

import "pocketmine-go/pocketmine/math"

// Azalea is a port of pocketmine\block\Azalea.
type Azalea struct {
	Flowable
}

func NewAzalea(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Azalea {
	a := &Azalea{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	a.Init(a)
	return a
}

func (a *Azalea) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

// OnInteract should grow into a tree (TreeFactory/TreeType.AZALEA) when fertilized — needs a
// Fertilizer item marker and the whole world-gen tree subsystem, none ported yet, so this is a
// no-op for now.
func (a *Azalea) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

func (a *Azalea) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	top := math.OneAABB().SquashedCopy(math.AxisX, 6.0/16).SquashedCopy(math.AxisZ, 6.0/16).TrimmedCopy(math.Up, 8.0/16)
	bottom := math.OneAABB().TrimmedCopy(math.Down, 8.0/16)
	return []math.AxisAlignedBB{top, bottom}
}

// canBeSupportedAt: TODO (from the PHP original) moss block support.
func (a *Azalea) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return support.GetTypeId() == CLAY || geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud)
}

func (a *Azalea) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return a.canBeSupportedAt(blockReplace) && a.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (a *Azalea) OnNearbyBlockChange() {
	if !a.canBeSupportedAt(a.self) {
		if world, err := a.position.GetWorld(); err == nil {
			world.UseBreakOn(a.position.AsVector3())
		}
	} else {
		a.Flowable.OnNearbyBlockChange()
	}
}
