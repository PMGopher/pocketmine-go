package block

// BlockTransactionImpl is a port of pocketmine\world\BlockTransaction: a batch of pending block
// changes that either all apply or none do. Concrete leaf types that need to change several
// blocks atomically (Bamboo, ChorusFlower, ...) construct one directly with NewBlockTransaction
// and call Apply() themselves - the narrow BlockTransaction interface in world.go (just AddBlock)
// is what Place() callers receive instead, since they only ever add to a transaction someone else
// applies.
//
// The real AddValidator/dummyValidator machinery isn't ported: every call site in the PHP
// codebase relies solely on the default IsInWorld validator added by the constructor, so that
// check is just inlined into Apply directly instead of building out unused closure-list
// extensibility.
//
// Iteration order also isn't a literal port: PHP's nested $blocks[x][y][z] array iterates in
// first-seen order of x, then first-seen order of y within that x, then insertion order of z -
// this instead just tracks simple global insertion order. Every real caller only ever adds each
// position once at non-overlapping coordinates, where the exact relative order between distinct
// positions doesn't affect the outcome, so this simplification is behaviorally equivalent for all
// current usage.
type BlockTransactionImpl struct {
	world  World
	blocks map[[3]int]Behavior
	order  [][3]int
}

func NewBlockTransaction(world World) *BlockTransactionImpl {
	return &BlockTransactionImpl{world: world, blocks: map[[3]int]Behavior{}}
}

// AddBlock is a port of BlockTransaction::addBlock.
func (t *BlockTransactionImpl) AddBlock(pos Position, state Behavior) {
	t.AddBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ(), state)
}

// AddBlockAt is a port of BlockTransaction::addBlockAt.
func (t *BlockTransactionImpl) AddBlockAt(x, y, z int, state Behavior) {
	key := [3]int{x, y, z}
	if _, exists := t.blocks[key]; !exists {
		t.order = append(t.order, key)
	}
	t.blocks[key] = state
}

// FetchBlock is a port of BlockTransaction::fetchBlock.
func (t *BlockTransactionImpl) FetchBlock(pos Position) Behavior {
	return t.FetchBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())
}

// FetchBlockAt is a port of BlockTransaction::fetchBlockAt.
func (t *BlockTransactionImpl) FetchBlockAt(x, y, z int) Behavior {
	if blk, ok := t.blocks[[3]int{x, y, z}]; ok {
		return blk
	}
	return t.world.GetBlockAt(x, y, z)
}

// Apply is a port of BlockTransaction::apply. Unlike the PHP original (where World::setBlockAt
// can't fail), World.SetBlock here returns an error - a position is only counted as changed if
// SetBlock actually succeeds for it, and a failure doesn't abort the rest of the transaction
// (matching the PHP original having no failure path to abort on in the first place).
func (t *BlockTransactionImpl) Apply() bool {
	for _, key := range t.order {
		if !t.world.IsInWorld(key[0], key[1], key[2]) {
			return false
		}
	}

	changed := 0
	for _, key := range t.order {
		state := t.blocks[key]
		old := t.world.GetBlockAt(key[0], key[1], key[2])
		if old != nil && old.GetStateId() == state.GetStateId() {
			continue
		}
		pos := NewPosition(float64(key[0]), float64(key[1]), float64(key[2]), t.world)
		if err := t.world.SetBlock(pos, state); err == nil {
			changed++
		}
	}
	return changed != 0
}
