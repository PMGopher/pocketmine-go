package blockconvert

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// The block interfaces, traits and classes the PHP property closures are typed with
// (fn(HorizontalFacing $b) => ...), as Go interfaces over the block methods they call. Using
// interfaces (rather than concrete *block.X types) keeps shared property sets working for every
// block that has the methods, like PHP's subclassing does (e.g. CopperDoor extends Door).

type facingHolder interface {
	GetFacing() math.Facing
	SetFacing(facing math.Facing)
}

type axisHolder interface {
	GetAxis() math.Axis
	SetAxis(axis math.Axis)
}

type multiFacingHolder interface {
	GetFaces() []math.Facing
	SetFaces(faces []math.Facing)
}

type rotationHolder interface {
	GetRotation() int
	SetRotation(rotation int)
}

type analogRedstoneSignalEmitter interface {
	GetOutputSignalStrength() int
	SetOutputSignalStrength(signalStrength int)
}

type ageable interface {
	GetAge() int
	SetAge(age int)
}

type lightable interface {
	IsLit() bool
	SetLit(lit bool)
}

type colored interface {
	GetColor() blockutils.DyeColor
	SetColor(color blockutils.DyeColor)
}

type topHalf interface {
	IsTop() bool
	SetTop(top bool)
}

type openable interface {
	IsOpen() bool
	SetOpen(open bool)
}

type poweredByRedstone interface {
	IsPowered() bool
	SetPowered(powered bool)
}

type slabLike interface {
	GetSlabType() blockutils.SlabType
	SetSlabType(slabType blockutils.SlabType)
}

type coralMaterial interface {
	IsDead() bool
	SetDead(dead bool)
	GetCoralType() blockutils.CoralType
	SetCoralType(coralType blockutils.CoralType)
}

type copperMaterial interface {
	IsWaxed() bool
	SetWaxed(waxed bool)
	GetOxidation() blockutils.CopperOxidation
	SetOxidation(oxidation blockutils.CopperOxidation)
}

type liquidLike interface {
	GetDecay() int
	SetDecay(decay int)
	IsFalling() bool
	SetFalling(falling bool)
	IsStill() bool
	SetStill(still bool)
}

type woodLike interface {
	IsStripped() bool
	SetStripped(stripped bool)
}

type buttonLike interface {
	IsPressed() bool
	SetPressed(pressed bool)
}

type doorLike interface {
	topHalf
	openable
	IsHingeRight() bool
	SetHingeRight(hingeRight bool)
}

type fenceGateLike interface {
	openable
	IsInWall() bool
	SetInWall(inWall bool)
}

type itemFrameLike interface {
	HasMap() bool
	SetHasMap(hasMap bool)
}

type stairLike interface {
	IsUpsideDown() bool
	SetUpsideDown(upsideDown bool)
}

type wallLike interface {
	IsPost() bool
	SetPost(post bool)
	GetConnection(face math.Facing) (blockutils.WallConnectionType, bool)
	SetConnection(face math.Facing, connType blockutils.WallConnectionType, present bool)
}

type hangingHolder interface {
	IsHanging() bool
	SetHanging(hanging bool)
}
