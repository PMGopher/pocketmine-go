package enchantment

// ItemFlags constants, a port of pocketmine\item\enchantment\ItemFlags.
//
// Deprecated (in PHP too): use ItemEnchantmentTags for enchantment applicability.
const (
	ItemFlagNone          = 0x0
	ItemFlagAll           = 0xffff
	ItemFlagArmor         = ItemFlagHead | ItemFlagTorso | ItemFlagLegs | ItemFlagFeet
	ItemFlagHead          = 0x1
	ItemFlagTorso         = 0x2
	ItemFlagLegs          = 0x4
	ItemFlagFeet          = 0x8
	ItemFlagSword         = 0x10
	ItemFlagBow           = 0x20
	ItemFlagTool          = ItemFlagHoe | ItemFlagShears | ItemFlagFlintAndSteel
	ItemFlagHoe           = 0x40
	ItemFlagShears        = 0x80
	ItemFlagFlintAndSteel = 0x100
	ItemFlagDig           = ItemFlagAxe | ItemFlagPickaxe | ItemFlagShovel
	ItemFlagAxe           = 0x200
	ItemFlagPickaxe       = 0x400
	ItemFlagShovel        = 0x800
	ItemFlagFishingRod    = 0x1000
	ItemFlagCarrotStick   = 0x2000
	ItemFlagElytra        = 0x4000
	ItemFlagTrident       = 0x8000
)
