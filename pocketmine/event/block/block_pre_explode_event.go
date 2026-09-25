package block

import "pocketmine-go/pocketmine/event"

// DefaultFireChance is Explosion::DEFAULT_FIRE_CHANCE.
const DefaultFireChance = 1.0 / 3.0

// BlockPreExplodeEvent is a port of pocketmine\event\block\BlockPreExplodeEvent: called when a
// block wants to explode, before the explosion impact is calculated.
type BlockPreExplodeEvent struct {
	BlockEvent
	event.CancellableTrait

	blockBreaking bool
	radius        float64
	player        Player
	fireChance    float64
}

// NewBlockPreExplodeEvent creates the event; player (who caused the explosion) may be nil.
func NewBlockPreExplodeEvent(block Block, radius float64, player Player, fireChance float64) *BlockPreExplodeEvent {
	checkRadius(radius)
	checkFireChance(fireChance)
	return &BlockPreExplodeEvent{BlockEvent: BlockEvent{block: block}, blockBreaking: true, radius: radius, player: player, fireChance: fireChance}
}

func checkRadius(radius float64) {
	checkFinite("radius", radius)
	if radius <= 0 {
		panic("Explosion radius must be positive")
	}
}

func checkFireChance(fireChance float64) {
	checkFinite("fireChance", fireChance)
	if fireChance < 0.0 || fireChance > 1.0 {
		panic("Fire chance must be a number between 0 and 1.")
	}
}

func (e *BlockPreExplodeEvent) GetRadius() float64 { return e.radius }

func (e *BlockPreExplodeEvent) SetRadius(radius float64) {
	checkRadius(radius)
	e.radius = radius
}

func (e *BlockPreExplodeEvent) IsBlockBreaking() bool { return e.blockBreaking }

func (e *BlockPreExplodeEvent) SetBlockBreaking(affectsBlocks bool) { e.blockBreaking = affectsBlocks }

// IsIncendiary returns whether the explosion will create a fire.
func (e *BlockPreExplodeEvent) IsIncendiary() bool { return e.fireChance > 0 }

// SetIncendiary sets whether the explosion will create a fire by filling fireChance with the
// default value.
func (e *BlockPreExplodeEvent) SetIncendiary(incendiary bool) {
	if !incendiary {
		e.fireChance = 0
	} else if e.fireChance <= 0 {
		e.fireChance = DefaultFireChance
	}
}

// GetFireChance returns a chance between 0 and 1 of creating a fire.
func (e *BlockPreExplodeEvent) GetFireChance() float64 { return e.fireChance }

// SetFireChance sets a chance between 0 and 1 of creating a fire. For example, if the chance is
// 1/3, then that amount of affected blocks will be ignited.
func (e *BlockPreExplodeEvent) SetFireChance(fireChance float64) {
	checkFireChance(fireChance)
	e.fireChance = fireChance
}

// GetPlayer returns the player who triggered the block explosion, or nil.
func (e *BlockPreExplodeEvent) GetPlayer() Player { return e.player }
