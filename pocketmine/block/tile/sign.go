package tile

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	SignTagTextBlob         = "Text"
	SignTagTextColor        = "SignTextColor"
	SignTagGlowingText      = "IgnoreLighting"
	SignTagPersistFormat    = "PersistFormatting"
	SignTagLegacyBugResolve = "TextIgnoreLegacyBugResolved"
	SignTagFrontText        = "FrontText"
	SignTagBackText         = "BackText"
	SignTagWaxed            = "IsWaxed"
	SignTagLockedForEditing = "LockedForEditingBy"
)

// Sign is a port of pocketmine\block\tile\Sign.
//
// Deprecated in the PHP original too - see block.BaseSign (not ported yet).
type Sign struct {
	SpawnableBase

	Text                   blockutils.SignText
	BackText               blockutils.SignText
	Waxed                  bool
	EditorEntityRuntimeID  int64
	HasEditorEntityRuntime bool
}

func NewSign(world World, pos math.Vector3) *Sign {
	s := &Sign{
		SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)},
		Text:          blockutils.NewSignText(nil, nil, false),
		BackText:      blockutils.NewSignText(nil, nil, false),
	}
	s.Init(s)
	return s
}

func (s *Sign) SaveID() string { return "Sign" }

func signReadTextTag(tag *nbt.CompoundTag, lightingBugResolved bool) blockutils.SignText {
	baseColor := color.NewColor(0, 0, 0)
	if baseColorTag, ok := tag.GetTag(SignTagTextColor); ok {
		if intTag, ok := baseColorTag.(nbt.IntTag); ok {
			// PHP round-trips this through Binary::unsignInt/signInt for its 64-bit int semantics;
			// Go's int32<->uint32 conversions already preserve the bit pattern exactly, so the
			// shifts in FromARGB/ToARGB produce the same result without that round trip.
			baseColor = color.FromARGB(int32(intTag))
		}
	}
	glowingText := false
	if lightingBugResolved {
		if glowTag, ok := tag.GetTag(SignTagGlowingText); ok {
			if byteTag, ok := glowTag.(nbt.ByteTag); ok {
				glowingText = byteTag != 0
			}
		}
	}
	blob, _ := tag.GetString(SignTagTextBlob)
	return blockutils.SignTextFromBlob(string(blob), &baseColor, glowingText)
}

func signWriteTextTag(text blockutils.SignText) *nbt.CompoundTag {
	lines := text.GetLines()
	blob := ""
	for i, line := range lines {
		if i > 0 {
			blob += "\n"
		}
		blob += line
	}
	for len(blob) > 0 && blob[len(blob)-1] == '\n' {
		blob = blob[:len(blob)-1]
	}
	tag := nbt.NewCompoundTag()
	tag.SetString(SignTagTextBlob, nbt.StringTag(blob))
	tag.SetInt(SignTagTextColor, nbt.IntTag(text.GetBaseColor().ToARGB()))
	glowing := nbt.ByteTag(0)
	if text.IsGlowing() {
		glowing = 1
	}
	tag.SetByte(SignTagGlowingText, glowing)
	tag.SetByte(SignTagPersistFormat, 1)
	return tag
}

func (s *Sign) ReadSaveData(tag *nbt.CompoundTag) error {
	if frontTextTag, hasCompound, err := tag.GetCompoundTag(SignTagFrontText); err == nil && hasCompound {
		s.Text = signReadTextTag(frontTextTag, true)
	} else if _, ok := tag.GetTag(SignTagTextBlob); ok {
		// MCPE 1.2 save format.
		lightingBugResolved := false
		if resolvedTag, ok := tag.GetTag(SignTagLegacyBugResolve); ok {
			if byteTag, ok := resolvedTag.(nbt.ByteTag); ok {
				lightingBugResolved = byteTag != 0
			}
		}
		s.Text = signReadTextTag(tag, lightingBugResolved)
	} else {
		var lines []string
		for i := 0; i < blockutils.SignTextLineCount; i++ {
			key := fmt.Sprintf("Text%d", i+1)
			if lineTag, ok := tag.GetTag(key); ok {
				if strTag, ok := lineTag.(nbt.StringTag); ok {
					for len(lines) <= i {
						lines = append(lines, "")
					}
					lines[i] = string(strTag)
				}
			}
		}
		s.Text = blockutils.NewSignText(lines, nil, false)
	}

	if backTextTag, hasCompound, err := tag.GetCompoundTag(SignTagBackText); err == nil && hasCompound {
		s.BackText = signReadTextTag(backTextTag, true)
	} else {
		s.BackText = blockutils.NewSignText(nil, nil, false)
	}

	s.Waxed = tag.GetByteOr(SignTagWaxed, 0) != 0
	return nil
}

func (s *Sign) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetTag(SignTagFrontText, signWriteTextTag(s.Text))
	tag.SetTag(SignTagBackText, signWriteTextTag(s.BackText))
	waxed := nbt.ByteTag(0)
	if s.Waxed {
		waxed = 1
	}
	tag.SetByte(SignTagWaxed, waxed)
}

func (s *Sign) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetTag(SignTagFrontText, signWriteTextTag(s.Text))
	tag.SetTag(SignTagBackText, signWriteTextTag(s.BackText))
	waxed := nbt.ByteTag(0)
	if s.Waxed {
		waxed = 1
	}
	tag.SetByte(SignTagWaxed, waxed)
	editorID := int64(-1)
	if s.HasEditorEntityRuntime {
		editorID = s.EditorEntityRuntimeID
	}
	tag.SetLong(SignTagLockedForEditing, nbt.LongTag(editorID))
}

func (s *Sign) GetText() blockutils.SignText { return s.Text }

func (s *Sign) SetText(text blockutils.SignText) { s.Text = text }

func (s *Sign) GetBackText() blockutils.SignText { return s.BackText }

func (s *Sign) SetBackText(text blockutils.SignText) { s.BackText = text }

func (s *Sign) IsWaxed() bool { return s.Waxed }

func (s *Sign) SetWaxed(waxed bool) { s.Waxed = waxed }

// GetEditorEntityRuntimeID returns the entity runtime ID of the player who placed this sign (only
// that player may edit the sign text - see the PHP original's doc comment for the full
// reasoning), and false if no editor is currently tracked.
func (s *Sign) GetEditorEntityRuntimeID() (int64, bool) {
	return s.EditorEntityRuntimeID, s.HasEditorEntityRuntime
}

func (s *Sign) SetEditorEntityRuntimeID(id int64, has bool) {
	s.EditorEntityRuntimeID = id
	s.HasEditorEntityRuntime = has
}
