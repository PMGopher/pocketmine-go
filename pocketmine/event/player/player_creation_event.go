package player

// PlayerCreationEvent is a port of pocketmine\event\player\PlayerCreationEvent: allows the use of
// custom Player classes.
//
// Go can't instantiate a type chosen at runtime by name, so the base/player class setters
// (setBaseClass/setPlayerClass) have no counterpart: every player is a *player.Player. The event
// is still fired, so handlers can inspect the connecting session.
type PlayerCreationEvent struct {
	session NetworkSession
}

func NewPlayerCreationEvent(session NetworkSession) *PlayerCreationEvent {
	return &PlayerCreationEvent{session: session}
}

func (e *PlayerCreationEvent) GetNetworkSession() NetworkSession { return e.session }

func (e *PlayerCreationEvent) GetAddress() string { return e.session.GetIp() }

func (e *PlayerCreationEvent) GetPort() int { return e.session.GetPort() }
