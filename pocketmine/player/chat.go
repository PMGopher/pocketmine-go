package player

import (
	"strings"
	"unicode/utf8"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/player/chat"
	"pocketmine-go/pocketmine/timings"
	"pocketmine-go/pocketmine/utils"
)

// Chat is a port of Player::chat: sends a chat message as this player. If the message begins
// with a / (forward-slash) it will be treated as a command.
func (p *Player) Chat(message string) bool {
	p.RemoveCurrentWindow()

	if p.messageCounter <= 0 {
		//the check below would take care of this (0 * (maxlen + 1) = 0), but it's better be explicit
		return false
	}

	//Fast length check, to make sure we don't get hung trying to explode MBs of string ...
	maxTotalLength := p.messageCounter * (MaxChatByteLength + 1)
	if len(message) > maxTotalLength {
		return false
	}

	message = utils.Clean(message, false)
	for _, messagePart := range strings.SplitN(message, "\n", p.messageCounter+1) {
		if strings.TrimSpace(messagePart) == "" || len(messagePart) > MaxChatByteLength || utf8.RuneCountInString(messagePart) > MaxChatCharLength {
			continue
		}
		counter := p.messageCounter
		p.messageCounter--
		if counter <= 0 {
			continue
		}
		if strings.HasPrefix(messagePart, "./") {
			messagePart = messagePart[1:]
		}

		if strings.HasPrefix(messagePart, "/") {
			timings.Init()
			timings.PlayerCommand.StartTiming()
			p.server.DispatchCommand(p, messagePart[1:], false)
			timings.PlayerCommand.StopTiming()
		} else {
			ev := playerevent.NewPlayerChatEvent(p, messagePart, p.server.GetBroadcastChannelSubscribers(BroadcastChannelUsers), chat.StandardChatFormatter{})
			event.Call(ev)
			if !ev.IsCancelled() {
				p.server.BroadcastMessage(ev.GetFormatter().Format(ev.GetPlayer().GetDisplayName(), ev.GetMessage()), ev.GetRecipients())
			}
		}
	}

	return true
}

var _ command.Sender = (*Player)(nil)
