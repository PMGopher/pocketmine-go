package object

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
)

// GrowGrass is a port of pocketmine\world\generator\object\TallGrass::growGrass: count random
// flowers and tall grass on the grass blocks within radius of pos (PHP's defaults are 15 and 10).
func GrowGrass(world block.World, x, y, z int, random *utils.Random, count, radius int) {
	tallGrass := block.VanillaTallGrass()
	arr := []block.Behavior{
		block.VanillaBlock("dandelion"),
		block.VanillaBlock("poppy"),
		tallGrass,
		tallGrass,
		tallGrass,
		tallGrass,
	}
	arrC := len(arr) - 1
	for c := 0; c < count; c++ {
		xx := random.NextRange(x-radius, x+radius)
		zz := random.NextRange(z-radius, z+radius)
		if world.GetBlockAt(xx, y+1, zz).GetTypeId() == block.AIR && world.GetBlockAt(xx, y, zz).GetTypeId() == block.GRASS {
			_ = world.SetBlock(block.NewPosition(float64(xx), float64(y+1), float64(zz), world), arr[random.NextRange(0, arrC)].Clone())
		}
	}
}
