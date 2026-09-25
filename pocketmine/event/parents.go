package event

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// This file ports the event-inheritance half of pocketmine\event\HandlerListManager: in PHP, a
// handler registered for an event class also receives every subclass of it
// (HandlerListManager::resolveNearestHandleableParent + HandlerList::$parentList). Go has no class
// inheritance, so concrete event types declare their nearest handleable parent explicitly with
// DeclareParent, in the same init() that defines them. The child struct must embed the parent
// struct (directly or through other embedded structs), which is how the ported events model PHP's
// `extends` anyway; a parent handler then receives a pointer to that embedded parent value, so
// changes it makes (cancelling, setting damage modifiers, ...) are seen by the child too.

type parentLink struct {
	parent reflect.Type // pointer type of the parent event
	index  []int        // field index path from the child struct to the embedded parent struct
}

var (
	parentsMu sync.RWMutex
	parents   = map[reflect.Type]parentLink{}
	// hierarchyVersion changes whenever a parent is declared, so cached listener lists that
	// include parent handlers are rebuilt.
	hierarchyVersion int
)

// DeclareParent records that events of type *C are also dispatched to handlers of *P (PHP's
// `class C extends P` where P isn't abstract, or is abstract with the @allowHandle tag). It
// panics if C doesn't embed P, since that's a programming error.
func DeclareParent[C, P any]() {
	child := reflect.TypeOf((*C)(nil))
	parent := reflect.TypeOf((*P)(nil))
	index, ok := findEmbedded(child.Elem(), parent.Elem())
	if !ok {
		panic(fmt.Sprintf("event: %s does not embed %s", child.Elem(), parent.Elem()))
	}
	parentsMu.Lock()
	defer parentsMu.Unlock()
	parents[child] = parentLink{parent: parent, index: index}
	hierarchyVersion++
}

// findEmbedded returns the field index path of the struct type target embedded (anonymously, by
// value) in t, searching breadth-first like Go's own field promotion.
func findEmbedded(t, target reflect.Type) ([]int, bool) {
	type node struct {
		t     reflect.Type
		index []int
	}
	queue := []node{{t: t}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n.t.Kind() != reflect.Struct {
			continue
		}
		for i := 0; i < n.t.NumField(); i++ {
			f := n.t.Field(i)
			if !f.Anonymous || f.Type.Kind() != reflect.Struct {
				continue
			}
			index := append(append([]int(nil), n.index...), i)
			if f.Type == target {
				return index, true
			}
			queue = append(queue, node{t: f.Type, index: index})
		}
	}
	return nil, false
}

// parentOf returns the declared parent of the event pointer type t.
func parentOf(t reflect.Type) (parentLink, bool) {
	parentsMu.RLock()
	defer parentsMu.RUnlock()
	link, ok := parents[t]
	return link, ok
}

func currentHierarchyVersion() int {
	parentsMu.RLock()
	defer parentsMu.RUnlock()
	return hierarchyVersion
}

// ancestry is t followed by its chain of declared parents, each with the conversion from a *t
// value to that ancestor's pointer (nil for t itself).
type ancestor struct {
	t       reflect.Type
	convert func(e any) any
}

func ancestry(t reflect.Type) []ancestor {
	result := []ancestor{{t: t}}
	var path []int
	for current := t; ; {
		link, ok := parentOf(current)
		if !ok {
			return result
		}
		path = append(append([]int(nil), path...), link.index...)
		fieldPath := path
		result = append(result, ancestor{t: link.parent, convert: func(e any) any {
			// NewAt instead of Addr().Interface(): the embedded parent may be an unexported
			// field (reflect refuses to hand those out otherwise), and it's the same memory.
			f := reflect.ValueOf(e).Elem().FieldByIndex(fieldPath)
			return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Interface()
		}})
		current = link.parent
	}
}
