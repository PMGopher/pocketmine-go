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

	// cacheMu guards cache/version: the merged, priority-ordered handler lists including parent
	// event handlers (RegisteredListenerCache), dropped whenever any list changes.
	cacheMu sync.Mutex
	cache   map[reflect.Type][]boundListener
	// cacheHierarchy is the DeclareParent hierarchy version the cache was built against.
	cacheHierarchy int
}

func NewManager() *Manager {
	return &Manager{lists: map[reflect.Type]*handlerList{}, cache: map[reflect.Type][]boundListener{}}
}

var globalManager = NewManager()

// Global returns the process-wide Manager, mirroring HandlerListManager::global().
func Global() *Manager { return globalManager }

func (m *Manager) listFor(t reflect.Type) *handlerList {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.lists[t]
	if !ok {
		l = newHandlerList(m.invalidateCaches)
		m.lists[t] = l
	}
	return l
}

// invalidateCaches is HandlerList::invalidateAffectedCaches: any list change can affect the
// merged list of the event type itself and of every child type, so all of them are dropped.
func (m *Manager) invalidateCaches() {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	clear(m.cache)
}

// boundListener is a listener together with the conversion from the dispatched event to the type
// the listener was registered for (nil when they're the same type).
type boundListener struct {
	listener *registeredListener
	convert  func(e any) any
}

func (b boundListener) call(e any) {
	if b.convert != nil {
		e = b.convert(e)
	}
	b.listener.call(e)
}

// handlersFor is a port of HandlerListManager::getHandlersFor + HandlerList::getListenerList: the
// listeners for t and every declared parent of t, ordered by priority (Lowest first); within one
// priority, t's own listeners come before its parents'.
func (m *Manager) handlersFor(t reflect.Type) []boundListener {
	version := currentHierarchyVersion()
	m.cacheMu.Lock()
	if m.cacheHierarchy != version {
		clear(m.cache)
		m.cacheHierarchy = version
	}
	if cached, ok := m.cache[t]; ok {
		m.cacheMu.Unlock()
		return cached
	}
	m.cacheMu.Unlock()

	chain := ancestry(t)
	lists := make([]*handlerList, len(chain))
	for i, a := range chain {
		lists[i] = m.listFor(a.t)
	}
	result := make([]boundListener, 0)
	for _, p := range AllPriorities {
		for i, l := range lists {
			for _, rl := range l.slotsByPriority(p) {
				result = append(result, boundListener{listener: rl, convert: chain[i].convert})
			}
		}
	}

	m.cacheMu.Lock()
	m.cache[t] = result
	m.cacheMu.Unlock()
	return result
}

func (m *Manager) nextListenerID() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	return m.nextID
}

// UnregisterAllForPlugin unregisters every listener registered by plugin, across every event type.
func (m *Manager) UnregisterAllForPlugin(plugin PluginRef) {
	for _, l := range m.allLists() {
		l.unregisterMatching(func(r *registeredListener) bool { return r.plugin == plugin })
	}
}

// UnregisterAll unregisters every listener for every event type, regardless of owner.
func (m *Manager) UnregisterAll() {
	for _, l := range m.allLists() {
		l.clear()
	}
}

func (m *Manager) allLists() []*handlerList {
	m.mu.Lock()
	defer m.mu.Unlock()
	lists := make([]*handlerList, 0, len(m.lists))
	for _, l := range m.lists {
		lists = append(lists, l)
	}
	return lists
}
