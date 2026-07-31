package event

// registeredListener is a port of pocketmine\event\RegisteredListener, minus the TimingsHandler
// integration (the timings package doesn't exist yet in this port — hook it in here once it does,
// the same way PHP wraps handler() in startTiming()/stopTiming()).
type registeredListener struct {
	id              int
	handler         func(event any)
	priority        Priority
	plugin          PluginRef
	handleCancelled bool
}

func (r *registeredListener) call(e any) {
	if c, ok := e.(Cancellable); ok && c.IsCancelled() && !r.handleCancelled {
		return
	}
	r.handler(e)
}
