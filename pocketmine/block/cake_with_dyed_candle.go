package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// CakeWithDyedCandle is a port of pocketmine\block\CakeWithDyedCandle.
type CakeWithDyedCandle struct {
	CakeWithCandle
	ColorComponent
}

func NewCakeWithDyedCandle(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CakeWithDyedCandle {
	c := &CakeWithDyedCandle{
		CakeWithCandle: CakeWithCandle{BaseCake: BaseCake{Transparent{NewBlock(idInfo, name, typeInfo)}}},
		ColorComponent: NewColorComponent(),
	}
	c.Init(c)
	return c
}

func (c *CakeWithDyedCandle) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CakeWithDyedCandle) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeColor(w) }

// GetCandle is a port of CakeWithDyedCandle::getCandle.
func (c *CakeWithDyedCandle) GetCandle() Behavior {
	candle := VanillaBlock("dyed_candle").(*DyedCandle)
	candle.Color = c.Color
	return candle
}
