package item

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
)

// This file holds the items whose use spawns an entity (spawn eggs, paintings, end crystals,
// fireworks). This package can't create entities (entity imports item), so the entity half of
// each is a hook installed by the entity/object package, like ThrowProjectileFunc.

// SpawnedEntity is what SpawnEgg::onInteractBlock does with the entity it creates.
type SpawnedEntity interface {
	SetNameTag(name string)
	SpawnToAll()
}

var (
	// CreateSpawnEggEntityFunc is SpawnEgg::createEntity: the egg's mob at pos (nil if the mob is
	// unknown).
	CreateSpawnEggEntityFunc func(egg *SpawnEgg, world block.World, pos math.Vector3, yaw, pitch float64) SpawnedEntity

	// PlacePaintingFunc is PaintingItem::onInteractBlock after the vertical-face check: it picks
	// one of the largest motives that fit, spawns the painting and plays the place sound. It
	// returns false if no motive fits.
	PlacePaintingFunc func(player Player, blockReplace, blockClicked block.Behavior, face math.Facing) bool

	// PlaceEndCrystalFunc is EndCrystal::onInteractBlock's placement check and spawn on the
	// clicked obsidian or bedrock block. It returns false if the crystal can't go there.
	PlaceEndCrystalFunc func(blockClicked block.Behavior) bool

	// LaunchFireworkFunc is FireworkRocket::onInteractBlock's entity creation: the firework
	// entity at position, owned by player, spawned.
	LaunchFireworkFunc func(rocket *FireworkRocket, player Player, position math.Vector3, randomDuration int)
)

// OnInteractBlock is a port of SpawnEgg::onInteractBlock.
func (s *SpawnEgg) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	world := entityWorld(player)
	if world == nil || CreateSpawnEggEntityFunc == nil {
		return ItemUseResultNone
	}
	entity := CreateSpawnEggEntityFunc(s, world, blockReplace.GetPosition().Add(0.5, 0, 0.5), utils.GetRandomFloat()*360, 0)
	if entity == nil {
		return ItemUseResultNone
	}

	if s.HasCustomName() {
		entity.SetNameTag(s.GetCustomName())
	}
	s.Pop()
	entity.SpawnToAll()
	//TODO: what if the entity was marked for deletion?
	return ItemUseResultSuccess
}

// OnInteractBlock is a port of PaintingItem::onInteractBlock.
func (p *PaintingItem) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if math.FacingAxis(face) == math.AxisY {
		return ItemUseResultNone
	}
	if PlacePaintingFunc == nil || !PlacePaintingFunc(player, blockReplace, blockClicked, face) { //No space available
		return ItemUseResultNone
	}
	p.Pop()
	return ItemUseResultSuccess
}

// OnInteractBlock is a port of EndCrystal::onInteractBlock.
func (e *EndCrystal) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if blockClicked.GetTypeId() == block.OBSIDIAN || blockClicked.GetTypeId() == block.BEDROCK {
		if PlaceEndCrystalFunc != nil && PlaceEndCrystalFunc(blockClicked) {
			e.Pop()
			return ItemUseResultSuccess
		}
	}
	return ItemUseResultNone
}

// OnInteractBlock is a port of FireworkRocket::onInteractBlock.
func (f *FireworkRocket) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if LaunchFireworkFunc == nil {
		return ItemUseResultNone
	}
	//TODO: this would be nicer if Vector3::getSide() accepted floats for distance
	position := blockClicked.GetPosition().Vector3.AddVector(clickVector).AddVector(math.Vector3{}.GetSide(face, 1).Multiply(0.15))

	randomDuration := ((f.FlightTimeMultiplier + 1) * 10) + rand.IntN(13) // mt_rand(0, 12)

	LaunchFireworkFunc(f, player, position, randomDuration)

	f.Pop()

	return ItemUseResultSuccess
}
