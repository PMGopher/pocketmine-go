package event

import (
	"reflect"
	"sync"
)

// Manager is a port of pocketmine\event\HandlerListManager (renamed since Go doesn't need the
// "HandlerList" prefix repeated — the package name already says that).
type Manager struct {
	mu     sync.Mutex
	lists  map[reflect.Type]*handlerList
	nextID int
}

func NewManager() *Manager {
	return &Manager{lists: map[reflect.Type]*handlerList{}}
}

var globalManager = NewManager()

// Global returns the process-wide Manager, mirroring HandlerListManager::global().
func Global() *Manager { return globalManager }

func (m *Manager) listFor(t reflect.Type) *handlerList {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.lists[t]
	if !ok {
		l = newHandlerList()
		m.lists[t] = l
	}
	return l
}

func (m *Manager) nextListenerID() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	return m.nextID
}

// UnregisterAllForPlugin unregisters every listener registered by plugin, across every event type.
func (m *Manager) UnregisterAllForPlugin(plugin PluginRef) {
	m.mu.Lock()
	lists := make([]*handlerList, 0, len(m.lists))
	for _, l := range m.lists {
		lists = append(lists, l)
	}
	m.mu.Unlock()

	for _, l := range lists {
		l.unregisterMatching(func(r *registeredListener) bool { return r.plugin == plugin })
	}
}

// UnregisterAll unregisters every listener for every event type, regardless of owner.
func (m *Manager) UnregisterAll() {
	m.mu.Lock()
	lists := make([]*handlerList, 0, len(m.lists))
	for _, l := range m.lists {
		lists = append(lists, l)
	}
	m.mu.Unlock()

	for _, l := range lists {
		l.clear()
	}
}
