package runtime

import "pocketmine-go/pocketmine/math"

// SizeCalculator is a port of pocketmine\data\runtime\RuntimeDataSizeCalculator: runs the same
// describe method as Reader/Writer, but only tallies how many bits it would need.
type SizeCalculator struct {
	bits int
}

func NewSizeCalculator() *SizeCalculator {
	return &SizeCalculator{}
}

func (c *SizeCalculator) addBits(bits int) { c.bits += bits }

func (c *SizeCalculator) GetBitsUsed() int { return c.bits }

func (c *SizeCalculator) Int(bits int, value *int) { c.addBits(bits) }

func (c *SizeCalculator) BoundedIntAuto(min, max int, value *int) {
	c.addBits(boundedIntAutoBits(min, max))
}

func (c *SizeCalculator) Bool(value *bool) { c.addBits(1) }

func (c *SizeCalculator) HorizontalFacing(facing *math.Facing) { c.addBits(2) }

func (c *SizeCalculator) FacingFlags(faces *[]math.Facing) { c.addBits(len(math.AllFacing)) }

func (c *SizeCalculator) HorizontalFacingFlags(faces *[]math.Facing) {
	c.addBits(len(math.HorizontalFacing))
}

func (c *SizeCalculator) Facing(facing *math.Facing) { c.addBits(3) }

func (c *SizeCalculator) FacingExcept(facing *math.Facing, except math.Facing) { c.Facing(facing) }

func (c *SizeCalculator) Axis(axis *math.Axis) { c.addBits(2) }

func (c *SizeCalculator) HorizontalAxis(axis *math.Axis) { c.addBits(1) }

var _ DataDescriber = (*SizeCalculator)(nil)
