package item

import "pocketmine-go/pocketmine/entity/effect"

// ItemCooldownTagChorusFruit mirrors ItemCooldownTags::CHORUS_FRUIT.
const ItemCooldownTagChorusFruit = "chorus_fruit"

// ChorusFruitTeleportFunc is ChorusFruit::onConsume's body (the random teleport): it needs the
// consumer's world, which this package can't import, so the entity package sets it in init().
var ChorusFruitTeleportFunc func(consumer effect.Living)

// ChorusFruit is a port of pocketmine\item\ChorusFruit.
type ChorusFruit struct {
	Food
}

func NewChorusFruit(identifier ItemIdentifier, name string) *ChorusFruit {
	c := &ChorusFruit{}
	c.Init(c, identifier, name)
	return c
}

func (c *ChorusFruit) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *ChorusFruit) GetFoodRestore() int { return 4 }

func (c *ChorusFruit) GetSaturationRestore() float64 { return 2.4 }

func (c *ChorusFruit) RequiresHunger() bool { return false }

// OnConsume is a port of ChorusFruit::onConsume: the consumer is teleported up to 8 blocks away.
func (c *ChorusFruit) OnConsume(consumer effect.Living) {
	if ChorusFruitTeleportFunc != nil {
		ChorusFruitTeleportFunc(consumer)
	}
}

func (c *ChorusFruit) GetCooldownTicks() int { return 20 }

func (c *ChorusFruit) GetCooldownTag() (string, bool) { return ItemCooldownTagChorusFruit, true }
