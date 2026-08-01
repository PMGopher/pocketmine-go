package blockutils

// MobHeadType is a port of pocketmine\block\utils\MobHeadType.
type MobHeadType int

const (
	MobHeadTypeSkeleton MobHeadType = iota
	MobHeadTypeWitherSkeleton
	MobHeadTypeZombie
	MobHeadTypePlayer
	MobHeadTypeCreeper
	MobHeadTypeDragon
	MobHeadTypePiglin
)
