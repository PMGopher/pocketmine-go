package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const comparatorTagOutputSignal = "OutputSignal"

// Comparator is a port of pocketmine\block\tile\Comparator.
//
// Deprecated in the PHP original too - see block.RedstoneComparator.
type Comparator struct {
	TileBase

	SignalStrength int
}

func NewComparator(world World, pos math.Vector3) *Comparator {
	c := &Comparator{TileBase: NewTileBase(world, pos)}
	c.Init(c)
	return c
}

func (c *Comparator) SaveID() string { return "Comparator" }

func (c *Comparator) GetSignalStrength() int { return c.SignalStrength }

func (c *Comparator) SetSignalStrength(signalStrength int) { c.SignalStrength = signalStrength }

func (c *Comparator) ReadSaveData(tag *nbt.CompoundTag) error {
	c.SignalStrength = int(tag.GetIntOr(comparatorTagOutputSignal, 0))
	return nil
}

func (c *Comparator) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetInt(comparatorTagOutputSignal, nbt.IntTag(c.SignalStrength))
}
