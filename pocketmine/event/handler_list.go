package event

import "sync"

// handlerList is a port of the essential parts of pocketmine\event\HandlerList.
//
// PHP's version also maintains a parentList link so that registering for a supertype (e.g.
// EntityDamageEvent) also receives every subclass (EntityDamageByEntityEvent, etc.), walked via
// runtime reflection over the class hierarchy (getParentClass()). Go has no class inheritance and
// no equivalent reflection over "parent type", and no concrete event types exist yet in this port
// to design that mechanism against — so this is intentionally per-concrete-Go-type only for now.
// When real event families with a shared supertype get ported, this will need an explicit
// "declares its supertypes" hook (events opting in, since Go can't discover it automatically).
type handlerList struct {
	mu     sync.Mutex
	slots  map[Priority][]*registeredListener
	cached []*registeredListener
}

func newHandlerList() *handlerList {
	return &handlerList{slots: map[Priority][]*registeredListener{}}
}

func (h *handlerList) register(l *registeredListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slots[l.priority] = append(h.slots[l.priority], l)
	h.cached = nil
}

func (h *handlerList) unregisterByID(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for p, list := range h.slots {
		for i, l := range list {
			if l.id == id {
				h.slots[p] = append(list[:i:i], list[i+1:]...)
				h.cached = nil
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
		h.cached = nil
	}
}

func (h *handlerList) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slots = map[Priority][]*registeredListener{}
	h.cached = nil
}

// listeners returns the merged, priority-ordered listener list (Lowest called first, Monitor last).
func (h *handlerList) listeners() []*registeredListener {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cached != nil {
		return h.cached
	}
	result := make([]*registeredListener, 0)
	for _, p := range AllPriorities {
		result = append(result, h.slots[p]...)
	}
	h.cached = result
	return result
}
