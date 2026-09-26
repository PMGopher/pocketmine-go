package blockinventory

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	inventoryevent "pocketmine-go/pocketmine/event/inventory"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/sound"
)

// CraftingManagerFunc is Server::getCraftingManager() for the furnace, brewing stand and campfire
// logic (PHP reaches it through $world->getServer()); the server sets it.
var CraftingManagerFunc func() *crafting.CraftingManager

// ToNetworkNbtFunc is TypeConverter's getItemTranslator()->toNetworkNbt(), for the campfire's
// spawn data (set by the network package).
var ToNetworkNbtFunc func(it item.Item) *nbt.CompoundTag

// tilePosition is the tile's position as a block.Position (PHP tiles hold a world Position).
func tilePosition(t tile.Tile) (block.Position, block.World, bool) {
	pos := t.GetPosition()
	tw, ok := pos.GetWorld()
	if !ok {
		return block.Position{}, nil, false
	}
	w, ok := tw.(block.World)
	if !ok {
		return block.Position{}, nil, false
	}
	return block.NewPosition(pos.X, pos.Y, pos.Z, w), w, true
}

// scheduleUpdateOnChange is the `$world->scheduleDelayedBlockUpdate($pos, 1)` listener the
// furnace and brewing stand tiles add to their inventory.
func scheduleUpdateOnChange(inv inventory.Inventory, w block.World, pos math.Vector3) {
	inv.GetListeners().Add(inventory.OnAnyChange(func(inventory.Inventory) {
		w.ScheduleDelayedBlockUpdate(pos, 1)
	}))
}

// newTileInventory is the inventory each container tile's constructor creates.
func newTileInventory(t tile.Tile) tile.Inventory {
	pos, w, ok := tilePosition(t)
	if !ok {
		return nil
	}
	switch tt := t.(type) {
	case *tile.Chest:
		return NewChestInventory(pos)
	case *tile.Barrel:
		return NewBarrelInventory(pos)
	case *tile.Hopper:
		return NewHopperInventory(pos, 5)
	case *tile.ShulkerBox:
		return NewShulkerBoxInventory(pos)
	case *tile.BrewingStand:
		inv := NewBrewingStandInventory(pos, 5)
		scheduleUpdateOnChange(inv, w, pos.Vector3)
		return inv
	case *tile.Campfire:
		inv := NewCampfireInventory(pos)
		inv.GetListeners().Add(inventory.OnAnyChange(func(inventory.Inventory) {
			if campfire, ok := w.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()).(*block.Campfire); ok {
				_ = w.SetBlock(pos, campfire)
			}
		}))
		return inv
	case *tile.ChiseledBookshelf:
		return inventory.NewSimpleInventory(blockutils.ChiseledBookshelfSlotCount)
	case interface{ GetFurnaceType() tile.FurnaceType }:
		inv := NewFurnaceInventory(pos, tt.GetFurnaceType())
		scheduleUpdateOnChange(inv, w, pos.Vector3)
		return inv
	}
	return nil
}

func init() {
	tile.NewInventoryFunc = newTileInventory
	tile.LoadInventoryItemsFunc = func(inv tile.Inventory, items []*nbt.CompoundTag, errorLogContext string) {
		i, ok := inv.(inventory.Inventory)
		if !ok {
			return
		}
		// ContainerTrait::loadItems: no events are fired by initialization.
		listeners := i.GetListeners().ToSlice()
		i.GetListeners().Remove(listeners...)
		newContents := map[int]item.Item{}
		for _, itemNBT := range items {
			slotID := int(itemNBT.GetByteOr("Slot", 0))
			newContents[slotID] = item.SafeNbtDeserialize(itemNBT, fmt.Sprintf("%s slot %d", errorLogContext, slotID), nil)
		}
		i.SetContents(newContents)
		i.GetListeners().Add(listeners...)
	}
	tile.SaveInventoryItemsFunc = func(inv tile.Inventory) []*nbt.CompoundTag {
		i, ok := inv.(inventory.Inventory)
		if !ok {
			return nil
		}
		contents := i.GetContents(false)
		var items []*nbt.CompoundTag
		for slot := 0; slot < i.GetSize(); slot++ {
			it, ok := contents[slot]
			if !ok {
				continue
			}
			if tag, err := item.NbtSerialize(it, slot); err == nil {
				items = append(items, tag)
			}
		}
		return items
	}
	tile.DropInventoryContentsFunc = func(inv tile.Inventory, world tile.World, pos math.Vector3) {
		i, ok := inv.(inventory.Inventory)
		if !ok {
			return
		}
		if dropper, ok := world.(interface {
			DropBlockItem(source math.Vector3, it block.Item)
		}); ok {
			contents := i.GetContents(false)
			for slot := 0; slot < i.GetSize(); slot++ {
				if it, ok := contents[slot]; ok {
					dropper.DropBlockItem(pos, it)
				}
			}
		}
		i.ClearAll()
	}
	tile.RemoveAllViewersFunc = func(inv tile.Inventory) {
		if i, ok := inv.(interface{ RemoveAllViewers() }); ok {
			i.RemoveAllViewers()
		}
	}
	tile.InventoryIsSlotEmptyFunc = func(inv tile.Inventory, slot int) bool {
		i, ok := inv.(inventory.Inventory)
		return !ok || i.IsSlotEmpty(slot)
	}
	tile.NewDoubleChestInventoryFunc = func(left, right tile.Inventory) tile.Inventory {
		l, okL := left.(*ChestInventory)
		r, okR := right.(*ChestInventory)
		if !okL || !okR {
			return nil
		}
		return NewDoubleChestInventory(l, r)
	}
	tile.CampfireSlotItemFunc = func(inv tile.Inventory, slot int) *nbt.CompoundTag {
		i, ok := inv.(inventory.Inventory)
		if !ok {
			return nil
		}
		it := i.GetItem(slot)
		if it.IsNull() {
			return nil
		}
		tag, err := item.NbtSerialize(it, -1)
		if err != nil {
			return nil
		}
		return tag
	}
	tile.CampfireSlotNetworkItemFunc = func(inv tile.Inventory, slot int) *nbt.CompoundTag {
		i, ok := inv.(inventory.Inventory)
		if !ok || ToNetworkNbtFunc == nil {
			return nil
		}
		it := i.GetItem(slot)
		if it.IsNull() {
			return nil
		}
		return ToNetworkNbtFunc(it)
	}
	block.ChiseledBookshelfInteractFunc = chiseledBookshelfInteract
	block.CampfireAddIngredientFunc = campfireAddIngredient
	block.CampfireCookFunc = campfireCook
	tile.FurnaceOnUpdateFunc = furnaceOnUpdate
	tile.BrewingStandOnUpdateFunc = brewingStandOnUpdate
}

// tileBlock adapts a tile to the event packages' Furnace/BrewingStand (Tile::getBlock).
type tileBlock struct{ t tile.Tile }

func (b tileBlock) GetBlock() blockevent.Block {
	pos, w, ok := tilePosition(b.t)
	if !ok {
		return nil
	}
	return w.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())
}

// syncData is InventoryManager::syncData for every viewer of inv.
func syncData(inv inventory.Inventory, property, value int) {
	for _, viewer := range inv.GetViewers() {
		nv, ok := viewer.(inventory.NetworkViewer)
		if !ok {
			continue
		}
		if syncer, ok := nv.GetInvManager().(interface {
			SyncData(inv inventory.Inventory, propertyID, value int)
		}); ok {
			syncer.SyncData(inv, property, value)
		}
	}
}

// setFurnaceLit is Furnace::onStartSmelting/onStopSmelting.
func setFurnaceLit(f *tile.Furnace, lit bool) {
	pos, w, ok := tilePosition(f)
	if !ok {
		return
	}
	if furnace, ok := w.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()).(*block.Furnace); ok && furnace.IsLit() != lit {
		furnace.SetLit(lit)
		_ = w.SetBlock(pos, furnace)
	}
}

// furnaceCheckFuel is a port of Furnace::checkFuel.
func furnaceCheckFuel(f *tile.Furnace, inv *FurnaceInventory, fuel item.Item) {
	ev := inventoryevent.NewFurnaceBurnEvent(tileBlock{f}, fuel, fuel.GetFuelTime())
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}
	f.RemainingFuelTime = ev.GetBurnTime()
	f.MaxFuelTime = f.RemainingFuelTime
	setFurnaceLit(f, true)

	if f.RemainingFuelTime > 0 && ev.IsBurning() {
		inv.SetFuel(fuel.GetFuelResidue())
	}
}

// furnaceOnUpdate is a port of Furnace::onUpdate.
func furnaceOnUpdate(f *tile.Furnace) bool {
	inv, ok := f.GetInventory().(*FurnaceInventory)
	if !ok || CraftingManagerFunc == nil {
		return false
	}
	prevCookTime := f.CookTime
	prevRemainingFuelTime := f.RemainingFuelTime
	prevMaxFuelTime := f.MaxFuelTime

	ret := false

	fuel := inv.GetFuel()
	raw := inv.GetSmelting()
	product := inv.GetResult()

	furnaceType := f.GetFurnaceType()
	smelt := CraftingManagerFunc().GetFurnaceRecipeManager(furnaceType).Match(raw)
	canSmelt := smelt != nil && raw.GetCount() > 0 &&
		((smelt.GetResult().CanStackWith(product) && product.GetCount() < product.GetMaxStackSize()) || product.IsNull())

	if f.RemainingFuelTime <= 0 && canSmelt && fuel.GetFuelTime() > 0 && fuel.GetCount() > 0 {
		furnaceCheckFuel(f, inv, fuel)
	}

	if f.RemainingFuelTime > 0 {
		f.RemainingFuelTime--

		if smelt != nil && canSmelt {
			f.CookTime++

			if f.CookTime >= furnaceType.GetCookDurationTicks() {
				product = smelt.GetResult()
				product.SetCount(inv.GetResult().GetCount() + 1)

				ev := inventoryevent.NewFurnaceSmeltEvent(tileBlock{f}, raw, product)
				event.Call(ev)

				if !ev.IsCancelled() {
					inv.SetResult(ev.GetResult().(item.Item))
					raw.Pop()
					inv.SetSmelting(raw)
				}
				f.CookTime -= furnaceType.GetCookDurationTicks()
			}
		} else if f.RemainingFuelTime <= 0 {
			f.RemainingFuelTime, f.CookTime, f.MaxFuelTime = 0, 0, 0
		} else {
			f.CookTime = 0
		}
		ret = true
	} else {
		setFurnaceLit(f, false)
		f.RemainingFuelTime, f.CookTime, f.MaxFuelTime = 0, 0, 0
	}

	if prevCookTime != f.CookTime {
		syncData(inv, packet.ContainerDataFurnaceTickCount, f.CookTime)
	}
	if prevRemainingFuelTime != f.RemainingFuelTime {
		syncData(inv, packet.ContainerDataFurnaceLitTime, f.RemainingFuelTime)
	}
	if prevMaxFuelTime != f.MaxFuelTime {
		syncData(inv, packet.ContainerDataFurnaceLitDuration, f.MaxFuelTime)
	}
	return ret
}

// brewingStandCheckFuel is a port of BrewingStand::checkFuel.
func brewingStandCheckFuel(b *tile.BrewingStand, inv *BrewingStandInventory, fuel item.Item) {
	ev := blockevent.NewBrewingFuelUseEvent(tileBlock{b})
	if !fuel.Equals(item.VanillaItem("blaze_powder"), false) {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}
	fuel.Pop()
	inv.SetItem(BrewingStandSlotFuel, fuel)
	b.RemainingFuelTime = ev.GetFuelTime()
	b.MaxFuelTime = b.RemainingFuelTime
}

// brewingStandRecipes is a port of BrewingStand::getBrewableRecipes: slot => recipe.
func brewingStandRecipes(inv *BrewingStandInventory) map[int]crafting.BrewingRecipe {
	ingredient := inv.GetItem(BrewingStandSlotIngredient)
	if ingredient.IsNull() {
		return nil
	}
	recipes := map[int]crafting.BrewingRecipe{}
	craftingManager := CraftingManagerFunc()
	for _, slot := range []int{BrewingStandSlotBottleLeft, BrewingStandSlotBottleMiddle, BrewingStandSlotBottleRight} {
		input := inv.GetItem(slot)
		if input.IsNull() {
			continue
		}
		if recipe := craftingManager.MatchBrewingRecipe(input, ingredient); recipe != nil {
			recipes[slot] = recipe
		}
	}
	return recipes
}

// brewingStandOnUpdate is a port of BrewingStand::onUpdate.
func brewingStandOnUpdate(b *tile.BrewingStand) bool {
	inv, ok := b.GetInventory().(*BrewingStandInventory)
	if !ok || CraftingManagerFunc == nil {
		return false
	}
	prevBrewTime := b.BrewTime
	prevRemainingFuelTime := b.RemainingFuelTime
	prevMaxFuelTime := b.MaxFuelTime

	ret := false

	fuel := inv.GetItem(BrewingStandSlotFuel)
	ingredient := inv.GetItem(BrewingStandSlotIngredient)

	recipes := brewingStandRecipes(inv)
	canBrew := len(recipes) != 0

	if b.RemainingFuelTime <= 0 && canBrew {
		brewingStandCheckFuel(b, inv, fuel)
	}

	if b.RemainingFuelTime > 0 {
		if canBrew {
			if b.BrewTime == 0 {
				b.BrewTime = tile.BrewingStandBrewTimeTicks
				b.RemainingFuelTime--
			}
			b.BrewTime--

			if b.BrewTime <= 0 {
				anythingBrewed := false
				for _, slot := range []int{BrewingStandSlotBottleLeft, BrewingStandSlotBottleMiddle, BrewingStandSlotBottleRight} {
					recipe, ok := recipes[slot]
					if !ok {
						continue
					}
					input := inv.GetItem(slot)
					output := recipe.GetResultFor(input)
					if output == nil {
						continue
					}
					ev := blockevent.NewBrewItemEvent(tileBlock{b}, slot, input, output, recipe)
					event.Call(ev)
					if ev.IsCancelled() {
						continue
					}
					inv.SetItem(slot, ev.GetResult().(item.Item))
					anythingBrewed = true
				}

				if anythingBrewed {
					if pos, w, ok := tilePosition(b); ok {
						w.AddSound(pos.Add(0.5, 0.5, 0.5), sound.PotionFinishBrewingSound{})
					}
				}
				ingredient.Pop()
				inv.SetItem(BrewingStandSlotIngredient, ingredient)

				b.BrewTime = 0
			} else {
				ret = true
			}
		} else {
			b.BrewTime = 0
		}
	} else {
		b.BrewTime, b.RemainingFuelTime, b.MaxFuelTime = 0, 0, 0
	}

	if prevBrewTime != b.BrewTime {
		syncData(inv, packet.ContainerDataBrewingStandBrewTime, b.BrewTime)
	}
	if prevRemainingFuelTime != b.RemainingFuelTime {
		syncData(inv, packet.ContainerDataBrewingStandFuelAmount, b.RemainingFuelTime)
	}
	if prevMaxFuelTime != b.MaxFuelTime {
		syncData(inv, packet.ContainerDataBrewingStandFuelTotal, b.MaxFuelTime)
	}
	return ret
}

// campfireAddIngredient is Campfire::onInteract's recipe branch (the caller pops the item).
func campfireAddIngredient(c *block.Campfire, heldItem block.Item) bool {
	it, ok := heldItem.(item.Item)
	inv, okInv := c.GetInventory().(inventory.Inventory)
	if !ok || !okInv || CraftingManagerFunc == nil {
		return false
	}
	if CraftingManagerFunc().GetFurnaceRecipeManager(c.GetFurnaceType()).Match(it) == nil {
		return false
	}
	ingredient := it.Clone()
	ingredient.SetCount(1)
	return len(inv.AddItem(ingredient)) == 0
}

// campfireCook is the cooking part of Campfire::onScheduledUpdate: whether there were items.
func campfireCook(c *block.Campfire) bool {
	inv, ok := c.GetInventory().(inventory.Inventory)
	if !ok || CraftingManagerFunc == nil {
		return false
	}
	pos := c.GetPosition()
	items := inv.GetContents(false)
	furnaceType := c.GetFurnaceType()
	maxCookDuration := furnaceType.GetCookDurationTicks()
	for slot := 0; slot < inv.GetSize(); slot++ {
		it, ok := items[slot]
		if !ok {
			continue
		}
		c.SetCookingTime(slot, min(maxCookDuration, c.GetCookingTime(slot)+10)) // Campfire::UPDATE_INTERVAL_TICKS
		if c.GetCookingTime(slot) < maxCookDuration {
			continue
		}
		var result item.Item = item.VanillaAir()
		if recipe := CraftingManagerFunc().GetFurnaceRecipeManager(furnaceType).Match(it); recipe != nil {
			result = recipe.GetResult()
		}
		ev := blockevent.NewCampfireCookEvent(c, slot, it, result)
		event.Call(ev)
		if ev.IsCancelled() {
			continue
		}
		inv.SetItem(slot, item.VanillaAir())
		c.SetCookingTime(slot, 0)
		if w, err := pos.GetWorld(); err == nil {
			if dropper, ok := w.(interface {
				DropBlockItem(source math.Vector3, it block.Item)
			}); ok {
				if cooked, ok := ev.GetResult().(block.Item); ok {
					dropper.DropBlockItem(pos.Add(0.5, 1, 0.5), cooked)
				}
			}
		}
	}
	return len(items) > 0
}

// chiseledBookshelfInteract is the inventory part of ChiseledBookshelf::onInteract.
func chiseledBookshelfInteract(inv tile.Inventory, slot int, held block.Item) (block.Item, bool, bool) {
	i, ok := inv.(inventory.Inventory)
	if !ok {
		return nil, false, false
	}
	if !i.IsSlotEmpty(slot) {
		returned := i.GetItem(slot)
		i.Clear(slot)
		return returned, false, true
	}
	it, ok := held.(item.Item)
	if !ok {
		return nil, false, false
	}
	switch it.GetTypeId() {
	//TODO: type tags like blocks would be better for this
	case item.WRITABLE_BOOK, item.WRITTEN_BOOK, item.BOOK, item.ENCHANTED_BOOK:
		i.SetItem(slot, it.PopCount(1))
		return nil, true, true
	}
	return nil, false, false
}
