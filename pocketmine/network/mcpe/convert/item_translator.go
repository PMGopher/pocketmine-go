package convert

import (
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
)

// noBlockRuntimeID mirrors ItemTranslator::NO_BLOCK_RUNTIME_ID - a technically-valid block runtime
// ID (0) used to mean "this item isn't a blockitem".
const noBlockRuntimeID = 0

// itemTypeNames is a port of the slice of real ItemSerializerDeserializerRegistrar/ItemTypeNames
// this port's item.VanillaItems covers - each item.ItemTypeIds constant mapped to its real Bedrock
// network item string ID. Verified against the vendored bedrock.ItemTypes() table (see
// item_translator_test.go's exhaustive resolution check), not guessed - several diverge from what
// the Go type/constant name alone would suggest, matching real PocketMine-MP's own historical
// naming (kept for save compatibility) versus modern Bedrock's flattened item IDs:
//   - item.STEAK is Bedrock's "minecraft:cooked_beef" (PMMP kept the old "Steak" class name).
//   - item.RAW_FISH/COOKED_FISH are Bedrock's "minecraft:cod"/"minecraft:cooked_cod" (the item
//     PMMP still calls "fish" was renamed to Cod in vanilla around 1.13).
//   - item.RAW_BEEF/RAW_CHICKEN/RAW_MUTTON/RAW_PORKCHOP/RAW_RABBIT/RAW_SALMON have no "raw_"
//     prefix on the network side (Bedrock just calls them beef/chicken/mutton/porkchop/rabbit/
//     salmon - only the cooked forms get a distinguishing prefix-free name of their own).
//   - item.TOTEM is Bedrock's "minecraft:totem_of_undying".
//   - the RECORD_* constants are Bedrock's "minecraft:music_disc_*".
//
// item.CLOWNFISH is deliberately not mapped: Bedrock has no distinct network item for it at all -
// vanilla represents clownfish as a "minecraft:tropical_fish" variant selected by NBT, which this
// port's bare Clownfish type doesn't carry (no tropical-fish-variant data model exists anywhere in
// this port yet) - a documented gap, not an oversight.
var itemTypeNames = map[int]string{
	item.APPLE:                  "minecraft:apple",
	item.BAKED_POTATO:           "minecraft:baked_potato",
	item.BEETROOT_SEEDS:         "minecraft:beetroot_seeds",
	item.BEETROOT_SOUP:          "minecraft:beetroot_soup",
	item.BLAZE_ROD:              "minecraft:blaze_rod",
	item.BONE_MEAL:              "minecraft:bone_meal",
	item.BOOK:                   "minecraft:book",
	item.BOWL:                   "minecraft:bowl",
	item.BREAD:                  "minecraft:bread",
	item.BUCKET:                 "minecraft:bucket",
	item.CARROT:                 "minecraft:carrot",
	item.CHARCOAL:               "minecraft:charcoal",
	item.CLOCK:                  "minecraft:clock",
	item.COAL:                   "minecraft:coal",
	item.COCOA_BEANS:            "minecraft:cocoa_beans",
	item.COMPASS:                "minecraft:compass",
	item.COOKED_CHICKEN:         "minecraft:cooked_chicken",
	item.COOKED_FISH:            "minecraft:cooked_cod",
	item.COOKED_MUTTON:          "minecraft:cooked_mutton",
	item.COOKED_PORKCHOP:        "minecraft:cooked_porkchop",
	item.COOKED_RABBIT:          "minecraft:cooked_rabbit",
	item.COOKED_SALMON:          "minecraft:cooked_salmon",
	item.COOKIE:                 "minecraft:cookie",
	item.DRIED_KELP:             "minecraft:dried_kelp",
	item.DYE:                    "minecraft:dye",
	item.ENCHANTED_BOOK:         "minecraft:enchanted_book",
	item.ENCHANTED_GOLDEN_APPLE: "minecraft:enchanted_golden_apple",
	item.FIREWORK_STAR:          "minecraft:firework_star",
	item.FISHING_ROD:            "minecraft:fishing_rod",
	item.FLINT_AND_STEEL:        "minecraft:flint_and_steel",
	item.GLASS_BOTTLE:           "minecraft:glass_bottle",
	item.GLOW_BERRIES:           "minecraft:glow_berries",
	item.GOAT_HORN:              "minecraft:goat_horn",
	item.GOLDEN_APPLE:           "minecraft:golden_apple",
	item.GOLDEN_CARROT:          "minecraft:golden_carrot",
	item.HONEY_BOTTLE:           "minecraft:honey_bottle",
	item.LINGERING_POTION:       "minecraft:lingering_potion",
	item.MEDICINE:               "minecraft:medicine",
	item.MELON_SEEDS:            "minecraft:melon_seeds",
	item.MILK_BUCKET:            "minecraft:milk_bucket",
	item.MINECART:               "minecraft:minecart",
	item.MUSHROOM_STEW:          "minecraft:mushroom_stew",
	item.PITCHER_POD:            "minecraft:pitcher_pod",
	item.POISONOUS_POTATO:       "minecraft:poisonous_potato",
	item.POTATO:                 "minecraft:potato",
	item.POTION:                 "minecraft:potion",
	item.PUFFERFISH:             "minecraft:pufferfish",
	item.PUMPKIN_PIE:            "minecraft:pumpkin_pie",
	item.PUMPKIN_SEEDS:          "minecraft:pumpkin_seeds",
	item.RABBIT_STEW:            "minecraft:rabbit_stew",
	item.RAW_BEEF:               "minecraft:beef",
	item.RAW_CHICKEN:            "minecraft:chicken",
	item.RAW_FISH:               "minecraft:cod",
	item.RAW_MUTTON:             "minecraft:mutton",
	item.RAW_PORKCHOP:           "minecraft:porkchop",
	item.RAW_RABBIT:             "minecraft:rabbit",
	item.RAW_SALMON:             "minecraft:salmon",
	item.REDSTONE_DUST:          "minecraft:redstone",
	item.ROTTEN_FLESH:           "minecraft:rotten_flesh",
	item.SHEARS:                 "minecraft:shears",
	item.SPIDER_EYE:             "minecraft:spider_eye",
	item.SPLASH_POTION:          "minecraft:splash_potion",
	item.SPYGLASS:               "minecraft:spyglass",
	item.STEAK:                  "minecraft:cooked_beef",
	item.STICK:                  "minecraft:stick",
	item.STRING:                 "minecraft:string",
	item.SUSPICIOUS_STEW:        "minecraft:suspicious_stew",
	item.SWEET_BERRIES:          "minecraft:sweet_berries",
	item.TORCHFLOWER_SEEDS:      "minecraft:torchflower_seeds",
	item.TOTEM:                  "minecraft:totem_of_undying",
	item.TRIDENT:                "minecraft:trident",
	item.WHEAT_SEEDS:            "minecraft:wheat_seeds",
	item.WRITABLE_BOOK:          "minecraft:writable_book",
	item.WRITTEN_BOOK:           "minecraft:written_book",

	item.RECORD_11:                "minecraft:music_disc_11",
	item.RECORD_13:                "minecraft:music_disc_13",
	item.RECORD_5:                 "minecraft:music_disc_5",
	item.RECORD_BLOCKS:            "minecraft:music_disc_blocks",
	item.RECORD_CAT:               "minecraft:music_disc_cat",
	item.RECORD_CHIRP:             "minecraft:music_disc_chirp",
	item.RECORD_CREATOR:           "minecraft:music_disc_creator",
	item.RECORD_CREATOR_MUSIC_BOX: "minecraft:music_disc_creator_music_box",
	item.RECORD_FAR:               "minecraft:music_disc_far",
	item.RECORD_LAVA_CHICKEN:      "minecraft:music_disc_lava_chicken",
	item.RECORD_MALL:              "minecraft:music_disc_mall",
	item.RECORD_MELLOHI:           "minecraft:music_disc_mellohi",
	item.RECORD_OTHERSIDE:         "minecraft:music_disc_otherside",
	item.RECORD_PIGSTEP:           "minecraft:music_disc_pigstep",
	item.RECORD_PRECIPICE:         "minecraft:music_disc_precipice",
	item.RECORD_RELIC:             "minecraft:music_disc_relic",
	item.RECORD_STAL:              "minecraft:music_disc_stal",
	item.RECORD_STRAD:             "minecraft:music_disc_strad",
	item.RECORD_WAIT:              "minecraft:music_disc_wait",
	item.RECORD_WARD:              "minecraft:music_disc_ward",
}

// ItemTranslator is a port of a slice of pocketmine\network\mcpe\convert\ItemTranslator - only the
// item types item.VanillaItems covers so far (see itemTypeNames' own doc comment); block items and
// meta-carrying items (dyes with a real color, potions with a real type) aren't handled yet, since
// none of the item types this port has are either of those (documented scope, not a guess).
type ItemTranslator struct{}

func NewItemTranslator() *ItemTranslator { return &ItemTranslator{} }

// ToNetworkID is a port of ItemTranslator::toNetworkId, returning the network numeric ID, meta
// (always 0 for this slice), and block runtime ID (always noBlockRuntimeID - none of these item
// types carry a blockstate).
func (t *ItemTranslator) ToNetworkID(it item.Item) (networkID int32, meta int16, blockRuntimeID int32, ok bool) {
	name, ok := itemTypeNames[it.GetTypeId()]
	if !ok {
		return 0, 0, 0, false
	}
	networkID, ok = bedrock.ItemRuntimeIDFor(name)
	if !ok {
		return 0, 0, 0, false
	}
	return networkID, 0, noBlockRuntimeID, true
}
