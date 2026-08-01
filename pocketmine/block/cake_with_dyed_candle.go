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

// GetCandle should return VanillaBlocks.DYED_CANDLE().SetColor(c.Color) - needs the unported
// block registry, same gap as CakeWithCandle.GetCandle.
func (c *CakeWithDyedCandle) GetCandle() *Candle {
	return &Candle{Count: candleMinCount}
}
