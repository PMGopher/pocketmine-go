package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// DyedCandle is a port of pocketmine\block\DyedCandle.
type DyedCandle struct {
	Candle
	ColorComponent
}

func NewDyedCandle(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DyedCandle {
	d := &DyedCandle{
		Candle:         Candle{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Count: candleMinCount},
		ColorComponent: NewColorComponent(),
	}
	d.Init(d)
	return d
}

func (d *DyedCandle) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DyedCandle) DescribeBlockItemState(w runtime.DataDescriber) { d.DescribeColor(w) }

// GetCandleIfCompatibleType is a port of DyedCandle::getCandleIfCompatibleType — differently
// coloured candles can't be combined in the same block.
func (d *DyedCandle) GetCandleIfCompatibleType(blk Behavior) *Candle {
	result := d.Candle.GetCandleIfCompatibleType(blk)
	if result == nil {
		return nil
	}
	if dyed, ok := blk.(*DyedCandle); ok && dyed.Color == d.Color {
		return result
	}
	return nil
}
