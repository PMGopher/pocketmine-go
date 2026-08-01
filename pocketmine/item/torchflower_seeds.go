package item

// TorchflowerSeeds is a port of pocketmine\item\TorchflowerSeeds. GetBlock (should return VanillaBlocks.TORCHFLOWER_CROP())
// isn't ported - see StringItem's doc comment for why.
type TorchflowerSeeds struct {
	ItemBase
}

func NewTorchflowerSeeds(identifier ItemIdentifier, name string) *TorchflowerSeeds {
	t := &TorchflowerSeeds{}
	t.Init(t, identifier, name)
	return t
}

func (t *TorchflowerSeeds) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}
