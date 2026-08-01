package block

import "pocketmine-go/pocketmine/entity"

// Magma is a port of pocketmine\block\Magma.
type Magma struct {
	Opaque
}

func NewMagma(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Magma {
	m := &Magma{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *Magma) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Magma) GetLightLevel() int { return 3 }

func (m *Magma) HasEntityCollision() bool { return true }

// OnEntityInside is a port of Magma::onEntityInside, minus the frost-walker-level check (not on
// the minimal Living interface yet, and no boots/enchantment system is ported anyway) - so this
// damages any non-sneaking Living entity, frost walker or not.
func (m *Magma) OnEntityInside(e Entity) bool {
	if living, ok := e.(Living); ok && !living.IsSneaking() {
		ev := entity.NewEntityDamageByBlockEvent(m.self, e, entity.EntityDamageCauseFire, 1, nil)
		living.Attack(ev)
	}
	return true
}

func (m *Magma) BurnsForever() bool { return true }
