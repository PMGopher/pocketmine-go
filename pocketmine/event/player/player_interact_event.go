package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

// PlayerInteractEvent actions, PlayerInteractEvent::LEFT_CLICK_BLOCK/RIGHT_CLICK_BLOCK.
const (
	InteractLeftClickBlock  = 0
	InteractRightClickBlock = 1
)

// PlayerInteractEvent is a port of pocketmine\event\player\PlayerInteractEvent: called when a
// player interacts or touches a block. This is called for both left click (start break) and
// right click (use).
type PlayerInteractEvent struct {
	PlayerEvent
	event.CancellableTrait

	item         Item
	blockTouched Block
	touchVector  math.Vector3
	blockFace    int
	action       int
	useItem      bool
	useBlock     bool
}

// NewPlayerInteractEvent creates the event; a nil touchVector is Vector3::zero().
func NewPlayerInteractEvent(player Player, item Item, blockTouched Block, touchVector *math.Vector3, blockFace int, action int) *PlayerInteractEvent {
	e := &PlayerInteractEvent{PlayerEvent: PlayerEvent{player: player}, item: item, blockTouched: blockTouched, blockFace: blockFace, action: action, useItem: true, useBlock: true}
	if touchVector != nil {
		e.touchVector = *touchVector
	}
	return e
}

func (e *PlayerInteractEvent) GetAction() int { return e.action }

func (e *PlayerInteractEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

func (e *PlayerInteractEvent) GetBlock() Block { return e.blockTouched }

func (e *PlayerInteractEvent) GetTouchVector() math.Vector3 { return e.touchVector }

func (e *PlayerInteractEvent) GetFace() int { return e.blockFace }

// UseItem returns whether the item may react to the interaction. If disabled, items such as
// spawn eggs will not activate. This does NOT prevent blocks from being placed - it makes the
// item behave as if the player is sneaking.
func (e *PlayerInteractEvent) UseItem() bool { return e.useItem }

// SetUseItem sets whether the used item may react to the interaction.
func (e *PlayerInteractEvent) SetUseItem(useItem bool) { e.useItem = useItem }

// UseBlock returns whether the block may react to the interaction. If disabled, doors, fence
// gates and trapdoors will not respond, containers will not open, etc.
func (e *PlayerInteractEvent) UseBlock() bool { return e.useBlock }

// SetUseBlock sets whether the block may react to the interaction.
func (e *PlayerInteractEvent) SetUseBlock(useBlock bool) { e.useBlock = useBlock }
