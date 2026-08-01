package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Note pitch range mirrors pocketmine\block\Note::MIN_PITCH/MAX_PITCH - duplicated here rather
// than imported (the tile package can't import block, see tile.go's package doc comment).
const (
	NoteMinPitch = 0
	NoteMaxPitch = 24
)

// Note is a port of pocketmine\block\tile\Note.
//
// Deprecated in the PHP original too.
type Note struct {
	TileBase

	Pitch int
}

func NewNote(world World, pos math.Vector3) *Note {
	n := &Note{TileBase: NewTileBase(world, pos)}
	n.Init(n)
	return n
}

func (n *Note) SaveID() string { return "Music" }

func (n *Note) ReadSaveData(tag *nbt.CompoundTag) error {
	pitch := int(tag.GetByteOr("note", nbt.ByteTag(n.Pitch)))
	if pitch > NoteMinPitch && pitch <= NoteMaxPitch {
		n.Pitch = pitch
	}
	return nil
}

func (n *Note) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetByte("note", nbt.ByteTag(n.Pitch))
}

func (n *Note) GetPitch() int { return n.Pitch }

// SetPitch panics if out of range, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (n *Note) SetPitch(pitch int) {
	if pitch < NoteMinPitch || pitch > NoteMaxPitch {
		panic("Pitch must be in range 0 - 24")
	}
	n.Pitch = pitch
}
