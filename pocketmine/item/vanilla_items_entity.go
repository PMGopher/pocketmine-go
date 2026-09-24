package item

// These VanillaItems entries are the ones entity classes reference (drops, picked items, pickups
// and particles). Every name and type ID is copied from VanillaItemsInputs.php.
var (
	vanillaInkSac           Item
	vanillaIronIngot        Item
	vanillaArrow            Item
	vanillaEgg              Item
	vanillaSnowball         Item
	vanillaIceBomb          Item
	vanillaEnderPearl       Item
	vanillaExperienceBottle Item
	vanillaPainting         Item
	vanillaEndCrystal       Item
	vanillaZombieSpawnEgg   Item
	vanillaSquidSpawnEgg    Item
	vanillaVillagerSpawnEgg Item
)

func VanillaInkSac() Item {
	if vanillaInkSac == nil {
		vanillaInkSac = NewItem(NewItemIdentifier(INK_SAC), "Ink Sac")
	}
	return vanillaInkSac.Clone()
}

func VanillaIronIngot() Item {
	if vanillaIronIngot == nil {
		vanillaIronIngot = NewItem(NewItemIdentifier(IRON_INGOT), "Iron Ingot")
	}
	return vanillaIronIngot.Clone()
}

func VanillaArrow() Item {
	if vanillaArrow == nil {
		vanillaArrow = NewArrow(NewItemIdentifier(ARROW), "Arrow")
	}
	return vanillaArrow.Clone()
}

func VanillaEgg() Item {
	if vanillaEgg == nil {
		vanillaEgg = NewEgg(NewItemIdentifier(EGG), "Egg")
	}
	return vanillaEgg.Clone()
}

func VanillaSnowball() Item {
	if vanillaSnowball == nil {
		vanillaSnowball = NewSnowball(NewItemIdentifier(SNOWBALL), "Snowball")
	}
	return vanillaSnowball.Clone()
}

func VanillaIceBomb() Item {
	if vanillaIceBomb == nil {
		vanillaIceBomb = NewIceBomb(NewItemIdentifier(ICE_BOMB), "Ice Bomb")
	}
	return vanillaIceBomb.Clone()
}

func VanillaEnderPearl() Item {
	if vanillaEnderPearl == nil {
		vanillaEnderPearl = NewEnderPearl(NewItemIdentifier(ENDER_PEARL), "Ender Pearl")
	}
	return vanillaEnderPearl.Clone()
}

func VanillaExperienceBottle() Item {
	if vanillaExperienceBottle == nil {
		vanillaExperienceBottle = NewExperienceBottle(NewItemIdentifier(EXPERIENCE_BOTTLE), "Bottle o' Enchanting")
	}
	return vanillaExperienceBottle.Clone()
}

func VanillaPainting() Item {
	if vanillaPainting == nil {
		vanillaPainting = NewPaintingItem(NewItemIdentifier(PAINTING), "Painting")
	}
	return vanillaPainting.Clone()
}

func VanillaEndCrystal() Item {
	if vanillaEndCrystal == nil {
		vanillaEndCrystal = NewEndCrystal(NewItemIdentifier(END_CRYSTAL), "End Crystal")
	}
	return vanillaEndCrystal.Clone()
}

func VanillaZombieSpawnEgg() Item {
	if vanillaZombieSpawnEgg == nil {
		vanillaZombieSpawnEgg = NewSpawnEgg(NewItemIdentifier(ZOMBIE_SPAWN_EGG), "Zombie Spawn Egg", "Zombie")
	}
	return vanillaZombieSpawnEgg.Clone()
}

func VanillaSquidSpawnEgg() Item {
	if vanillaSquidSpawnEgg == nil {
		vanillaSquidSpawnEgg = NewSpawnEgg(NewItemIdentifier(SQUID_SPAWN_EGG), "Squid Spawn Egg", "Squid")
	}
	return vanillaSquidSpawnEgg.Clone()
}

func VanillaVillagerSpawnEgg() Item {
	if vanillaVillagerSpawnEgg == nil {
		vanillaVillagerSpawnEgg = NewSpawnEgg(NewItemIdentifier(VILLAGER_SPAWN_EGG), "Villager Spawn Egg", "Villager")
	}
	return vanillaVillagerSpawnEgg.Clone()
}
