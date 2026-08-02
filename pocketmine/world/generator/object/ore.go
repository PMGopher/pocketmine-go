package object

import (
	"math"

	"pocketmine-go/pocketmine/block"
	pmmath "pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
)

// Ore is a port of pocketmine\world\generator\object\Ore (aliased ObjectOre in PHP to avoid a
// name clash with populator\Ore - no such clash exists here since it's a different Go package).
type Ore struct {
	random *utils.Random
	Type   *OreType
}

func NewOre(random *utils.Random, oreType *OreType) *Ore {
	return &Ore{random: random, Type: oreType}
}

func (o *Ore) CanPlaceObject(world block.World, x, y, z int) bool {
	return world.GetBlockAt(x, y, z).GetTypeId() == o.Type.Replaces.GetTypeId()
}

// PlaceObject is a port of Ore::placeObject. The real PHP uses a SubChunkExplorer to cache the
// current subchunk across placeSphere's tight loop - this port doesn't have that abstraction, so
// placeSphere goes through World.GetBlockAt/SetBlock directly on every visited position instead;
// behaviourally identical, just without the cross-call subchunk-pointer-caching optimisation.
func (o *Ore) PlaceObject(world block.World, x, y, z int) {
	clusterSize := o.Type.ClusterSize
	angle := o.random.NextFloat() * math.Pi
	offset := pmmath.GetDirection2D(angle).Multiply(float64(clusterSize) / 8)
	x1 := float64(x) + 8 + offset.X
	x2 := float64(x) + 8 - offset.X
	z1 := float64(z) + 8 + offset.Y
	z2 := float64(z) + 8 - offset.Y
	y1 := float64(y + o.random.NextBoundedInt(3) + 2)
	y2 := float64(y + o.random.NextBoundedInt(3) + 2)

	visited := map[[3]int]bool{}
	replaceableStateIDs := map[int32]bool{}

	for count := 0; count <= clusterSize; count++ {
		centerX := x1 + (x2-x1)*float64(count)/float64(clusterSize)
		centerY := y1 + (y2-y1)*float64(count)/float64(clusterSize)
		centerZ := z1 + (z2-z1)*float64(count)/float64(clusterSize)
		radius := ((math.Sin(float64(count)*(math.Pi/float64(clusterSize)))+1)*o.random.NextFloat()*float64(clusterSize)/16 + 1) / 2

		o.placeSphere(world, centerX, centerY, centerZ, radius, visited, replaceableStateIDs)
	}
}

// placeSphere is a port of Ore::placeSphere.
func (o *Ore) placeSphere(world block.World, centerX, centerY, centerZ, radius float64, visited map[[3]int]bool, replaceableStateIDs map[int32]bool) {
	startX, startY, startZ := int(centerX-radius), int(centerY-radius), int(centerZ-radius)
	endX, endY, endZ := int(centerX+radius), int(centerY+radius), int(centerZ+radius)

	for xx := startX; xx <= endX; xx++ {
		sizeX := (float64(xx) + 0.5 - centerX) / radius
		sizeX *= sizeX
		if sizeX >= 1 {
			continue
		}
		for yy := startY; yy <= endY; yy++ {
			sizeY := (float64(yy) + 0.5 - centerY) / radius
			sizeY *= sizeY
			if yy <= 0 || sizeX+sizeY >= 1 {
				continue
			}
			for zz := startZ; zz <= endZ; zz++ {
				sizeZ := (float64(zz) + 0.5 - centerZ) / radius
				sizeZ *= sizeZ
				if sizeX+sizeY+sizeZ >= 1 {
					continue
				}

				pos := [3]int{xx, yy, zz}
				if visited[pos] {
					continue
				}
				visited[pos] = true

				stateID := int32(world.GetBlockAt(xx, yy, zz).GetStateId())
				replaceable, ok := replaceableStateIDs[stateID]
				if !ok {
					replaceable = world.GetBlockAt(xx, yy, zz).GetTypeId() == o.Type.Replaces.GetTypeId()
					replaceableStateIDs[stateID] = replaceable
				}
				if replaceable {
					_ = world.SetBlock(block.NewPosition(float64(xx), float64(yy), float64(zz), world), o.Type.Material.Clone())
				}
			}
		}
	}
}
