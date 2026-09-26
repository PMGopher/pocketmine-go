package block

import (
	"fmt"
	"math/rand"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const campfireUpdateIntervalTicks = 10

// Shovel is a forward-compatible marker for pocketmine\item\Shovel - same pattern (and same
// caveat: it's structurally identical to Durable, so any Durable item satisfies it too) as the
// Axe interface in wood.go.
type Shovel interface {
	ApplyDamage(amount int) bool
}

// campfireMarker lets Campfire.Place check "is the block below any kind of campfire" the way
// PHP's `getSide(Facing::DOWN) instanceof Campfire` does. SoulCampfire IS-A Campfire in PHP, but
// Go's type assertions on structs don't follow embedding the same way instanceof follows class
// hierarchy - SoulCampfire gets isCampfire() for free via embedding Campfire, so no separate
// SoulCampfire-specific handling is needed here (unlike TrappedChest.OnPostPlace, which needed a
// same-EXACT-type check and so couldn't rely on this trick).
type campfireMarker interface{ isCampfire() }

func (c *Campfire) isCampfire() {}

// Campfire cooking hooks: the parts of Campfire that need furnace recipes (the crafting package)
// and the tile's item inventory, which this package can't import. block/inventory sets them.
var (
	// CampfireAddIngredientFunc is Campfire::onInteract's recipe branch: if item can be cooked,
	// one of it goes into the campfire's inventory (true).
	CampfireAddIngredientFunc func(c *Campfire, item Item) bool
	// CampfireCookFunc is Campfire::onScheduledUpdate's cooking of the items in the campfire's
	// inventory (cooking times, CampfireCookEvent, dropping the results). It returns whether the
	// campfire had items.
	CampfireCookFunc func(c *Campfire) bool
)

// Campfire is a port of pocketmine\block\Campfire.
type Campfire struct {
	Transparent
	HorizontalFacingComponent
	LightableComponent

	// inventory is the tile's CampfireInventory (see tile.Inventory); cookingTimes is slot =>
	// ticks.
	inventory    tile.Inventory
	cookingTimes map[int]int
}

func NewCampfire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Campfire {
	c := &Campfire{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		cookingTimes:              map[int]int{},
	}
	c.Init(c)
	return c
}

func (c *Campfire) Clone() Behavior {
	cl := *c
	cl.cookingTimes = make(map[int]int, len(c.cookingTimes))
	for k, v := range c.cookingTimes {
		cl.cookingTimes[k] = v
	}
	cl.rebind(&cl)
	return &cl
}

// ReadStateFromWorld is a port of Campfire::readStateFromWorld.
func (c *Campfire) ReadStateFromWorld() Behavior {
	c.Block.ReadStateFromWorld()
	c.inventory = nil
	if t, ok := c.tileAt(); ok {
		if campfireTile, ok := t.(*tile.Campfire); ok {
			c.inventory = campfireTile.GetInventory()
			c.cookingTimes = map[int]int{}
			for k, v := range campfireTile.GetCookingTimes() {
				c.cookingTimes[k] = v
			}
		}
	}
	return c.self
}

// WriteStateToWorld is a port of Campfire::writeStateToWorld.
func (c *Campfire) WriteStateToWorld() {
	c.Block.WriteStateToWorld()
	if t, ok := c.tileAt(); ok {
		if campfireTile, ok := t.(*tile.Campfire); ok {
			times := make(map[int]int, len(c.cookingTimes))
			for k, v := range c.cookingTimes {
				times[k] = v
			}
			campfireTile.SetCookingTimes(times)
		}
	}
}

// GetInventory is a port of Campfire::getInventory: the tile's CampfireInventory (nil if the block
// was never read from a world, like PHP's uninitialized property).
func (c *Campfire) GetInventory() tile.Inventory { return c.inventory }

// campfireFurnaceTyper lets Campfire reach SoulCampfire's getFurnaceType override.
type campfireFurnaceTyper interface{ GetFurnaceType() tile.FurnaceType }

// GetFurnaceType is a port of Campfire::getFurnaceType.
func (c *Campfire) GetFurnaceType() tile.FurnaceType { return tile.FurnaceTypeCampfire }

func (c *Campfire) furnaceType() tile.FurnaceType {
	return c.self.(campfireFurnaceTyper).GetFurnaceType()
}

// SetCookingTime is a port of Campfire::setCookingTime. Panics for an invalid slot or time.
func (c *Campfire) SetCookingTime(slot, time int) {
	if slot < 0 || slot > 3 {
		panic("Slot must be in range 0-3")
	}
	if max := c.furnaceType().GetCookDurationTicks(); time < 0 || time > max {
		panic(fmt.Sprintf("CookingTime must be in range 0-%d", max))
	}
	c.cookingTimes[slot] = time
}

// GetCookingTime is a port of Campfire::getCookingTime.
func (c *Campfire) GetCookingTime(slot int) int { return c.cookingTimes[slot] }

func (c *Campfire) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
	c.DescribeLit(w)
}

func (c *Campfire) HasEntityCollision() bool { return true }

func (c *Campfire) GetLightLevel() int {
	if c.Lit {
		return 15
	}
	return 0
}

func (c *Campfire) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool is a port of Campfire::getDropsForCompatibleTool.
func (c *Campfire) GetDropsForCompatibleTool(item Item) []Item {
	charcoal := vanillaItem("charcoal")
	if charcoal == nil {
		return nil
	}
	charcoal.SetCount(2)
	return []Item{charcoal}
}

func (c *Campfire) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (c *Campfire) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 9.0/16)}
}

// campfireEntityDamageShaper lets OnEntityInside reach a concrete leaf's (SoulCampfire's)
// GetEntityCollisionDamage override - same self-dispatch shape as fireShaper.
type campfireEntityDamageShaper interface {
	GetEntityCollisionDamage() int
}

// GetEntityCollisionDamage is a port of Campfire::getEntityCollisionDamage (SoulCampfire returns
// 2).
func (c *Campfire) GetEntityCollisionDamage() int { return 1 }

// Place is a port of Campfire::place.
func (c *Campfire) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if _, ok := c.self.(blockGeometry).GetSide(math.Down, 1).(campfireMarker); ok {
		return false
	}
	if player != nil {
		c.Facing = player.GetHorizontalFacing()
	}
	c.Lit = true
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of Campfire::onInteract.
func (c *Campfire) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if !c.Lit {
		if item.GetTypeId() == itemTypeIDsFireCharge {
			item.Pop()
			c.ignite()
			c.addSound(sound.BlazeShootSound{})
			return true
		}
		if item.GetTypeId() == itemTypeIDsFlintAndSteel || hasEnchantment(item, enchantment.VanillaFireAspect()) {
			if durable, ok := item.(Durable); ok {
				durable.ApplyDamage(1)
			}
			c.ignite()
			return true
		}
	} else if isShovel(item) {
		applyDamage(item, 1)
		c.extinguish()
		return true
	}

	if CampfireAddIngredientFunc != nil && CampfireAddIngredientFunc(c, item) {
		item.Pop()
		c.addSound(sound.ItemFrameAddItemSound{})
		return true
	}
	return false
}

// OnNearbyBlockChange is a port of Campfire::onNearbyBlockChange, minus waterlogging (marked
// //TODO in the PHP original too).
func (c *Campfire) OnNearbyBlockChange() {
	if c.Lit && c.self.(blockGeometry).GetSide(math.Up, 1).GetTypeId() == WATER {
		c.extinguish()
	}
}

// OnEntityInside is a port of Campfire::onEntityInside.
func (c *Campfire) OnEntityInside(e Entity) bool {
	if !c.Lit {
		if e.IsOnFire() {
			c.ignite()
			return false
		}
	} else if living, ok := e.(Living); ok {
		damage := c.self.(campfireEntityDamageShaper).GetEntityCollisionDamage()
		ev := entityevent.NewEntityDamageByBlockEvent(c.self, e, entityevent.CauseFire, float64(damage), nil)
		living.Attack(ev)
	}
	return true
}

// OnProjectileHit is a port of Campfire::onProjectileHit: water splash potions put it out.
func (c *Campfire) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	if potion, ok := projectile.(interface{ IsWaterPotion() bool }); ok && c.Lit && potion.IsWaterPotion() {
		c.extinguish()
	}
}

// OnScheduledUpdate is a port of Campfire::onScheduledUpdate.
func (c *Campfire) OnScheduledUpdate() {
	if !c.Lit {
		return
	}
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	if CampfireCookFunc != nil && CampfireCookFunc(c) {
		_ = world.SetBlock(c.position, c.self)
	}
	if rand.Intn(6) == 0 {
		world.AddSound(c.position.Vector3, furnaceCookSound(c.furnaceType()))
	}
	world.ScheduleDelayedBlockUpdate(c.position.Vector3, campfireUpdateIntervalTicks)
}

func (c *Campfire) addSound(s sound.Sound) {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(c.position.Vector3, s)
}

func (c *Campfire) extinguish() {
	c.addSound(sound.FireExtinguishSound{})
	c.Lit = false
	c.setSelf()
}

func (c *Campfire) ignite() {
	c.addSound(sound.FlintSteelSound{})
	c.Lit = true
	c.setSelf()
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	world.ScheduleDelayedBlockUpdate(c.position.Vector3, campfireUpdateIntervalTicks)
}

func (c *Campfire) setSelf() {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(c.position, c.self)
}
