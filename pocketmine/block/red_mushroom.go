package block

import "pocketmine-go/pocketmine/math"

// RedMushroom is a port of pocketmine\block\RedMushroom.
type RedMushroom struct {
	Flowable
}

func NewRedMushroom(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedMushroom {
	r := &RedMushroom{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	r.Init(r)
	return r
}

func (r *RedMushroom) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedMushroom) TicksRandomly() bool { return true }

func (r *RedMushroom) OnNearbyBlockChange() {
	if r.GetSide(math.Down, 1).IsTransparent() {
		if world, err := r.position.GetWorld(); err == nil {
			world.UseBreakOn(r.position.AsVector3())
		}
	}
}

func (r *RedMushroom) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	down := r.GetSide(math.Down, 1)
	world, err := r.position.GetWorld()
	if err != nil {
		panic(err)
	}
	pos := r.position.AsVector3()
	lightLevel := world.GetFullLightAt(int(pos.X), int(pos.Y), int(pos.Z))
	downID := down.GetTypeId()
	if (lightLevel <= 12 && !down.IsTransparent()) || downID == MYCELIUM || downID == PODZOL ||
		down.(blockGeometry).HasTypeTag(BlockTypeTagsNylium) {
		return r.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}
