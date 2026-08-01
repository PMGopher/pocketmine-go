package block

import "pocketmine-go/pocketmine/math"

// SoulFire is a port of pocketmine\block\SoulFire.
type SoulFire struct {
	BaseFire
}

func NewSoulFire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SoulFire {
	s := &SoulFire{BaseFire{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	s.Init(s)
	return s
}

func (s *SoulFire) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SoulFire) GetLightLevel() int { return 10 }

func (s *SoulFire) GetFireDamage() int { return 2 }

// soulFireCanBeSupportedBy is a port of SoulFire::canBeSupportedBy (a static method in the PHP
// original, since Fire also calls it to decide whether to convert into SoulFire).
//
// TODO (from the PHP original): this really ought to use some kind of tag system.
func soulFireCanBeSupportedBy(blk Behavior) bool {
	id := blk.GetTypeId()
	return id == SOUL_SAND || id == SOUL_SOIL
}

// OnNearbyBlockChange is a port of SoulFire::onNearbyBlockChange. The PHP original replaces
// itself with VanillaBlocks.AIR() specifically; this uses UseBreakOn as the practical equivalent
// (breaking via the world, rather than constructing an Air instance from the unported block
// registry) - same simplification used wherever else a block "becomes air".
func (s *SoulFire) OnNearbyBlockChange() {
	if !soulFireCanBeSupportedBy(s.self.(blockGeometry).GetSide(math.Down, 1)) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	}
}
