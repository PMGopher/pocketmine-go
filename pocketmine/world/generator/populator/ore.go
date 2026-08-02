package populator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/object"
)

// Ore is a port of pocketmine\world\generator\populator\Ore.
type Ore struct {
	OreTypes []*object.OreType
}

func (o *Ore) SetOreTypes(types []*object.OreType) { o.OreTypes = types }

func (o *Ore) Populate(world block.World, chunkX, chunkZ int, random *utils.Random) {
	for _, oreType := range o.OreTypes {
		ore := object.NewOre(random, oreType)
		for i := 0; i < ore.Type.ClusterCount; i++ {
			x := random.NextRange(chunkX<<4, (chunkX<<4)+format.SubChunkEdgeLength-1)
			y := random.NextRange(ore.Type.MinHeight, ore.Type.MaxHeight)
			z := random.NextRange(chunkZ<<4, (chunkZ<<4)+format.SubChunkEdgeLength-1)
			if ore.CanPlaceObject(world, x, y, z) {
				ore.PlaceObject(world, x, y, z)
			}
		}
	}
}
