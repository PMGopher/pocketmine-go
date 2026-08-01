package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	campfireTagFirstCookingTime  = "ItemTime1"
	campfireTagSecondCookingTime = "ItemTime2"
	campfireTagThirdCookingTime  = "ItemTime3"
	campfireTagFourthCookingTime = "ItemTime4"
)

var campfireCookingTimeTags = [4]string{
	campfireTagFirstCookingTime,
	campfireTagSecondCookingTime,
	campfireTagThirdCookingTime,
	campfireTagFourthCookingTime,
}

// Campfire is a port of pocketmine\block\tile\Campfire, minus its inventory/Container half - see
// ContainerComponent's doc comment for why the inventory package can't be imported here. Since
// there's no real inventory, the item half of each slot's save data (Item1-4) is never populated,
// so only the cooking-time half round-trips through NBT - a slot's cooking time surviving a
// reload with no item to go with it is harmless (onScheduledUpdate, which would consume it, isn't
// ported either - see block.Campfire's doc comment).
type Campfire struct {
	SpawnableBase

	cookingTimes map[int]int
}

func NewCampfire(world World, pos math.Vector3) *Campfire {
	c := &Campfire{cookingTimes: map[int]int{}}
	c.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	c.Init(c)
	return c
}

func (c *Campfire) SaveID() string { return "Campfire" }

// AddAdditionalSpawnData is a port of Campfire::addAdditionalSpawnData, minus the item-to-network-
// NBT translation (TypeConverter isn't ported) - with no real inventory to read items from
// anyway, this is a no-op, same as EnderChest's.
func (c *Campfire) AddAdditionalSpawnData(tag *nbt.CompoundTag) {}

// GetCookingTimes is a port of pocketmine\block\tile\Campfire::getCookingTimes.
func (c *Campfire) GetCookingTimes() map[int]int { return c.cookingTimes }

// SetCookingTimes is a port of pocketmine\block\tile\Campfire::setCookingTimes.
func (c *Campfire) SetCookingTimes(cookingTimes map[int]int) { c.cookingTimes = cookingTimes }

func (c *Campfire) ReadSaveData(tag *nbt.CompoundTag) error {
	for slot, tagName := range campfireCookingTimeTags {
		if t, ok := tag.GetTag(tagName); ok {
			if v, ok := t.(nbt.IntTag); ok {
				c.cookingTimes[slot] = int(v)
			}
		}
	}
	return nil
}

func (c *Campfire) WriteSaveData(tag *nbt.CompoundTag) {
	for slot, tagName := range campfireCookingTimeTags {
		if t, ok := c.cookingTimes[slot]; ok {
			tag.SetInt(tagName, nbt.IntTag(t))
		}
	}
}
