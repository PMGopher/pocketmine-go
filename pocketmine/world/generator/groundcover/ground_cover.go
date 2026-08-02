// Package groundcover is a port of pocketmine\world\generator\populator\GroundCover. It lives
// outside the populator package because it needs *biome.Registry, and biome already imports
// populator (for TallGrass/Ore) - putting GroundCover in populator too would cycle. It doesn't
// need to import populator at all: Go interfaces are satisfied structurally, so this still counts
// as a populator.Populator (Normal wires it into a []populator.Populator) purely by having a
// matching Populate method.
package groundcover

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/format"
)

// GroundCover is a port of pocketmine\world\generator\populator\GroundCover.
type GroundCover struct {
	Registry *biome.Registry
}

func New(registry *biome.Registry) *GroundCover {
	return &GroundCover{Registry: registry}
}

func (g *GroundCover) Populate(world block.World, chunkX, chunkZ int, random *utils.Random) {
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			worldX, worldZ := chunkX*format.SubChunkEdgeLength+x, chunkZ*format.SubChunkEdgeLength+z

			pos := block.NewPosition(float64(worldX), 0, float64(worldZ), world)
			chunk, ok := world.GetOrLoadChunkAtPosition(pos)
			if !ok {
				continue
			}

			b := g.Registry.GetBiome(int(chunk.GetBiomeID(x, 0, z)))
			cover := b.GroundCover()
			if len(cover) == 0 {
				continue
			}

			diffY := 0
			if !cover[0].IsSolid() {
				diffY = 1
			}

			startY, ok := chunk.GetHighestBlockAt(x, z)
			if !ok {
				// ground cover is supposed to replace preexisting blocks - nothing to do if there
				// are none.
				continue
			}

			for ; startY > 0; startY-- {
				if !world.GetBlockAt(worldX, startY, worldZ).IsTransparent() {
					break
				}
			}
			if startY+diffY < 127 {
				startY += diffY
			} else {
				startY = 127
			}
			endY := startY - len(cover)

			for y := startY; y > endY && y >= 0; y-- {
				cb := cover[startY-y]
				existing := world.GetBlockAt(worldX, y, worldZ)

				if existing.GetTypeId() == block.AIR && cb.IsSolid() {
					break
				}
				if cb.CanBeFlowedInto() && block.IsLiquid(existing) {
					continue
				}

				_ = world.SetBlock(block.NewPosition(float64(worldX), float64(y), float64(worldZ), world), cb.Clone())
			}
		}
	}
}
