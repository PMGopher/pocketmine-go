package item

import (
	"fmt"
	"sync"

	"pocketmine-go/pocketmine/world/sound"
)

// This file is a port of pocketmine\item\VanillaArmorMaterialsInputs (VanillaArmorMaterials).

var (
	vanillaArmorMaterials     map[string]ArmorMaterial
	vanillaArmorMaterialsOnce sync.Once
)

// VanillaArmorMaterial is VanillaArmorMaterials::NAME().
func VanillaArmorMaterial(name string) ArmorMaterial {
	vanillaArmorMaterialsOnce.Do(func() {
		vanillaArmorMaterials = map[string]ArmorMaterial{
			"leather":   NewArmorMaterial(15, sound.ArmorEquipLeatherSound{}),
			"chainmail": NewArmorMaterial(12, sound.ArmorEquipChainSound{}),
			"copper":    NewArmorMaterial(8, sound.ArmorEquipCopperSound{}),
			"iron":      NewArmorMaterial(9, sound.ArmorEquipIronSound{}),
			"turtle":    NewArmorMaterial(9, sound.ArmorEquipGenericSound{}),
			"gold":      NewArmorMaterial(25, sound.ArmorEquipGoldSound{}),
			"diamond":   NewArmorMaterial(10, sound.ArmorEquipDiamondSound{}),
			"netherite": NewArmorMaterial(15, sound.ArmorEquipNetheriteSound{}),
		}
	})
	m, ok := vanillaArmorMaterials[name]
	if !ok {
		panic(fmt.Sprintf("VanillaArmorMaterials: no material %q", name))
	}
	return m
}
