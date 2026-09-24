package object

import (
	stdmath "math"
	"math/rand/v2"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// init registers this package's entity types with EntityFactory (the object half of PHP's
// EntityFactory constructor - FireworkRocket is never saved, so like PHP it isn't registered) and
// installs the world/block hooks that create these entities.
func init() {
	f := entity.GetEntityFactory()

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*AreaEffectCloud, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewAreaEffectCloud(loc, tag), nil
	}, []string{"AreaEffectCloud", "minecraft:area_effect_cloud"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*EndCrystal, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewEndCrystal(loc, tag), nil
	}, []string{"EnderCrystal", "minecraft:ender_crystal"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*ExperienceOrb, error) {
		value := 1
		if valuePC, ok := tag.GetTag(TagValuePC); ok {
			if v, ok := valuePC.(nbt.ShortTag); ok { //PC
				value = int(v)
			}
		} else if valuePE, ok := tag.GetTag(TagValuePE); ok {
			if v, ok := valuePE.(nbt.IntTag); ok { //PE save format
				value = int(v)
			}
		}
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewExperienceOrb(loc, value, tag), nil
	}, []string{"XPOrb", "minecraft:xp_orb"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*FallingBlock, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		blk, err := ParseBlockNBT(w, tag)
		if err != nil {
			return nil, err
		}
		return NewFallingBlock(loc, blk, tag), nil
	}, []string{"FallingSand", "minecraft:falling_block"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*ItemEntity, error) {
		if _, ok, _ := tag.GetCompoundTag(TagItem); !ok {
			return nil, data.NewSavedDataLoadingError(`Expected "` + TagItem + `" NBT tag not found`)
		}
		// Item::nbtDeserialize isn't ported (see ItemEntity's doc comment).
		return nil, data.NewSavedDataLoadingError("item entity data can't be loaded: item NBT deserialization isn't ported")
	}, []string{"Item", "minecraft:item"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Painting, error) {
		motive := GetPaintingMotiveByName(string(tag.GetStringOr(TagMotive, "")))
		if motive == nil {
			return nil, data.NewSavedDataLoadingError("Unknown painting motive")
		}
		blockIn := math.NewVector3(float64(tag.GetIntOr(TagTileX, 0)), float64(tag.GetIntOr(TagTileY, 0)), float64(tag.GetIntOr(TagTileZ, 0)))
		var facing math.Facing
		if directionTag, ok := tag.GetTag(TagDirectionBE); ok && isByteTag(directionTag) {
			facing = paintingFacingFromData(int(directionTag.(nbt.ByteTag)))
		} else if facingTag, ok := tag.GetTag(TagFacingJE); ok && isByteTag(facingTag) {
			facing = paintingFacingFromData(int(facingTag.(nbt.ByteTag)))
		} else {
			return nil, data.NewSavedDataLoadingError("Missing facing info")
		}
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewPainting(loc, blockIn, facing, motive, tag), nil
	}, []string{"Painting", "minecraft:painting"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*PrimedTNT, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewPrimedTNT(loc, tag), nil
	}, []string{"PrimedTnt", "PrimedTNT", "minecraft:tnt"})

	world.DropItemFunc = dropItem
	world.DropExperienceFunc = dropExperience
	block.SpawnFallingBlockFunc = spawnFallingBlock
	block.SpawnPrimedTNTFunc = spawnPrimedTNT
}

func isByteTag(t nbt.Tag) bool {
	_, ok := t.(nbt.ByteTag)
	return ok
}

func paintingFacingFromData(data int) math.Facing {
	if facing, ok := PaintingDataToFacing[data]; ok {
		return facing
	}
	return math.North
}

// dropItem is the body of World::dropItem.
func dropItem(w *world.World, source math.Vector3, it item.Item, motion *math.Vector3, delay int) world.Entity {
	itemEntity := NewItemEntity(entity.LocationFromObject(source, w, utils.GetRandomFloat()*360, 0), it, nil)

	itemEntity.SetPickupDelay(delay)
	if motion != nil {
		itemEntity.SetMotion(*motion)
	} else {
		itemEntity.SetMotion(math.NewVector3(utils.GetRandomFloat()*0.2-0.1, 0.2, utils.GetRandomFloat()*0.2-0.1))
	}
	itemEntity.SpawnToAll()

	return itemEntity
}

// dropExperience is the body of World::dropExperience.
func dropExperience(w *world.World, pos math.Vector3, amount int) []world.Entity {
	var orbs []world.Entity

	for _, split := range SplitIntoOrbSizes(amount) {
		orb := NewExperienceOrb(entity.LocationFromObject(pos, w, utils.GetRandomFloat()*360, 0), split, nil)

		orb.SetMotion(math.NewVector3((utils.GetRandomFloat()*0.2-0.1)*2, utils.GetRandomFloat()*0.4, (utils.GetRandomFloat()*0.2-0.1)*2))
		orb.SpawnToAll()

		orbs = append(orbs, orb)
	}

	return orbs
}

// spawnFallingBlock is the entity half of FallableTrait::onNearbyBlockChange.
func spawnFallingBlock(blk block.Behavior) {
	pos := blk.GetPosition()
	blockWorld, err := pos.GetWorld()
	if err != nil {
		return
	}
	w, ok := blockWorld.(*world.World)
	if !ok {
		return
	}
	fall := NewFallingBlock(entity.LocationFromObject(pos.AsVector3().Add(0.5, 0, 0.5), w, 0, 0), blk, nil)
	fall.SpawnToAll()
}

// spawnPrimedTNT is the entity half of TNT::ignite.
func spawnPrimedTNT(pos block.Position, worksUnderwater bool, fuse int) {
	blockWorld, err := pos.GetWorld()
	if err != nil {
		return
	}
	w, ok := blockWorld.(*world.World)
	if !ok {
		return
	}

	mot := (rand.Float64()*2 - 1) * stdmath.Pi * 2 // Random::nextSignedFloat()

	tnt := NewPrimedTNT(entity.LocationFromObject(pos.AsVector3().Add(0.5, 0, 0.5), w, 0, 0), nil)
	tnt.SetFuse(fuse)
	tnt.SetWorksUnderwater(worksUnderwater)
	tnt.SetMotion(math.NewVector3(-stdmath.Sin(mot)*0.02, 0.2, -stdmath.Cos(mot)*0.02))

	tnt.SpawnToAll()
	tnt.BroadcastSound(sound.IgniteSound{})
}
