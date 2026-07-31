package event

import (
	"fmt"
	"reflect"
)

const maxEventCallDepth = 50

// eventCallDepth guards against runaway recursive event calls (e.g. a handler that re-fires the
// same event unconditionally).
//
// PHP's version is a class-static counter, safe because a single PocketMine process's game logic
// (including all event dispatch) runs on one thread. This port keeps the same assumption — event
// dispatch is expected to happen from a single "main" goroutine, matching the rest of the engine
// state (world/entity data etc.) that isn't safe for concurrent access either — so a plain int
// is used rather than an atomic counter. If that assumption ever changes, this needs to become
// goroutine-local (e.g. threaded through a context), not just atomic, since the limit is meant to
// catch one call chain recursing into itself, not the total number of concurrent dispatches.
var eventCallDepth int

// ListenerHandle identifies a single registered listener for later Unregister.
//
// PHP's HandlerList::unregister() also accepts a Plugin or Listener object to bulk-remove every
// listener belonging to it; see Manager.UnregisterAllForPlugin for that use case here instead.
type ListenerHandle struct {
	manager *Manager
	typ     reflect.Type
	id      int
}

func (h ListenerHandle) Unregister() {
	h.manager.listFor(h.typ).unregisterByID(h.id)
}

// RegisterListener registers handler to be called whenever an *E event is dispatched via Call or
// CallOn(m, ...). plugin identifies the owner for bulk unregistration; handleCancelled matches
// the PHP @handleCancelled tag — if false, handler is skipped once the event is cancelled.
func RegisterListener[E any](m *Manager, plugin PluginRef, priority Priority, handleCancelled bool, handler func(e *E)) ListenerHandle {
	t := reflect.TypeOf((*E)(nil))
	id := m.nextListenerID()
	rl := &registeredListener{
		id:              id,
		handler:         func(e any) { handler(e.(*E)) },
		priority:        priority,
		plugin:          plugin,
		handleCancelled: handleCancelled,
	}
	m.listFor(t).register(rl)
	return ListenerHandle{manager: m, typ: t, id: id}
}

// Call dispatches e to every registered handler of *E on the global Manager, in priority order.
// Equivalent to PHP's $event->call().
func Call[E any](e *E) {
	CallOn(Global(), e)
}

// CallOn is Call against a specific Manager, e.g. a per-plugin-test Manager in unit tests.
func CallOn[E any](m *Manager, e *E) {
	if eventCallDepth >= maxEventCallDepth {
		// This mirrors PHP's \RuntimeException here: a bug (unconditional handler recursion), not
		// a normal runtime condition callers are expected to handle, so it panics rather than
		// returning an error every caller would have to check.
		panic(fmt.Sprintf("Recursive event call detected (reached max depth of %d calls)", maxEventCallDepth))
	}

	t := reflect.TypeOf(e)
	handlers := m.listFor(t).listeners()

	eventCallDepth++
	defer func() { eventCallDepth-- }()

	for _, h := range handlers {
		h.call(e)
	}
}

// HasHandlers returns whether *E has any registered handlers on the global Manager. Useful in hot
// code paths to skip constructing an event object nobody's listening for.
func HasHandlers[E any]() bool {
	return HasHandlersOn[E](Global())
}

func HasHandlersOn[E any](m *Manager) bool {
	t := reflect.TypeOf((*E)(nil))
	return len(m.listFor(t).listeners()) > 0
}
