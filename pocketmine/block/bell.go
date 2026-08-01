package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Bell is a port of pocketmine\block\Bell.
type Bell struct {
	Transparent
	HorizontalFacingComponent

	AttachmentType blockutils.BellAttachmentType
}

func NewBell(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Bell {
	b := &Bell{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		AttachmentType:            blockutils.BellAttachmentTypeFloor,
	}
	b.Init(b)
	return b
}

func (b *Bell) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bell) DescribeBlockOnlyState(w runtime.DataDescriber) {
	t := int(b.AttachmentType)
	w.BoundedIntAuto(int(blockutils.BellAttachmentTypeCeiling), int(blockutils.BellAttachmentTypeTwoWalls), &t)
	b.AttachmentType = blockutils.BellAttachmentType(t)
	b.DescribeHorizontalFacing(w)
}

func (b *Bell) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	switch b.AttachmentType {
	case blockutils.BellAttachmentTypeFloor:
		return []math.AxisAlignedBB{
			math.OneAABB().SquashedCopy(math.FacingAxis(b.Facing), 1.0/4).TrimmedCopy(math.Up, 3.0/16),
		}
	case blockutils.BellAttachmentTypeCeiling:
		return []math.AxisAlignedBB{
			math.OneAABB().ContractedCopy(1.0/4, 0, 1.0/4).TrimmedCopy(math.Down, 1.0/4),
		}
	}

	box := math.OneAABB().
		SquashedCopy(math.FacingAxis(math.RotateY(b.Facing, true)), 1.0/4).
		TrimmedCopy(math.Up, 1.0/16).
		TrimmedCopy(math.Down, 1.0/4)
	if b.AttachmentType == blockutils.BellAttachmentTypeOneWall {
		box = box.TrimmedCopy(b.Facing, 3.0/16)
	}
	return []math.AxisAlignedBB{box}
}

func (b *Bell) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *Bell) GetAttachmentType() blockutils.BellAttachmentType { return b.AttachmentType }

func (b *Bell) SetAttachmentType(attachmentType blockutils.BellAttachmentType) {
	b.AttachmentType = attachmentType
}

// bellCanBeSupportedAt is a port of Bell::canBeSupportedAt.
func bellCanBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(face) != blockutils.SupportTypeNone
}

func (b *Bell) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !bellCanBeSupportedAt(blockReplace, math.Opposite(face)) {
		return false
	}
	switch face {
	case math.Up:
		if player != nil {
			b.Facing = math.Opposite(player.GetHorizontalFacing())
		}
		b.AttachmentType = blockutils.BellAttachmentTypeFloor
	case math.Down:
		b.AttachmentType = blockutils.BellAttachmentTypeCeiling
	default:
		b.Facing = face
		if bellCanBeSupportedAt(blockReplace, face) {
			b.AttachmentType = blockutils.BellAttachmentTypeTwoWalls
		} else {
			b.AttachmentType = blockutils.BellAttachmentTypeOneWall
		}
	}
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (b *Bell) OnNearbyBlockChange() {
	var directions []math.Facing
	switch b.AttachmentType {
	case blockutils.BellAttachmentTypeCeiling:
		directions = []math.Facing{math.Up}
	case blockutils.BellAttachmentTypeFloor:
		directions = []math.Facing{math.Down}
	case blockutils.BellAttachmentTypeOneWall:
		directions = []math.Facing{math.Opposite(b.Facing)}
	case blockutils.BellAttachmentTypeTwoWalls:
		directions = []math.Facing{b.Facing, math.Opposite(b.Facing)}
	}

	for _, dir := range directions {
		if !bellCanBeSupportedAt(b.self, dir) {
			world, err := b.position.GetWorld()
			if err != nil {
				return
			}
			world.UseBreakOn(b.position.AsVector3())
			break
		}
	}
}

func (b *Bell) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player != nil {
		faceHit := math.Opposite(player.GetHorizontalFacing())
		if b.isValidFaceToRing(faceHit) {
			b.Ring(faceHit)
			return true
		}
	}
	return false
}

func (b *Bell) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	faceHit := math.Opposite(projectile.GetHorizontalFacing())
	if b.isValidFaceToRing(faceHit) {
		b.Ring(faceHit)
	}
}

// Ring is a port of Bell::ring. Broadcasting the fake update packet (for the visual swing) to
// viewers needs block/tile.Bell and the network protocol layer, neither ported yet, so only the
// sound is played for now - see Block.GetDropsForCompatibleTool's doc comment for the same
// category of gap.
func (b *Bell) Ring(faceHit math.Facing) {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(b.position.AsVector3(), sound.BellRingSound{})
}

func (b *Bell) isValidFaceToRing(faceHit math.Facing) bool {
	switch b.AttachmentType {
	case blockutils.BellAttachmentTypeCeiling:
		return true
	case blockutils.BellAttachmentTypeFloor:
		return math.FacingAxis(faceHit) == math.FacingAxis(b.Facing)
	case blockutils.BellAttachmentTypeOneWall, blockutils.BellAttachmentTypeTwoWalls:
		return faceHit == math.RotateY(b.Facing, false) || faceHit == math.RotateY(b.Facing, true)
	}
	return false
}
