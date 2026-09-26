package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Sapling is a port of pocketmine\block\Sapling.
type Sapling struct {
	Flowable

	Ready       bool
	SaplingType blockutils.SaplingType
}

func NewSapling(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, saplingType blockutils.SaplingType) *Sapling {
	s := &Sapling{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, SaplingType: saplingType}
	s.Init(s)
	return s
}

func (s *Sapling) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Sapling) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&s.Ready) }

func (s *Sapling) IsReady() bool { return s.Ready }

func (s *Sapling) SetReady(ready bool) { s.Ready = ready }

func (s *Sapling) GetSaplingType() blockutils.SaplingType { return s.SaplingType }

func (s *Sapling) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud)
}

func (s *Sapling) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return s.canBeSupportedAt(blockReplace) && s.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (s *Sapling) OnNearbyBlockChange() {
	if !s.canBeSupportedAt(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract is a port of Sapling::onInteract.
func (s *Sapling) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if isFertilizer(item) && s.grow(player) {
		item.Pop()
		return true
	}
	return false
}

func (s *Sapling) TicksRandomly() bool { return true }

func (s *Sapling) OnRandomTick() {
	world, err := s.position.GetWorld()
	if err != nil {
		return
	}
	pos := s.position.AsVector3()
	if world.GetFullLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) < 8 || rand.Intn(7) != 0 {
		return
	}
	if s.Ready {
		s.grow(nil)
	} else {
		s.Ready = true
		if err := world.SetBlock(s.position, s.self); err != nil {
			panic(err)
		}
	}
}

// grow is a port of Sapling::grow.
func (s *Sapling) grow(player Player) bool {
	return growStructure(s.self, saplingTreeType(s.SaplingType), player)
}

// saplingTreeType is a port of SaplingType::getTreeType.
func saplingTreeType(t blockutils.SaplingType) TreeType {
	switch t {
	case blockutils.SaplingTypeSpruce:
		return TreeTypeSpruce
	case blockutils.SaplingTypeBirch:
		return TreeTypeBirch
	case blockutils.SaplingTypeJungle:
		return TreeTypeJungle
	case blockutils.SaplingTypeAcacia:
		return TreeTypeAcacia
	case blockutils.SaplingTypeDarkOak:
		return TreeTypeDarkOak
	}
	return TreeTypeOak
}

func (s *Sapling) GetFuelTime() int { return 100 }
