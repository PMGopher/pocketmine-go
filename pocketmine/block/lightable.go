package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// LightableComponent is a port of pocketmine\block\utils\LightableTrait.
type LightableComponent struct {
	Lit bool
}

func (l *LightableComponent) DescribeLit(w runtime.DataDescriber) { w.Bool(&l.Lit) }

func (l *LightableComponent) IsLit() bool { return l.Lit }

func (l *LightableComponent) SetLit(lit bool) { l.Lit = lit }
