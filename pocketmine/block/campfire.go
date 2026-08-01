package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
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

// Campfire is a port of pocketmine\block\Campfire, minus its cooking half: matching held items
// against furnace recipes and simulating cook progress needs the crafting package's
// FurnaceRecipeManager/CraftingManager (not ported) plus a real cooking-item inventory on the
// tile side (not ported either - see tile.Campfire's doc comment for why). OnScheduledUpdate is
// therefore a no-op. The ignite/extinguish state machine (Place, OnInteract's item-based
// branches, OnNearbyBlockChange, OnProjectileHit) is fully real.
type Campfire struct {
	Transparent
	HorizontalFacingComponent
	LightableComponent
}

func NewCampfire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Campfire {
	c := &Campfire{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *Campfire) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

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

// GetDropsForCompatibleTool should return [VanillaItems.CHARCOAL().SetCount(2)] - needs the
// unported item registry (see Block.GetDropsForCompatibleTool's doc comment), so it's left as
// Block's default for now.

func (c *Campfire) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (c *Campfire) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 9.0/16)}
}

// GetEntityCollisionDamage is a port of Campfire::getEntityCollisionDamage. It isn't called
// anywhere yet - the damage half of OnEntityInside needs Entity.Attack/EntityDamageByBlockEvent,
// neither ported (see OnEntityInside's doc comment) - but it's kept as a real overridable method
// (SoulCampfire returns 2) so it's ready once that gap closes.
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

// OnInteract is a port of Campfire::onInteract, minus the furnace-recipe-matching branch (see
// type doc comment). The ignite/extinguish branches are fully real.
func (c *Campfire) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if !c.Lit {
		if item.GetTypeId() == itemTypeIDsFireCharge {
			item.Pop()
			c.ignite()
			c.addSound(sound.BlazeShootSound{})
			return true
		}
		if item.GetTypeId() == itemTypeIDsFlintAndSteel {
			if durable, ok := item.(Durable); ok {
				durable.ApplyDamage(1)
			}
			c.ignite()
			return true
		}
	} else if shovel, ok := item.(Shovel); ok {
		shovel.ApplyDamage(1)
		c.extinguish()
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

// OnEntityInside is a port of Campfire::onEntityInside. The Living-entity damage branch needs
// Entity.Attack/EntityDamageByBlockEvent, neither ported (same gap category as BaseFire's doc
// comment), so it's skipped; the ignite branch is real.
func (c *Campfire) OnEntityInside(entity Entity) bool {
	if !c.Lit {
		if entity.IsOnFire() {
			c.ignite()
			return false
		}
	}
	return true
}

func (c *Campfire) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {}

// OnScheduledUpdate is a no-op - see type doc comment.
func (c *Campfire) OnScheduledUpdate() {}

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
