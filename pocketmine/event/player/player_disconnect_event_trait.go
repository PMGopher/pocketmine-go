package player

// PlayerDisconnectEventTrait is a port of pocketmine\event\player\PlayerDisconnectEventTrait:
// the disconnect reason (shown in the server log) and disconnection screen message of an event
// that disconnects a player. Messages are strings or *lang.Translatable.
type PlayerDisconnectEventTrait struct {
	disconnectReason        any
	disconnectScreenMessage any
}

// SetDisconnectReason sets the kick reason shown in the server log and on the disconnection
// screen.
func (t *PlayerDisconnectEventTrait) SetDisconnectReason(disconnectReason any) {
	t.disconnectReason = disconnectReason
}

// GetDisconnectReason returns the kick reason shown in the server log and on the disconnection
// screen. Note: this reason is also used in PlayerQuitEvent if the player is kicked.
func (t *PlayerDisconnectEventTrait) GetDisconnectReason() any { return t.disconnectReason }

// SetDisconnectScreenMessage sets the message shown on the player's disconnection screen. This
// can be as long as you like. If nil, the disconnect reason will be used as the disconnect screen
// message.
func (t *PlayerDisconnectEventTrait) SetDisconnectScreenMessage(disconnectScreenMessage any) {
	t.disconnectScreenMessage = disconnectScreenMessage
}

// GetDisconnectScreenMessage returns the message shown on the player's disconnection screen
// (the disconnect reason if none was set).
func (t *PlayerDisconnectEventTrait) GetDisconnectScreenMessage() any {
	if t.disconnectScreenMessage != nil {
		return t.disconnectScreenMessage
	}
	return t.disconnectReason
}
