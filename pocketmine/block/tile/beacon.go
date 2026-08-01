package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	beaconTagPrimary   = "primary"
	beaconTagSecondary = "secondary"
)

// Beacon is a port of pocketmine\block\tile\Beacon.
type Beacon struct {
	SpawnableBase

	PrimaryEffect   int
	SecondaryEffect int
}

func NewBeacon(world World, pos math.Vector3) *Beacon {
	b := &Beacon{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}}
	b.Init(b)
	return b
}

func (b *Beacon) SaveID() string { return "Beacon" }

func (b *Beacon) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetInt(beaconTagPrimary, nbt.IntTag(b.PrimaryEffect))
	tag.SetInt(beaconTagSecondary, nbt.IntTag(b.SecondaryEffect))
}

// ReadSaveData is a port of Beacon::readSaveData.
//
// TODO (from the PHP original): PC uses "Primary"/"Secondary" (capitalized first letter), not
// read here because the effect IDs would be different.
func (b *Beacon) ReadSaveData(tag *nbt.CompoundTag) error {
	b.PrimaryEffect = int(tag.GetIntOr(beaconTagPrimary, 0))
	b.SecondaryEffect = int(tag.GetIntOr(beaconTagSecondary, 0))
	return nil
}

func (b *Beacon) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetInt(beaconTagPrimary, nbt.IntTag(b.PrimaryEffect))
	tag.SetInt(beaconTagSecondary, nbt.IntTag(b.SecondaryEffect))
}

func (b *Beacon) GetPrimaryEffect() int { return b.PrimaryEffect }

func (b *Beacon) SetPrimaryEffect(primaryEffect int) { b.PrimaryEffect = primaryEffect }

func (b *Beacon) GetSecondaryEffect() int { return b.SecondaryEffect }

func (b *Beacon) SetSecondaryEffect(secondaryEffect int) { b.SecondaryEffect = secondaryEffect }
