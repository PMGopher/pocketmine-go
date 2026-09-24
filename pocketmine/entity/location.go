package entity

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
)

// Location is a port of pocketmine\entity\Location: a position in a world plus a yaw and pitch.
// It's a value type (PHP code clones Locations wherever it would otherwise share one).
type Location struct {
	math.Vector3
	World *world.World
	Yaw   float64
	Pitch float64
}

// NewLocation is a port of Location::__construct.
func NewLocation(x, y, z float64, w *world.World, yaw, pitch float64) Location {
	return Location{Vector3: math.NewVector3(x, y, z), World: w, Yaw: yaw, Pitch: pitch}
}

// LocationFromObject is a port of Location::fromObject.
func LocationFromObject(pos math.Vector3, w *world.World, yaw, pitch float64) Location {
	return Location{Vector3: pos, World: w, Yaw: yaw, Pitch: pitch}
}

// AsLocation is a port of Location::asLocation (a copy).
func (l Location) AsLocation() Location { return l }

// AsVector3 is a port of Vector3::asVector3.
func (l Location) AsVector3() math.Vector3 { return l.Vector3 }

func (l Location) GetYaw() float64 { return l.Yaw }

func (l Location) GetPitch() float64 { return l.Pitch }

// IsValid is a port of Position::isValid: whether the location has a world.
func (l Location) IsValid() bool { return l.World != nil }

// GetWorld is a port of Position::getWorld. Panics if the location has no world (PHP's
// AssumptionFailedError).
func (l Location) GetWorld() *world.World {
	if l.World == nil {
		panic("Position world is null")
	}
	return l.World
}

// Equals is a port of Location::equals.
func (l Location) Equals(o Location) bool {
	return l.Vector3.Equals(o.Vector3) && l.World == o.World && l.Yaw == o.Yaw && l.Pitch == o.Pitch
}

func (l Location) String() string {
	worldName := "null"
	if l.World != nil {
		worldName = l.World.GetDisplayName()
	}
	return fmt.Sprintf("Location (world=%s, x=%v, y=%v, z=%v, yaw=%v, pitch=%v)", worldName, l.X, l.Y, l.Z, l.Yaw, l.Pitch)
}
