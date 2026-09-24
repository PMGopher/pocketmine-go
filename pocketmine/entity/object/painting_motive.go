// Package object is a port of pocketmine\entity\object: non-living entities (item entities,
// falling blocks, primed TNT, experience orbs, paintings, area effect clouds, end crystals,
// firework rockets).
package object

import (
	"fmt"
	"sync"
)

// PaintingMotive is a port of pocketmine\entity\object\PaintingMotive.
type PaintingMotive struct {
	width  int
	height int
	name   string
}

func NewPaintingMotive(width, height int, name string) *PaintingMotive {
	return &PaintingMotive{width: width, height: height, name: name}
}

var (
	paintingMotivesOnce sync.Once
	paintingMotives     map[string]*PaintingMotive
	paintingMotiveOrder []string
)

func initPaintingMotives() {
	paintingMotivesOnce.Do(func() {
		paintingMotives = map[string]*PaintingMotive{}
		for _, m := range []*PaintingMotive{
			NewPaintingMotive(1, 1, "Alban"),
			NewPaintingMotive(1, 1, "Aztec"),
			NewPaintingMotive(1, 1, "Aztec2"),
			NewPaintingMotive(1, 1, "Bomb"),
			NewPaintingMotive(1, 1, "Kebab"),
			NewPaintingMotive(1, 1, "meditative"),
			NewPaintingMotive(1, 1, "Plant"),
			NewPaintingMotive(1, 1, "Wasteland"),
			NewPaintingMotive(1, 2, "Graham"),
			NewPaintingMotive(1, 2, "prairie_ride"),
			NewPaintingMotive(1, 2, "Wanderer"),
			NewPaintingMotive(2, 1, "Courbet"),
			NewPaintingMotive(2, 1, "Creebet"),
			NewPaintingMotive(2, 1, "Pool"),
			NewPaintingMotive(2, 1, "Sea"),
			NewPaintingMotive(2, 1, "Sunset"),
			NewPaintingMotive(2, 2, "Bust"),
			NewPaintingMotive(2, 2, "baroque"),
			NewPaintingMotive(2, 2, "Earth"),
			NewPaintingMotive(2, 2, "Fire"),
			NewPaintingMotive(2, 2, "humble"),
			NewPaintingMotive(2, 2, "Match"),
			NewPaintingMotive(2, 2, "SkullAndRoses"),
			NewPaintingMotive(2, 2, "Stage"),
			NewPaintingMotive(2, 2, "Void"),
			NewPaintingMotive(2, 2, "Water"),
			NewPaintingMotive(2, 2, "Wind"),
			NewPaintingMotive(2, 2, "Wither"),
			NewPaintingMotive(3, 3, "bouquet"),
			NewPaintingMotive(3, 3, "cavebird"),
			NewPaintingMotive(3, 3, "cotan"),
			NewPaintingMotive(3, 3, "endboss"),
			NewPaintingMotive(3, 3, "fern"),
			NewPaintingMotive(3, 3, "owlemons"),
			NewPaintingMotive(3, 3, "sunflowers"),
			NewPaintingMotive(3, 3, "tides"),
			NewPaintingMotive(3, 4, "backyard"),
			NewPaintingMotive(3, 4, "pond"),
			NewPaintingMotive(4, 2, "changing"),
			NewPaintingMotive(4, 2, "Fighters"),
			NewPaintingMotive(4, 2, "finding"),
			NewPaintingMotive(4, 2, "lowmist"),
			NewPaintingMotive(4, 2, "passage"),
			NewPaintingMotive(4, 3, "DonkeyKong"),
			NewPaintingMotive(4, 3, "Skeleton"),
			NewPaintingMotive(4, 4, "BurningSkull"),
			NewPaintingMotive(4, 4, "orb"),
			NewPaintingMotive(4, 4, "Pigscene"),
			NewPaintingMotive(4, 4, "Pointer"),
			NewPaintingMotive(4, 4, "unpacked"),
		} {
			registerMotive(m)
		}
	})
}

func registerMotive(motive *PaintingMotive) {
	if _, exists := paintingMotives[motive.GetName()]; !exists {
		paintingMotiveOrder = append(paintingMotiveOrder, motive.GetName())
	}
	paintingMotives[motive.GetName()] = motive
}

// RegisterPaintingMotive is a port of PaintingMotive::registerMotive.
func RegisterPaintingMotive(motive *PaintingMotive) {
	initPaintingMotives()
	registerMotive(motive)
}

// GetPaintingMotiveByName is a port of PaintingMotive::getMotiveByName (nil if unknown).
func GetPaintingMotiveByName(name string) *PaintingMotive {
	initPaintingMotives()
	return paintingMotives[name]
}

// GetAllPaintingMotives is a port of PaintingMotive::getAll (registration order).
func GetAllPaintingMotives() []*PaintingMotive {
	initPaintingMotives()
	result := make([]*PaintingMotive, 0, len(paintingMotiveOrder))
	for _, name := range paintingMotiveOrder {
		result = append(result, paintingMotives[name])
	}
	return result
}

func (m *PaintingMotive) GetName() string { return m.name }

func (m *PaintingMotive) GetWidth() int { return m.width }

func (m *PaintingMotive) GetHeight() int { return m.height }

func (m *PaintingMotive) String() string {
	return fmt.Sprintf("PaintingMotive(name: %s, height: %d, width: %d)", m.GetName(), m.GetHeight(), m.GetWidth())
}
