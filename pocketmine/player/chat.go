package player

import (
	"strings"
	"unicode/utf8"

	"pocketmine-go/pocketmine/player/chat"
	"pocketmine-go/pocketmine/utils"
)

// Chat length limits, Player::MAX_CHAT_CHAR_LENGTH/MAX_CHAT_BYTE_LENGTH.
const (
	MaxChatCharLength = 512
	MaxChatByteLength = MaxChatCharLength * 4 //max bytes per character in UTF-8
)

// Chat is a port of Player::chat: sends a chat message (or runs a command, for lines starting with
// "/") as this player. At most 2 messages per tick are accepted (messageCounter).
//
// Not ported: removeCurrentWindow (no windows yet), the Timings and the PlayerChatEvent (player
// events aren't ported), so every message is broadcast with the StandardChatFormatter to every
// online player (Server::BROADCAST_CHANNEL_USERS subscribers).
func (p *Player) Chat(message string) bool {
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
			if p.server != nil {
				p.server.DispatchCommand(p, messagePart[1:])
			}
		} else if p.server != nil {
			formatter := chat.StandardChatFormatter{}
			p.server.BroadcastMessage(formatter.Format(p.GetDisplayName(), messagePart), nil)
		}
	}

	return true
}

// SendMessage is a port of Player::sendMessage: message is a string or *lang.Translatable.
func (p *Player) SendMessage(message any) {
	if p.networkSession != nil {
		p.networkSession.OnChatMessage(message)
	}
}

// SelectHotbarSlot is a port of Player::selectHotbarSlot, minus the cancellable PlayerItemHeldEvent
// (player events aren't ported).
func (p *Player) SelectHotbarSlot(hotbarSlot int) bool {
	inventory := p.GetInventory()
	if !inventory.IsHotbarSlot(hotbarSlot) { //TODO: exception here?
		return false
	}
	if hotbarSlot == inventory.GetHeldItemIndex() {
		return true
	}

	inventory.SetHeldItemIndex(hotbarSlot)
	p.SetUsingItem(false)

	return true
}

// IsUsingItem is a port of Player::isUsingItem.
func (p *Player) IsUsingItem() bool { return p.startAction > -1 }

// SetUsingItem is a port of Player::setUsingItem.
func (p *Player) SetUsingItem(value bool) {
	p.startAction = -1
	if value && p.server != nil {
		p.startAction = p.server.GetTick()
	}
	p.MarkNetworkPropertiesDirty()
}

// GetItemUseDuration is a port of Player::getItemUseDuration: ticks since item use started, or -1.
func (p *Player) GetItemUseDuration() int64 {
	if p.startAction == -1 || p.server == nil {
		return -1
	}
	return p.server.GetTick() - p.startAction
}
