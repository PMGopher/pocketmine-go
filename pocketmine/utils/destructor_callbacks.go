package utils

// DestructorCallbacks is a port of pocketmine\utils\DestructorCallbackTrait.
//
// PHP's trait hooks __destruct(), running registered callbacks when the object is garbage
// collected. Go has no reliable deterministic destructor — runtime.SetFinalizer exists, but it
// fires at an unpredictable time (or never, if the program exits first) and is explicitly
// discouraged for anything but freeing non-memory resources as a last resort. The idiomatic Go
// replacement is explicit cleanup: the type embedding this must call Run() itself from its own
// Close()/Dispose() method, at the point it knows it's actually being torn down.
//
// The PHP original stores callbacks in an ObjectSet, keyed by spl_object_id so the same
// Closure object can't be registered twice. Go func values aren't comparable (they can't
// satisfy ObjectSet's `comparable` constraint), so identity-based dedup isn't available here;
// this just keeps an ordered slice, which matches how these callbacks are actually used in
// PocketMine (added once, never removed by reference).
type DestructorCallbacks struct {
	callbacks []func()
}

// Add registers a callback to run when Run is called.
func (d *DestructorCallbacks) Add(callback func()) {
	d.callbacks = append(d.callbacks, callback)
}

// Run invokes every registered callback, in registration order.
func (d *DestructorCallbacks) Run() {
	for _, cb := range d.callbacks {
		cb()
	}
}
