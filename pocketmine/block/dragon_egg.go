package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/particle"
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

// Teleport is a port of DragonEgg::teleport.
func (d *DragonEgg) Teleport() {
	world, err := d.position.GetWorld()
	if err != nil {
		return
	}
	const yMin = format.MinSubChunkIndex * format.SubChunkEdgeLength
	const yMax = (format.MaxSubChunkIndex + 1) * format.SubChunkEdgeLength
	x, y, z := d.position.FloorX(), d.position.FloorY(), d.position.FloorZ()
	for tries := 0; tries < 16; tries++ {
		blk := world.GetBlockAt(x+mtRand(-16, 16), max(yMin, min(yMax-1, y+mtRand(-8, 8))), z+mtRand(-16, 16))
		if blk.GetTypeId() != AIR {
			continue
		}
		ev := blockevent.NewBlockTeleportEvent(d.self, blk.GetPosition().Vector3)
		event.Call(ev)
		if ev.IsCancelled() {
			break
		}
		blockPos := ev.GetTo()
		if p, err := particle.NewDragonEggTeleportParticle(int(d.position.X-blockPos.X), int(d.position.Y-blockPos.Y), int(d.position.Z-blockPos.Z)); err == nil {
			if w, ok := world.(interface {
				AddParticle(pos math.Vector3, p particle.Particle)
			}); ok {
				w.AddParticle(d.position.Vector3, p)
			}
		}
		_ = world.SetBlock(d.position, VanillaAir())
		_ = world.SetBlock(NewPosition(blockPos.X, blockPos.Y, blockPos.Z, world), d.self)
		break
	}
}

func (d *DragonEgg) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	d.Teleport()
	return true
}

// OnAttack is a port of DragonEgg::onAttack: creative players break the egg instead.
func (d *DragonEgg) OnAttack(item Item, face math.Facing, player Player) bool {
	if player == nil {
		return false
	}
	if p, ok := player.(interface{ IsCreativeLiteral() bool }); ok && p.IsCreativeLiteral() {
		return false
	}
	d.Teleport()
	return true
}

// OnNearbyBlockChange is FallableTrait::onNearbyBlockChange.
func (d *DragonEgg) OnNearbyBlockChange() { FallableOnNearbyBlockChange(d.self) }
