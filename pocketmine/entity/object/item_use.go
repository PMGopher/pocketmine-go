package object

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// This file is the entity half of the items whose use spawns an entity (see
// item/entity_item_use.go), installed into the item package which can't create entities.

func init() {
	item.CreateSpawnEggEntityFunc = createSpawnEggEntity
	item.PlacePaintingFunc = placePainting
	item.PlaceEndCrystalFunc = placeEndCrystal
	item.LaunchFireworkFunc = launchFirework
}

// createSpawnEggEntity is the createEntity of the anonymous SpawnEgg subclasses VanillaItems
// registers (zombie, squid and villager).
func createSpawnEggEntity(egg *item.SpawnEgg, blockWorld block.World, pos math.Vector3, yaw, pitch float64) item.SpawnedEntity {
	w, ok := blockWorld.(*world.World)
	if !ok {
		return nil
	}
	location := entity.LocationFromObject(pos, w, yaw, pitch)
	switch egg.EntityTypeName {
	case "Zombie":
		return entity.NewZombie(location, nil)
	case "Squid":
		return entity.NewSquid(location, nil)
	case "Villager":
		return entity.NewVillager(location, nil)
	}
	return nil
}

// placePainting is PaintingItem::onInteractBlock from the motive selection on (the item pops
// itself).
func placePainting(player item.Player, blockReplace, blockClicked block.Behavior, face math.Facing) bool {
	replacePos := blockReplace.GetPosition()
	w, ok := item.EntityWorldFunc(player).(*world.World)
	if !ok {
		return false
	}

	var motives []*PaintingMotive

	totalDimension := 0
	for _, motive := range GetAllPaintingMotives() {
		currentTotalDimension := motive.GetHeight() + motive.GetWidth()
		if currentTotalDimension < totalDimension {
			continue
		}

		if CanPaintingFit(w, replacePos.Vector3, face, true, motive) {
			if currentTotalDimension > totalDimension {
				totalDimension = currentTotalDimension
				/*
				 * This drops all motive possibilities smaller than this
				 * We use the total of height + width to allow equal chance of horizontal/vertical paintings
				 * when there is an L-shape of space available.
				 */
				motives = nil
			}

			motives = append(motives, motive)
		}
	}

	if len(motives) == 0 { //No space available
		return false
	}

	motive := motives[rand.IntN(len(motives))]

	clickedPos := blockClicked.GetPosition()

	replaceWorld, _ := replacePos.GetWorld()
	locationWorld, _ := replaceWorld.(*world.World)
	painting := NewPainting(entity.LocationFromObject(replacePos.Vector3, locationWorld, 0, 0), clickedPos.Vector3, face, motive, nil)
	painting.SpawnToAll()

	w.AddSound(replacePos.Add(0.5, 0.5, 0.5), sound.PaintingPlaceSound{})
	return true
}

// placeEndCrystal is EndCrystal::onInteractBlock's check and spawn on the clicked obsidian or
// bedrock (the item checks the block type and pops itself).
func placeEndCrystal(blockClicked block.Behavior) bool {
	pos := blockClicked.GetPosition()
	blockWorld, err := pos.GetWorld()
	if err != nil {
		return false
	}
	w, ok := blockWorld.(*world.World)
	if !ok {
		return false
	}
	bb := math.OneAABB()
	bb.Offset(pos.X, pos.Y, pos.Z).Extend(math.Up, 1)
	sides, ok := blockClicked.(sideGetter)
	if !ok {
		return false
	}
	if len(w.GetNearbyEntities(bb)) == 0 &&
		sides.GetSide(math.Up, 1).GetTypeId() == block.AIR &&
		sides.GetSide(math.Up, 2).GetTypeId() == block.AIR {
		crystal := NewEndCrystal(entity.LocationFromObject(pos.Add(0.5, 1, 0.5), w, 0, 0), nil)
		crystal.SpawnToAll()
		return true
	}
	return false
}

// launchFirework is FireworkRocket::onInteractBlock's entity creation (the item works out the
// position and duration, and pops itself).
func launchFirework(rocket *item.FireworkRocket, player item.Player, position math.Vector3, randomDuration int) {
	w, _ := item.EntityWorldFunc(player).(*world.World)
	fireworkEntity := NewFireworkRocket(entity.LocationFromObject(position, w, utils.GetRandomFloat()*360, 90), randomDuration, rocket.GetExplosions(), nil)
	if owner, ok := player.(world.Entity); ok {
		fireworkEntity.SetOwningEntity(owner)
	}
	fireworkEntity.SetMotion(math.NewVector3(
		(utils.GetRandomFloat()-utils.GetRandomFloat())*0.0023,
		0.05,
		(utils.GetRandomFloat()-utils.GetRandomFloat())*0.0023,
	))
	fireworkEntity.SpawnToAll()
}
