package player

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// signEditorTarget is block.BaseSign as Player::openSignEditor needs it (every sign embeds it).
type signEditorTarget interface {
	block.Behavior
	SetEditorEntityRuntimeID(id int, has bool)
}

// OpenSignEditor is a port of Player::openSignEditor: the sign at position remembers this player as
// its editor (BaseSign::updateText only accepts text from them) and the client opens the editor.
func (p *Player) OpenSignEditor(position math.Vector3, frontFace bool) {
	w := p.GetWorld()
	sign, ok := w.GetBlock(position).(signEditorTarget)
	if !ok {
		panic(fmt.Sprintf("Block at %v is not a sign", position))
	}
	sign.SetEditorEntityRuntimeID(p.GetID(), true)
	_ = w.SetBlock(block.NewPosition(position.X, position.Y, position.Z, w), sign)
	p.GetNetworkSession().OnOpenSignEditor(position, frontFace)
}
