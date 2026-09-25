package player

import "pocketmine-go/pocketmine/event"

// PlayerTransferEvent is a port of pocketmine\event\player\PlayerTransferEvent: called when a
// player attempts to be transferred to another server, e.g. by using /transferserver.
type PlayerTransferEvent struct {
	PlayerEvent
	event.CancellableTrait

	address string
	port    int
	message any
}

func NewPlayerTransferEvent(player Player, address string, port int, message any) *PlayerTransferEvent {
	return &PlayerTransferEvent{PlayerEvent: PlayerEvent{player: player}, address: address, port: port, message: message}
}

// GetAddress returns the destination server address. This could be an IP or a domain name.
func (e *PlayerTransferEvent) GetAddress() string { return e.address }

func (e *PlayerTransferEvent) SetAddress(address string) { e.address = address }

// GetPort returns the destination server port.
func (e *PlayerTransferEvent) GetPort() int { return e.port }

func (e *PlayerTransferEvent) SetPort(port int) { e.port = port }

// GetMessage returns the disconnect reason shown in the server log and on the console.
func (e *PlayerTransferEvent) GetMessage() any { return e.message }

func (e *PlayerTransferEvent) SetMessage(message any) { e.message = message }
