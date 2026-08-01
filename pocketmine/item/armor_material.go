package item

import "pocketmine-go/pocketmine/world/sound"

// ArmorMaterial is a port of pocketmine\item\ArmorMaterial. sound.Sound is an interface, so its
// nil value already represents the PHP original's nullable ?Sound with no extra bool needed.
type ArmorMaterial struct {
	Enchantability int
	EquipSound     sound.Sound
}

func NewArmorMaterial(enchantability int, equipSound sound.Sound) ArmorMaterial {
	return ArmorMaterial{Enchantability: enchantability, EquipSound: equipSound}
}

func (m ArmorMaterial) GetEnchantability() int { return m.Enchantability }

func (m ArmorMaterial) GetEquipSound() sound.Sound { return m.EquipSound }
