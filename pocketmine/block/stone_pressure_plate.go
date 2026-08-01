package block

// StonePressurePlate is a port of pocketmine\block\StonePressurePlate.
type StonePressurePlate struct {
	SimplePressurePlate
}

func NewStonePressurePlate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, deactivationDelayTicks int) *StonePressurePlate {
	s := &StonePressurePlate{
		SimplePressurePlate: SimplePressurePlate{
			PressurePlate: PressurePlate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, DeactivationDelayTicks: deactivationDelayTicks},
		},
	}
	s.Init(s)
	return s
}

func (s *StonePressurePlate) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

// filterIrrelevantEntities: TODO (from the PHP original) armor stands should activate stone
// plates too.
func (s *StonePressurePlate) filterIrrelevantEntities(entities []Entity) []Entity {
	var result []Entity
	for _, e := range entities {
		if _, ok := e.(Living); ok {
			result = append(result, e)
		}
	}
	return result
}
