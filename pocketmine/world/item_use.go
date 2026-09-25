package world

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// ItemUser is what World::useBreakOn/useItemOn need from pocketmine\player\Player.
type ItemUser interface {
	block.Player
	GetName() string
	GetDisplayName() string
	HasFiniteResources() bool
	IsCreative() bool
	IsSpectator() bool
	// IsAdventureLiteral is Player::isAdventure(true).
	IsAdventureLiteral() bool
	IsSneakPressed() bool
}

// UseBreakOnWith is a port of World::useBreakOn: tries to break the block at vector with item
// (nil means an empty hand) as player (nil for a non-player break), dropping the drops and XP.
// Items given back by the break (e.g. a bucket) are appended to returnedItems. Returns whether the
// block was broken; item is changed in place (durability).
func (w *World) UseBreakOnWith(vector math.Vector3, it item.Item, player ItemUser, createParticles bool, returnedItems *[]item.Item) bool {
	vector = vector.Floor()

	chunkX, chunkZ := vector.FloorX()>>4, vector.FloorZ()>>4
	if !w.IsChunkLoaded(chunkX, chunkZ) {
		return false
	}

	target := w.GetBlock(vector)
	affectedBlocks := target.GetAffectedBlocks()

	if it == nil {
		it = item.VanillaAir()
	}
	if returnedItems == nil {
		returnedItems = &[]item.Item{}
	}

	var drops []item.Item
	if player == nil || player.HasFiniteResources() {
		for _, b := range affectedBlocks {
			for _, drop := range b.GetDrops(it) {
				if d, ok := drop.(item.Item); ok {
					drops = append(drops, d)
				}
			}
		}
	}

	xpDrop := 0
	if player != nil && player.HasFiniteResources() {
		for _, b := range affectedBlocks {
			xpDrop += block.GetXpDropForTool(b, it)
		}
	}

	if player != nil {
		eventDrops := make([]blockevent.Item, len(drops))
		for i, d := range drops {
			eventDrops[i] = d
		}
		ev := blockevent.NewBlockBreakEvent(player, target, it, player.IsCreative(), eventDrops, xpDrop)

		if target.GetTypeId() == block.AIR || (player.IsSurvival() && !target.GetBreakInfo().IsBreakable()) || player.IsSpectator() {
			ev.Cancel()
		}

		if player.IsAdventureLiteral() && !ev.IsCancelled() {
			// LegacyStringToItemParser (resolving the CanDestroy names to blocks) isn't ported, so no
			// entry can match and adventure players can't break anything, as when none is set.
			ev.Cancel()
		}

		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}

		drops = drops[:0]
		for _, d := range ev.GetDrops() {
			if di, ok := d.(item.Item); ok {
				drops = append(drops, di)
			}
		}
		xpDrop = ev.GetXpDropAmount()
	} else if !target.GetBreakInfo().IsBreakable() {
		return false
	}

	for _, t := range affectedBlocks {
		w.destroyBlockInternal(t, it, player, createParticles, returnedItems)
	}

	it.OnDestroyBlock(target, returnedItems)

	if len(drops) > 0 {
		dropPos := vector.Add(0.5, 0.5, 0.5)
		for _, drop := range drops {
			if !drop.IsNull() {
				w.DropItem(dropPos, drop, nil, 10)
			}
		}
	}

	if xpDrop > 0 {
		w.DropExperience(vector.Add(0.5, 0.5, 0.5), xpDrop)
	}

	return true
}

// destroyBlockInternal is a port of World::destroyBlockInternal.
func (w *World) destroyBlockInternal(target block.Behavior, it item.Item, player ItemUser, createParticles bool, returnedItems *[]item.Item) {
	pos := target.GetPosition()
	if createParticles {
		w.AddParticle(pos.Add(0.5, 0.5, 0.5), particle.BlockBreakParticle{BlockStateID: target.GetStateId()})
	}

	var blockPlayer block.Player
	if player != nil {
		blockPlayer = player
	}
	var blockReturned []block.Item
	target.OnBreak(it, blockPlayer, &blockReturned)
	for _, r := range blockReturned {
		if ri, ok := r.(item.Item); ok {
			*returnedItems = append(*returnedItems, ri)
		}
	}

	if t, ok := w.GetTile(pos); ok {
		if destroyed, ok := t.(interface{ OnBlockDestroyed() }); ok {
			destroyed.OnBlockDestroyed()
		}
	}
}

// UseItemOn is a port of World::useItemOn: uses an item on a block at vector, clicking face at
// clickVector (nil means the block's origin). This is how blocks are placed and how blocks and
// items react to being clicked. player may be nil. Returns whether something happened; it is
// changed in place (placing pops one).
func (w *World) UseItemOn(vector math.Vector3, it item.Item, face math.Facing, clickVector *math.Vector3, player ItemUser, playSound bool, returnedItems *[]item.Item) bool {
	blockClicked := w.GetBlock(vector)
	blockReplace := blockClicked.(interface {
		GetSide(side math.Facing, step int) block.Behavior
	}).GetSide(face, 1)

	var click math.Vector3
	if clickVector != nil {
		click = math.NewVector3(
			stdmath.Min(1.0, stdmath.Max(0.0, clickVector.X)),
			stdmath.Min(1.0, stdmath.Max(0.0, clickVector.Y)),
			stdmath.Min(1.0, stdmath.Max(0.0, clickVector.Z)),
		)
	}
	if returnedItems == nil {
		returnedItems = &[]item.Item{}
	}

	replacePos := blockReplace.GetPosition()
	if !w.IsInWorld(replacePos.FloorX(), replacePos.FloorY(), replacePos.FloorZ()) {
		//TODO: build height limit messages for custom world heights and mcregion cap
		return false
	}
	if !w.IsChunkLoaded(replacePos.FloorX()>>4, replacePos.FloorZ()>>4) {
		return false
	}

	if blockClicked.GetTypeId() == block.AIR {
		return false
	}

	var blockPlayer block.Player
	if player != nil {
		blockPlayer = player
	}
	blockReturned := func(onInteract func(returned *[]block.Item) bool) bool {
		var returned []block.Item
		result := onInteract(&returned)
		for _, r := range returned {
			if ri, ok := r.(item.Item); ok {
				*returnedItems = append(*returnedItems, ri)
			}
		}
		return result
	}

	if player != nil {
		ev := playerevent.NewPlayerInteractEvent(player, it, blockClicked, &click, int(face), playerevent.InteractRightClickBlock)
		if player.IsSneakPressed() {
			ev.SetUseItem(false)
			ev.SetUseBlock(it.IsNull()) //opening doors is still possible when sneaking if using an empty hand
		}
		if player.IsSpectator() {
			ev.Cancel() //set it to cancelled so plugins can bypass this
		}

		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
		if ev.UseBlock() && blockReturned(func(r *[]block.Item) bool { return blockClicked.OnInteract(it, face, click, blockPlayer, r) }) {
			return true
		}
		if ev.UseItem() {
			if result := it.OnInteractBlock(player, blockReplace, blockClicked, face, click, returnedItems); result != item.ItemUseResultNone {
				return result == item.ItemUseResultSuccess
			}
		}
	} else if blockReturned(func(r *[]block.Item) bool { return blockClicked.OnInteract(it, face, click, nil, r) }) {
		return true
	}

	if it.IsNull() || !it.CanBePlaced() {
		return false
	}

	//TODO: while passing Facing::UP mimics the vanilla behaviour with replaceable blocks, we should really pass
	//some other value like NULL and let place() deal with it. This will look like a bug to anyone who doesn't know
	//about the vanilla behaviour.
	tx := it.GetPlacementTransaction(blockClicked, blockClicked, math.Up, click, blockPlayer)
	if tx == nil {
		tx = it.GetPlacementTransaction(blockReplace, blockClicked, face, click, blockPlayer)
	}
	if tx == nil {
		//no placement options available
		return false
	}

	for _, entry := range tx.GetBlocks() {
		if p, ok := entry.Block.(positionable); ok {
			p.SetPosition(w, entry.X, entry.Y, entry.Z)
		}
		for _, collisionBox := range entry.Block.(interface {
			GetCollisionBoxes() []math.AxisAlignedBB
		}).GetCollisionBoxes() {
			if len(w.GetCollidingEntities(collisionBox, nil)) > 0 {
				return false //Entity in block
			}
		}
	}

	if player != nil {
		ev := blockevent.NewBlockPlaceEvent(player, tx, blockClicked, it)
		if player.IsSpectator() {
			ev.Cancel()
		}
		if player.IsAdventureLiteral() && !ev.IsCancelled() {
			// See UseBreakOnWith: CanPlaceOn entries can't be resolved without LegacyStringToItemParser.
			ev.Cancel()
		}
		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
	}

	if !tx.Apply() {
		return false
	}
	first := true
	for _, entry := range tx.GetBlocks() {
		if t, ok := w.GetTileAt(entry.X, entry.Y, entry.Z); ok {
			//TODO: seal this up inside block placement
			if copier, ok := t.(interface{ CopyDataFromItem(item tile.Item) }); ok {
				copier.CopyDataFromItem(tileItem{it})
			}
		}

		placed := w.GetBlockAt(entry.X, entry.Y, entry.Z)
		placed.OnPostPlace()
		if first && playSound {
			w.AddSound(placed.GetPosition().Vector3, sound.BlockPlaceSound{BlockStateID: placed.GetStateId()})
		}
		first = false
	}

	it.Pop()

	return true
}

// tileItem adapts item.Item to tile.Item, whose GetCustomBlockData also reports presence.
type tileItem struct{ item.Item }

func (t tileItem) GetCustomBlockData() (*nbt.CompoundTag, bool) {
	tag := t.Item.GetCustomBlockData()
	return tag, tag != nil
}
