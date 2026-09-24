package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// TNTBlock is a port of pocketmine\block\TNT. Named TNTBlock rather than TNT to avoid colliding
// with the TNT block-type-ID constant in type_ids.go (the only PHP block class name that's
// already all-uppercase, so it's the only one that collides with its own BlockTypeIds constant).
type TNTBlock struct {
	Opaque

	Unstable        bool
	WorksUnderwater bool
}

func NewTNT(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *TNTBlock {
	t := &TNTBlock{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	t.Init(t)
	return t
}

func (t *TNTBlock) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *TNTBlock) DescribeBlockItemState(w runtime.DataDescriber) { w.Bool(&t.WorksUnderwater) }

func (t *TNTBlock) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&t.Unstable) }

func (t *TNTBlock) IsUnstable() bool { return t.Unstable }

func (t *TNTBlock) SetUnstable(unstable bool) { t.Unstable = unstable }

func (t *TNTBlock) DoesWorkUnderwater() bool { return t.WorksUnderwater }

func (t *TNTBlock) SetWorksUnderwater(worksUnderwater bool) { t.WorksUnderwater = worksUnderwater }

func (t *TNTBlock) OnBreak(item Item, player Player, returnedItems *[]Item) bool {
	if t.Unstable {
		t.Ignite(80)
		return true
	}
	return t.Opaque.OnBreak(item, player, returnedItems)
}

func (t *TNTBlock) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() == itemTypeIDsFireCharge {
		item.Pop()
		t.Ignite(80)
		return true
	}
	if item.GetTypeId() == itemTypeIDsFlintAndSteel {
		if durable, ok := item.(Durable); ok {
			durable.ApplyDamage(1)
		}
		t.Ignite(80)
		return true
	}
	return false
}

// SpawnPrimedTNTFunc creates, launches and spawns a PrimedTNT entity for TNT ignited at pos (the
// entity half of TNT::ignite). pocketmine/entity/object installs it from its init() - the same
// dependency-inversion hook as NewItemBlockFunc.
var SpawnPrimedTNTFunc func(pos Position, worksUnderwater bool, fuse int)

// Ignite is a port of TNT::ignite (PHP's default fuse is 80).
func (t *TNTBlock) Ignite(fuse int) {
	world, err := t.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(t.position, VanillaAir())

	if SpawnPrimedTNTFunc != nil {
		SpawnPrimedTNTFunc(t.position, t.WorksUnderwater, fuse)
	}
}

func (t *TNTBlock) GetFlameEncouragement() int { return 15 }

func (t *TNTBlock) GetFlammability() int { return 100 }

func (t *TNTBlock) OnIncinerate() { t.Ignite(80) }

func (t *TNTBlock) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	if projectile.IsOnFire() {
		t.Ignite(80)
	}
}
