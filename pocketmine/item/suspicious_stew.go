package item

import runtime "pocketmine-go/pocketmine/data/runtime"

// SuspiciousStew is a port of pocketmine\item\SuspiciousStew. GetAdditionalEffects (delegating to
// SuspiciousStewType.GetEffects) and GetResidue (VanillaItems.BOWL()) aren't ported - see
// SuspiciousStewType's and Food's doc comments for why.
type SuspiciousStew struct {
	Food

	StewType SuspiciousStewType
}

func NewSuspiciousStew(identifier ItemIdentifier, name string) *SuspiciousStew {
	s := &SuspiciousStew{StewType: SuspiciousStewTypePoppy}
	s.Init(s, identifier, name)
	return s
}

func (s *SuspiciousStew) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SuspiciousStew) GetType() SuspiciousStewType { return s.StewType }

func (s *SuspiciousStew) SetType(t SuspiciousStewType) { s.StewType = t }

func (s *SuspiciousStew) GetMaxStackSize() int { return 1 }

func (s *SuspiciousStew) RequiresHunger() bool { return false }

func (s *SuspiciousStew) GetFoodRestore() int { return 6 }

func (s *SuspiciousStew) GetSaturationRestore() float64 { return 7.2 }

func (s *SuspiciousStew) describeState(w runtime.DataDescriber) {
	t := int(s.StewType)
	w.BoundedIntAuto(int(SuspiciousStewTypePoppy), int(SuspiciousStewTypeWitherRose), &t)
	s.StewType = SuspiciousStewType(t)
}
