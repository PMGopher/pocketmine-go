package object

import (
	stdmath "math"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
)

// Painting NBT keys, a port of Painting's TAG_* constants.
const (
	TagTileX       = "TileX"     //TAG_Int
	TagTileY       = "TileY"     //TAG_Int
	TagTileZ       = "TileZ"     //TAG_Int
	TagFacingJE    = "Facing"    //TAG_Byte
	TagDirectionBE = "Direction" //TAG_Byte
	TagMotive      = "Motive"    //TAG_String
)

// PaintingDataToFacing mirrors Painting::DATA_TO_FACING.
var PaintingDataToFacing = map[int]math.Facing{
	0: math.South,
	1: math.West,
	2: math.North,
	3: math.East,
}

// paintingFacingToData mirrors Painting::FACING_TO_DATA.
var paintingFacingToData = map[math.Facing]int{
	math.South: 0,
	math.West:  1,
	math.North: 2,
	math.East:  3,
}

// Painting is a port of pocketmine\entity\object\Painting.
type Painting struct {
	entity.Entity

	blockIn math.Vector3
	facing  math.Facing
	motive  *PaintingMotive
}

// NewPainting is a port of Painting::__construct.
func NewPainting(location entity.Location, blockIn math.Vector3, facing math.Facing, motive *PaintingMotive, tag *nbt.CompoundTag) *Painting {
	p := &Painting{motive: motive, blockIn: blockIn, facing: facing}
	p.Construct(p, location, tag)
	return p
}

func (p *Painting) GetNetworkTypeID() string { return entity.EntityIDPainting }

func (p *Painting) GetInitialSizeInfo() entity.EntitySizeInfo {
	//these aren't accurate, but it doesn't matter since they aren't used (vanilla PC does something similar)
	return entity.NewEntitySizeInfo(0.5, 0.5)
}

func (p *Painting) GetInitialDragMultiplier() float64 { return 1.0 }

func (p *Painting) GetInitialGravity() float64 { return 0.0 }

// InitEntity is a port of Painting::initEntity.
func (p *Painting) InitEntity(tag *nbt.CompoundTag) {
	p.SetMaxHealth(1)
	p.SetHealth(1)
	p.Entity.InitEntity(tag)
}

// SaveNBT is a port of Painting::saveNBT.
func (p *Painting) SaveNBT() *nbt.CompoundTag {
	tag := p.Entity.SaveNBT()
	tag.SetInt(TagTileX, nbt.IntTag(int(p.blockIn.X)))
	tag.SetInt(TagTileY, nbt.IntTag(int(p.blockIn.Y)))
	tag.SetInt(TagTileZ, nbt.IntTag(int(p.blockIn.Z)))

	tag.SetByte(TagFacingJE, nbt.ByteTag(paintingFacingToData[p.facing]))
	tag.SetByte(TagDirectionBE, nbt.ByteTag(paintingFacingToData[p.facing])) //Save both for full compatibility

	tag.SetString(TagMotive, nbt.StringTag(p.motive.GetName()))

	return tag
}

// OnDeath is a port of Painting::onDeath.
func (p *Painting) OnDeath() {
	p.Entity.OnDeath()

	drops := true

	if byEntity, ok := entity.AsDamageByEntity(p.GetLastDamageCause()); ok {
		if killer, isPlayer := entity.AsPlayer(byEntity.GetDamager()); isPlayer && !killer.HasFiniteResources() {
			drops = false
		}
	}

	if drops {
		//non-living entities don't have a way to create drops generically yet
		p.GetWorld().DropItem(p.GetPosition(), item.VanillaPainting(), nil, 10)
	}
	p.GetWorld().AddParticle(p.GetPosition().Add(0.5, 0.5, 0.5), particle.BlockBreakParticle{BlockStateID: block.VanillaOakPlanks().GetStateId()})
}

// RecalculateBoundingBox is a port of Painting::recalculateBoundingBox.
func (p *Painting) RecalculateBoundingBox() {
	side := p.blockIn.GetSide(p.facing, 1)
	p.BoundingBox = paintingBB(p.facing, p.motive).OffsetCopy(side.X, side.Y, side.Z)
}

// OnNearbyBlockChange is a port of Painting::onNearbyBlockChange: a painting that no longer fits
// breaks.
func (p *Painting) OnNearbyBlockChange() {
	p.Entity.OnNearbyBlockChange()

	if !CanPaintingFit(p.GetWorld(), p.blockIn.GetSide(p.facing, 1), p.facing, false, p.motive) {
		p.Kill()
	}
}

// OnRandomUpdate is a no-op for paintings.
func (p *Painting) OnRandomUpdate() {}

func (p *Painting) HasMovementUpdate() bool { return false }

func (p *Painting) UpdateMovement(teleport bool) {}

func (p *Painting) CanBeCollidedWith() bool { return false }

// SendSpawnPacket is a port of Painting::sendSpawnPacket.
func (p *Painting) SendSpawnPacket(player world.EntityViewer) {
	bb := p.BoundingBox
	player.SendPacket(&packet.AddPainting{
		EntityUniqueID:  int64(p.GetID()), //TODO: entity unique ID
		EntityRuntimeID: uint64(p.GetID()),
		Position: vec32(math.NewVector3(
			(bb.MinX+bb.MaxX)/2,
			(bb.MinY+bb.MaxY)/2,
			(bb.MinZ+bb.MaxZ)/2,
		)),
		Direction: int32(paintingFacingToData[p.facing]),
		Title:     p.motive.GetName(),
	})
}

func (p *Painting) GetPickedItem() item.Item { return item.VanillaPainting() }

func (p *Painting) GetMotive() *PaintingMotive { return p.motive }

func (p *Painting) GetFacing() math.Facing { return p.facing }

// paintingBB is a port of Painting::getPaintingBB: the bounding box of a painting with the given
// motive, on a block facing the given direction.
func paintingBB(facing math.Facing, motive *PaintingMotive) math.AxisAlignedBB {
	width := motive.GetWidth()
	height := motive.GetHeight()

	horizontalStart := int(stdmath.Ceil(float64(width)/2) - 1)
	verticalStart := int(stdmath.Ceil(float64(height)/2) - 1)

	bb := math.OneAABB()
	bb.Trim(facing, 15.0/16.0).
		Extend(math.RotateY(facing, true), float64(horizontalStart)).
		Extend(math.RotateY(facing, false), float64(-horizontalStart+width-1)).
		Extend(math.Down, float64(verticalStart)).
		Extend(math.Up, float64(-verticalStart+height-1))
	return bb
}

// sideGetter is the promoted-from-*block.Block surface CanPaintingFit needs.
type sideGetter interface {
	GetSide(side math.Facing, step int) block.Behavior
}

// CanPaintingFit is a port of Painting::canFit: whether a painting with the given motive can be
// placed on the given block face.
func CanPaintingFit(w *world.World, blockIn math.Vector3, facing math.Facing, checkOverlap bool, motive *PaintingMotive) bool {
	width := motive.GetWidth()
	height := motive.GetHeight()

	horizontalStart := int(stdmath.Ceil(float64(width)/2) - 1)
	verticalStart := int(stdmath.Ceil(float64(height)/2) - 1)

	rotatedFace := math.RotateY(facing, false)

	oppositeSide := math.Opposite(facing)

	startPos := blockIn.GetSide(math.Opposite(rotatedFace), horizontalStart).GetSide(math.Down, verticalStart)

	for wi := 0; wi < width; wi++ {
		for h := 0; h < height; h++ {
			pos := startPos.GetSide(rotatedFace, wi).GetSide(math.Up, h)
			blk := w.GetBlockAtIfLoaded(pos.FloorX(), pos.FloorY(), pos.FloorZ())
			if blk.IsSolid() {
				return false
			}
			if sides, ok := blk.(sideGetter); ok && !sides.GetSide(oppositeSide, 1).IsSolid() {
				return false
			}
		}
	}

	if checkOverlap {
		bb := paintingBB(facing, motive).OffsetCopy(blockIn.X, blockIn.Y, blockIn.Z)

		for _, e := range w.GetNearbyEntitiesExcept(bb, nil) {
			if _, ok := e.(*Painting); ok {
				return false
			}
		}
	}

	return true
}
