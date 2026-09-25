package mcpe

import (
	"fmt"
	"time"

	"pocketmine-go/pocketmine/network"
)

// PacketRateLimiter is a port of pocketmine\network\mcpe\PacketRateLimiter: a token bucket that
// limits how many packets a client may send per tick on average, allowing short bursts.
type PacketRateLimiter struct {
	name            string
	averagePerTick  int
	maxBudget       int
	budget          int
	lastUpdate      time.Time
	updateFrequency time.Duration
}

// NewPacketRateLimiter is a port of PacketRateLimiter::__construct (update frequency: one tick).
func NewPacketRateLimiter(name string, averagePerTick, maxBufferTicks int) *PacketRateLimiter {
	return &PacketRateLimiter{
		name:            name,
		averagePerTick:  averagePerTick,
		maxBudget:       averagePerTick * maxBufferTicks,
		budget:          averagePerTick * maxBufferTicks,
		lastUpdate:      time.Now(),
		updateFrequency: 50 * time.Millisecond,
	}
}

// Decrement is a port of PacketRateLimiter::decrement: it returns a *network.PacketHandlingError
// once the budget is exhausted.
func (r *PacketRateLimiter) Decrement(amount int) error {
	if r.budget <= 0 {
		r.Update()
		if r.budget <= 0 {
			return &network.PacketHandlingError{Message: fmt.Sprintf("Exceeded rate limit for %q", r.name)}
		}
	}
	r.budget -= amount
	return nil
}

// Update is a port of PacketRateLimiter::update.
func (r *PacketRateLimiter) Update() {
	now := time.Now()
	since := now.Sub(r.lastUpdate)
	if since > r.updateFrequency {
		ticksSinceLastUpdate := int(since / r.updateFrequency)
		/*
		 * If the server takes an abnormally long time to process a tick, add the budget for time difference to
		 * compensate. This extra budget may be very large, but it will disappear the next time a normal update
		 * occurs. This ensures that backlogs during a large lag spike don't cause everyone to get kicked.
		 * As long as all the backlogged packets are processed before the next tick, everything should be OK for
		 * clients behaving normally.
		 */
		r.budget = min(r.budget, r.maxBudget) + (r.averagePerTick * 2 * ticksSinceLastUpdate)
		r.lastUpdate = now
	}
}
