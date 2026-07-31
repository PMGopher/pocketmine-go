package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/world/sound"
)

// itemTypeIDsHoneycomb mirrors item.HONEYCOMB (pocketmine-go/pocketmine/item.HONEYCOMB = 20238).
// Not imported directly: the item package will eventually need to import block (for ItemBlock),
// which would make an import back the other way a cycle - same reasoning as
// world/sound.ItemUseOnBlockSound storing a state ID instead of a whole Behavior.
const itemTypeIDsHoneycomb = 20238

// CopperMaterial is a port of pocketmine\block\utils\CopperMaterial.
type CopperMaterial interface {
	GetOxidation() blockutils.CopperOxidation
	SetOxidation(oxidation blockutils.CopperOxidation)
	IsWaxed() bool
	SetWaxed(waxed bool)
}

// CopperComponent is a port of pocketmine\block\utils\CopperTrait.
type CopperComponent struct {
	Oxidation blockutils.CopperOxidation
	Waxed     bool
}

func (c *CopperComponent) DescribeCopper(w runtime.DataDescriber) {
	oxidation := int(c.Oxidation)
	w.BoundedIntAuto(int(blockutils.CopperOxidationNone), int(blockutils.CopperOxidationOxidized), &oxidation)
	c.Oxidation = blockutils.CopperOxidation(oxidation)
	w.Bool(&c.Waxed)
}

func (c *CopperComponent) GetOxidation() blockutils.CopperOxidation { return c.Oxidation }

func (c *CopperComponent) SetOxidation(oxidation blockutils.CopperOxidation) { c.Oxidation = oxidation }

func (c *CopperComponent) IsWaxed() bool { return c.Waxed }

func (c *CopperComponent) SetWaxed(waxed bool) { c.Waxed = waxed }

// OnInteractCopper is a port of pocketmine\block\utils\CopperTrait::onInteract. Concrete copper
// block types call this from their own OnInteract (Go has no way to share the method itself
// across the embedding chain the way a PHP trait's onInteract, copy-pasted into the class, calls
// $this->position implicitly - see FacingComponent's doc comment for the same reasoning).
func (c *CopperComponent) OnInteractCopper(self Behavior, position Position, item Item) bool {
	world, err := position.GetWorld()
	if err != nil {
		return false
	}

	if !c.Waxed && item.GetTypeId() == itemTypeIDsHoneycomb {
		c.Waxed = true
		if err := world.SetBlock(position, self); err != nil {
			panic(err)
		}
		// TODO: orange particles are supposed to appear when applying wax
		world.AddSound(position.AsVector3(), sound.CopperWaxApplySound{})
		item.Pop()
		return true
	}

	if axe, ok := item.(Axe); ok {
		if c.Waxed {
			c.Waxed = false
			if err := world.SetBlock(position, self); err != nil {
				panic(err)
			}
			// TODO: white particles are supposed to appear when removing wax
			world.AddSound(position.AsVector3(), sound.CopperWaxRemoveSound{})
			axe.ApplyDamage(1)
			return true
		}

		if previous, ok := c.Oxidation.GetPrevious(); ok {
			c.Oxidation = previous
			if err := world.SetBlock(position, self); err != nil {
				panic(err)
			}
			// TODO: turquoise particles are supposed to appear when removing oxidation
			world.AddSound(position.AsVector3(), sound.ScrapeSound{})
			axe.ApplyDamage(1)
			return true
		}
	}

	return false
}
