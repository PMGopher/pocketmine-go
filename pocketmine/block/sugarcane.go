package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const SugarcaneMaxAge = 15

// Sugarcane is a port of pocketmine\block\Sugarcane.
type Sugarcane struct {
	Flowable
	AgeComponent
}

func NewSugarcane(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Sugarcane {
	s := &Sugarcane{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(SugarcaneMaxAge),
	}
	s.Init(s)
	return s
}

func (s *Sugarcane) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Sugarcane) DescribeBlockOnlyState(w runtime.DataDescriber) { s.DescribeAge(w) }

// seekToBottom is a port of Sugarcane::seekToBottom.
func (s *Sugarcane) seekToBottom() Position {
	bottom := s.position
	for step := 1; ; step++ {
		next := s.self.(blockGeometry).GetSide(math.Down, step)
		if !next.(blockGeometry).HasSameTypeId(s.self) {
			break
		}
		bottom = next.GetPosition()
	}
	return bottom
}

// grow is a port of Sugarcane::grow. The loop that spawns new sugarcane above via
// BlockEventHelper.Grow needs BlockEventHelper and World.IsInWorld, neither ported yet, so it's
// skipped - this always resets age and returns false (never grew) for now.
func (s *Sugarcane) grow(pos Position) bool {
	world, err := pos.GetWorld()
	if err != nil {
		return false
	}
	s.Age = 0
	if err := world.SetBlock(pos, s.self); err != nil {
		panic(err)
	}
	return false
}

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, not ported yet. Block's
// default OnInteract (return false) already matches this gap, so there's nothing to override
// here.

func (s *Sugarcane) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	return support.(blockGeometry).HasSameTypeId(s.self) ||
		support.(blockGeometry).HasTypeTag(BlockTypeTagsMud) ||
		support.(blockGeometry).HasTypeTag(BlockTypeTagsDirt) ||
		support.(blockGeometry).HasTypeTag(BlockTypeTagsSand)
}

func (s *Sugarcane) TicksRandomly() bool { return true }

func (s *Sugarcane) hasNearbyWater(down Behavior) bool {
	geo := down.(blockGeometry)
	for _, side := range math.HorizontalFacing {
		typeID := geo.GetSide(side, 1).GetTypeId()
		if typeID == WATER || typeID == FROSTED_ICE {
			return true
		}
	}
	return false
}

func (s *Sugarcane) OnRandomTick() {
	down := s.self.(blockGeometry).GetSide(math.Down, 1)
	if down.(blockGeometry).HasSameTypeId(s.self) {
		return
	}
	world, err := s.position.GetWorld()
	if err != nil {
		return
	}
	if !s.hasNearbyWater(down) {
		world.UseBreakOn(s.position.AsVector3())
		return
	}
	if s.Age == SugarcaneMaxAge {
		s.grow(s.position)
	} else {
		s.Age++
		if err := world.SetBlock(s.position, s.self); err != nil {
			panic(err)
		}
	}
}

func (s *Sugarcane) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	down := blockReplace.(blockGeometry).GetSide(math.Down, 1)
	if down.(blockGeometry).HasSameTypeId(s.self) {
		return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}

	// Support criteria are checked by canBeSupportedAt/OnNearbyBlockChange - this part applies to
	// placement only.
	for _, side := range math.HorizontalFacing {
		typeID := down.(blockGeometry).GetSide(side, 1).GetTypeId()
		if typeID == WATER || typeID == FROSTED_ICE {
			return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
		}
	}

	return false
}

func (s *Sugarcane) onSupportBlockChange() {
	if !s.canBeSupportedAt(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.Flowable.OnNearbyBlockChange()
	}
}

func (s *Sugarcane) OnNearbyBlockChange() {
	down := s.self.(blockGeometry).GetSide(math.Down, 1)
	if !down.(blockGeometry).HasSameTypeId(s.self) && !s.hasNearbyWater(down) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.onSupportBlockChange()
	}
}
