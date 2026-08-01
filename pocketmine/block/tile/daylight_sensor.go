package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// DaylightSensor is a port of pocketmine\block\tile\DaylightSensor.
//
// Deprecated in the PHP original too - as per the wiki, this is an old hack to force daylight
// sensors to get updated every game tick, kept only so vanilla can understand daylight sensors in
// worlds created by PM.
type DaylightSensor struct {
	TileBase
}

func NewDaylightSensor(world World, pos math.Vector3) *DaylightSensor {
	d := &DaylightSensor{TileBase: NewTileBase(world, pos)}
	d.Init(d)
	return d
}

func (d *DaylightSensor) SaveID() string { return "DaylightDetector" }

func (d *DaylightSensor) ReadSaveData(nbt *nbt.CompoundTag) error { return nil }

func (d *DaylightSensor) WriteSaveData(nbt *nbt.CompoundTag) {}
