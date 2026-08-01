package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// CoralMaterial is a port of pocketmine\block\utils\CoralMaterial.
type CoralMaterial interface {
	GetCoralType() blockutils.CoralType
	SetCoralType(coralType blockutils.CoralType)
	IsDead() bool
	SetDead(dead bool)
}

// CoralComponent is a port of pocketmine\block\utils\CoralTypeTrait.
type CoralComponent struct {
	CoralType blockutils.CoralType
	Dead      bool
}

func (c *CoralComponent) DescribeCoral(w runtime.DataDescriber) {
	t := int(c.CoralType)
	w.BoundedIntAuto(int(blockutils.CoralTypeTube), int(blockutils.CoralTypeHorn), &t)
	c.CoralType = blockutils.CoralType(t)
	w.Bool(&c.Dead)
}

func (c *CoralComponent) GetCoralType() blockutils.CoralType { return c.CoralType }

func (c *CoralComponent) SetCoralType(coralType blockutils.CoralType) { c.CoralType = coralType }

func (c *CoralComponent) IsDead() bool { return c.Dead }

func (c *CoralComponent) SetDead(dead bool) { c.Dead = dead }
