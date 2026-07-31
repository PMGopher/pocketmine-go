package runtime

import "pocketmine-go/pocketmine/math"

// DataDescriber is a port of pocketmine\data\runtime\RuntimeDataDescriber, implemented by
// Reader, Writer and SizeCalculator.
//
// PHP passes state fields by reference (`int &$value`) so the exact same describe method,
// written once per block type, can either read or write depending on which concrete describer
// is used. Go has no reference parameters, but a pointer gives the identical "read-or-write
// through this pointer" capability, so every field here is a pointer parameter instead.
//
// PHP's enum()/enumSet() (and the deprecated per-named-enum-type shim methods in
// RuntimeEnumDescriber) exist to work around \UnitEnum having no usable native integer value —
// RuntimeEnumMetadata reflects over an enum's cases, sorts them by name, and caches an index
// assignment for each. Go enums are typically just int-based types declared with iota, which
// already have a stable native integer value with no reflection needed, so that whole mechanism
// isn't ported: block state descriptions for a small fixed set of cases just use BoundedIntAuto
// directly on the enum's underlying int value.
//
// This interface isn't sealed the way PHP's is (nothing stops another package from implementing
// it) — but by the same token, it's expected to grow more methods over time as more of block/utils
// (WallConnectionType, rail shapes, etc.) gets ported, the same way PHP's grew.
type DataDescriber interface {
	Int(bits int, value *int)
	BoundedIntAuto(min, max int, value *int)
	Bool(value *bool)
	HorizontalFacing(facing *math.Facing)
	FacingFlags(faces *[]math.Facing)
	HorizontalFacingFlags(faces *[]math.Facing)
	Facing(facing *math.Facing)
	FacingExcept(facing *math.Facing, except math.Facing)
	Axis(axis *math.Axis)
	HorizontalAxis(axis *math.Axis)
}

// InvalidSerializedRuntimeDataError is a port of pocketmine\data\runtime\InvalidSerializedRuntimeDataException.
type InvalidSerializedRuntimeDataError struct{ Message string }

func (e *InvalidSerializedRuntimeDataError) Error() string { return e.Message }
