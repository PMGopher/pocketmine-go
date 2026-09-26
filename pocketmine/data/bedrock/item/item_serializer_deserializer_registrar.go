package bedrockitem

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// ItemSerializerDeserializerRegistrar is a port of
// pocketmine\data\bedrock\item\ItemSerializerDeserializerRegistrar: registers the serializer and
// deserializer of every item with a dedicated item ID. Plain block items don't need one (their block
// state is saved with them).
type ItemSerializerDeserializerRegistrar struct {
	deserializer *ItemDeserializer
	serializer   *ItemSerializer
}

// NewItemSerializerDeserializerRegistrar is a port of ItemSerializerDeserializerRegistrar::__construct
// (either side may be nil).
func NewItemSerializerDeserializerRegistrar(deserializer *ItemDeserializer, serializer *ItemSerializer) *ItemSerializerDeserializerRegistrar {
	r := &ItemSerializerDeserializerRegistrar{deserializer: deserializer, serializer: serializer}
	r.register1to1BlockMappings()
	r.register1to1ItemMappings()
	r.register1to1BlockWithMetaMappings()
	r.register1to1ItemWithMetaMappings()
	r.register1ToNItemMappings()
	r.registerMiscBlockMappings()
	r.registerMiscItemMappings()
	return r
}

// Map1to1Item is a port of ItemSerializerDeserializerRegistrar::map1to1Item.
func (r *ItemSerializerDeserializerRegistrar) Map1to1Item(id string, it item.Item) {
	if r.deserializer != nil {
		r.deserializer.Map(id, func(SavedItemData) item.Item { return it.Clone() })
	}
	if r.serializer != nil {
		r.serializer.Map(it, func(item.Item) SavedItemData { return SavedItemData{Name: id} })
	}
}

// Map1to1ItemWithMeta is a port of ItemSerializerDeserializerRegistrar::map1to1ItemWithMeta.
func Map1to1ItemWithMeta[T item.Item](r *ItemSerializerDeserializerRegistrar, id string, it T, deserializeMeta func(T, int), serializeMeta func(T) int) {
	if r.deserializer != nil {
		r.deserializer.Map(id, func(data SavedItemData) item.Item {
			result := it.Clone().(T)
			deserializeMeta(result, data.Meta)
			return result
		})
	}
	if r.serializer != nil {
		r.serializer.Map(it, func(i item.Item) SavedItemData { return SavedItemData{Name: id, Meta: serializeMeta(i.(T))} })
	}
}

// Map1to1Block is a port of ItemSerializerDeserializerRegistrar::map1to1Block.
func (r *ItemSerializerDeserializerRegistrar) Map1to1Block(id string, blk block.Behavior) {
	if r.deserializer != nil {
		r.deserializer.MapBlock(id, func(SavedItemData) block.Behavior { return blk.Clone() })
	}
	if r.serializer != nil {
		r.serializer.MapBlock(blk, func(block.Behavior) SavedItemData { return SavedItemData{Name: id} })
	}
}

// Map1to1BlockWithMeta is a port of ItemSerializerDeserializerRegistrar::map1to1BlockWithMeta.
func Map1to1BlockWithMeta[T block.Behavior](r *ItemSerializerDeserializerRegistrar, id string, blk T, deserializeMeta func(T, int), serializeMeta func(T) int) {
	if r.deserializer != nil {
		r.deserializer.MapBlock(id, func(data SavedItemData) block.Behavior {
			result := blk.Clone().(T)
			deserializeMeta(result, data.Meta)
			return result
		})
	}
	if r.serializer != nil {
		r.serializer.MapBlock(blk, func(b block.Behavior) SavedItemData { return SavedItemData{Name: id, Meta: serializeMeta(b.(T))} })
	}
}

// Map1ToNItem is a port of ItemSerializerDeserializerRegistrar::map1ToNItem.
func (r *ItemSerializerDeserializerRegistrar) Map1ToNItem(id string, items map[int]item.Item) {
	if r.deserializer != nil {
		r.deserializer.Map(id, func(data SavedItemData) item.Item {
			result, ok := items[data.Meta]
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unhandled meta value %d for item ID %s", data.Meta, data.Name)})
			}
			return result.Clone()
		})
	}
	if r.serializer != nil {
		for meta, it := range items {
			r.serializer.Map(it, func(item.Item) SavedItemData { return SavedItemData{Name: id, Meta: meta} })
		}
	}
}

func v(name string) item.Item { return item.VanillaItem(name) }

func b(name string) block.Behavior { return block.VanillaBlock(name) }

// register1to1BlockMappings registers mappings for item IDs which directly correspond to
// PocketMine-MP blockitems. Mappings here are only necessary when the item has a dedicated item
// ID; in these cases, the blockstate is not included in the itemstack, and the item ID may be
// different from the block ID.
func (r *ItemSerializerDeserializerRegistrar) register1to1BlockMappings() {
	r.Map1to1Block(ACACIA_DOOR, b("acacia_door"))
	r.Map1to1Block(BAMBOO_DOOR, b("bamboo_door"))
	r.Map1to1Block(BIRCH_DOOR, b("birch_door"))
	r.Map1to1Block(BREWING_STAND, b("brewing_stand"))
	r.Map1to1Block(CAKE, b("cake"))
	r.Map1to1Block(CAMPFIRE, b("campfire"))
	r.Map1to1Block(CAULDRON, b("cauldron"))
	r.Map1to1Block(CHERRY_DOOR, b("cherry_door"))
	r.Map1to1Block(COMPARATOR, b("redstone_comparator"))
	r.Map1to1Block(CRIMSON_DOOR, b("crimson_door"))
	r.Map1to1Block(DARK_OAK_DOOR, b("dark_oak_door"))
	r.Map1to1Block(FLOWER_POT, b("flower_pot"))
	r.Map1to1Block(FRAME, b("item_frame"))
	r.Map1to1Block(GLOW_FRAME, b("glowing_item_frame"))
	r.Map1to1Block(HOPPER, b("hopper"))
	r.Map1to1Block(IRON_DOOR, b("iron_door"))
	r.Map1to1Block(JUNGLE_DOOR, b("jungle_door"))
	r.Map1to1Block(MANGROVE_DOOR, b("mangrove_door"))
	r.Map1to1Block(NETHER_SPROUTS, b("nether_sprouts"))
	r.Map1to1Block(NETHER_WART, b("nether_wart"))
	r.Map1to1Block(PALE_OAK_DOOR, b("pale_oak_door"))
	r.Map1to1Block(REPEATER, b("redstone_repeater"))
	r.Map1to1Block(SOUL_CAMPFIRE, b("soul_campfire"))
	r.Map1to1Block(SPRUCE_DOOR, b("spruce_door"))
	r.Map1to1Block(SUGAR_CANE, b("sugarcane"))
	r.Map1to1Block(WARPED_DOOR, b("warped_door"))
	r.Map1to1Block(WOODEN_DOOR, b("oak_door"))
}

// register1to1ItemMappings registers mappings for item IDs which directly correspond to PocketMine-MP items.
func (r *ItemSerializerDeserializerRegistrar) register1to1ItemMappings() {
	r.Map1to1Item(ACACIA_BOAT, v("acacia_boat"))
	r.Map1to1Item(ACACIA_HANGING_SIGN, v("acacia_hanging_sign"))
	r.Map1to1Item(ACACIA_SIGN, v("acacia_sign"))
	r.Map1to1Item(AMETHYST_SHARD, v("amethyst_shard"))
	r.Map1to1Item(APPLE, v("apple"))
	r.Map1to1Item(BAKED_POTATO, v("baked_potato"))
	r.Map1to1Item(BAMBOO_HANGING_SIGN, v("bamboo_hanging_sign"))
	r.Map1to1Item(BAMBOO_SIGN, v("bamboo_sign"))
	r.Map1to1Item(BEEF, v("raw_beef"))
	r.Map1to1Item(BEETROOT, v("beetroot"))
	r.Map1to1Item(BEETROOT_SEEDS, v("beetroot_seeds"))
	r.Map1to1Item(BEETROOT_SOUP, v("beetroot_soup"))
	r.Map1to1Item(BIRCH_BOAT, v("birch_boat"))
	r.Map1to1Item(BIRCH_HANGING_SIGN, v("birch_hanging_sign"))
	r.Map1to1Item(BIRCH_SIGN, v("birch_sign"))
	r.Map1to1Item(BLAZE_POWDER, v("blaze_powder"))
	r.Map1to1Item(BLAZE_ROD, v("blaze_rod"))
	r.Map1to1Item(BLEACH, v("bleach"))
	r.Map1to1Item(BONE, v("bone"))
	r.Map1to1Item(BONE_MEAL, v("bone_meal"))
	r.Map1to1Item(BOOK, v("book"))
	r.Map1to1Item(BOW, v("bow"))
	r.Map1to1Item(BOWL, v("bowl"))
	r.Map1to1Item(BREAD, v("bread"))
	r.Map1to1Item(BRICK, v("brick"))
	r.Map1to1Item(BUCKET, v("bucket"))
	r.Map1to1Item(CARROT, v("carrot"))
	r.Map1to1Item(CHAINMAIL_BOOTS, v("chainmail_boots"))
	r.Map1to1Item(CHAINMAIL_CHESTPLATE, v("chainmail_chestplate"))
	r.Map1to1Item(CHAINMAIL_HELMET, v("chainmail_helmet"))
	r.Map1to1Item(CHAINMAIL_LEGGINGS, v("chainmail_leggings"))
	r.Map1to1Item(CHARCOAL, v("charcoal"))
	r.Map1to1Item(CHERRY_HANGING_SIGN, v("cherry_hanging_sign"))
	r.Map1to1Item(CHERRY_SIGN, v("cherry_sign"))
	r.Map1to1Item(CHICKEN, v("raw_chicken"))
	r.Map1to1Item(CHORUS_FRUIT, v("chorus_fruit"))
	r.Map1to1Item(CLAY_BALL, v("clay"))
	r.Map1to1Item(CLOCK, v("clock"))
	r.Map1to1Item(COAL, v("coal"))
	r.Map1to1Item(COAST_ARMOR_TRIM_SMITHING_TEMPLATE, v("coast_armor_trim_smithing_template"))
	r.Map1to1Item(COCOA_BEANS, v("cocoa_beans"))
	r.Map1to1Item(COD, v("raw_fish"))
	r.Map1to1Item(COMPASS, v("compass"))
	r.Map1to1Item(COOKED_BEEF, v("steak"))
	r.Map1to1Item(COOKED_CHICKEN, v("cooked_chicken"))
	r.Map1to1Item(COOKED_COD, v("cooked_fish"))
	r.Map1to1Item(COOKED_MUTTON, v("cooked_mutton"))
	r.Map1to1Item(COOKED_PORKCHOP, v("cooked_porkchop"))
	r.Map1to1Item(COOKED_RABBIT, v("cooked_rabbit"))
	r.Map1to1Item(COOKED_SALMON, v("cooked_salmon"))
	r.Map1to1Item(COOKIE, v("cookie"))
	r.Map1to1Item(COPPER_AXE, v("copper_axe"))
	r.Map1to1Item(COPPER_BOOTS, v("copper_boots"))
	r.Map1to1Item(COPPER_CHESTPLATE, v("copper_chestplate"))
	r.Map1to1Item(COPPER_HELMET, v("copper_helmet"))
	r.Map1to1Item(COPPER_HOE, v("copper_hoe"))
	r.Map1to1Item(COPPER_INGOT, v("copper_ingot"))
	r.Map1to1Item(COPPER_LEGGINGS, v("copper_leggings"))
	r.Map1to1Item(COPPER_NUGGET, v("copper_nugget"))
	r.Map1to1Item(COPPER_PICKAXE, v("copper_pickaxe"))
	r.Map1to1Item(COPPER_SHOVEL, v("copper_shovel"))
	r.Map1to1Item(COPPER_SWORD, v("copper_sword"))
	r.Map1to1Item(CRIMSON_HANGING_SIGN, v("crimson_hanging_sign"))
	r.Map1to1Item(CRIMSON_SIGN, v("crimson_sign"))
	r.Map1to1Item(DARK_OAK_BOAT, v("dark_oak_boat"))
	r.Map1to1Item(DARK_OAK_HANGING_SIGN, v("dark_oak_hanging_sign"))
	r.Map1to1Item(DARK_OAK_SIGN, v("dark_oak_sign"))
	r.Map1to1Item(DIAMOND, v("diamond"))
	r.Map1to1Item(DIAMOND_AXE, v("diamond_axe"))
	r.Map1to1Item(DIAMOND_BOOTS, v("diamond_boots"))
	r.Map1to1Item(DIAMOND_CHESTPLATE, v("diamond_chestplate"))
	r.Map1to1Item(DIAMOND_HELMET, v("diamond_helmet"))
	r.Map1to1Item(DIAMOND_HOE, v("diamond_hoe"))
	r.Map1to1Item(DIAMOND_LEGGINGS, v("diamond_leggings"))
	r.Map1to1Item(DIAMOND_PICKAXE, v("diamond_pickaxe"))
	r.Map1to1Item(DIAMOND_SHOVEL, v("diamond_shovel"))
	r.Map1to1Item(DIAMOND_SWORD, v("diamond_sword"))
	r.Map1to1Item(DISC_FRAGMENT_5, v("disc_fragment_5"))
	r.Map1to1Item(DRAGON_BREATH, v("dragon_breath"))
	r.Map1to1Item(DRIED_KELP, v("dried_kelp"))
	r.Map1to1Item(DUNE_ARMOR_TRIM_SMITHING_TEMPLATE, v("dune_armor_trim_smithing_template"))
	r.Map1to1Item(ECHO_SHARD, v("echo_shard"))
	r.Map1to1Item(EGG, v("egg"))
	r.Map1to1Item(EMERALD, v("emerald"))
	r.Map1to1Item(ENCHANTED_BOOK, v("enchanted_book"))
	r.Map1to1Item(ENCHANTED_GOLDEN_APPLE, v("enchanted_golden_apple"))
	r.Map1to1Item(END_CRYSTAL, v("end_crystal"))
	r.Map1to1Item(ENDER_PEARL, v("ender_pearl"))
	r.Map1to1Item(EXPERIENCE_BOTTLE, v("experience_bottle"))
	r.Map1to1Item(EYE_ARMOR_TRIM_SMITHING_TEMPLATE, v("eye_armor_trim_smithing_template"))
	r.Map1to1Item(FEATHER, v("feather"))
	r.Map1to1Item(FERMENTED_SPIDER_EYE, v("fermented_spider_eye"))
	r.Map1to1Item(FIREWORK_ROCKET, v("firework_rocket"))
	r.Map1to1Item(FIRE_CHARGE, v("fire_charge"))
	r.Map1to1Item(FISHING_ROD, v("fishing_rod"))
	r.Map1to1Item(FLINT, v("flint"))
	r.Map1to1Item(FLINT_AND_STEEL, v("flint_and_steel"))
	r.Map1to1Item(GHAST_TEAR, v("ghast_tear"))
	r.Map1to1Item(GLASS_BOTTLE, v("glass_bottle"))
	r.Map1to1Item(GLISTERING_MELON_SLICE, v("glistering_melon"))
	r.Map1to1Item(GLOW_BERRIES, v("glow_berries"))
	r.Map1to1Item(GLOW_INK_SAC, v("glow_ink_sac"))
	r.Map1to1Item(GLOWSTONE_DUST, v("glowstone_dust"))
	r.Map1to1Item(GOLD_INGOT, v("gold_ingot"))
	r.Map1to1Item(GOLD_NUGGET, v("gold_nugget"))
	r.Map1to1Item(GOLDEN_APPLE, v("golden_apple"))
	r.Map1to1Item(GOLDEN_AXE, v("golden_axe"))
	r.Map1to1Item(GOLDEN_BOOTS, v("golden_boots"))
	r.Map1to1Item(GOLDEN_CARROT, v("golden_carrot"))
	r.Map1to1Item(GOLDEN_CHESTPLATE, v("golden_chestplate"))
	r.Map1to1Item(GOLDEN_HELMET, v("golden_helmet"))
	r.Map1to1Item(GOLDEN_HOE, v("golden_hoe"))
	r.Map1to1Item(GOLDEN_LEGGINGS, v("golden_leggings"))
	r.Map1to1Item(GOLDEN_PICKAXE, v("golden_pickaxe"))
	r.Map1to1Item(GOLDEN_SHOVEL, v("golden_shovel"))
	r.Map1to1Item(GOLDEN_SWORD, v("golden_sword"))
	r.Map1to1Item(GUNPOWDER, v("gunpowder"))
	r.Map1to1Item(HEART_OF_THE_SEA, v("heart_of_the_sea"))
	r.Map1to1Item(HONEY_BOTTLE, v("honey_bottle"))
	r.Map1to1Item(HONEYCOMB, v("honeycomb"))
	r.Map1to1Item(HOST_ARMOR_TRIM_SMITHING_TEMPLATE, v("host_armor_trim_smithing_template"))
	r.Map1to1Item(ICE_BOMB, v("ice_bomb"))
	r.Map1to1Item(INK_SAC, v("ink_sac"))
	r.Map1to1Item(IRON_AXE, v("iron_axe"))
	r.Map1to1Item(IRON_BOOTS, v("iron_boots"))
	r.Map1to1Item(IRON_CHESTPLATE, v("iron_chestplate"))
	r.Map1to1Item(IRON_HELMET, v("iron_helmet"))
	r.Map1to1Item(IRON_HOE, v("iron_hoe"))
	r.Map1to1Item(IRON_INGOT, v("iron_ingot"))
	r.Map1to1Item(IRON_LEGGINGS, v("iron_leggings"))
	r.Map1to1Item(IRON_NUGGET, v("iron_nugget"))
	r.Map1to1Item(IRON_PICKAXE, v("iron_pickaxe"))
	r.Map1to1Item(IRON_SHOVEL, v("iron_shovel"))
	r.Map1to1Item(IRON_SWORD, v("iron_sword"))
	r.Map1to1Item(JUNGLE_BOAT, v("jungle_boat"))
	r.Map1to1Item(JUNGLE_HANGING_SIGN, v("jungle_hanging_sign"))
	r.Map1to1Item(JUNGLE_SIGN, v("jungle_sign"))
	r.Map1to1Item(LAPIS_LAZULI, v("lapis_lazuli"))
	r.Map1to1Item(LAVA_BUCKET, v("lava_bucket"))
	r.Map1to1Item(LEATHER, v("leather"))
	r.Map1to1Item(LEATHER_BOOTS, v("leather_boots"))
	r.Map1to1Item(LEATHER_CHESTPLATE, v("leather_tunic"))
	r.Map1to1Item(LEATHER_HELMET, v("leather_cap"))
	r.Map1to1Item(LEATHER_LEGGINGS, v("leather_pants"))
	r.Map1to1Item(MAGMA_CREAM, v("magma_cream"))
	r.Map1to1Item(MANGROVE_BOAT, v("mangrove_boat"))
	r.Map1to1Item(MANGROVE_HANGING_SIGN, v("mangrove_hanging_sign"))
	r.Map1to1Item(MANGROVE_SIGN, v("mangrove_sign"))
	r.Map1to1Item(MELON_SEEDS, v("melon_seeds"))
	r.Map1to1Item(MELON_SLICE, v("melon"))
	r.Map1to1Item(MILK_BUCKET, v("milk_bucket"))
	r.Map1to1Item(MINECART, v("minecart"))
	r.Map1to1Item(MUSHROOM_STEW, v("mushroom_stew"))
	r.Map1to1Item(MUSIC_DISC_11, v("record_11"))
	r.Map1to1Item(MUSIC_DISC_13, v("record_13"))
	r.Map1to1Item(MUSIC_DISC_5, v("record_5"))
	r.Map1to1Item(MUSIC_DISC_BLOCKS, v("record_blocks"))
	r.Map1to1Item(MUSIC_DISC_CAT, v("record_cat"))
	r.Map1to1Item(MUSIC_DISC_CHIRP, v("record_chirp"))
	r.Map1to1Item(MUSIC_DISC_CREATOR, v("record_creator"))
	r.Map1to1Item(MUSIC_DISC_CREATOR_MUSIC_BOX, v("record_creator_music_box"))
	r.Map1to1Item(MUSIC_DISC_FAR, v("record_far"))
	r.Map1to1Item(MUSIC_DISC_LAVA_CHICKEN, v("record_lava_chicken"))
	r.Map1to1Item(MUSIC_DISC_MALL, v("record_mall"))
	r.Map1to1Item(MUSIC_DISC_MELLOHI, v("record_mellohi"))
	r.Map1to1Item(MUSIC_DISC_OTHERSIDE, v("record_otherside"))
	r.Map1to1Item(MUSIC_DISC_PIGSTEP, v("record_pigstep"))
	r.Map1to1Item(MUSIC_DISC_PRECIPICE, v("record_precipice"))
	r.Map1to1Item(MUSIC_DISC_RELIC, v("record_relic"))
	r.Map1to1Item(MUSIC_DISC_STAL, v("record_stal"))
	r.Map1to1Item(MUSIC_DISC_STRAD, v("record_strad"))
	r.Map1to1Item(MUSIC_DISC_WAIT, v("record_wait"))
	r.Map1to1Item(MUSIC_DISC_WARD, v("record_ward"))
	r.Map1to1Item(MUTTON, v("raw_mutton"))
	r.Map1to1Item(NAME_TAG, v("name_tag"))
	r.Map1to1Item(NAUTILUS_SHELL, v("nautilus_shell"))
	r.Map1to1Item(NETHER_STAR, v("nether_star"))
	r.Map1to1Item(NETHERBRICK, v("nether_brick"))
	r.Map1to1Item(NETHERITE_AXE, v("netherite_axe"))
	r.Map1to1Item(NETHERITE_BOOTS, v("netherite_boots"))
	r.Map1to1Item(NETHERITE_CHESTPLATE, v("netherite_chestplate"))
	r.Map1to1Item(NETHERITE_HELMET, v("netherite_helmet"))
	r.Map1to1Item(NETHERITE_HOE, v("netherite_hoe"))
	r.Map1to1Item(NETHERITE_INGOT, v("netherite_ingot"))
	r.Map1to1Item(NETHERITE_LEGGINGS, v("netherite_leggings"))
	r.Map1to1Item(NETHERITE_PICKAXE, v("netherite_pickaxe"))
	r.Map1to1Item(NETHERITE_SCRAP, v("netherite_scrap"))
	r.Map1to1Item(NETHERITE_SHOVEL, v("netherite_shovel"))
	r.Map1to1Item(NETHERITE_SWORD, v("netherite_sword"))
	r.Map1to1Item(NETHERITE_UPGRADE_SMITHING_TEMPLATE, v("netherite_upgrade_smithing_template"))
	r.Map1to1Item(OAK_BOAT, v("oak_boat"))
	r.Map1to1Item(OAK_HANGING_SIGN, v("oak_hanging_sign"))
	r.Map1to1Item(OAK_SIGN, v("oak_sign"))
	r.Map1to1Item(PAINTING, v("painting"))
	r.Map1to1Item(PALE_OAK_HANGING_SIGN, v("pale_oak_hanging_sign"))
	r.Map1to1Item(PALE_OAK_SIGN, v("pale_oak_sign"))
	r.Map1to1Item(PAPER, v("paper"))
	r.Map1to1Item(PHANTOM_MEMBRANE, v("phantom_membrane"))
	r.Map1to1Item(PITCHER_POD, v("pitcher_pod"))
	r.Map1to1Item(POISONOUS_POTATO, v("poisonous_potato"))
	r.Map1to1Item(POPPED_CHORUS_FRUIT, v("popped_chorus_fruit"))
	r.Map1to1Item(PORKCHOP, v("raw_porkchop"))
	r.Map1to1Item(POTATO, v("potato"))
	r.Map1to1Item(PRISMARINE_CRYSTALS, v("prismarine_crystals"))
	r.Map1to1Item(PRISMARINE_SHARD, v("prismarine_shard"))
	r.Map1to1Item(PUFFERFISH, v("pufferfish"))
	r.Map1to1Item(PUMPKIN_PIE, v("pumpkin_pie"))
	r.Map1to1Item(PUMPKIN_SEEDS, v("pumpkin_seeds"))
	r.Map1to1Item(QUARTZ, v("nether_quartz"))
	r.Map1to1Item(RABBIT, v("raw_rabbit"))
	r.Map1to1Item(RABBIT_FOOT, v("rabbit_foot"))
	r.Map1to1Item(RABBIT_HIDE, v("rabbit_hide"))
	r.Map1to1Item(RABBIT_STEW, v("rabbit_stew"))
	r.Map1to1Item(RAISER_ARMOR_TRIM_SMITHING_TEMPLATE, v("raiser_armor_trim_smithing_template"))
	r.Map1to1Item(RAW_COPPER, v("raw_copper"))
	r.Map1to1Item(RAW_GOLD, v("raw_gold"))
	r.Map1to1Item(RAW_IRON, v("raw_iron"))
	r.Map1to1Item(RECOVERY_COMPASS, v("recovery_compass"))
	r.Map1to1Item(REDSTONE, v("redstone_dust"))
	r.Map1to1Item(RESIN_BRICK, v("resin_brick"))
	r.Map1to1Item(RIB_ARMOR_TRIM_SMITHING_TEMPLATE, v("rib_armor_trim_smithing_template"))
	r.Map1to1Item(ROTTEN_FLESH, v("rotten_flesh"))
	r.Map1to1Item(SALMON, v("raw_salmon"))
	r.Map1to1Item(TURTLE_SCUTE, v("scute"))
	r.Map1to1Item(SENTRY_ARMOR_TRIM_SMITHING_TEMPLATE, v("sentry_armor_trim_smithing_template"))
	r.Map1to1Item(SHAPER_ARMOR_TRIM_SMITHING_TEMPLATE, v("shaper_armor_trim_smithing_template"))
	r.Map1to1Item(SHEARS, v("shears"))
	r.Map1to1Item(SHULKER_SHELL, v("shulker_shell"))
	r.Map1to1Item(SILENCE_ARMOR_TRIM_SMITHING_TEMPLATE, v("silence_armor_trim_smithing_template"))
	r.Map1to1Item(SLIME_BALL, v("slimeball"))
	r.Map1to1Item(SNOUT_ARMOR_TRIM_SMITHING_TEMPLATE, v("snout_armor_trim_smithing_template"))
	r.Map1to1Item(SNOWBALL, v("snowball"))
	r.Map1to1Item(SPIDER_EYE, v("spider_eye"))
	r.Map1to1Item(SPIRE_ARMOR_TRIM_SMITHING_TEMPLATE, v("spire_armor_trim_smithing_template"))
	r.Map1to1Item(SPRUCE_BOAT, v("spruce_boat"))
	r.Map1to1Item(SPRUCE_HANGING_SIGN, v("spruce_hanging_sign"))
	r.Map1to1Item(SPRUCE_SIGN, v("spruce_sign"))
	r.Map1to1Item(SPYGLASS, v("spyglass"))
	r.Map1to1Item(SQUID_SPAWN_EGG, v("squid_spawn_egg"))
	r.Map1to1Item(STICK, v("stick"))
	r.Map1to1Item(STONE_AXE, v("stone_axe"))
	r.Map1to1Item(STONE_HOE, v("stone_hoe"))
	r.Map1to1Item(STONE_PICKAXE, v("stone_pickaxe"))
	r.Map1to1Item(STONE_SHOVEL, v("stone_shovel"))
	r.Map1to1Item(STONE_SWORD, v("stone_sword"))
	r.Map1to1Item(STRING, v("string"))
	r.Map1to1Item(SUGAR, v("sugar"))
	r.Map1to1Item(SWEET_BERRIES, v("sweet_berries"))
	r.Map1to1Item(TORCHFLOWER_SEEDS, v("torchflower_seeds"))
	r.Map1to1Item(TIDE_ARMOR_TRIM_SMITHING_TEMPLATE, v("tide_armor_trim_smithing_template"))
	r.Map1to1Item(TOTEM_OF_UNDYING, v("totem"))
	r.Map1to1Item(TRIDENT, v("trident"))
	r.Map1to1Item(TROPICAL_FISH, v("clownfish"))
	r.Map1to1Item(TURTLE_HELMET, v("turtle_helmet"))
	r.Map1to1Item(VEX_ARMOR_TRIM_SMITHING_TEMPLATE, v("vex_armor_trim_smithing_template"))
	r.Map1to1Item(VILLAGER_SPAWN_EGG, v("villager_spawn_egg"))
	r.Map1to1Item(WARD_ARMOR_TRIM_SMITHING_TEMPLATE, v("ward_armor_trim_smithing_template"))
	r.Map1to1Item(WARPED_HANGING_SIGN, v("warped_hanging_sign"))
	r.Map1to1Item(WARPED_SIGN, v("warped_sign"))
	r.Map1to1Item(WATER_BUCKET, v("water_bucket"))
	r.Map1to1Item(WAYFINDER_ARMOR_TRIM_SMITHING_TEMPLATE, v("wayfinder_armor_trim_smithing_template"))
	r.Map1to1Item(WHEAT, v("wheat"))
	r.Map1to1Item(WHEAT_SEEDS, v("wheat_seeds"))
	r.Map1to1Item(WILD_ARMOR_TRIM_SMITHING_TEMPLATE, v("wild_armor_trim_smithing_template"))
	r.Map1to1Item(WOODEN_AXE, v("wooden_axe"))
	r.Map1to1Item(WOODEN_HOE, v("wooden_hoe"))
	r.Map1to1Item(WOODEN_PICKAXE, v("wooden_pickaxe"))
	r.Map1to1Item(WOODEN_SHOVEL, v("wooden_shovel"))
	r.Map1to1Item(WOODEN_SWORD, v("wooden_sword"))
	r.Map1to1Item(WRITABLE_BOOK, v("writable_book"))
	r.Map1to1Item(WRITTEN_BOOK, v("written_book"))
	r.Map1to1Item(ZOMBIE_SPAWN_EGG, v("zombie_spawn_egg"))
}

func (r *ItemSerializerDeserializerRegistrar) register1ToNItemMappings() {
	r.Map1ToNItem(ARROW, map[int]item.Item{
		0: v("arrow"),
		//TODO: tipped arrows
	})
	r.Map1ToNItem(COMPOUND, map[int]item.Item{
		0:  v("chemical_salt"),                // CompoundTypeIds::SALT
		1:  v("chemical_sodium_oxide"),        // CompoundTypeIds::SODIUM_OXIDE
		2:  v("chemical_sodium_hydroxide"),    // CompoundTypeIds::SODIUM_HYDROXIDE
		3:  v("chemical_magnesium_nitrate"),   // CompoundTypeIds::MAGNESIUM_NITRATE
		4:  v("chemical_iron_sulphide"),       // CompoundTypeIds::IRON_SULPHIDE
		5:  v("chemical_lithium_hydride"),     // CompoundTypeIds::LITHIUM_HYDRIDE
		6:  v("chemical_sodium_hydride"),      // CompoundTypeIds::SODIUM_HYDRIDE
		7:  v("chemical_calcium_bromide"),     // CompoundTypeIds::CALCIUM_BROMIDE
		8:  v("chemical_magnesium_oxide"),     // CompoundTypeIds::MAGNESIUM_OXIDE
		9:  v("chemical_sodium_acetate"),      // CompoundTypeIds::SODIUM_ACETATE
		10: v("chemical_luminol"),             // CompoundTypeIds::LUMINOL
		11: v("chemical_charcoal"),            // CompoundTypeIds::CHARCOAL
		12: v("chemical_sugar"),               // CompoundTypeIds::SUGAR
		13: v("chemical_aluminium_oxide"),     // CompoundTypeIds::ALUMINIUM_OXIDE
		14: v("chemical_boron_trioxide"),      // CompoundTypeIds::BORON_TRIOXIDE
		15: v("chemical_soap"),                // CompoundTypeIds::SOAP
		16: v("chemical_polyethylene"),        // CompoundTypeIds::POLYETHYLENE
		17: v("chemical_rubbish"),             // CompoundTypeIds::RUBBISH
		18: v("chemical_magnesium_salts"),     // CompoundTypeIds::MAGNESIUM_SALTS
		19: v("chemical_sulphate"),            // CompoundTypeIds::SULPHATE
		20: v("chemical_barium_sulphate"),     // CompoundTypeIds::BARIUM_SULPHATE
		21: v("chemical_potassium_chloride"),  // CompoundTypeIds::POTASSIUM_CHLORIDE
		22: v("chemical_mercuric_chloride"),   // CompoundTypeIds::MERCURIC_CHLORIDE
		23: v("chemical_cerium_chloride"),     // CompoundTypeIds::CERIUM_CHLORIDE
		24: v("chemical_tungsten_chloride"),   // CompoundTypeIds::TUNGSTEN_CHLORIDE
		25: v("chemical_calcium_chloride"),    // CompoundTypeIds::CALCIUM_CHLORIDE
		26: v("chemical_water"),               // CompoundTypeIds::WATER
		27: v("chemical_glue"),                // CompoundTypeIds::GLUE
		28: v("chemical_hypochlorite"),        // CompoundTypeIds::HYPOCHLORITE
		29: v("chemical_crude_oil"),           // CompoundTypeIds::CRUDE_OIL
		30: v("chemical_latex"),               // CompoundTypeIds::LATEX
		31: v("chemical_potassium_iodide"),    // CompoundTypeIds::POTASSIUM_IODIDE
		32: v("chemical_sodium_fluoride"),     // CompoundTypeIds::SODIUM_FLUORIDE
		33: v("chemical_benzene"),             // CompoundTypeIds::BENZENE
		34: v("chemical_ink"),                 // CompoundTypeIds::INK
		35: v("chemical_hydrogen_peroxide"),   // CompoundTypeIds::HYDROGEN_PEROXIDE
		36: v("chemical_ammonia"),             // CompoundTypeIds::AMMONIA
		37: v("chemical_sodium_hypochlorite"), // CompoundTypeIds::SODIUM_HYPOCHLORITE
	})
}

// Save IDs of the meta-carrying item types: ports of GoatHornTypeIdMap, MedicineTypeIdMap and
// SuspiciousStewTypeIdMap (with GoatHornTypeIds, MedicineTypeIds and SuspiciousStewTypeIds). They
// live here because their enums are in the item package, which imports data/bedrock.
var (
	goatHornTypeIdMap = func() *bedrock.IntSaveIdMap[item.GoatHornType] {
		m := bedrock.NewIntSaveIdMap[item.GoatHornType]()
		for id, t := range []item.GoatHornType{item.GoatHornTypePonder, item.GoatHornTypeSing, item.GoatHornTypeSeek, item.GoatHornTypeFeel, item.GoatHornTypeAdmire, item.GoatHornTypeCall, item.GoatHornTypeYearn, item.GoatHornTypeDream} {
			m.Register(id, t)
		}
		return m
	}()
	medicineTypeIdMap = func() *bedrock.IntSaveIdMap[item.MedicineType] {
		m := bedrock.NewIntSaveIdMap[item.MedicineType]()
		m.Register(0, item.MedicineTypeEyeDrops)
		m.Register(1, item.MedicineTypeTonic)
		m.Register(2, item.MedicineTypeAntidote)
		m.Register(3, item.MedicineTypeElixir)
		return m
	}()
	suspiciousStewTypeIdMap = func() *bedrock.IntSaveIdMap[item.SuspiciousStewType] {
		m := bedrock.NewIntSaveIdMap[item.SuspiciousStewType]()
		for id, t := range []item.SuspiciousStewType{item.SuspiciousStewTypePoppy, item.SuspiciousStewTypeCornflower, item.SuspiciousStewTypeTulip, item.SuspiciousStewTypeAzureBluet, item.SuspiciousStewTypeLilyOfTheValley, item.SuspiciousStewTypeDandelion, item.SuspiciousStewTypeBlueOrchid, item.SuspiciousStewTypeAllium, item.SuspiciousStewTypeOxeyeDaisy, item.SuspiciousStewTypeWitherRose} {
			m.Register(id, t)
		}
		return m
	}()
)

// register1to1BlockWithMetaMappings registers mappings for item IDs which map to single
// blockitems, and have meta values that alter their properties.
func (r *ItemSerializerDeserializerRegistrar) register1to1BlockWithMetaMappings() {
	Map1to1BlockWithMeta(
		r,
		BED,
		b("bed").(*block.Bed),
		func(blk *block.Bed, meta int) {
			color, ok := bedrock.GetDyeColorIdMap().FromID(meta)
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown bed color ID %d", meta)})
			}
			blk.SetColor(color)
		},
		func(blk *block.Bed) int { return bedrock.GetDyeColorIdMap().ToID(blk.GetColor()) },
	)
}

// register1to1ItemWithMetaMappings registers mappings for item IDs which map to single items, and
// have meta values that alter their properties.
func (r *ItemSerializerDeserializerRegistrar) register1to1ItemWithMetaMappings() {
	Map1to1ItemWithMeta(
		r,
		FIREWORK_STAR,
		v("firework_star").(*item.FireworkStar),
		func(it *item.FireworkStar, meta int) {
			// Colors will be defined by CompoundTag deserialization.
		},
		func(it *item.FireworkStar) int {
			return bedrock.GetDyeColorIdMap().ToInvertedID(it.GetExplosion().GetFlashColor())
		},
	)
	Map1to1ItemWithMeta(
		r,
		GOAT_HORN,
		v("goat_horn").(*item.GoatHorn),
		func(it *item.GoatHorn, meta int) {
			t, ok := goatHornTypeIdMap.FromID(meta)
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown goat horn type ID %d", meta)})
			}
			it.SetHornType(t)
		},
		func(it *item.GoatHorn) int { return goatHornTypeIdMap.ToID(it.GetHornType()) },
	)
	potionType := func(meta int) item.PotionType {
		t, ok := item.PotionTypeIdMapInstance.FromID(meta)
		if !ok {
			panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown potion type ID %d", meta)})
		}
		return t
	}
	Map1to1ItemWithMeta(
		r,
		LINGERING_POTION,
		v("lingering_potion").(*item.SplashPotion),
		func(it *item.SplashPotion, meta int) { it.SetType(potionType(meta)) },
		func(it *item.SplashPotion) int { return item.PotionTypeIdMapInstance.ToID(it.GetType()) },
	)
	Map1to1ItemWithMeta(
		r,
		MEDICINE,
		v("medicine").(*item.Medicine),
		func(it *item.Medicine, meta int) {
			t, ok := medicineTypeIdMap.FromID(meta)
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown medicine type ID %d", meta)})
			}
			it.SetType(t)
		},
		func(it *item.Medicine) int { return medicineTypeIdMap.ToID(it.GetType()) },
	)
	Map1to1ItemWithMeta(
		r,
		POTION,
		v("potion").(*item.Potion),
		func(it *item.Potion, meta int) { it.SetType(potionType(meta)) },
		func(it *item.Potion) int { return item.PotionTypeIdMapInstance.ToID(it.GetType()) },
	)
	Map1to1ItemWithMeta(
		r,
		SPLASH_POTION,
		v("splash_potion").(*item.SplashPotion),
		func(it *item.SplashPotion, meta int) { it.SetType(potionType(meta)) },
		func(it *item.SplashPotion) int { return item.PotionTypeIdMapInstance.ToID(it.GetType()) },
	)
	Map1to1ItemWithMeta(
		r,
		SUSPICIOUS_STEW,
		v("suspicious_stew").(*item.SuspiciousStew),
		func(it *item.SuspiciousStew, meta int) {
			t, ok := suspiciousStewTypeIdMap.FromID(meta)
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown suspicious stew type ID %d", meta)})
			}
			it.SetType(t)
		},
		func(it *item.SuspiciousStew) int { return suspiciousStewTypeIdMap.ToID(it.GetType()) },
	)
}

// Banner tile constants (tile\Banner::TAG_TYPE, TYPE_NORMAL, TYPE_OMINOUS).
const (
	bannerTagType     = "Type"
	bannerTypeNormal  = 0
	bannerTypeOminous = 1
)

// registerMiscItemMappings registers serializers and deserializers for items that don't fit any
// other pattern. Most of these are single PocketMine-MP items which map to multiple IDs depending
// on their properties, which is complex to implement in a generic way.
func (r *ItemSerializerDeserializerRegistrar) registerMiscItemMappings() {
	dyeMap := bedrock.GetDyeColorIdMap()
	for _, color := range blockutils.AllDyeColors {
		id := dyeMap.ToItemID(color)
		if r.deserializer != nil {
			r.deserializer.Map(id, func(SavedItemData) item.Item {
				dye := v("dye").(*item.Dye)
				dye.SetColor(color)
				return dye
			})
		}
	}
	if r.serializer != nil {
		r.serializer.Map(v("dye"), func(it item.Item) SavedItemData {
			return SavedItemData{Name: dyeMap.ToItemID(it.(*item.Dye).GetColor())}
		})
	}

	if r.deserializer != nil {
		r.deserializer.Map(BANNER, func(data SavedItemData) item.Item {
			bannerType := bannerTypeNormal
			if data.Tag != nil {
				bannerType = int(data.Tag.GetIntOr(bannerTagType, bannerTypeNormal))
			}
			if bannerType == bannerTypeOminous {
				return v("ominous_banner")
			}
			color, ok := dyeMap.FromInvertedID(data.Meta)
			if !ok {
				panic(&ItemTypeDeserializeError{Message: fmt.Sprintf("Unknown banner meta %d", data.Meta)})
			}
			banner := v("banner").(*item.Banner)
			banner.SetColor(color)
			return banner
		})
	}
	if r.serializer != nil {
		r.serializer.Map(v("ominous_banner"), func(item.Item) SavedItemData {
			return SavedItemData{Name: BANNER, Tag: nbt.NewCompoundTag().SetInt(bannerTagType, bannerTypeOminous)}
		})
		r.serializer.Map(v("banner"), func(it item.Item) SavedItemData {
			return SavedItemData{Name: BANNER, Meta: dyeMap.ToInvertedID(it.(*item.Banner).GetColor())}
		})
	}
}

// registerMiscBlockMappings registers serializers and deserializers for PocketMine-MP blockitems
// that don't fit any other pattern.
func (r *ItemSerializerDeserializerRegistrar) registerMiscBlockMappings() {
	copperDoorStateIdMap := map[blockutils.CopperOxidation][2]string{}
	for _, e := range []struct {
		id        string
		oxidation blockutils.CopperOxidation
		waxed     bool
	}{
		{COPPER_DOOR, blockutils.CopperOxidationNone, false},
		{EXPOSED_COPPER_DOOR, blockutils.CopperOxidationExposed, false},
		{WEATHERED_COPPER_DOOR, blockutils.CopperOxidationWeathered, false},
		{OXIDIZED_COPPER_DOOR, blockutils.CopperOxidationOxidized, false},
		{WAXED_COPPER_DOOR, blockutils.CopperOxidationNone, true},
		{WAXED_EXPOSED_COPPER_DOOR, blockutils.CopperOxidationExposed, true},
		{WAXED_WEATHERED_COPPER_DOOR, blockutils.CopperOxidationWeathered, true},
		{WAXED_OXIDIZED_COPPER_DOOR, blockutils.CopperOxidationOxidized, true},
	} {
		ids := copperDoorStateIdMap[e.oxidation]
		if e.waxed {
			ids[1] = e.id
		} else {
			ids[0] = e.id
		}
		copperDoorStateIdMap[e.oxidation] = ids
		if r.deserializer != nil {
			oxidation, waxed := e.oxidation, e.waxed
			r.deserializer.MapBlock(e.id, func(SavedItemData) block.Behavior {
				door := b("copper_door").(*block.CopperDoor)
				door.SetOxidation(oxidation)
				door.SetWaxed(waxed)
				return door
			})
		}
	}
	if r.serializer != nil {
		r.serializer.MapBlock(b("copper_door"), func(blk block.Behavior) SavedItemData {
			door := blk.(*block.CopperDoor)
			waxed := 0
			if door.IsWaxed() {
				waxed = 1
			}
			return SavedItemData{Name: copperDoorStateIdMap[door.GetOxidation()][waxed]}
		})
	}
}
