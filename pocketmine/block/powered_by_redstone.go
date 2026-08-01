package block

// PoweredByRedstone is a port of pocketmine\block\utils\PoweredByRedstone.
type PoweredByRedstone interface {
	IsPowered() bool
	SetPowered(powered bool)
}

// PoweredByRedstoneComponent is a port of pocketmine\block\utils\PoweredByRedstoneTrait.
type PoweredByRedstoneComponent struct {
	Powered bool
}

func (p *PoweredByRedstoneComponent) IsPowered() bool { return p.Powered }

func (p *PoweredByRedstoneComponent) SetPowered(powered bool) { p.Powered = powered }

// Lightable is a port of pocketmine\block\utils\Lightable.
type Lightable interface {
	IsLit() bool
	SetLit(lit bool)
}
