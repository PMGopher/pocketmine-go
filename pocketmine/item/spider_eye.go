package item

// SpiderEye is a port of pocketmine\item\SpiderEye. GetAdditionalEffects (a Poison effect) isn't
// ported - see GoldenApple's doc comment for why.
type SpiderEye struct {
	Food
}

func NewSpiderEye(identifier ItemIdentifier, name string) *SpiderEye {
	s := &SpiderEye{}
	s.Init(s, identifier, name)
	return s
}

func (s *SpiderEye) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SpiderEye) GetFoodRestore() int { return 2 }

func (s *SpiderEye) GetSaturationRestore() float64 { return 3.2 }
