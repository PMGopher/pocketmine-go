package block

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// The asItem() overrides of blocks whose item isn't a plain block item (crops drop seeds, signs and
// banners have dedicated items, ...). The items come from VanillaItems through VanillaItemFunc.

func vanillaItemOrError(name string) (Item, error) {
	it := vanillaItem(name)
	if it == nil {
		return nil, fmt.Errorf("block: VanillaItemFunc isn't set (the item package was never imported)")
	}
	return it, nil
}

// AsItem is a port of Bamboo::asItem.
func (b *Bamboo) AsItem() (Item, error) { return vanillaItemOrError("bamboo") }

// AsItem is a port of BambooSapling::asItem.
func (b *BambooSapling) AsItem() (Item, error) { return vanillaItemOrError("bamboo") }

// AsItem is a port of BaseOminousBanner::asItem.
func (b *BaseOminousBanner) AsItem() (Item, error) { return vanillaItemOrError("ominous_banner") }

// AsItem is a port of BaseSign::asItem ($this->asItemCallback: the sign or hanging sign item of the
// wood type, as VanillaBlocksInputs passes it).
func (b *BaseSign) AsItem() (Item, error) {
	switch b.self.(type) {
	case *CeilingCenterHangingSign, *CeilingEdgesHangingSign, *WallHangingSign:
		return vanillaItemOrError(b.WoodType.IDName() + "_hanging_sign")
	}
	return vanillaItemOrError(b.WoodType.IDName() + "_sign")
}

// itemColorSetter is item.Banner's SetColor.
type itemColorSetter interface {
	SetColor(color blockutils.DyeColor)
}

// AsItem is a port of BaseBanner::asItem.
func (b *BaseBanner) AsItem() (Item, error) {
	it, err := vanillaItemOrError("banner")
	if err != nil {
		return nil, err
	}
	it.(itemColorSetter).SetColor(b.Color)
	return it, nil
}

// AsItem is a port of Beetroot::asItem.
func (b *Beetroot) AsItem() (Item, error) { return vanillaItemOrError("beetroot_seeds") }

// AsItem is a port of Carrot::asItem.
func (c *Carrot) AsItem() (Item, error) { return vanillaItemOrError("carrot") }

// AsItem is a port of Potato::asItem.
func (p *Potato) AsItem() (Item, error) { return vanillaItemOrError("potato") }

// AsItem is a port of Wheat::asItem.
func (w *Wheat) AsItem() (Item, error) { return vanillaItemOrError("wheat_seeds") }

// AsItem is a port of MelonStem::asItem.
func (m *MelonStem) AsItem() (Item, error) { return vanillaItemOrError("melon_seeds") }

// AsItem is a port of PumpkinStem::asItem.
func (p *PumpkinStem) AsItem() (Item, error) { return vanillaItemOrError("pumpkin_seeds") }

// AsItem is a port of PitcherCrop::asItem.
func (p *PitcherCrop) AsItem() (Item, error) { return vanillaItemOrError("pitcher_pod") }

// AsItem is a port of DoublePitcherCrop::asItem.
func (d *DoublePitcherCrop) AsItem() (Item, error) { return vanillaItemOrError("pitcher_pod") }

// AsItem is a port of RedstoneWire::asItem.
func (r *RedstoneWire) AsItem() (Item, error) { return vanillaItemOrError("redstone_dust") }

// AsItem is a port of BigDripleafStem::asItem.
func (b *BigDripleafStem) AsItem() (Item, error) {
	return VanillaBlock("big_dripleaf_head").(interface{ AsItem() (Item, error) }).AsItem()
}

// AsItem is a port of CaveVines::asItem.
func (c *CaveVines) AsItem() (Item, error) { return vanillaItemOrError("glow_berries") }

// AsItem is a port of CocoaBlock::asItem.
func (c *CocoaBlock) AsItem() (Item, error) { return vanillaItemOrError("cocoa_beans") }

// AsItem is a port of SweetBerryBush::asItem.
func (s *SweetBerryBush) AsItem() (Item, error) { return vanillaItemOrError("sweet_berries") }

// AsItem is a port of TorchflowerCrop::asItem.
func (t *TorchflowerCrop) AsItem() (Item, error) { return vanillaItemOrError("torchflower_seeds") }

// AsItem is a port of Tripwire::asItem.
func (t *Tripwire) AsItem() (Item, error) { return vanillaItemOrError("string") }

// itemCoralSetter is item.CoralFan's coral setters.
type itemCoralSetter interface {
	SetCoralType(coralType blockutils.CoralType)
	SetDead(dead bool)
}

func coralFanItem(coral CoralComponent) (Item, error) {
	it, err := vanillaItemOrError("coral_fan")
	if err != nil {
		return nil, err
	}
	setter := it.(itemCoralSetter)
	setter.SetCoralType(coral.CoralType)
	setter.SetDead(coral.Dead)
	return it, nil
}

// AsItem is a port of FloorCoralFan::asItem.
func (f *FloorCoralFan) AsItem() (Item, error) { return coralFanItem(f.CoralComponent) }

// AsItem is a port of WallCoralFan::asItem.
func (w *WallCoralFan) AsItem() (Item, error) { return coralFanItem(w.CoralComponent) }
