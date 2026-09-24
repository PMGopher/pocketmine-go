package entity

import (
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// Villager profession constants, a port of Villager::PROFESSION_*.
const (
	VillagerProfessionFarmer     = 0
	VillagerProfessionLibrarian  = 1
	VillagerProfessionPriest     = 2
	VillagerProfessionBlacksmith = 3
	VillagerProfessionButcher    = 4
)

const tagProfession = "Profession" //TAG_Int

// Villager is a port of pocketmine\entity\Villager (also Ageable).
type Villager struct {
	Living

	baby       bool
	profession int
}

// NewVillager is a port of Villager::__construct (Living's constructor).
func NewVillager(location Location, tag *nbt.CompoundTag) *Villager {
	v := &Villager{profession: VillagerProfessionFarmer}
	v.ConstructLiving(v, location, tag)
	return v
}

func (v *Villager) GetNetworkTypeID() string { return EntityIDVillager }

func (v *Villager) GetInitialSizeInfo() EntitySizeInfo {
	return NewEntitySizeInfo(1.9, 0.6) //TODO: eye height??
}

func (v *Villager) GetName() string { return "Villager" }

// InitEntity is a port of Villager::initEntity.
func (v *Villager) InitEntity(tag *nbt.CompoundTag) {
	v.Living.InitEntity(tag)

	profession := int(tag.GetIntOr(tagProfession, VillagerProfessionFarmer))

	if profession > 4 || profession < 0 {
		profession = VillagerProfessionFarmer
	}

	v.SetProfession(profession)
}

// SaveNBT is a port of Villager::saveNBT.
func (v *Villager) SaveNBT() *nbt.CompoundTag {
	tag := v.Living.SaveNBT()
	tag.SetInt(tagProfession, nbt.IntTag(v.GetProfession()))

	return tag
}

// SetProfession sets the villager profession.
func (v *Villager) SetProfession(profession int) {
	v.profession = profession //TODO: validation
	v.networkPropertiesDirty = true
}

func (v *Villager) GetProfession() int { return v.profession }

func (v *Villager) IsBaby() bool { return v.baby }

func (v *Villager) GetPickedItem() item.Item { return item.VanillaVillagerSpawnEgg() }

// SyncNetworkData is a port of Villager::syncNetworkData.
func (v *Villager) SyncNetworkData(properties *MetadataCollection) {
	v.Living.SyncNetworkData(properties)
	properties.SetGenericFlag(FlagBaby, v.baby)

	properties.SetInt(MetadataVariant, int32(v.profession))
}
