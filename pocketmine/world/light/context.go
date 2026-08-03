// Package light is a port of pocketmine\world\light: the sky/block light propagation engine.
//
// Real PHP arrays only support int/string keys, so LightUpdate itself needs World::blockHash to
// pack an (x,y,z) position into a single int before it can use one as an array key. Go map keys
// have no such restriction - a plain [3]int compares by value and is directly usable as a map key,
// so this port uses that instead of porting blockHash at all; it's simpler, always collision-free,
// and behaviourally identical (blockHash's only job was working around a PHP limitation this
// language doesn't have).
package light

// pos is this package's position-as-map-key type - see this file's doc comment on why it replaces
// World::blockHash.
type pos struct{ X, Y, Z int }

// spreadFromTrue is the sentinel PropagationContext.SpreadVisited uses for "this node was lit
// directly (setAndUpdateLight/computeRemoveLight), not received via propagation from a specific
// neighbour" - a port of PHP's `$context->spreadVisited[$index] = true` (as opposed to
// `= Facing::opposite($side)`, an actual Facing constant 0-5). PHP's `$from === $side` comparison
// is always false when $from is the boolean true (strict comparison against an int), meaning no
// side gets excluded from propagation; -1 achieves the same "never equals a real Facing value"
// property here, since every real math.Facing constant is >= 0.
const spreadFromTrue = -1

// PropagationContext is a port of pocketmine\world\light\LightPropagationContext. The two queues
// are plain slices used FIFO (append to enqueue, take index 0 to dequeue) - PHP's \SplQueue
// doesn't need a dedicated port, a slice used this way behaves identically for this package's
// purposes.
type PropagationContext struct {
	SpreadQueue   []pos
	SpreadVisited map[pos]int

	RemovalQueue   []removalEntry
	RemovalVisited map[pos]bool
}

type removalEntry struct {
	pos
	OldLevel int
}

func NewPropagationContext() *PropagationContext {
	return &PropagationContext{
		SpreadVisited:  map[pos]int{},
		RemovalVisited: map[pos]bool{},
	}
}
