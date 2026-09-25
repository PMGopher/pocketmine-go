package network

import "sync"

// Session is what NetworkSessionManager needs from pocketmine\network\mcpe\NetworkSession.
type Session interface {
	Tick()
	IsConnected() bool
	// Disconnect is NetworkSession::disconnect(reason, disconnectScreenMessage); messages are
	// strings or *lang.Translatable, and a nil screen message uses the reason.
	Disconnect(reason, disconnectScreenMessage any)
}

// NetworkSessionManager is a port of pocketmine\network\NetworkSessionManager. Sessions are added
// from connection goroutines, so it's safe for concurrent use.
type NetworkSessionManager struct {
	mu                   sync.Mutex
	sessions             []Session
	pendingLoginSessions map[Session]bool
}

func NewNetworkSessionManager() *NetworkSessionManager {
	return &NetworkSessionManager{pendingLoginSessions: map[Session]bool{}}
}

// Add adds a network session to the manager. This should only be called on session creation.
func (m *NetworkSessionManager) Add(session Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = append(m.sessions, session)
	m.pendingLoginSessions[session] = true
}

// MarkLoginReceived marks the session as having sent a login request. After this point, the
// session will count towards the total player count.
func (m *NetworkSessionManager) MarkLoginReceived(session Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pendingLoginSessions, session)
}

// Remove removes the given network session, due to disconnect. This should only be called by a
// network session on disconnection.
func (m *NetworkSessionManager) Remove(session Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeLocked(session)
}

func (m *NetworkSessionManager) removeLocked(session Session) {
	for i, s := range m.sessions {
		if s == session {
			m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
			break
		}
	}
	delete(m.pendingLoginSessions, session)
}

// GetSessionCount returns the number of known connected sessions, including sessions which have
// not yet sent a login request.
func (m *NetworkSessionManager) GetSessionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// GetValidSessionCount returns the number of connected sessions which have either sent a login
// request, or have already completed the login sequence.
func (m *NetworkSessionManager) GetValidSessionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions) - len(m.pendingLoginSessions)
}

// GetSessions returns the connected sessions.
func (m *NetworkSessionManager) GetSessions() []Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Session(nil), m.sessions...)
}

// Tick updates all sessions which need it, and drops disconnected ones.
func (m *NetworkSessionManager) Tick() {
	for _, session := range m.GetSessions() {
		session.Tick()
		if !session.IsConnected() {
			m.Remove(session)
		}
	}
}

// Close terminates all connected sessions with the given reason.
func (m *NetworkSessionManager) Close(reason, disconnectScreenMessage any) {
	for _, session := range m.GetSessions() {
		session.Disconnect(reason, disconnectScreenMessage)
	}
	m.mu.Lock()
	m.sessions = nil
	m.pendingLoginSessions = map[Session]bool{}
	m.mu.Unlock()
}
