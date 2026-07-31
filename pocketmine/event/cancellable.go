package event

// Cancellable is a port of pocketmine\event\Cancellable: implemented by an event type if and
// only if it can be cancelled. A cancelled event is still passed to downstream handlers that
// registered with handleCancelled=true; others are skipped.
type Cancellable interface {
	IsCancelled() bool
}

// CancellableTrait is a port of pocketmine\event\CancellableTrait: an embeddable struct giving a
// concrete event type Cancellable for free. Embed it by value and it just works, since Go method
// promotion (unlike the SimpleLogger/PrefixedLogger case elsewhere in this port) needs no
// self-dispatch here — Cancel/Uncancel/IsCancelled don't call back into other overridable methods.
type CancellableTrait struct {
	cancelled bool
}

func (c *CancellableTrait) IsCancelled() bool { return c.cancelled }
func (c *CancellableTrait) Cancel()           { c.cancelled = true }
func (c *CancellableTrait) Uncancel()         { c.cancelled = false }
