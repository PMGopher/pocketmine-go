package block

import (
	stdmath "math"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// DaylightSensor is a port of pocketmine\block\DaylightSensor.
type DaylightSensor struct {
	Transparent
	AnalogRedstoneSignalEmitterComponent

	Inverted bool
}

func NewDaylightSensor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DaylightSensor {
	d := &DaylightSensor{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}
	d.Init(d)
	return d
}

func (d *DaylightSensor) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DaylightSensor) DescribeBlockOnlyState(w runtime.DataDescriber) {
	d.DescribeSignalStrength(w)
	w.Bool(&d.Inverted)
}

func (d *DaylightSensor) IsInverted() bool { return d.Inverted }

func (d *DaylightSensor) SetInverted(inverted bool) { d.Inverted = inverted }

func (d *DaylightSensor) GetFuelTime() int { return 300 }

func (d *DaylightSensor) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 10.0/16)}
}

func (d *DaylightSensor) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (d *DaylightSensor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	d.Inverted = !d.Inverted
	d.SignalStrength = d.recalculateSignalStrength()
	if world, err := d.position.GetWorld(); err == nil {
		if err := world.SetBlock(d.position, d.self); err != nil {
			panic(err)
		}
	}
	return true
}

func (d *DaylightSensor) OnScheduledUpdate() {
	world, err := d.position.GetWorld()
	if err != nil {
		return
	}
	signalStrength := d.recalculateSignalStrength()
	if d.SignalStrength != signalStrength {
		d.SignalStrength = signalStrength
		if err := world.SetBlock(d.position, d.self); err != nil {
			panic(err)
		}
	}
	world.ScheduleDelayedBlockUpdate(d.position.AsVector3(), 20)
}

func (d *DaylightSensor) recalculateSignalStrength() int {
	world, err := d.position.GetWorld()
	if err != nil {
		return 0
	}
	pos := d.position.AsVector3()
	lightLevel := world.GetRealBlockSkyLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())
	if d.Inverted {
		return 15 - lightLevel
	}

	sunAngle := world.GetSunAnglePercentage()
	offset := 0.0
	if sunAngle >= 0.5 {
		offset = 1.0
	}
	adjustedAngle := (sunAngle + ((offset - sunAngle) / 5)) * 2 * stdmath.Pi

	result := int(stdmath.Round(float64(lightLevel) * stdmath.Cos(adjustedAngle)))
	if result < 0 {
		result = 0
	}
	return result
}
