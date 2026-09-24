package entity

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// Zombie is a port of pocketmine\entity\Zombie.
type Zombie struct {
	Living
}

// NewZombie is a port of Zombie::__construct (Living's constructor).
func NewZombie(location Location, tag *nbt.CompoundTag) *Zombie {
	z := &Zombie{}
	z.ConstructLiving(z, location, tag)
	return z
}

func (z *Zombie) GetNetworkTypeID() string { return EntityIDZombie }

func (z *Zombie) GetInitialSizeInfo() EntitySizeInfo {
	return NewEntitySizeInfo(1.9, 0.6) //TODO: eye height ??
}

func (z *Zombie) GetName() string { return "Zombie" }

// GetDrops is a port of Zombie::getDrops.
func (z *Zombie) GetDrops() []item.Item {
	rottenFlesh := item.VanillaRottenFlesh()
	rottenFlesh.SetCount(rand.IntN(3))
	drops := []item.Item{rottenFlesh}

	if rand.IntN(200) < 5 {
		switch rand.IntN(3) {
		case 0:
			drops = append(drops, item.VanillaIronIngot())
		case 1:
			drops = append(drops, item.VanillaCarrot())
		case 2:
			drops = append(drops, item.VanillaPotato())
		}
	}

	return drops
}

// GetXpDropAmount is a port of Zombie::getXpDropAmount.
func (z *Zombie) GetXpDropAmount() int {
	//TODO: check for equipment and whether it's a baby
	return 5
}

func (z *Zombie) GetPickedItem() item.Item { return item.VanillaZombieSpawnEgg() }
