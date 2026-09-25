package block

import "pocketmine-go/pocketmine/event"

// SignChangeEvent is a port of pocketmine\event\block\SignChangeEvent: called when a sign's text
// is changed by a player.
type SignChangeEvent struct {
	BlockEvent
	event.CancellableTrait

	sign      Block
	player    Player
	oldText   any
	text      any
	frontFace bool
}

// NewSignChangeEvent creates the event. sign is the block.BaseSign and the texts are
// blockutils.SignText values; oldText is the sign's current text on that face (PHP's constructor
// reads it with $sign->getFaceText($frontFace)).
func NewSignChangeEvent(sign Block, player Player, oldText, text any, frontFace bool) *SignChangeEvent {
	return &SignChangeEvent{BlockEvent: BlockEvent{block: sign}, sign: sign, player: player, oldText: oldText, text: text, frontFace: frontFace}
}

func (e *SignChangeEvent) GetSign() Block { return e.sign }

func (e *SignChangeEvent) GetPlayer() Player { return e.player }

// GetOldText returns the text currently on the sign.
func (e *SignChangeEvent) GetOldText() any { return e.oldText }

// GetNewText returns the text which will be on the sign after the event.
func (e *SignChangeEvent) GetNewText() any { return e.text }

// SetNewText sets the text to be written on the sign after the event.
func (e *SignChangeEvent) SetNewText(text any) { e.text = text }

func (e *SignChangeEvent) IsFrontFace() bool { return e.frontFace }
