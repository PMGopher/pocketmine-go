package object

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// FallingBlock NBT keys, a port of FallingBlock's TAG_* constants.
const (
	tagFallingBlock = "FallingBlock" //TAG_Compound
	tagTileID       = "TileID"       //TAG_Int
	tagTile         = "Tile"         //TAG_Byte
	tagData         = "Data"         //TAG_Byte
)

// FallingBlock is a port of pocketmine\entity\object\FallingBlock.
type FallingBlock struct {
	entity.Entity

	block block.Behavior
}

// NewFallingBlock is a port of FallingBlock::__construct.
func NewFallingBlock(location entity.Location, blk block.Behavior, tag *nbt.CompoundTag) *FallingBlock {
	f := &FallingBlock{block: blk}
	f.Construct(f, location, tag)
	return f
}

func (f *FallingBlock) GetNetworkTypeID() string { return entity.EntityIDFallingBlock }

func (f *FallingBlock) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.98, 0.98)
}

func (f *FallingBlock) GetInitialDragMultiplier() float64 { return 0.02 }

func (f *FallingBlock) GetInitialGravity() float64 { return 0.04 }

// ParseBlockNBT is a port of FallingBlock::parseBlockNBT. The legacy TileID/Tile + Data format
// needs the block data upgrader, which isn't ported, so it's reported as unloadable data.
func ParseBlockNBT(w *world.World, tag *nbt.CompoundTag) (block.Behavior, error) {
	//TODO: 1.8+ save format
	if fallingBlockTag, ok, _ := tag.GetCompoundTag(tagFallingBlock); ok {
		blk, err := w.DeserializeBlockState(fallingBlockTag)
		if err != nil {
			return nil, &data.SavedDataLoadingError{Message: "Invalid falling block blockstate: " + err.Error(), Cause: err}
		}
		return blk, nil
	}
	_, hasTileID := tag.GetTag(tagTileID)
	_, hasTile := tag.GetTag(tagTile)
	if !hasTileID && !hasTile {
		return nil, data.NewSavedDataLoadingError("Missing legacy falling block info")
	}
	return nil, data.NewSavedDataLoadingError("Invalid legacy falling block data: the legacy block ID upgrader isn't ported")
}

func (f *FallingBlock) CanCollideWith(other world.Entity) bool { return false }

func (f *FallingBlock) CanBeMovedByCurrents() bool { return false }

// Attack is a port of FallingBlock::attack: only the void can damage a falling block.
func (f *FallingBlock) Attack(source entityevent.DamageSource) {
	if source.GetCause() == entityevent.CauseVoid {
		f.Entity.Attack(source)
	}
}

// positionable is the promoted-from-*block.Block surface FallingBlock needs (Block::position()).
type positionable interface {
	SetPosition(w block.World, x, y, z int)
}

// asItemable is the promoted-from-*block.Block surface FallingBlock needs (Block::asItem()).
type asItemable interface {
	AsItem() (block.Item, error)
}

func blockAsItem(blk block.Behavior) item.Item {
	if a, ok := blk.(asItemable); ok {
		if it, err := a.AsItem(); err == nil {
			if converted, ok := it.(item.Item); ok {
				return converted
			}
		}
	}
	return nil
}

// EntityBaseTick is a port of FallingBlock::entityBaseTick: place the block on landing, or drop it
// as an item if it can't be placed.
func (f *FallingBlock) EntityBaseTick(tickDiff int) bool {
	if f.IsClosed() {
		return false
	}

	hasUpdate := f.Entity.EntityBaseTick(tickDiff)

	if !f.IsFlaggedForDespawn() {
		w := f.GetWorld()
		pos := f.GetPosition().Add(-f.Size.GetWidth()/2, f.Size.GetHeight(), -f.Size.GetWidth()/2).Floor()

		if p, ok := f.block.(positionable); ok {
			p.SetPosition(w, pos.FloorX(), pos.FloorY(), pos.FloorZ())
		}

		var blockTarget block.Behavior
		if fallable, ok := f.block.(block.Fallable); ok {
			if replacement, ok := fallable.TickFalling(); ok {
				blockTarget = replacement
			}
		}

		if f.OnGround || blockTarget != nil {
			f.FlagForDespawn()

			blockResult := f.block
			if blockTarget != nil {
				blockResult = blockTarget
			}
			existing := w.GetBlockAtIfLoaded(pos.FloorX(), pos.FloorY(), pos.FloorZ())
			if !existing.CanBeReplaced() || !w.IsInWorld(pos.FloorX(), pos.FloorY(), pos.FloorZ()) || (f.OnGround && stdmath.Abs(f.GetPosition().Y-float64(f.GetPosition().FloorY())) > 0.001) {
				if it := blockAsItem(f.block); it != nil {
					w.DropItem(f.GetPosition(), it, nil, 10)
				}
				w.AddSound(pos.Add(0.5, 0.5, 0.5), sound.BlockBreakSound{BlockStateID: blockResult.GetStateId()})
			} else {
				ev := entityevent.NewEntityBlockChangeEvent(f, existing, blockResult)
				ev.Call()
				if !ev.IsCancelled() {
					if b, ok := ev.GetTo().(block.Behavior); ok {
						_ = w.SetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ(), b)
						if fallable, ok := b.(block.Fallable); ok && f.OnGround {
							if landSound, ok := fallable.GetLandSound(); ok {
								w.AddSound(pos.Add(0.5, 0.5, 0.5), landSound)
							}
						}
					}
				}
			}
			hasUpdate = true
		}
	}

	return hasUpdate
}

// livingTarget is the surface falling-block damage needs to recognise a Living (PHP's
// `$entity instanceof Living`).
type livingTarget interface {
	world.Entity
	IsLiving() bool
}

// OnHitGround is a port of FallingBlock::onHitGround: falling anvils and the like hurt the living
// entities they land on.
func (f *FallingBlock) OnHitGround() *float64 {
	if fallable, ok := f.block.(block.Fallable); ok {
		damagePerBlock := fallable.GetFallDamagePerBlock()
		if fallenBlocks := stdmath.Round(f.FallDistance) - 1; damagePerBlock > 0 && fallenBlocks > 0 {
			damage := min(fallenBlocks*damagePerBlock, fallable.GetMaxFallDamage())
			for _, e := range f.GetWorld().GetCollidingEntities(f.GetBoundingBox(), nil) {
				if living, ok := e.(livingTarget); ok {
					ev := entityevent.NewEntityDamageByEntityEvent(f, living, entityevent.CauseFallingBlock, damage, nil)
					living.Attack(ev)
				}
			}
		}
		if !fallable.OnHitGround(f) {
			f.FlagForDespawn()
		}
	}
	return nil
}

func (f *FallingBlock) GetBlock() block.Behavior { return f.block }

// SaveNBT is a port of FallingBlock::saveNBT.
func (f *FallingBlock) SaveNBT() *nbt.CompoundTag {
	tag := f.Entity.SaveNBT()
	if blockTag, err := f.GetWorld().SerializeBlockState(f.block); err == nil {
		tag.SetTag(tagFallingBlock, blockTag)
	}

	return tag
}

func (f *FallingBlock) GetPickedItem() item.Item { return blockAsItem(f.block) }

// SyncNetworkData is a port of FallingBlock::syncNetworkData.
func (f *FallingBlock) SyncNetworkData(properties *entity.MetadataCollection) {
	f.Entity.SyncNetworkData(properties)

	properties.SetInt(entity.MetadataVariant, f.GetWorld().Translator().InternalIDToNetworkID(f.block))
}

func (f *FallingBlock) GetOffsetPosition(v math.Vector3) math.Vector3 {
	return v.Add(0, 0.49, 0) //TODO: check if height affects this
}
