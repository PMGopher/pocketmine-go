package block

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// itemTypeIDsBoneMeal etc. mirror item.* constants (pocketmine-go/pocketmine/item, not yet
// ported) - same reasoning as itemTypeIDsHoneycomb in copper_material.go.
const (
	itemTypeIDsBoneMeal    = 20017
	itemTypeIDsCocoaBeans  = 20073
	itemTypeIDsInkSac      = 20127
	itemTypeIDsLapisLazuli = 20141
	itemTypeIDsGlowInkSac  = 20236
)

// Dye is a forward-compatible marker for pocketmine\item\Dye - same pattern as the Axe interface
// in wood.go.
type Dye interface {
	GetColor() blockutils.DyeColor
}

// signShaper lets concrete leaf types (FloorSign, WallSign) report which side supports them and
// how the sign is oriented - same self-dispatch shape as bannerShaper.
type signShaper interface {
	GetSupportingFace() math.Facing
	GetFacingDegrees() float64
}

// BaseSign is a port of pocketmine\block\BaseSign. Like BaseBanner, this isn't meant to be
// instantiated directly - a concrete leaf type (FloorSign, WallSign) must embed it, implement
// Clone, and satisfy signShaper.
//
// The PHP constructor's `asItemCallback` closure is dropped: it's only used by AsItem(), which
// needs real Item construction from the unported item package regardless (see
// Block.GetDropsForCompatibleTool's doc comment), so AsItem is left as Block's default here too.
type BaseSign struct {
	Transparent
	WoodTypeComponent

	Text                  blockutils.SignText
	BackText              blockutils.SignText
	Waxed                 bool
	EditorEntityRuntimeID int
	HasEditor             bool
}

// ReadStateFromWorld is a port of BaseSign::readStateFromWorld.
func (b *BaseSign) ReadStateFromWorld() Behavior {
	b.Block.ReadStateFromWorld()

	world, err := b.position.GetWorld()
	if err != nil {
		return b.self
	}
	t, _ := world.GetTile(b.position)
	if signTile, ok := t.(*tile.Sign); ok {
		b.Text = signTile.GetText()
		b.BackText = signTile.GetBackText()
		b.Waxed = signTile.IsWaxed()
		editorID, hasEditor := signTile.GetEditorEntityRuntimeID()
		b.EditorEntityRuntimeID, b.HasEditor = int(editorID), hasEditor
	}
	return b.self
}

func (b *BaseSign) IsSolid() bool { return false }

func (b *BaseSign) GetMaxStackSize() int { return 16 }

func (b *BaseSign) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (b *BaseSign) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *BaseSign) OnNearbyBlockChange() {
	supportingFace := b.self.(signShaper).GetSupportingFace()
	if b.self.(blockGeometry).GetSide(supportingFace, 1).GetTypeId() == AIR {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	}
}

func (b *BaseSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		b.EditorEntityRuntimeID = player.GetID()
		b.HasEditor = true
	}
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnPostPlace should open a sign editor GUI for the placing player - needs World.GetEntity and
// the network layer (Player.OpenSignEditor), neither ported yet, so this is a documented no-op
// for now.

func (b *BaseSign) getHitboxCenter() math.Vector3 {
	pos := b.position.AsVector3()
	return pos.Add(0.5, 0.5, 0.5)
}

// doSignChange is a port of BaseSign::doSignChange. SignChangeEvent is treated as always
// uncancelled, matching every other deferred concrete event in this port.
func (b *BaseSign) doSignChange(newText blockutils.SignText, item Item, frontFace bool) bool {
	if frontFace {
		b.Text = newText
	} else {
		b.BackText = newText
	}
	world, err := b.position.GetWorld()
	if err != nil {
		return false
	}
	if err := world.SetBlock(b.position, b.self); err != nil {
		panic(err)
	}
	item.Pop()
	return true
}

func (b *BaseSign) getFaceText(frontFace bool) blockutils.SignText {
	if frontFace {
		return b.Text
	}
	return b.BackText
}

func (b *BaseSign) changeSignGlowingState(glowing bool, item Item, frontFace bool) bool {
	text := b.getFaceText(frontFace)
	if text.IsGlowing() == glowing {
		return false
	}
	baseColor := text.GetBaseColor()
	if !b.doSignChange(blockutils.NewSignText(sliceOfSignTextLines(text), &baseColor, glowing), item, frontFace) {
		return false
	}
	if world, err := b.position.GetWorld(); err == nil {
		world.AddSound(b.position.AsVector3(), sound.InkSacUseSound{})
	}
	return true
}

func sliceOfSignTextLines(text blockutils.SignText) []string {
	lines := text.GetLines()
	return lines[:]
}

func (b *BaseSign) wax(item Item) bool {
	if b.Waxed {
		return false
	}
	b.Waxed = true
	world, err := b.position.GetWorld()
	if err != nil {
		return false
	}
	if err := world.SetBlock(b.position, b.self); err != nil {
		panic(err)
	}
	item.Pop()
	return true
}

func (b *BaseSign) interactsFront(hitboxCenter, playerPosition math.Vector3, signFacingDegrees float64) bool {
	playerCenterDiffX := playerPosition.X - hitboxCenter.X
	playerCenterDiffZ := playerPosition.Z - hitboxCenter.Z

	f1 := stdmath.Atan2(playerCenterDiffZ, playerCenterDiffX)*180/stdmath.Pi - 90.0

	rotationDiff := signFacingDegrees - f1
	rotation := stdmath.Mod(rotationDiff+180.0, 360.0) - 180.0
	return stdmath.Abs(rotation) <= 90.0
}

// OnInteract is a port of BaseSign::onInteract, minus the final openSignEditor call (needs the
// network layer, not ported yet - documented below). The dye-colouring/glow-toggle/waxing logic
// is fully functional.
func (b *BaseSign) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return false
	}
	if b.Waxed {
		return true
	}

	shaper := b.self.(signShaper)
	frontFace := b.interactsFront(b.self.(interface{ getHitboxCenter() math.Vector3 }).getHitboxCenter(), player.GetPosition(), shaper.GetFacingDegrees())

	var dyeColor blockutils.DyeColor
	hasDyeColor := false
	if dye, ok := item.(Dye); ok {
		dyeColor = dye.GetColor()
		hasDyeColor = true
	} else {
		switch item.GetTypeId() {
		case itemTypeIDsBoneMeal:
			dyeColor, hasDyeColor = blockutils.DyeColorWhite, true
		case itemTypeIDsLapisLazuli:
			dyeColor, hasDyeColor = blockutils.DyeColorBlue, true
		case itemTypeIDsCocoaBeans:
			dyeColor, hasDyeColor = blockutils.DyeColorBrown, true
		}
	}

	if hasDyeColor {
		rgb := dyeColor.GetRgbValue()
		text := b.getFaceText(frontFace)
		if rgb.ToARGB() != text.GetBaseColor().ToARGB() {
			if b.doSignChange(blockutils.NewSignText(sliceOfSignTextLines(text), &rgb, text.IsGlowing()), item, frontFace) {
				if world, err := b.position.GetWorld(); err == nil {
					world.AddSound(b.position.AsVector3(), sound.DyeUseSound{})
				}
				return true
			}
		}
	} else {
		handled := false
		switch item.GetTypeId() {
		case itemTypeIDsInkSac:
			handled = b.changeSignGlowingState(false, item, frontFace)
		case itemTypeIDsGlowInkSac:
			handled = b.changeSignGlowingState(true, item, frontFace)
		case itemTypeIDsHoneycomb:
			handled = b.wax(item)
		}
		if handled {
			return true
		}
	}

	// PHP falls through to player.openSignEditor(...) here and returns true - needs the network
	// layer, not ported yet.
	return true
}

func (b *BaseSign) GetText() blockutils.SignText { return b.Text }

func (b *BaseSign) SetText(text blockutils.SignText) { b.Text = text }

func (b *BaseSign) GetFaceText(frontFace bool) blockutils.SignText { return b.getFaceText(frontFace) }

func (b *BaseSign) SetFaceText(frontFace bool, text blockutils.SignText) {
	if frontFace {
		b.Text = text
	} else {
		b.BackText = text
	}
}

func (b *BaseSign) IsWaxed() bool { return b.Waxed }

func (b *BaseSign) SetWaxed(waxed bool) { b.Waxed = waxed }

func (b *BaseSign) GetEditorEntityRuntimeID() (int, bool) {
	return b.EditorEntityRuntimeID, b.HasEditor
}

func (b *BaseSign) SetEditorEntityRuntimeID(id int, has bool) {
	b.EditorEntityRuntimeID = id
	b.HasEditor = has
}

func (b *BaseSign) GetFuelTime() int {
	if b.WoodType.IsFlammable() {
		return 200
	}
	return 0
}
