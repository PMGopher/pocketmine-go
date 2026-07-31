package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Axe is a forward-compatible marker for pocketmine\item\Axe — declared here since Wood.OnInteract
// is the only current consumer, matching the local-interface pattern used elsewhere for
// not-yet-ported types. The future Axe item type just needs to satisfy this.
type Axe interface {
	ApplyDamage(amount int)
}

// Wood is a port of pocketmine\block\Wood.
type Wood struct {
	Opaque
	PillarRotationComponent
	WoodTypeComponent

	Stripped bool
}

func NewWood(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *Wood {
	w := &Wood{
		Opaque:                  Opaque{NewBlock(idInfo, name, typeInfo)},
		PillarRotationComponent: NewPillarRotationComponent(),
		WoodTypeComponent:       NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *Wood) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *Wood) DescribeBlockItemState(d runtime.DataDescriber) { d.Bool(&w.Stripped) }

func (w *Wood) DescribeBlockOnlyState(d runtime.DataDescriber) { w.DescribeAxis(d) }

func (w *Wood) IsStripped() bool { return w.Stripped }

func (w *Wood) SetStripped(stripped bool) { w.Stripped = stripped }

func (w *Wood) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	w.SetAxisFromFace(face)
	return w.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (w *Wood) GetFuelTime() int {
	if w.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

func (w *Wood) GetFlameEncouragement() int {
	if w.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (w *Wood) GetFlammability() int {
	if w.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (w *Wood) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if axe, ok := item.(Axe); !w.Stripped && ok {
		axe.ApplyDamage(1)
		w.Stripped = true
		if world, err := w.position.GetWorld(); err == nil {
			if err := world.SetBlock(w.position, w.self); err != nil {
				panic(err)
			}
			world.AddSound(w.position.AsVector3(), sound.ItemUseOnBlockSound{BlockStateID: w.GetStateId()})
		}
		return true
	}
	return false
}
