package item

import "pocketmine-go/pocketmine/block"

// SweetBerries is a port of pocketmine\item\SweetBerries.
type SweetBerries struct {
	Food
}

func NewSweetBerries(identifier ItemIdentifier, name string) *SweetBerries {
	s := &SweetBerries{}
	s.Init(s, identifier, name)
	return s
}

func (s *SweetBerries) Clone() Item {
	cl := *s
	cl.rebind(&cl)
	return &cl
}

func (s *SweetBerries) GetFoodRestore() int { return 2 }

func (s *SweetBerries) GetSaturationRestore() float64 { return 1.2 }

// GetBlock is a port of SweetBerries::getBlock.
func (x *SweetBerries) GetBlock() block.Behavior { return block.VanillaBlock("sweet_berry_bush") }
