package player

import "pocketmine-go/pocketmine/event"

// PlayerBucketEvent is a port of pocketmine\event\player\PlayerBucketEvent (abstract, but tagged
// @allowHandle: its handlers receive both PlayerBucketFillEvent and PlayerBucketEmptyEvent).
type PlayerBucketEvent struct {
	PlayerEvent
	event.CancellableTrait

	blockClicked Block
	blockFace    int
	bucket       Item
	itemInHand   Item
}

func newPlayerBucketEvent(who Player, blockClicked Block, blockFace int, bucket, itemInHand Item) PlayerBucketEvent {
	return PlayerBucketEvent{PlayerEvent: PlayerEvent{player: who}, blockClicked: blockClicked, blockFace: blockFace, bucket: bucket, itemInHand: itemInHand}
}

// GetBucket returns the bucket used in this event.
func (e *PlayerBucketEvent) GetBucket() Item { return e.bucket }

// GetItem returns the item in hand after the event.
func (e *PlayerBucketEvent) GetItem() Item { return e.itemInHand }

func (e *PlayerBucketEvent) SetItem(item Item) { e.itemInHand = item }

func (e *PlayerBucketEvent) GetBlockClicked() Block { return e.blockClicked }

func (e *PlayerBucketEvent) GetBlockFace() int { return e.blockFace }

// PlayerBucketEmptyEvent is a port of pocketmine\event\player\PlayerBucketEmptyEvent.
type PlayerBucketEmptyEvent struct {
	PlayerBucketEvent
}

func NewPlayerBucketEmptyEvent(who Player, blockClicked Block, blockFace int, bucket, itemInHand Item) *PlayerBucketEmptyEvent {
	return &PlayerBucketEmptyEvent{PlayerBucketEvent: newPlayerBucketEvent(who, blockClicked, blockFace, bucket, itemInHand)}
}

// PlayerBucketFillEvent is a port of pocketmine\event\player\PlayerBucketFillEvent.
type PlayerBucketFillEvent struct {
	PlayerBucketEvent
}

func NewPlayerBucketFillEvent(who Player, blockClicked Block, blockFace int, bucket, itemInHand Item) *PlayerBucketFillEvent {
	return &PlayerBucketFillEvent{PlayerBucketEvent: newPlayerBucketEvent(who, blockClicked, blockFace, bucket, itemInHand)}
}

func init() {
	event.DeclareParent[PlayerBucketEmptyEvent, PlayerBucketEvent]()
	event.DeclareParent[PlayerBucketFillEvent, PlayerBucketEvent]()
}
