package event

import "sync"

// handlerList is a port of the essential parts of pocketmine\event\HandlerList: the listeners
// registered for exactly one event type. The parentList half (handlers of a parent event class
// also receiving subclasses) is resolved by Manager.handlersFor from the DeclareParent hierarchy
// (see parents.go).
type handlerList struct {
	mu    sync.Mutex
	slots map[Priority][]*registeredListener
	// changed is called whenever the list's contents change, so the Manager can drop cached
	// merged lists that include it (HandlerList::invalidateAffectedCaches).
	changed func()
}

func newHandlerList(changed func()) *handlerList {
	return &handlerList{slots: map[Priority][]*registeredListener{}, changed: changed}
}

// slotsByPriority returns a copy of the listeners registered at priority.
func (h *handlerList) slotsByPriority(p Priority) []*registeredListener {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*registeredListener(nil), h.slots[p]...)
}

func (h *handlerList) register(l *registeredListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slots[l.priority] = append(h.slots[l.priority], l)
	h.changed()
}

func (h *handlerList) unregisterByID(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for p, list := range h.slots {
		for i, l := range list {
			if l.id == id {
				h.slots[p] = append(list[:i:i], list[i+1:]...)
				h.changed()
				return
			}
		}
	}
}

// unregisterMatching removes every listener for which pred returns true.
func (h *handlerList) unregisterMatching(pred func(*registeredListener) bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	changed := false
	for p, list := range h.slots {
		kept := list[:0]
		for _, l := range list {
			if pred(l) {
				changed = true
				continue
			}
			kept = append(kept, l)
		}
		h.slots[p] = kept
	}
	if changed {
		h.changed()
	}
}

func (h *handlerList) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slots = map[Priority][]*registeredListener{}
	h.changed()
}
