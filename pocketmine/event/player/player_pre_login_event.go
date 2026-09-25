package player

// PlayerPreLoginEvent kick flags, PlayerPreLoginEvent::KICK_FLAG_*.
const (
	KickFlagPlugin            = 0
	KickFlagServerFull        = 1
	KickFlagServerWhitelisted = 2
	KickFlagBanned            = 3
)

// KickFlagPriority is PlayerPreLoginEvent::KICK_FLAG_PRIORITY: the plugin reason always takes
// priority over anything else.
var KickFlagPriority = []int{KickFlagPlugin, KickFlagServerFull, KickFlagServerWhitelisted, KickFlagBanned}

// PlayerPreLoginEvent is a port of pocketmine\event\player\PlayerPreLoginEvent: called before a
// player is created and spawned, after login and authentication. It can be used to disconnect
// players who don't meet some criteria (full server, whitelist, bans).
type PlayerPreLoginEvent struct {
	playerInfo   PlayerInfo
	ip           string
	port         int
	authRequired bool

	// flags keeps the kick flags in the order they were set (PHP array order).
	flags                    []int
	disconnectReasons        map[int]any
	disconnectScreenMessages map[int]any
}

func NewPlayerPreLoginEvent(playerInfo PlayerInfo, ip string, port int, authRequired bool) *PlayerPreLoginEvent {
	return &PlayerPreLoginEvent{playerInfo: playerInfo, ip: ip, port: port, authRequired: authRequired, disconnectReasons: map[int]any{}, disconnectScreenMessages: map[int]any{}}
}

// GetPlayerInfo returns an object containing self-proclaimed information about the connecting
// player. WARNING: THE PLAYER IS NOT VERIFIED DURING THIS EVENT.
func (e *PlayerPreLoginEvent) GetPlayerInfo() PlayerInfo { return e.playerInfo }

func (e *PlayerPreLoginEvent) GetIp() string { return e.ip }

func (e *PlayerPreLoginEvent) GetPort() int { return e.port }

func (e *PlayerPreLoginEvent) IsAuthRequired() bool { return e.authRequired }

func (e *PlayerPreLoginEvent) SetAuthRequired(v bool) { e.authRequired = v }

// GetKickFlags returns an array of kick flags currently assigned.
func (e *PlayerPreLoginEvent) GetKickFlags() []int { return append([]int(nil), e.flags...) }

// IsKickFlagSet returns whether the given kick flag has been assigned.
func (e *PlayerPreLoginEvent) IsKickFlagSet(flag int) bool {
	_, ok := e.disconnectReasons[flag]
	return ok
}

// SetKickFlag sets a reason to disallow the player to continue authenticating, with a
// message. This can also be used to change kick messages for already-set flags. A nil
// disconnectScreenMessage uses the disconnect reason.
func (e *PlayerPreLoginEvent) SetKickFlag(flag int, disconnectReason, disconnectScreenMessage any) {
	if !e.IsKickFlagSet(flag) {
		e.flags = append(e.flags, flag)
	}
	e.disconnectReasons[flag] = disconnectReason
	if disconnectScreenMessage == nil {
		disconnectScreenMessage = disconnectReason
	}
	e.disconnectScreenMessages[flag] = disconnectScreenMessage
}

// ClearKickFlag clears a specific kick flag if it was set. This allows fine-tuned kick
// control.
func (e *PlayerPreLoginEvent) ClearKickFlag(flag int) {
	if !e.IsKickFlagSet(flag) {
		return
	}
	delete(e.disconnectReasons, flag)
	delete(e.disconnectScreenMessages, flag)
	for i, f := range e.flags {
		if f == flag {
			e.flags = append(e.flags[:i], e.flags[i+1:]...)
			break
		}
	}
}

// ClearAllKickFlags clears all pre-assigned kick reasons, allowing the player to pass through.
func (e *PlayerPreLoginEvent) ClearAllKickFlags() {
	e.flags = nil
	e.disconnectReasons = map[int]any{}
	e.disconnectScreenMessages = map[int]any{}
}

// IsAllowed returns whether the player is allowed to continue logging in.
func (e *PlayerPreLoginEvent) IsAllowed() bool { return len(e.disconnectReasons) == 0 }

// GetDisconnectReason returns the disconnect reason provided for the given kick flag, or nil if
// not set. This is the message which will be shown in the server log and on the console.
func (e *PlayerPreLoginEvent) GetDisconnectReason(flag int) any { return e.disconnectReasons[flag] }

// GetDisconnectScreenMessage returns the disconnect screen message provided for the given kick
// flag, or nil if not set. This is the message shown to the player on the disconnect screen.
func (e *PlayerPreLoginEvent) GetDisconnectScreenMessage(flag int) any {
	return e.disconnectScreenMessages[flag]
}

// GetFinalDisconnectReason returns the reason for the player being disconnected: the one with
// the highest priority (see KickFlagPriority), or "" if none are set.
func (e *PlayerPreLoginEvent) GetFinalDisconnectReason() any {
	for _, p := range KickFlagPriority {
		if r, ok := e.disconnectReasons[p]; ok {
			return r
		}
	}
	return ""
}

// GetFinalDisconnectScreenMessage returns the disconnect screen message with the highest
// priority, or "" if none are set.
func (e *PlayerPreLoginEvent) GetFinalDisconnectScreenMessage() any {
	for _, p := range KickFlagPriority {
		if r, ok := e.disconnectScreenMessages[p]; ok {
			return r
		}
	}
	return ""
}
