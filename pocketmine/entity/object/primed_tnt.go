package object

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

const tagFuse = "Fuse" //TAG_Short

// PrimedTNT is a port of pocketmine\entity\object\PrimedTNT (also Explosive).
type PrimedTNT struct {
	entity.Entity

	fuse            int
	worksUnderwater bool
}

// NewPrimedTNT is a port of PrimedTNT::__construct.
func NewPrimedTNT(location entity.Location, tag *nbt.CompoundTag) *PrimedTNT {
	t := &PrimedTNT{}
	t.Construct(t, location, tag)
	return t
}

func (t *PrimedTNT) GetNetworkTypeID() string { return entity.EntityIDTNT }

func (t *PrimedTNT) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.98, 0.98)
}

func (t *PrimedTNT) GetInitialDragMultiplier() float64 { return 0.02 }

func (t *PrimedTNT) GetInitialGravity() float64 { return 0.04 }

func (t *PrimedTNT) GetFuse() int { return t.fuse }

// SetFuse panics outside 0-32767 (PHP's InvalidArgumentException).
func (t *PrimedTNT) SetFuse(fuse int) {
	if fuse < 0 || fuse > 32767 {
		panic("Fuse must be in the range 0-32767")
	}
	t.fuse = fuse
	t.MarkNetworkPropertiesDirty()
}

func (t *PrimedTNT) WorksUnderwater() bool { return t.worksUnderwater }

func (t *PrimedTNT) SetWorksUnderwater(worksUnderwater bool) {
	t.worksUnderwater = worksUnderwater
	t.MarkNetworkPropertiesDirty()
}

// Attack is a port of PrimedTNT::attack: only the void can damage primed TNT.
func (t *PrimedTNT) Attack(source entityevent.DamageSource) {
	if source.GetCause() == entityevent.CauseVoid {
		t.Entity.Attack(source)
	}
}

// InitEntity is a port of PrimedTNT::initEntity.
func (t *PrimedTNT) InitEntity(tag *nbt.CompoundTag) {
	t.Entity.InitEntity(tag)

	t.fuse = int(tag.GetShortOr(tagFuse, 80))
}

func (t *PrimedTNT) CanCollideWith(other world.Entity) bool { return false }

// SaveNBT is a port of PrimedTNT::saveNBT.
func (t *PrimedTNT) SaveNBT() *nbt.CompoundTag {
	tag := t.Entity.SaveNBT()
	tag.SetShort(tagFuse, nbt.ShortTag(t.fuse))

	return tag
}

// EntityBaseTick is a port of PrimedTNT::entityBaseTick.
func (t *PrimedTNT) EntityBaseTick(tickDiff int) bool {
	if t.IsClosed() {
		return false
	}

	hasUpdate := t.Entity.EntityBaseTick(tickDiff)

	if !t.IsFlaggedForDespawn() {
		t.fuse -= tickDiff
		t.MarkNetworkPropertiesDirty()

		if t.fuse <= 0 {
			t.FlagForDespawn()
			t.Explode()
		}
	}

	return hasUpdate || t.fuse >= 0
}

// Explode is a port of PrimedTNT::explode.
func (t *PrimedTNT) Explode() {
	ev := entityevent.NewEntityPreExplodeEvent(t, 4, 0.0)
	ev.Call()
	if !ev.IsCancelled() {
		//TODO: deal with underwater TNT (underwater TNT treats water as if it has a blast resistance of 0)
		center := t.GetPosition().Add(0, t.Size.GetHeight()/2, 0)
		explosion, err := world.NewEntityExplosion(block.NewPosition(center.X, center.Y, center.Z, t.GetWorld()), ev.GetRadius(), t, ev.GetFireChance())
		if err != nil {
			panic(fmt.Sprintf("PrimedTNT: %v", err))
		}
		if ev.IsBlockBreaking() {
			explosion.ExplodeA()
		}
		explosion.ExplodeB()
	}
}

// GetPickedItem is a port of PrimedTNT::getPickedItem.
func (t *PrimedTNT) GetPickedItem() item.Item {
	tnt := block.VanillaTNT()
	if b, ok := tnt.(*block.TNTBlock); ok {
		b.SetWorksUnderwater(t.worksUnderwater)
	}
	return blockAsItem(tnt)
}

// SyncNetworkData is a port of PrimedTNT::syncNetworkData.
func (t *PrimedTNT) SyncNetworkData(properties *entity.MetadataCollection) {
	t.Entity.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagIgnited, true)
	variant := int32(0)
	if t.worksUnderwater {
		variant = 1
	}
	properties.SetInt(entity.MetadataVariant, variant)
	properties.SetInt(entity.MetadataFuseLength, int32(t.fuse))
}

func (t *PrimedTNT) GetOffsetPosition(v math.Vector3) math.Vector3 { return v.Add(0, 0.49, 0) }
