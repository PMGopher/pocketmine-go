package block

// ChemicalHeat is a port of pocketmine\block\ChemicalHeat.
//
// TODO (from the PHP original): this block causes melting of nearby ice and snow within 2 blocks
// taxicab distance. Since it doesn't emit any light, the mechanics of this block's behaviour are
// not currently clear even upstream, so - like the original - no such behavior is implemented.
type ChemicalHeat struct {
	Transparent
}

func NewChemicalHeat(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ChemicalHeat {
	c := &ChemicalHeat{Transparent{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *ChemicalHeat) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}
