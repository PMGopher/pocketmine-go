package world

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// signAuthor is the player writing on a sign; only its ID is used.
type signAuthor struct {
	block.Player
	id int
}

func (a signAuthor) GetID() int { return a.id }

type writableSign interface {
	block.Behavior
	SetEditorEntityRuntimeID(id int, has bool)
	UpdateFaceText(author block.Player, authorName string, frontFace bool, text blockutils.SignText) (bool, error)
}

// Writing on a sign (BaseSign::updateFaceText, what InGamePacketHandler calls) must reach the
// players: the block update carries the sign tile with the new text.
func TestSignTextIsSentAfterWriting(t *testing.T) {
	for _, name := range []string{"oak_sign", "oak_wall_hanging_sign"} {
		w := newTestWorld()
		w.GetOrLoadChunk(0, 0)
		pos := block.NewPosition(3, 70, 3, w)
		if err := w.SetBlock(pos, block.VanillaBlock(name)); err != nil {
			t.Fatal(err)
		}
		sign := w.GetBlockAt(3, 70, 3).(writableSign)
		sign.SetEditorEntityRuntimeID(7, true) // Player::openSignEditor
		_ = w.SetBlock(pos, sign)

		sign = w.GetBlockAt(3, 70, 3).(writableSign)
		updated, err := sign.UpdateFaceText(signAuthor{id: 7}, "Steve", true, blockutils.SignTextFromBlob("hello\nworld", nil, false))
		if err != nil || !updated {
			t.Fatalf("%s: UpdateFaceText = %v, %v", name, updated, err)
		}
		found := false
		for _, pk := range w.CreateBlockUpdatePackets([]math.Vector3{pos.Vector3}) {
			if d, ok := pk.(*packet.BlockActorData); ok {
				front, _ := d.NBTData["FrontText"].(map[string]any)
				t.Logf("%s: %v", name, d.NBTData)
				found = front["Text"] == "hello\nworld"
				// new SignText(): opaque black (ARGB 0xff000000). Alpha 0 made the text invisible.
				if front["SignTextColor"] != int32(-16777216) {
					t.Errorf("%s: text colour = %v, want opaque black (-16777216)", name, front["SignTextColor"])
				}
			}
		}
		if !found {
			t.Errorf("%s: the block update doesn't carry the new text", name)
		}
	}
}
