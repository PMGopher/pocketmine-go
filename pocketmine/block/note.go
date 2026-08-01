package block

import "pocketmine-go/pocketmine/block/tile"

const (
	NoteMinPitch = 0
	NoteMaxPitch = 24
)

// Note is a port of pocketmine\block\Note.
//
// Deprecated in the PHP original too.
//
// WriteStateToWorld's tile sync (writing Pitch back to the tile.Note on placement) is skipped:
// there's no WriteStateToWorld hook in the Behavior interface yet - same gap already documented
// on RedstoneComparator for its tile-backed signal strength.
type Note struct {
	Opaque

	Pitch int
}

func NewNote(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Note {
	n := &Note{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, Pitch: NoteMinPitch}
	n.Init(n)
	return n
}

func (n *Note) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *Note) ReadStateFromWorld() Behavior {
	n.Block.ReadStateFromWorld()

	world, err := n.position.GetWorld()
	if err != nil {
		n.Pitch = NoteMinPitch
		return n.self
	}
	t, _ := world.GetTile(n.position)
	if noteTile, ok := t.(*tile.Note); ok {
		n.Pitch = noteTile.GetPitch()
	} else {
		n.Pitch = NoteMinPitch
	}
	return n.self
}

func (n *Note) GetFuelTime() int { return 300 }

func (n *Note) GetPitch() int { return n.Pitch }

// SetPitch panics if out of range, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (n *Note) SetPitch(pitch int) {
	if pitch < NoteMinPitch || pitch > NoteMaxPitch {
		panic("Pitch must be in range 0 - 24")
	}
	n.Pitch = pitch
}
