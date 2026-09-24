package object

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// EndCrystal NBT keys, a port of EndCrystal's TAG_* constants.
const (
	tagShowBase     = "ShowBottom"   //TAG_Byte
	tagBlockTargetX = "BlockTargetX" //TAG_Int
	tagBlockTargetY = "BlockTargetY" //TAG_Int
	tagBlockTargetZ = "BlockTargetZ" //TAG_Int
)

// EndCrystal is a port of pocketmine\entity\object\EndCrystal (also Explosive).
type EndCrystal struct {
	entity.Entity

	showBase   bool
	beamTarget *math.Vector3
	primed     bool
}

// NewEndCrystal is a port of EndCrystal::__construct.
func NewEndCrystal(location entity.Location, tag *nbt.CompoundTag) *EndCrystal {
	c := &EndCrystal{}
	c.Construct(c, location, tag)
	return c
}

func (c *EndCrystal) GetNetworkTypeID() string { return entity.EntityIDEnderCrystal }

func (c *EndCrystal) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(2.0, 2.0)
}

func (c *EndCrystal) GetInitialDragMultiplier() float64 { return 1.0 }

func (c *EndCrystal) GetInitialGravity() float64 { return 0.0 }

func (c *EndCrystal) IsFireProof() bool { return true }

func (c *EndCrystal) GetPickedItem() item.Item { return item.VanillaEndCrystal() }

func (c *EndCrystal) ShowBase() bool { return c.showBase }

func (c *EndCrystal) SetShowBase(showBase bool) {
	c.showBase = showBase
	c.MarkNetworkPropertiesDirty()
}

// GetBeamTarget returns the beam target, or nil.
func (c *EndCrystal) GetBeamTarget() *math.Vector3 { return c.beamTarget }

func (c *EndCrystal) SetBeamTarget(beamTarget *math.Vector3) {
	c.beamTarget = beamTarget
	c.MarkNetworkPropertiesDirty()
}

// Attack is a port of EndCrystal::attack: any damage except the void primes the crystal.
func (c *EndCrystal) Attack(source entityevent.DamageSource) {
	c.Entity.Attack(source)
	if source.GetCause() != entityevent.CauseVoid && !source.IsCancelled() {
		c.primed = true
	}
}

// InitEntity is a port of EndCrystal::initEntity.
func (c *EndCrystal) InitEntity(tag *nbt.CompoundTag) {
	c.Entity.InitEntity(tag)

	c.SetMaxHealth(1)
	c.SetHealth(1)

	c.SetShowBase(tag.GetByteOr(tagShowBase, 0) == 1)

	x, xOK := tag.GetTag(tagBlockTargetX)
	y, yOK := tag.GetTag(tagBlockTargetY)
	z, zOK := tag.GetTag(tagBlockTargetZ)
	if xOK && yOK && zOK {
		bx, xInt := x.(nbt.IntTag)
		by, yInt := y.(nbt.IntTag)
		bz, zInt := z.(nbt.IntTag)
		if xInt && yInt && zInt {
			target := math.NewVector3(float64(bx), float64(by), float64(bz))
			c.SetBeamTarget(&target)
		}
	}
}

// SaveNBT is a port of EndCrystal::saveNBT.
func (c *EndCrystal) SaveNBT() *nbt.CompoundTag {
	tag := c.Entity.SaveNBT()

	showBase := nbt.ByteTag(0)
	if c.showBase {
		showBase = 1
	}
	tag.SetByte(tagShowBase, showBase)
	if c.beamTarget != nil {
		tag.SetInt(tagBlockTargetX, nbt.IntTag(c.beamTarget.FloorX()))
		tag.SetInt(tagBlockTargetY, nbt.IntTag(c.beamTarget.FloorY()))
		tag.SetInt(tagBlockTargetZ, nbt.IntTag(c.beamTarget.FloorZ()))
	}
	return tag
}

// OnDeathUpdate is a port of EndCrystal::onDeathUpdate: a primed crystal explodes on death.
func (c *EndCrystal) OnDeathUpdate(tickDiff int) bool {
	if c.primed {
		c.Explode()
	}
	return true
}

// Explode is a port of EndCrystal::explode.
func (c *EndCrystal) Explode() {
	ev := entityevent.NewEntityPreExplodeEvent(c, 6, 0.0)
	ev.Call()
	if !ev.IsCancelled() {
		pos := c.GetPosition()
		explosion, err := world.NewEntityExplosion(block.NewPosition(pos.X, pos.Y, pos.Z, c.GetWorld()), ev.GetRadius(), c, ev.GetFireChance())
		if err != nil {
			panic(fmt.Sprintf("EndCrystal: %v", err))
		}
		if ev.IsBlockBreaking() {
			explosion.ExplodeA()
		}
		explosion.ExplodeB()
	}
}

// SyncNetworkData is a port of EndCrystal::syncNetworkData.
func (c *EndCrystal) SyncNetworkData(properties *entity.MetadataCollection) {
	c.Entity.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagShowBase, c.showBase)
	target := math.Vector3Zero()
	if c.beamTarget != nil {
		target = *c.beamTarget
	}
	properties.SetBlockPos(entity.MetadataBlockTarget, protocol.BlockPos{int32(target.FloorX()), int32(target.FloorY()), int32(target.FloorZ())})
}
