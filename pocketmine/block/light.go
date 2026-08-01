package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	LightMinLightLevel = 0
	LightMaxLightLevel = 15
)

// Light is a port of pocketmine\block\Light.
type Light struct {
	Flowable

	Level int
}

func NewLight(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Light {
	l := &Light{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Level: LightMaxLightLevel}
	l.Init(l)
	return l
}

func (l *Light) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Light) DescribeBlockItemState(w runtime.DataDescriber) {
	w.BoundedIntAuto(LightMinLightLevel, LightMaxLightLevel, &l.Level)
}

func (l *Light) GetLightLevel() int { return l.Level }

// SetLightLevel panics if level is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (l *Light) SetLightLevel(level int) {
	if level < LightMinLightLevel || level > LightMaxLightLevel {
		panic("Light level must be in the range 0 ... 15")
	}
	l.Level = level
}

func (l *Light) CanBeReplaced() bool { return true }

// CanBePlacedAt: light blocks behave like solid blocks when placing them on another light block.
func (l *Light) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return blockReplace.CanBeReplaced() && blockReplace.GetTypeId() != l.GetTypeId()
}

func (l *Light) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if l.Level == LightMaxLightLevel {
		l.Level = LightMinLightLevel
	} else {
		l.Level++
	}

	if world, err := l.position.GetWorld(); err == nil {
		if err := world.SetBlock(l.position, l.self); err != nil {
			panic(err)
		}
	}

	return true
}
