package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// DragonEgg is a port of pocketmine\block\DragonEgg.
type DragonEgg struct {
	Transparent
	FallableComponent
}

func NewDragonEgg(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DragonEgg {
	d := &DragonEgg{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}
	d.Init(d)
	return d
}

func (d *DragonEgg) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DragonEgg) GetLightLevel() int { return 1 }

func (d *DragonEgg) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// Teleport is a port of DragonEgg::teleport. It needs the unported BlockTeleportEvent, particle
// system (DragonEggTeleportParticle) and the block registry (VanillaBlocks) to pick a random
// nearby Air block and move itself there, so this is a no-op for now - see
// Block.GetDropsForCompatibleTool's doc comment for the same category of gap.
func (d *DragonEgg) Teleport() {}

func (d *DragonEgg) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	d.Teleport()
	return true
}

// OnAttack should also teleport unless the attacking player is in creative mode - GameMode isn't
// on the minimal Player interface yet, so this always teleports for now (same gap as Teleport
// itself being a no-op).
func (d *DragonEgg) OnAttack(item Item, face math.Facing, player Player) bool {
	if player != nil {
		d.Teleport()
		return true
	}
	return false
}
