package player

import (
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/player/chat"
)

// PlayerChatEvent is a port of pocketmine\event\player\PlayerChatEvent: called when a player
// chats something.
type PlayerChatEvent struct {
	PlayerEvent
	event.CancellableTrait

	message    string
	recipients []command.Sender
	formatter  chat.Formatter
}

func NewPlayerChatEvent(player Player, message string, recipients []command.Sender, formatter chat.Formatter) *PlayerChatEvent {
	return &PlayerChatEvent{PlayerEvent: PlayerEvent{player: player}, message: message, recipients: recipients, formatter: formatter}
}

func (e *PlayerChatEvent) GetMessage() string { return e.message }

func (e *PlayerChatEvent) SetMessage(message string) { e.message = message }

// SetPlayer changes the player that is sending the message.
func (e *PlayerChatEvent) SetPlayer(player Player) { e.player = player }

func (e *PlayerChatEvent) GetFormatter() chat.Formatter { return e.formatter }

func (e *PlayerChatEvent) SetFormatter(formatter chat.Formatter) { e.formatter = formatter }

func (e *PlayerChatEvent) GetRecipients() []command.Sender { return e.recipients }

func (e *PlayerChatEvent) SetRecipients(recipients []command.Sender) { e.recipients = recipients }
