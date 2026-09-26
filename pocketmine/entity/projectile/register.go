package projectile

import (
	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// init registers this package's entity types with EntityFactory (the projectile half of PHP's
// EntityFactory constructor).
func init() {
	f := entity.GetEntityFactory()

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Arrow, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewArrow(loc, nil, tag.GetByteOr(TagCrit, 0) == 1, tag), nil
	}, []string{"Arrow", "minecraft:arrow"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Egg, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewEgg(loc, nil, tag), nil
	}, []string{"Egg", "minecraft:egg"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*EnderPearl, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewEnderPearl(loc, nil, tag), nil
	}, []string{"ThrownEnderpearl", "minecraft:ender_pearl"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*ExperienceBottle, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewExperienceBottle(loc, nil, tag), nil
	}, []string{"ThrownExpBottle", "minecraft:xp_bottle"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*IceBomb, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewIceBomb(loc, nil, tag), nil
	}, []string{"minecraft:ice_bomb"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Snowball, error) {
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewSnowball(loc, nil, tag), nil
	}, []string{"Snowball", "minecraft:snowball"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*SplashPotion, error) {
		potionType, ok := item.PotionTypeIdMapInstance.FromID(int(tag.GetShortOr(TagPotionID, nbt.ShortTag(item.PotionTypeIdMapInstance.ToID(item.PotionTypeWater)))))
		if !ok {
			return nil, data.NewSavedDataLoadingError("No such potion type")
		}
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewSplashPotion(loc, nil, potionType, tag), nil
	}, []string{"ThrownPotion", "minecraft:potion", "thrownpotion"})

	entity.RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Trident, error) {
		if _, ok, _ := tag.GetCompoundTag(TagTridentItem); !ok {
			return nil, data.NewSavedDataLoadingError(`Expected "` + TagTridentItem + `" NBT tag not found`)
		}
		itemTag, _, _ := tag.GetCompoundTag(TagTridentItem)
		it, err := item.NbtDeserialize(itemTag)
		if err != nil {
			return nil, err
		}
		if it.IsNull() {
			return nil, data.NewSavedDataLoadingError("Trident item is invalid")
		}
		loc, err := entity.ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewTrident(loc, it, nil, tag), nil
	}, []string{
		"minecraft:trident",        //java
		"minecraft:thrown_trident", //bedrock
		"Trident",                  //backwards compat for people who used #4547 before it was merged, since it was sitting around for 4 years...
		"ThrownTrident",            //as above
	})
}
