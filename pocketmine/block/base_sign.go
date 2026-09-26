package block

import (
	"fmt"
	stdmath "math"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/utils"

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
// The PHP constructor's `asItemCallback` closure is replaced by BaseSign.AsItem, which builds the
// sign item of the wood type.
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

// signEditor is the surface of pocketmine\player\Player the sign editor needs.
type signEditor interface {
	IsConnected() bool
	OpenSignEditor(position math.Vector3, frontFace bool)
}

// OnPostPlace is a port of BaseSign::onPostPlace: the placing player gets the sign editor.
func (b *BaseSign) OnPostPlace() {
	if !b.HasEditor || GetEntityFunc == nil {
		return
	}
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	//TODO: HACK! We really shouldn't be keeping disconnected players (and generally flagged-for-despawn entities)
	//in the world's entity table, but changing that is too risky for a hotfix. This workaround will do for now.
	if e, ok := GetEntityFunc(world, b.EditorEntityRuntimeID); ok {
		if p, ok := e.(signEditor); ok && p.IsConnected() {
			p.OpenSignEditor(b.position.AsVector3(), true)
		}
	}
}

func (b *BaseSign) getHitboxCenter() math.Vector3 {
	pos := b.position.AsVector3()
	return pos.Add(0.5, 0.5, 0.5)
}

// doSignChange is a port of BaseSign::doSignChange.
func (b *BaseSign) doSignChange(newText blockutils.SignText, player Player, item Item, frontFace bool) bool {
	ev := blockevent.NewSignChangeEvent(b.self, player, b.getFaceText(frontFace), newText, frontFace)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	b.setFaceText(frontFace, ev.GetNewText().(blockutils.SignText))
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

func (b *BaseSign) setFaceText(frontFace bool, text blockutils.SignText) {
	if frontFace {
		b.Text = text
	} else {
		b.BackText = text
	}
}

func (b *BaseSign) getFaceText(frontFace bool) blockutils.SignText {
	if frontFace {
		return b.Text
	}
	return b.BackText
}

func (b *BaseSign) changeSignGlowingState(glowing bool, player Player, item Item, frontFace bool) bool {
	text := b.getFaceText(frontFace)
	if text.IsGlowing() == glowing {
		return false
	}
	baseColor := text.GetBaseColor()
	if !b.doSignChange(blockutils.NewSignText(sliceOfSignTextLines(text), &baseColor, glowing), player, item, frontFace) {
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

// OnInteract is a port of BaseSign::onInteract.
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
			if b.doSignChange(blockutils.NewSignText(sliceOfSignTextLines(text), &rgb, text.IsGlowing()), player, item, frontFace) {
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
			handled = b.changeSignGlowingState(false, player, item, frontFace)
		case itemTypeIDsGlowInkSac:
			handled = b.changeSignGlowingState(true, player, item, frontFace)
		case itemTypeIDsHoneycomb:
			handled = b.wax(item)
		}
		if handled {
			return true
		}
	}

	if editor, ok := player.(signEditor); ok {
		editor.OpenSignEditor(b.position.AsVector3(), frontFace)
	}
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

// UpdateFaceText is a port of BaseSign::updateFaceText: called by the player controller (network
// session) to update the sign text, firing events as appropriate. It reports whether the sign
// update was successful, and returns an error if the text payload is too large.
func (b *BaseSign) UpdateFaceText(author Player, authorName string, frontFace bool, text blockutils.SignText) (bool, error) {
	size := 0
	for _, line := range text.GetLines() {
		size += len(line)
	}
	if size > 1000 {
		return false, fmt.Errorf("%s tried to write %d bytes of text onto a sign (bigger than max 1000)", authorName, size)
	}
	oldText := b.getFaceText(frontFace)
	lines := text.GetLines()
	cleaned := make([]string, len(lines))
	for i, line := range lines {
		cleaned[i] = utils.Clean(line, false)
	}
	baseColor := oldText.GetBaseColor()
	ev := blockevent.NewSignChangeEvent(b.self, author, oldText, blockutils.NewSignText(cleaned, &baseColor, oldText.IsGlowing()), frontFace)
	if b.Waxed || !b.HasEditor || b.EditorEntityRuntimeID != author.GetID() {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return false, nil
	}
	b.setFaceText(frontFace, ev.GetNewText().(blockutils.SignText))
	b.SetEditorEntityRuntimeID(0, false)
	world, err := b.position.GetWorld()
	if err != nil {
		return false, nil
	}
	if err := world.SetBlock(b.position, b.self); err != nil {
		return false, err
	}
	return true, nil
}

// WriteStateToWorld is a port of BaseSign::writeStateToWorld.
func (b *BaseSign) WriteStateToWorld() {
	b.Block.WriteStateToWorld()
	if t, ok := b.tileAt(); ok {
		if signTile, ok := t.(*tile.Sign); ok {
			signTile.SetText(b.Text)
			signTile.SetBackText(b.BackText)
			signTile.SetWaxed(b.Waxed)
			signTile.SetEditorEntityRuntimeID(int64(b.EditorEntityRuntimeID), b.HasEditor)
		}
	}
}
