package block

import (
	"math/rand"
	"pocketmine-go/pocketmine/world/sound"

	runtime "pocketmine-go/pocketmine/data/runtime"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

const (
	SweetBerryBushStageSapling         = 0
	SweetBerryBushStageBushNoBerries   = 1
	SweetBerryBushStageBushSomeBerries = 2
	SweetBerryBushStageMature          = 3
	SweetBerryBushMaxAge               = SweetBerryBushStageMature
)

// SweetBerryBush is a port of pocketmine\block\SweetBerryBush.
type SweetBerryBush struct {
	Flowable
	AgeComponent
}

func NewSweetBerryBush(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SweetBerryBush {
	s := &SweetBerryBush{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(SweetBerryBushMaxAge),
	}
	s.Init(s)
	return s
}

func (s *SweetBerryBush) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SweetBerryBush) DescribeBlockOnlyState(w runtime.DataDescriber) { s.DescribeAge(w) }

func (s *SweetBerryBush) GetBerryDropAmount() int {
	switch {
	case s.Age == SweetBerryBushStageMature:
		return rand.Intn(2) + 2 // 2-3
	case s.Age >= SweetBerryBushStageBushSomeBerries:
		return rand.Intn(2) + 1 // 1-2
	default:
		return 0
	}
}

// canBeSupportedBy is a port of the (deprecated in the PHP original) canBeSupportedBy.
func (s *SweetBerryBush) canBeSupportedBy(blk Behavior) bool {
	geo := blk.(blockGeometry)
	return blk.GetTypeId() != FARMLAND && (geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud))
}

func (s *SweetBerryBush) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	return s.canBeSupportedBy(support)
}

func (s *SweetBerryBush) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return s.canBeSupportedAt(blockReplace) && s.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (s *SweetBerryBush) OnNearbyBlockChange() {
	if !s.canBeSupportedAt(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract is a port of SweetBerryBush::onInteract: bone meal grows it, otherwise ripe berries
// are picked.
func (s *SweetBerryBush) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := s.position.GetWorld()
	if err != nil {
		return true
	}
	if s.Age < SweetBerryBushStageMature && item.GetTypeId() == itemTypeIDsBoneMeal {
		next := s.self.Clone().(*SweetBerryBush)
		next.Age++
		if Grow(s.self, next, player) {
			item.Pop()
		}
	} else if dropAmount := s.GetBerryDropAmount(); dropAmount > 0 {
		s.Age = SweetBerryBushStageBushNoBerries
		_ = world.SetBlock(s.position, s.self)
		if berries := asItemOrNil(s.self); berries != nil {
			berries.SetCount(dropAmount)
			if dropper, ok := world.(blockItemDropper); ok {
				dropper.DropBlockItem(s.position.Vector3, berries)
			}
		}
		world.AddSound(s.position.Vector3, sound.SweetBerriesPickSound{})
	}
	return true
}

// GetDropsForCompatibleTool is a port of SweetBerryBush::getDropsForCompatibleTool.
func (s *SweetBerryBush) GetDropsForCompatibleTool(item Item) []Item {
	count := 0
	switch s.Age {
	case SweetBerryBushStageMature:
		count = FortuneDiscrete(item, 2, 3)
	case SweetBerryBushStageBushSomeBerries:
		count = FortuneDiscrete(item, 1, 2)
	}
	drop := asItemOrNil(s.self)
	if drop == nil {
		return nil
	}
	drop.SetCount(count)
	return []Item{drop}
}

func (s *SweetBerryBush) TicksRandomly() bool { return s.Age < SweetBerryBushStageMature }

// OnRandomTick is a port of SweetBerryBush::onRandomTick.
func (s *SweetBerryBush) OnRandomTick() {
	if s.Age < SweetBerryBushStageMature && rand.Intn(3) == 1 { // mt_rand(0, 2) === 1
		s.growOnce()
	}
}

// growOnce is the deterministic clone-increment-Grow act from OnRandomTick, split out from the
// random roll above so it's directly testable - same pattern as CocoaBlock.grow.
func (s *SweetBerryBush) growOnce() bool {
	next := s.self.Clone().(*SweetBerryBush)
	next.Age++
	return Grow(s.self, next, nil)
}

func (s *SweetBerryBush) HasEntityCollision() bool { return true }

// OnEntityInside is a port of SweetBerryBush::onEntityInside. The PHP original's TODO about only
// triggering while moving through the bush isn't addressed here either (no such movement-tracking
// system is ported).
func (s *SweetBerryBush) OnEntityInside(e Entity) bool {
	if s.Age >= SweetBerryBushStageBushNoBerries {
		if living, ok := e.(Living); ok {
			living.ResetFallDistance()
			ev := entityevent.NewEntityDamageByBlockEvent(s.self, e, entityevent.CauseContact, 1, nil)
			living.Attack(ev)
		}
	}
	return true
}
