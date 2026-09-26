package item

import (
	"fmt"
	"strings"
	"sync"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/item/enchantment"
)

// This file is a port of pocketmine\item\VanillaItemsInputs (the source PocketMine-MP generates
// VanillaItems from): every vanilla item that isn't a plain block item, registered with the same
// names. VanillaItem(name) is VanillaItems::NAME(): it returns a new instance (a clone).

// Armor slots (ArmorInventory::SLOT_*); inventory imports this package, so they're repeated here.
const (
	ArmorSlotHead  = 0
	ArmorSlotChest = 1
	ArmorSlotLegs  = 2
	ArmorSlotFeet  = 3
)

// armorInfo is `new ArmorTypeInfo(defensePoints, maxDurability, armorSlot, toughness, fireProof, material)`.
func armorInfo(defensePoints, maxDurability, armorSlot, toughness int, fireProof bool, material ArmorMaterial) ArmorTypeInfo {
	info := NewArmorTypeInfo(defensePoints, maxDurability, armorSlot, material)
	info.Toughness = toughness
	info.FireProof = fireProof
	return info
}

type vanillaItemsRegistry struct {
	items   map[string]Item
	names   []string
	delayed []func()
	nextID  int
}

var (
	vanillaItems     *vanillaItemsRegistry
	vanillaItemsOnce sync.Once
)

func getVanillaItems() *vanillaItemsRegistry {
	vanillaItemsOnce.Do(func() {
		r := &vanillaItemsRegistry{items: map[string]Item{}, nextID: FIRST_UNUSED_ITEM_ID}
		r.setup()
		for _, register := range r.delayed {
			register()
		}
		r.delayed = nil
		vanillaItems = r
	})
	return vanillaItems
}

// VanillaItem returns a new instance of the vanilla item registered under name (VanillaItems::NAME()).
// Panics if there is no such item.
func VanillaItem(name string) Item {
	it, ok := getVanillaItems().items[name]
	if !ok {
		panic(fmt.Sprintf("VanillaItems: no item registered as %q", name))
	}
	return it.Clone()
}

// GetAllVanillaItems is VanillaItems::getAll(): a new instance of every vanilla item, keyed by
// registry name.
func GetAllVanillaItems() map[string]Item {
	r := getVanillaItems()
	result := make(map[string]Item, len(r.items))
	for name, it := range r.items {
		result[name] = it.Clone()
	}
	return result
}

// GetVanillaItemNames returns the registry names in registration order.
func GetVanillaItemNames() []string { return append([]string(nil), getVanillaItems().names...) }

// makeIID is a port of VanillaItemsInputs::makeIID.
func (r *vanillaItemsRegistry) makeIID(name string) ItemIdentifier {
	typeID, ok := typeIDsByName[strings.ToUpper(name)]
	if !ok {
		//this allows registering new stuff without adding new type ID constants
		typeID = r.nextID
		r.nextID++
	}
	return NewItemIdentifier(typeID)
}

func (r *vanillaItemsRegistry) add(name string, it Item) {
	if _, exists := r.items[name]; exists {
		panic(fmt.Sprintf("VanillaItems: %q registered twice", name))
	}
	r.items[name] = it
	r.names = append(r.names, name)
}

// register is a port of VanillaItemsInputs::register.
func (r *vanillaItemsRegistry) register(name string, create func(id ItemIdentifier) Item) Item {
	it := create(r.makeIID(name))
	r.add(name, it)
	return it
}

// registerDelayed is RegistrySource::registerDelayed (the item is created after the others, when
// VanillaBlocks is available).
func (r *vanillaItemsRegistry) registerDelayed(name string, create func(id ItemIdentifier) Item) {
	r.delayed = append(r.delayed, func() { r.register(name, create) })
}

// fireProofItem is the anonymous Item subclass registered for "netherite_ingot" and
// "netherite_scrap".
type fireProofItem struct{ PlainItem }

func newFireProofItem(identifier ItemIdentifier, name string) *fireProofItem {
	f := &fireProofItem{}
	f.Init(f, identifier, name)
	return f
}

func (f *fireProofItem) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *fireProofItem) IsFireProof() bool { return true }

func (r *vanillaItemsRegistry) setup() {
	r.registerArmorItems()
	r.registerSpawnEggs()
	r.registerTierToolItems()
	r.registerSmithingTemplates()

	//this doesn't use the regular register() because it doesn't have an item typeID
	//in the future we'll probably want to dissociate this from the air block and make a proper null item
	r.delayed = append(r.delayed, func() {
		air := NewItemBlock(NewItemIdentifier(-block.AIR), block.VanillaAir())
		air.SetCount(0)
		r.add("air", air)
	})

	r.registerDelayed("acacia_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("acacia_sign"), block.VanillaBlock("acacia_wall_sign"))
	})
	r.registerDelayed("acacia_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Acacia Hanging Sign", block.VanillaBlock("acacia_ceiling_center_hanging_sign"), block.VanillaBlock("acacia_ceiling_edges_hanging_sign"), block.VanillaBlock("acacia_wall_hanging_sign"))
	})
	r.register("amethyst_shard", func(id ItemIdentifier) Item { return NewItem(id, "Amethyst Shard") })
	r.register("apple", func(id ItemIdentifier) Item { return NewApple(id, "Apple") })
	r.register("arrow", func(id ItemIdentifier) Item { return NewArrow(id, "Arrow") })
	r.register("baked_potato", func(id ItemIdentifier) Item { return NewBakedPotato(id, "Baked Potato") })
	r.register("bamboo", func(id ItemIdentifier) Item { return NewBamboo(id, "Bamboo") })
	r.registerDelayed("bamboo_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("bamboo_sign"), block.VanillaBlock("bamboo_wall_sign"))
	})
	r.registerDelayed("bamboo_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Bamboo Hanging Sign", block.VanillaBlock("bamboo_ceiling_center_hanging_sign"), block.VanillaBlock("bamboo_ceiling_edges_hanging_sign"), block.VanillaBlock("bamboo_wall_hanging_sign"))
	})
	r.register("beetroot", func(id ItemIdentifier) Item { return NewBeetroot(id, "Beetroot") })
	r.register("beetroot_seeds", func(id ItemIdentifier) Item { return NewBeetrootSeeds(id, "Beetroot Seeds") })
	r.register("beetroot_soup", func(id ItemIdentifier) Item { return NewBeetrootSoup(id, "Beetroot Soup") })
	r.registerDelayed("birch_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("birch_sign"), block.VanillaBlock("birch_wall_sign"))
	})
	r.registerDelayed("birch_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Birch Hanging Sign", block.VanillaBlock("birch_ceiling_center_hanging_sign"), block.VanillaBlock("birch_ceiling_edges_hanging_sign"), block.VanillaBlock("birch_wall_hanging_sign"))
	})
	r.register("blaze_powder", func(id ItemIdentifier) Item { return NewItem(id, "Blaze Powder") })
	r.register("blaze_rod", func(id ItemIdentifier) Item { return NewBlazeRod(id, "Blaze Rod") })
	r.register("bleach", func(id ItemIdentifier) Item { return NewItem(id, "Bleach") })
	r.register("bone", func(id ItemIdentifier) Item { return NewItem(id, "Bone") })
	r.register("bone_meal", func(id ItemIdentifier) Item { return NewFertilizer(id, "Bone Meal") })
	r.register("book", func(id ItemIdentifier) Item { return NewBook(id, "Book", enchantment.TagAll) })
	r.register("bow", func(id ItemIdentifier) Item { return NewBow(id, "Bow", enchantment.TagBow) })
	r.register("bowl", func(id ItemIdentifier) Item { return NewBowl(id, "Bowl") })
	r.register("bread", func(id ItemIdentifier) Item { return NewBread(id, "Bread") })
	r.register("brick", func(id ItemIdentifier) Item { return NewItem(id, "Brick") })
	r.register("bucket", func(id ItemIdentifier) Item { return NewBucket(id, "Bucket") })
	r.register("carrot", func(id ItemIdentifier) Item { return NewCarrot(id, "Carrot") })
	r.register("charcoal", func(id ItemIdentifier) Item { return NewCoal(id, "Charcoal") })
	r.registerDelayed("cherry_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("cherry_sign"), block.VanillaBlock("cherry_wall_sign"))
	})
	r.registerDelayed("cherry_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Cherry Hanging Sign", block.VanillaBlock("cherry_ceiling_center_hanging_sign"), block.VanillaBlock("cherry_ceiling_edges_hanging_sign"), block.VanillaBlock("cherry_wall_hanging_sign"))
	})
	r.register("chemical_aluminium_oxide", func(id ItemIdentifier) Item { return NewItem(id, "Aluminium Oxide") })
	r.register("chemical_ammonia", func(id ItemIdentifier) Item { return NewItem(id, "Ammonia") })
	r.register("chemical_barium_sulphate", func(id ItemIdentifier) Item { return NewItem(id, "Barium Sulphate") })
	r.register("chemical_benzene", func(id ItemIdentifier) Item { return NewItem(id, "Benzene") })
	r.register("chemical_boron_trioxide", func(id ItemIdentifier) Item { return NewItem(id, "Boron Trioxide") })
	r.register("chemical_calcium_bromide", func(id ItemIdentifier) Item { return NewItem(id, "Calcium Bromide") })
	r.register("chemical_calcium_chloride", func(id ItemIdentifier) Item { return NewItem(id, "Calcium Chloride") })
	r.register("chemical_cerium_chloride", func(id ItemIdentifier) Item { return NewItem(id, "Cerium Chloride") })
	r.register("chemical_charcoal", func(id ItemIdentifier) Item { return NewItem(id, "Charcoal") })
	r.register("chemical_crude_oil", func(id ItemIdentifier) Item { return NewItem(id, "Crude Oil") })
	r.register("chemical_glue", func(id ItemIdentifier) Item { return NewItem(id, "Glue") })
	r.register("chemical_hydrogen_peroxide", func(id ItemIdentifier) Item { return NewItem(id, "Hydrogen Peroxide") })
	r.register("chemical_hypochlorite", func(id ItemIdentifier) Item { return NewItem(id, "Hypochlorite") })
	r.register("chemical_ink", func(id ItemIdentifier) Item { return NewItem(id, "Ink") })
	r.register("chemical_iron_sulphide", func(id ItemIdentifier) Item { return NewItem(id, "Iron Sulphide") })
	r.register("chemical_latex", func(id ItemIdentifier) Item { return NewItem(id, "Latex") })
	r.register("chemical_lithium_hydride", func(id ItemIdentifier) Item { return NewItem(id, "Lithium Hydride") })
	r.register("chemical_luminol", func(id ItemIdentifier) Item { return NewItem(id, "Luminol") })
	r.register("chemical_magnesium_nitrate", func(id ItemIdentifier) Item { return NewItem(id, "Magnesium Nitrate") })
	r.register("chemical_magnesium_oxide", func(id ItemIdentifier) Item { return NewItem(id, "Magnesium Oxide") })
	r.register("chemical_magnesium_salts", func(id ItemIdentifier) Item { return NewItem(id, "Magnesium Salts") })
	r.register("chemical_mercuric_chloride", func(id ItemIdentifier) Item { return NewItem(id, "Mercuric Chloride") })
	r.register("chemical_polyethylene", func(id ItemIdentifier) Item { return NewItem(id, "Polyethylene") })
	r.register("chemical_potassium_chloride", func(id ItemIdentifier) Item { return NewItem(id, "Potassium Chloride") })
	r.register("chemical_potassium_iodide", func(id ItemIdentifier) Item { return NewItem(id, "Potassium Iodide") })
	r.register("chemical_rubbish", func(id ItemIdentifier) Item { return NewItem(id, "Rubbish") })
	r.register("chemical_salt", func(id ItemIdentifier) Item { return NewItem(id, "Salt") })
	r.register("chemical_soap", func(id ItemIdentifier) Item { return NewItem(id, "Soap") })
	r.register("chemical_sodium_acetate", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Acetate") })
	r.register("chemical_sodium_fluoride", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Fluoride") })
	r.register("chemical_sodium_hydride", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Hydride") })
	r.register("chemical_sodium_hydroxide", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Hydroxide") })
	r.register("chemical_sodium_hypochlorite", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Hypochlorite") })
	r.register("chemical_sodium_oxide", func(id ItemIdentifier) Item { return NewItem(id, "Sodium Oxide") })
	r.register("chemical_sugar", func(id ItemIdentifier) Item { return NewItem(id, "Sugar") })
	r.register("chemical_sulphate", func(id ItemIdentifier) Item { return NewItem(id, "Sulphate") })
	r.register("chemical_tungsten_chloride", func(id ItemIdentifier) Item { return NewItem(id, "Tungsten Chloride") })
	r.register("chemical_water", func(id ItemIdentifier) Item { return NewItem(id, "Water") })
	r.register("chorus_fruit", func(id ItemIdentifier) Item { return NewChorusFruit(id, "Chorus Fruit") })
	r.register("clay", func(id ItemIdentifier) Item { return NewItem(id, "Clay") })
	r.register("clock", func(id ItemIdentifier) Item { return NewClock(id, "Clock") })
	r.register("clownfish", func(id ItemIdentifier) Item { return NewClownfish(id, "Clownfish") })
	r.register("coal", func(id ItemIdentifier) Item { return NewCoal(id, "Coal") })
	r.register("cocoa_beans", func(id ItemIdentifier) Item { return NewCocoaBeans(id, "Cocoa Beans") })
	r.register("compass", func(id ItemIdentifier) Item { return NewCompass(id, "Compass", enchantment.TagCompass) })
	r.register("cooked_chicken", func(id ItemIdentifier) Item { return NewCookedChicken(id, "Cooked Chicken") })
	r.register("cooked_fish", func(id ItemIdentifier) Item { return NewCookedFish(id, "Cooked Fish") })
	r.register("cooked_mutton", func(id ItemIdentifier) Item { return NewCookedMutton(id, "Cooked Mutton") })
	r.register("cooked_porkchop", func(id ItemIdentifier) Item { return NewCookedPorkchop(id, "Cooked Porkchop") })
	r.register("cooked_rabbit", func(id ItemIdentifier) Item { return NewCookedRabbit(id, "Cooked Rabbit") })
	r.register("cooked_salmon", func(id ItemIdentifier) Item { return NewCookedSalmon(id, "Cooked Salmon") })
	r.register("cookie", func(id ItemIdentifier) Item { return NewCookie(id, "Cookie") })
	r.register("copper_ingot", func(id ItemIdentifier) Item { return NewItem(id, "Copper Ingot") })
	r.register("copper_nugget", func(id ItemIdentifier) Item { return NewItem(id, "Copper Nugget") })
	r.registerDelayed("crimson_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("crimson_sign"), block.VanillaBlock("crimson_wall_sign"))
	})
	r.registerDelayed("crimson_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Crimson Hanging Sign", block.VanillaBlock("crimson_ceiling_center_hanging_sign"), block.VanillaBlock("crimson_ceiling_edges_hanging_sign"), block.VanillaBlock("crimson_wall_hanging_sign"))
	})
	r.registerDelayed("dark_oak_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("dark_oak_sign"), block.VanillaBlock("dark_oak_wall_sign"))
	})
	r.registerDelayed("dark_oak_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Dark Oak Hanging Sign", block.VanillaBlock("dark_oak_ceiling_center_hanging_sign"), block.VanillaBlock("dark_oak_ceiling_edges_hanging_sign"), block.VanillaBlock("dark_oak_wall_hanging_sign"))
	})
	r.register("diamond", func(id ItemIdentifier) Item { return NewItem(id, "Diamond") })
	r.register("disc_fragment_5", func(id ItemIdentifier) Item { return NewItem(id, "Disc Fragment (5)") })
	r.register("dragon_breath", func(id ItemIdentifier) Item { return NewItem(id, "Dragon's Breath") })
	r.register("dried_kelp", func(id ItemIdentifier) Item { return NewDriedKelp(id, "Dried Kelp") })
	r.register("dye", func(id ItemIdentifier) Item { return NewDye(id, "Dye") })
	r.register("echo_shard", func(id ItemIdentifier) Item { return NewItem(id, "Echo Shard") })
	r.register("egg", func(id ItemIdentifier) Item { return NewEgg(id, "Egg") })
	r.register("emerald", func(id ItemIdentifier) Item { return NewItem(id, "Emerald") })
	r.register("enchanted_book", func(id ItemIdentifier) Item { return NewEnchantedBook(id, "Enchanted Book", enchantment.TagAll) })
	r.register("enchanted_golden_apple", func(id ItemIdentifier) Item { return NewGoldenAppleEnchanted(id, "Enchanted Golden Apple") })
	r.register("end_crystal", func(id ItemIdentifier) Item { return NewEndCrystal(id, "End Crystal") })
	r.register("ender_pearl", func(id ItemIdentifier) Item { return NewEnderPearl(id, "Ender Pearl") })
	r.register("experience_bottle", func(id ItemIdentifier) Item { return NewExperienceBottle(id, "Bottle o' Enchanting") })
	r.register("feather", func(id ItemIdentifier) Item { return NewItem(id, "Feather") })
	r.register("fermented_spider_eye", func(id ItemIdentifier) Item { return NewItem(id, "Fermented Spider Eye") })
	r.register("firework_rocket", func(id ItemIdentifier) Item { return NewFireworkRocket(id, "Firework Rocket") })
	r.register("firework_star", func(id ItemIdentifier) Item { return NewFireworkStar(id, "Firework Star") })
	r.register("fire_charge", func(id ItemIdentifier) Item { return NewFireCharge(id, "Fire Charge") })
	r.register("fishing_rod", func(id ItemIdentifier) Item { return NewFishingRod(id, "Fishing Rod", enchantment.TagFishingRod) })
	r.register("flint", func(id ItemIdentifier) Item { return NewItem(id, "Flint") })
	r.register("flint_and_steel", func(id ItemIdentifier) Item {
		return NewFlintSteel(id, "Flint and Steel", enchantment.TagFlintAndSteel)
	})
	r.register("ghast_tear", func(id ItemIdentifier) Item { return NewItem(id, "Ghast Tear") })
	r.register("glass_bottle", func(id ItemIdentifier) Item { return NewGlassBottle(id, "Glass Bottle") })
	r.register("glistering_melon", func(id ItemIdentifier) Item { return NewItem(id, "Glistering Melon") })
	r.register("glow_berries", func(id ItemIdentifier) Item { return NewGlowBerries(id, "Glow Berries") })
	r.register("glow_ink_sac", func(id ItemIdentifier) Item { return NewItem(id, "Glow Ink Sac") })
	r.register("glowstone_dust", func(id ItemIdentifier) Item { return NewItem(id, "Glowstone Dust") })
	r.register("goat_horn", func(id ItemIdentifier) Item { return NewGoatHorn(id, "Goat Horn") })
	r.register("gold_ingot", func(id ItemIdentifier) Item { return NewItem(id, "Gold Ingot") })
	r.register("gold_nugget", func(id ItemIdentifier) Item { return NewItem(id, "Gold Nugget") })
	r.register("golden_apple", func(id ItemIdentifier) Item { return NewGoldenApple(id, "Golden Apple") })
	r.register("golden_carrot", func(id ItemIdentifier) Item { return NewGoldenCarrot(id, "Golden Carrot") })
	r.register("gunpowder", func(id ItemIdentifier) Item { return NewItem(id, "Gunpowder") })
	r.register("heart_of_the_sea", func(id ItemIdentifier) Item { return NewItem(id, "Heart of the Sea") })
	r.register("honey_bottle", func(id ItemIdentifier) Item { return NewHoneyBottle(id, "Honey Bottle") })
	r.register("honeycomb", func(id ItemIdentifier) Item { return NewItem(id, "Honeycomb") })
	r.register("ice_bomb", func(id ItemIdentifier) Item { return NewIceBomb(id, "Ice Bomb") })
	r.register("ink_sac", func(id ItemIdentifier) Item { return NewItem(id, "Ink Sac") })
	r.register("iron_ingot", func(id ItemIdentifier) Item { return NewItem(id, "Iron Ingot") })
	r.register("iron_nugget", func(id ItemIdentifier) Item { return NewItem(id, "Iron Nugget") })
	r.registerDelayed("jungle_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("jungle_sign"), block.VanillaBlock("jungle_wall_sign"))
	})
	r.registerDelayed("jungle_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Jungle Hanging Sign", block.VanillaBlock("jungle_ceiling_center_hanging_sign"), block.VanillaBlock("jungle_ceiling_edges_hanging_sign"), block.VanillaBlock("jungle_wall_hanging_sign"))
	})
	r.register("lapis_lazuli", func(id ItemIdentifier) Item { return NewItem(id, "Lapis Lazuli") })
	r.register("leather", func(id ItemIdentifier) Item { return NewItem(id, "Leather") })
	r.register("lingering_potion", func(id ItemIdentifier) Item { return NewSplashPotion(id, "Lingering Potion", true) })
	r.register("magma_cream", func(id ItemIdentifier) Item { return NewItem(id, "Magma Cream") })
	r.registerDelayed("mangrove_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("mangrove_sign"), block.VanillaBlock("mangrove_wall_sign"))
	})
	r.registerDelayed("mangrove_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Mangrove Hanging Sign", block.VanillaBlock("mangrove_ceiling_center_hanging_sign"), block.VanillaBlock("mangrove_ceiling_edges_hanging_sign"), block.VanillaBlock("mangrove_wall_hanging_sign"))
	})
	r.register("medicine", func(id ItemIdentifier) Item { return NewMedicine(id, "Medicine") })
	r.register("melon", func(id ItemIdentifier) Item { return NewMelon(id, "Melon") })
	r.register("melon_seeds", func(id ItemIdentifier) Item { return NewMelonSeeds(id, "Melon Seeds") })
	r.register("milk_bucket", func(id ItemIdentifier) Item { return NewMilkBucket(id, "Milk Bucket") })
	r.register("minecart", func(id ItemIdentifier) Item { return NewMinecart(id, "Minecart") })
	r.register("mushroom_stew", func(id ItemIdentifier) Item { return NewMushroomStew(id, "Mushroom Stew") })
	r.register("name_tag", func(id ItemIdentifier) Item { return NewNameTag(id, "Name Tag") })
	r.register("nautilus_shell", func(id ItemIdentifier) Item { return NewItem(id, "Nautilus Shell") })
	r.register("nether_brick", func(id ItemIdentifier) Item { return NewItem(id, "Nether Brick") })
	r.register("nether_quartz", func(id ItemIdentifier) Item { return NewItem(id, "Nether Quartz") })
	r.register("nether_star", func(id ItemIdentifier) Item { return NewItem(id, "Nether Star") })
	r.registerDelayed("oak_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("oak_sign"), block.VanillaBlock("oak_wall_sign"))
	})
	r.registerDelayed("oak_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Oak Hanging Sign", block.VanillaBlock("oak_ceiling_center_hanging_sign"), block.VanillaBlock("oak_ceiling_edges_hanging_sign"), block.VanillaBlock("oak_wall_hanging_sign"))
	})
	r.registerDelayed("ominous_banner", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("ominous_banner"), block.VanillaBlock("ominous_wall_banner"))
	})
	r.register("painting", func(id ItemIdentifier) Item { return NewPaintingItem(id, "Painting") })
	r.registerDelayed("pale_oak_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("pale_oak_sign"), block.VanillaBlock("pale_oak_wall_sign"))
	})
	r.registerDelayed("pale_oak_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Pale Oak Hanging Sign", block.VanillaBlock("pale_oak_ceiling_center_hanging_sign"), block.VanillaBlock("pale_oak_ceiling_edges_hanging_sign"), block.VanillaBlock("pale_oak_wall_hanging_sign"))
	})
	r.register("paper", func(id ItemIdentifier) Item { return NewItem(id, "Paper") })
	r.register("phantom_membrane", func(id ItemIdentifier) Item { return NewItem(id, "Phantom Membrane") })
	r.register("pitcher_pod", func(id ItemIdentifier) Item { return NewPitcherPod(id, "Pitcher Pod") })
	r.register("poisonous_potato", func(id ItemIdentifier) Item { return NewPoisonousPotato(id, "Poisonous Potato") })
	r.register("popped_chorus_fruit", func(id ItemIdentifier) Item { return NewItem(id, "Popped Chorus Fruit") })
	r.register("potato", func(id ItemIdentifier) Item { return NewPotato(id, "Potato") })
	r.register("potion", func(id ItemIdentifier) Item { return NewPotion(id, "Potion") })
	r.register("prismarine_crystals", func(id ItemIdentifier) Item { return NewItem(id, "Prismarine Crystals") })
	r.register("prismarine_shard", func(id ItemIdentifier) Item { return NewItem(id, "Prismarine Shard") })
	r.register("pufferfish", func(id ItemIdentifier) Item { return NewPufferfish(id, "Pufferfish") })
	r.register("pumpkin_pie", func(id ItemIdentifier) Item { return NewPumpkinPie(id, "Pumpkin Pie") })
	r.register("pumpkin_seeds", func(id ItemIdentifier) Item { return NewPumpkinSeeds(id, "Pumpkin Seeds") })
	r.register("rabbit_foot", func(id ItemIdentifier) Item { return NewItem(id, "Rabbit's Foot") })
	r.register("rabbit_hide", func(id ItemIdentifier) Item { return NewItem(id, "Rabbit Hide") })
	r.register("rabbit_stew", func(id ItemIdentifier) Item { return NewRabbitStew(id, "Rabbit Stew") })
	r.register("raw_beef", func(id ItemIdentifier) Item { return NewRawBeef(id, "Raw Beef") })
	r.register("raw_chicken", func(id ItemIdentifier) Item { return NewRawChicken(id, "Raw Chicken") })
	r.register("raw_copper", func(id ItemIdentifier) Item { return NewItem(id, "Raw Copper") })
	r.register("raw_fish", func(id ItemIdentifier) Item { return NewRawFish(id, "Raw Fish") })
	r.register("raw_gold", func(id ItemIdentifier) Item { return NewItem(id, "Raw Gold") })
	r.register("raw_iron", func(id ItemIdentifier) Item { return NewItem(id, "Raw Iron") })
	r.register("raw_mutton", func(id ItemIdentifier) Item { return NewRawMutton(id, "Raw Mutton") })
	r.register("raw_porkchop", func(id ItemIdentifier) Item { return NewRawPorkchop(id, "Raw Porkchop") })
	r.register("raw_rabbit", func(id ItemIdentifier) Item { return NewRawRabbit(id, "Raw Rabbit") })
	r.register("raw_salmon", func(id ItemIdentifier) Item { return NewRawSalmon(id, "Raw Salmon") })
	r.register("record_11", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDisk11, "Record 11") })
	r.register("record_13", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDisk13, "Record 13") })
	r.register("record_5", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDisk5, "Record 5") })
	r.register("record_blocks", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskBlocks, "Record Blocks") })
	r.register("record_cat", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskCat, "Record Cat") })
	r.register("record_chirp", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskChirp, "Record Chirp") })
	r.register("record_creator", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskCreator, "Record Creator") })
	r.register("record_creator_music_box", func(id ItemIdentifier) Item {
		return NewRecord(id, blockutils.RecordTypeDiskCreatorMusicBox, "Record Creator (Music Box)")
	})
	r.register("record_far", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskFar, "Record Far") })
	r.register("record_lava_chicken", func(id ItemIdentifier) Item {
		return NewRecord(id, blockutils.RecordTypeDiskLavaChicken, "Record Lava Chicken")
	})
	r.register("record_mall", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskMall, "Record Mall") })
	r.register("record_mellohi", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskMellohi, "Record Mellohi") })
	r.register("record_otherside", func(id ItemIdentifier) Item {
		return NewRecord(id, blockutils.RecordTypeDiskOtherside, "Record Otherside")
	})
	r.register("record_pigstep", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskPigstep, "Record Pigstep") })
	r.register("record_precipice", func(id ItemIdentifier) Item {
		return NewRecord(id, blockutils.RecordTypeDiskPrecipice, "Record Precipice")
	})
	r.register("record_relic", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskRelic, "Record Relic") })
	r.register("record_stal", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskStal, "Record Stal") })
	r.register("record_strad", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskStrad, "Record Strad") })
	r.register("record_wait", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskWait, "Record Wait") })
	r.register("record_ward", func(id ItemIdentifier) Item { return NewRecord(id, blockutils.RecordTypeDiskWard, "Record Ward") })
	r.register("recovery_compass", func(id ItemIdentifier) Item { return NewItem(id, "Recovery Compass") })
	r.register("redstone_dust", func(id ItemIdentifier) Item { return NewRedstone(id, "Redstone") })
	r.register("resin_brick", func(id ItemIdentifier) Item { return NewItem(id, "Resin Brick") })
	r.register("rotten_flesh", func(id ItemIdentifier) Item { return NewRottenFlesh(id, "Rotten Flesh") })
	r.register("scute", func(id ItemIdentifier) Item { return NewItem(id, "Scute") })
	r.register("shears", func(id ItemIdentifier) Item { return NewShears(id, "Shears", enchantment.TagShears) })
	r.register("shulker_shell", func(id ItemIdentifier) Item { return NewItem(id, "Shulker Shell") })
	r.register("slimeball", func(id ItemIdentifier) Item { return NewItem(id, "Slimeball") })
	r.register("snowball", func(id ItemIdentifier) Item { return NewSnowball(id, "Snowball") })
	r.register("spider_eye", func(id ItemIdentifier) Item { return NewSpiderEye(id, "Spider Eye") })
	r.register("splash_potion", func(id ItemIdentifier) Item { return NewSplashPotion(id, "Splash Potion", false) })
	r.registerDelayed("spruce_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("spruce_sign"), block.VanillaBlock("spruce_wall_sign"))
	})
	r.registerDelayed("spruce_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Spruce Hanging Sign", block.VanillaBlock("spruce_ceiling_center_hanging_sign"), block.VanillaBlock("spruce_ceiling_edges_hanging_sign"), block.VanillaBlock("spruce_wall_hanging_sign"))
	})
	r.register("spyglass", func(id ItemIdentifier) Item { return NewSpyglass(id, "Spyglass") })
	r.register("steak", func(id ItemIdentifier) Item { return NewSteak(id, "Steak") })
	r.register("stick", func(id ItemIdentifier) Item { return NewStick(id, "Stick") })
	r.register("string", func(id ItemIdentifier) Item { return NewStringItem(id, "String") })
	r.register("sugar", func(id ItemIdentifier) Item { return NewItem(id, "Sugar") })
	r.register("suspicious_stew", func(id ItemIdentifier) Item { return NewSuspiciousStew(id, "Suspicious Stew") })
	r.register("sweet_berries", func(id ItemIdentifier) Item { return NewSweetBerries(id, "Sweet Berries") })
	r.register("torchflower_seeds", func(id ItemIdentifier) Item { return NewTorchflowerSeeds(id, "Torchflower Seeds") })
	r.register("totem", func(id ItemIdentifier) Item { return NewTotem(id, "Totem of Undying") })
	r.register("trident", func(id ItemIdentifier) Item { return NewTrident(id, "Trident") })
	r.registerDelayed("warped_sign", func(id ItemIdentifier) Item {
		return NewItemBlockWallOrFloor(id, block.VanillaBlock("warped_sign"), block.VanillaBlock("warped_wall_sign"))
	})
	r.registerDelayed("warped_hanging_sign", func(id ItemIdentifier) Item {
		return NewHangingSign(id, "Warped Hanging Sign", block.VanillaBlock("warped_ceiling_center_hanging_sign"), block.VanillaBlock("warped_ceiling_edges_hanging_sign"), block.VanillaBlock("warped_wall_hanging_sign"))
	})
	r.register("wheat", func(id ItemIdentifier) Item { return NewItem(id, "Wheat") })
	r.register("wheat_seeds", func(id ItemIdentifier) Item { return NewWheatSeeds(id, "Wheat Seeds") })
	r.register("writable_book", func(id ItemIdentifier) Item { return NewWritableBook(id, "Book & Quill") })
	r.register("written_book", func(id ItemIdentifier) Item { return NewWrittenBook(id, "Written Book") })
	r.registerDelayed("banner", func(id ItemIdentifier) Item {
		return NewBanner(id, block.VanillaBlock("banner"), block.VanillaBlock("wall_banner"))
	})
	r.registerDelayed("coral_fan", func(id ItemIdentifier) Item { return NewCoralFan(id, block.VanillaBlock("coral_fan").GetName()) }) //uses VanillaBlocks in constructor :(
	r.registerDelayed("lava_bucket", func(id ItemIdentifier) Item { return NewLiquidBucket(id, "Lava Bucket", block.VanillaBlock("lava")) })
	r.registerDelayed("water_bucket", func(id ItemIdentifier) Item { return NewLiquidBucket(id, "Water Bucket", block.VanillaBlock("water")) })
	r.register("netherite_ingot", func(id ItemIdentifier) Item { return newFireProofItem(id, "Netherite Ingot") })
	r.register("netherite_scrap", func(id ItemIdentifier) Item { return newFireProofItem(id, "Netherite Scrap") })

	for _, boatType := range []BoatType{BoatTypeOak, BoatTypeSpruce, BoatTypeBirch, BoatTypeJungle, BoatTypeAcacia, BoatTypeDarkOak, BoatTypeMangrove} {
		//boat type is static, because different types of wood may have different properties
		r.register(boatType.GetWoodType().IDName()+"_boat", func(id ItemIdentifier) Item {
			return NewBoat(id, boatType.GetDisplayName()+" Boat", boatType)
		})
	}

}

func (r *vanillaItemsRegistry) registerSpawnEggs() {
	r.register("zombie_spawn_egg", func(id ItemIdentifier) Item { return NewSpawnEgg(id, "Zombie Spawn Egg", "Zombie") })
	r.register("squid_spawn_egg", func(id ItemIdentifier) Item { return NewSpawnEgg(id, "Squid Spawn Egg", "Squid") })
	r.register("villager_spawn_egg", func(id ItemIdentifier) Item { return NewSpawnEgg(id, "Villager Spawn Egg", "Villager") })
}

func (r *vanillaItemsRegistry) registerTierToolItems() {
	for _, e := range []struct {
		tier                 ToolTier
		idPrefix, namePrefix string
	}{
		{ToolTierCopper, "copper", "Copper"},
		{ToolTierDiamond, "diamond", "Diamond"},
		{ToolTierGold, "golden", "Golden"},
		{ToolTierIron, "iron", "Iron"},
		{ToolTierNetherite, "netherite", "Netherite"},
		{ToolTierStone, "stone", "Stone"},
		{ToolTierWood, "wooden", "Wooden"},
	} {
		tier, namePrefix := e.tier, e.namePrefix
		r.register(e.idPrefix+"_axe", func(id ItemIdentifier) Item { return NewAxe(id, namePrefix+" Axe", tier, enchantment.TagAxe) })
		r.register(e.idPrefix+"_hoe", func(id ItemIdentifier) Item { return NewHoe(id, namePrefix+" Hoe", tier, enchantment.TagHoe) })
		r.register(e.idPrefix+"_pickaxe", func(id ItemIdentifier) Item {
			return NewPickaxe(id, namePrefix+" Pickaxe", tier, enchantment.TagPickaxe)
		})
		r.register(e.idPrefix+"_shovel", func(id ItemIdentifier) Item { return NewShovel(id, namePrefix+" Shovel", tier, enchantment.TagShovel) })
		r.register(e.idPrefix+"_sword", func(id ItemIdentifier) Item { return NewSword(id, namePrefix+" Sword", tier, enchantment.TagSword) })
	}
}

func (r *vanillaItemsRegistry) registerArmorItems() {
	r.registerDelayed("chainmail_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Chainmail Boots", armorInfo(1, 196, ArmorSlotFeet, 0, false, VanillaArmorMaterial("chainmail")), enchantment.TagBoots)
	})
	r.registerDelayed("copper_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Copper Boots", armorInfo(1, 144, ArmorSlotFeet, 0, false, VanillaArmorMaterial("copper")), enchantment.TagBoots)
	})
	r.registerDelayed("diamond_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Diamond Boots", armorInfo(3, 430, ArmorSlotFeet, 2, false, VanillaArmorMaterial("diamond")), enchantment.TagBoots)
	})
	r.registerDelayed("golden_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Golden Boots", armorInfo(1, 92, ArmorSlotFeet, 0, false, VanillaArmorMaterial("gold")), enchantment.TagBoots)
	})
	r.registerDelayed("iron_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Iron Boots", armorInfo(2, 196, ArmorSlotFeet, 0, false, VanillaArmorMaterial("iron")), enchantment.TagBoots)
	})
	r.registerDelayed("leather_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Leather Boots", armorInfo(1, 66, ArmorSlotFeet, 0, false, VanillaArmorMaterial("leather")), enchantment.TagBoots)
	})
	r.registerDelayed("netherite_boots", func(id ItemIdentifier) Item {
		return NewArmor(id, "Netherite Boots", armorInfo(3, 482, ArmorSlotFeet, 3, true, VanillaArmorMaterial("netherite")), enchantment.TagBoots)
	})
	r.registerDelayed("chainmail_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Chainmail Chestplate", armorInfo(5, 241, ArmorSlotChest, 0, false, VanillaArmorMaterial("chainmail")), enchantment.TagChestplate)
	})
	r.registerDelayed("copper_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Copper Chestplate", armorInfo(4, 177, ArmorSlotChest, 0, false, VanillaArmorMaterial("copper")), enchantment.TagChestplate)
	})
	r.registerDelayed("diamond_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Diamond Chestplate", armorInfo(8, 529, ArmorSlotChest, 2, false, VanillaArmorMaterial("diamond")), enchantment.TagChestplate)
	})
	r.registerDelayed("golden_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Golden Chestplate", armorInfo(5, 113, ArmorSlotChest, 0, false, VanillaArmorMaterial("gold")), enchantment.TagChestplate)
	})
	r.registerDelayed("iron_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Iron Chestplate", armorInfo(6, 241, ArmorSlotChest, 0, false, VanillaArmorMaterial("iron")), enchantment.TagChestplate)
	})
	r.registerDelayed("leather_tunic", func(id ItemIdentifier) Item {
		return NewArmor(id, "Leather Tunic", armorInfo(3, 81, ArmorSlotChest, 0, false, VanillaArmorMaterial("leather")), enchantment.TagChestplate)
	})
	r.registerDelayed("netherite_chestplate", func(id ItemIdentifier) Item {
		return NewArmor(id, "Netherite Chestplate", armorInfo(8, 593, ArmorSlotChest, 3, true, VanillaArmorMaterial("netherite")), enchantment.TagChestplate)
	})
	r.registerDelayed("chainmail_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Chainmail Helmet", armorInfo(2, 166, ArmorSlotHead, 0, false, VanillaArmorMaterial("chainmail")), enchantment.TagHelmet)
	})
	r.registerDelayed("copper_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Copper Helmet", armorInfo(2, 122, ArmorSlotHead, 0, false, VanillaArmorMaterial("copper")), enchantment.TagHelmet)
	})
	r.registerDelayed("diamond_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Diamond Helmet", armorInfo(3, 364, ArmorSlotHead, 2, false, VanillaArmorMaterial("diamond")), enchantment.TagHelmet)
	})
	r.registerDelayed("golden_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Golden Helmet", armorInfo(2, 78, ArmorSlotHead, 0, false, VanillaArmorMaterial("gold")), enchantment.TagHelmet)
	})
	r.registerDelayed("iron_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Iron Helmet", armorInfo(2, 166, ArmorSlotHead, 0, false, VanillaArmorMaterial("iron")), enchantment.TagHelmet)
	})
	r.registerDelayed("leather_cap", func(id ItemIdentifier) Item {
		return NewArmor(id, "Leather Cap", armorInfo(1, 56, ArmorSlotHead, 0, false, VanillaArmorMaterial("leather")), enchantment.TagHelmet)
	})
	r.registerDelayed("netherite_helmet", func(id ItemIdentifier) Item {
		return NewArmor(id, "Netherite Helmet", armorInfo(3, 408, ArmorSlotHead, 3, true, VanillaArmorMaterial("netherite")), enchantment.TagHelmet)
	})
	r.registerDelayed("turtle_helmet", func(id ItemIdentifier) Item {
		return NewTurtleHelmet(id, "Turtle Shell", armorInfo(2, 276, ArmorSlotHead, 0, false, VanillaArmorMaterial("turtle")), enchantment.TagHelmet)
	})
	r.registerDelayed("chainmail_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Chainmail Leggings", armorInfo(4, 226, ArmorSlotLegs, 0, false, VanillaArmorMaterial("chainmail")), enchantment.TagLeggings)
	})
	r.registerDelayed("copper_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Copper Leggings", armorInfo(3, 166, ArmorSlotLegs, 0, false, VanillaArmorMaterial("copper")), enchantment.TagLeggings)
	})
	r.registerDelayed("diamond_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Diamond Leggings", armorInfo(6, 496, ArmorSlotLegs, 2, false, VanillaArmorMaterial("diamond")), enchantment.TagLeggings)
	})
	r.registerDelayed("golden_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Golden Leggings", armorInfo(3, 106, ArmorSlotLegs, 0, false, VanillaArmorMaterial("gold")), enchantment.TagLeggings)
	})
	r.registerDelayed("iron_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Iron Leggings", armorInfo(5, 226, ArmorSlotLegs, 0, false, VanillaArmorMaterial("iron")), enchantment.TagLeggings)
	})
	r.registerDelayed("leather_pants", func(id ItemIdentifier) Item {
		return NewArmor(id, "Leather Pants", armorInfo(2, 76, ArmorSlotLegs, 0, false, VanillaArmorMaterial("leather")), enchantment.TagLeggings)
	})
	r.registerDelayed("netherite_leggings", func(id ItemIdentifier) Item {
		return NewArmor(id, "Netherite Leggings", armorInfo(6, 556, ArmorSlotLegs, 3, true, VanillaArmorMaterial("netherite")), enchantment.TagLeggings)
	})
}

func (r *vanillaItemsRegistry) registerSmithingTemplates() {
	r.register("netherite_upgrade_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Netherite Upgrade Smithing Template") })
	r.register("coast_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Coast Armor Trim Smithing Template") })
	r.register("dune_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Dune Armor Trim Smithing Template") })
	r.register("eye_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Eye Armor Trim Smithing Template") })
	r.register("host_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Host Armor Trim Smithing Template") })
	r.register("raiser_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Raiser Armor Trim Smithing Template") })
	r.register("rib_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Rib Armor Trim Smithing Template") })
	r.register("sentry_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Sentry Armor Trim Smithing Template") })
	r.register("shaper_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Shaper Armor Trim Smithing Template") })
	r.register("silence_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Silence Armor Trim Smithing Template") })
	r.register("snout_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Snout Armor Trim Smithing Template") })
	r.register("spire_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Spire Armor Trim Smithing Template") })
	r.register("tide_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Tide Armor Trim Smithing Template") })
	r.register("vex_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Vex Armor Trim Smithing Template") })
	r.register("ward_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Ward Armor Trim Smithing Template") })
	r.register("wayfinder_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Wayfinder Armor Trim Smithing Template") })
	r.register("wild_armor_trim_smithing_template", func(id ItemIdentifier) Item { return NewItem(id, "Wild Armor Trim Smithing Template") })
}

// VanillaAir is VanillaItems::AIR().
func VanillaAir() Item { return VanillaItem("air") }

// VanillaApple is VanillaItems::APPLE().
func VanillaApple() Item { return VanillaItem("apple") }

// VanillaArrow is VanillaItems::ARROW().
func VanillaArrow() Item { return VanillaItem("arrow") }

// VanillaBakedPotato is VanillaItems::BAKED_POTATO().
func VanillaBakedPotato() Item { return VanillaItem("baked_potato") }

// VanillaBeetrootSeeds is VanillaItems::BEETROOT_SEEDS().
func VanillaBeetrootSeeds() Item { return VanillaItem("beetroot_seeds") }

// VanillaBeetrootSoup is VanillaItems::BEETROOT_SOUP().
func VanillaBeetrootSoup() Item { return VanillaItem("beetroot_soup") }

// VanillaBlazeRod is VanillaItems::BLAZE_ROD().
func VanillaBlazeRod() Item { return VanillaItem("blaze_rod") }

// VanillaBoneMeal is VanillaItems::BONE_MEAL().
func VanillaBoneMeal() Item { return VanillaItem("bone_meal") }

// VanillaBook is VanillaItems::BOOK().
func VanillaBook() Item { return VanillaItem("book") }

// VanillaBow is VanillaItems::BOW().
func VanillaBow() Item { return VanillaItem("bow") }

// VanillaBowl is VanillaItems::BOWL().
func VanillaBowl() Item { return VanillaItem("bowl") }

// VanillaBread is VanillaItems::BREAD().
func VanillaBread() Item { return VanillaItem("bread") }

// VanillaBucket is VanillaItems::BUCKET().
func VanillaBucket() Item { return VanillaItem("bucket") }

// VanillaCarrot is VanillaItems::CARROT().
func VanillaCarrot() Item { return VanillaItem("carrot") }

// VanillaCharcoal is VanillaItems::CHARCOAL().
func VanillaCharcoal() Item { return VanillaItem("charcoal") }

// VanillaClock is VanillaItems::CLOCK().
func VanillaClock() Item { return VanillaItem("clock") }

// VanillaClownfish is VanillaItems::CLOWNFISH().
func VanillaClownfish() Item { return VanillaItem("clownfish") }

// VanillaCoal is VanillaItems::COAL().
func VanillaCoal() Item { return VanillaItem("coal") }

// VanillaCocoaBeans is VanillaItems::COCOA_BEANS().
func VanillaCocoaBeans() Item { return VanillaItem("cocoa_beans") }

// VanillaCompass is VanillaItems::COMPASS().
func VanillaCompass() Item { return VanillaItem("compass") }

// VanillaCookedChicken is VanillaItems::COOKED_CHICKEN().
func VanillaCookedChicken() Item { return VanillaItem("cooked_chicken") }

// VanillaCookedFish is VanillaItems::COOKED_FISH().
func VanillaCookedFish() Item { return VanillaItem("cooked_fish") }

// VanillaCookedMutton is VanillaItems::COOKED_MUTTON().
func VanillaCookedMutton() Item { return VanillaItem("cooked_mutton") }

// VanillaCookedPorkchop is VanillaItems::COOKED_PORKCHOP().
func VanillaCookedPorkchop() Item { return VanillaItem("cooked_porkchop") }

// VanillaCookedRabbit is VanillaItems::COOKED_RABBIT().
func VanillaCookedRabbit() Item { return VanillaItem("cooked_rabbit") }

// VanillaCookedSalmon is VanillaItems::COOKED_SALMON().
func VanillaCookedSalmon() Item { return VanillaItem("cooked_salmon") }

// VanillaCookie is VanillaItems::COOKIE().
func VanillaCookie() Item { return VanillaItem("cookie") }

// VanillaDriedKelp is VanillaItems::DRIED_KELP().
func VanillaDriedKelp() Item { return VanillaItem("dried_kelp") }

// VanillaDye is VanillaItems::DYE().
func VanillaDye() Item { return VanillaItem("dye") }

// VanillaEgg is VanillaItems::EGG().
func VanillaEgg() Item { return VanillaItem("egg") }

// VanillaEnchantedBook is VanillaItems::ENCHANTED_BOOK().
func VanillaEnchantedBook() Item { return VanillaItem("enchanted_book") }

// VanillaEnchantedGoldenApple is VanillaItems::ENCHANTED_GOLDEN_APPLE().
func VanillaEnchantedGoldenApple() Item { return VanillaItem("enchanted_golden_apple") }

// VanillaEndCrystal is VanillaItems::END_CRYSTAL().
func VanillaEndCrystal() Item { return VanillaItem("end_crystal") }

// VanillaEnderPearl is VanillaItems::ENDER_PEARL().
func VanillaEnderPearl() Item { return VanillaItem("ender_pearl") }

// VanillaExperienceBottle is VanillaItems::EXPERIENCE_BOTTLE().
func VanillaExperienceBottle() Item { return VanillaItem("experience_bottle") }

// VanillaFireworkStar is VanillaItems::FIREWORK_STAR().
func VanillaFireworkStar() Item { return VanillaItem("firework_star") }

// VanillaFishingRod is VanillaItems::FISHING_ROD().
func VanillaFishingRod() Item { return VanillaItem("fishing_rod") }

// VanillaFlintAndSteel is VanillaItems::FLINT_AND_STEEL().
func VanillaFlintAndSteel() Item { return VanillaItem("flint_and_steel") }

// VanillaGlassBottle is VanillaItems::GLASS_BOTTLE().
func VanillaGlassBottle() Item { return VanillaItem("glass_bottle") }

// VanillaGlowBerries is VanillaItems::GLOW_BERRIES().
func VanillaGlowBerries() Item { return VanillaItem("glow_berries") }

// VanillaGoatHorn is VanillaItems::GOAT_HORN().
func VanillaGoatHorn() Item { return VanillaItem("goat_horn") }

// VanillaGoldenApple is VanillaItems::GOLDEN_APPLE().
func VanillaGoldenApple() Item { return VanillaItem("golden_apple") }

// VanillaGoldenCarrot is VanillaItems::GOLDEN_CARROT().
func VanillaGoldenCarrot() Item { return VanillaItem("golden_carrot") }

// VanillaHoneyBottle is VanillaItems::HONEY_BOTTLE().
func VanillaHoneyBottle() Item { return VanillaItem("honey_bottle") }

// VanillaIceBomb is VanillaItems::ICE_BOMB().
func VanillaIceBomb() Item { return VanillaItem("ice_bomb") }

// VanillaInkSac is VanillaItems::INK_SAC().
func VanillaInkSac() Item { return VanillaItem("ink_sac") }

// VanillaIronIngot is VanillaItems::IRON_INGOT().
func VanillaIronIngot() Item { return VanillaItem("iron_ingot") }

// VanillaLapisLazuli is VanillaItems::LAPIS_LAZULI().
func VanillaLapisLazuli() Item { return VanillaItem("lapis_lazuli") }

// VanillaLingeringPotion is VanillaItems::LINGERING_POTION().
func VanillaLingeringPotion() Item { return VanillaItem("lingering_potion") }

// VanillaMedicine is VanillaItems::MEDICINE().
func VanillaMedicine() Item { return VanillaItem("medicine") }

// VanillaMelonSeeds is VanillaItems::MELON_SEEDS().
func VanillaMelonSeeds() Item { return VanillaItem("melon_seeds") }

// VanillaMilkBucket is VanillaItems::MILK_BUCKET().
func VanillaMilkBucket() Item { return VanillaItem("milk_bucket") }

// VanillaMinecart is VanillaItems::MINECART().
func VanillaMinecart() Item { return VanillaItem("minecart") }

// VanillaMushroomStew is VanillaItems::MUSHROOM_STEW().
func VanillaMushroomStew() Item { return VanillaItem("mushroom_stew") }

// VanillaPainting is VanillaItems::PAINTING().
func VanillaPainting() Item { return VanillaItem("painting") }

// VanillaPitcherPod is VanillaItems::PITCHER_POD().
func VanillaPitcherPod() Item { return VanillaItem("pitcher_pod") }

// VanillaPoisonousPotato is VanillaItems::POISONOUS_POTATO().
func VanillaPoisonousPotato() Item { return VanillaItem("poisonous_potato") }

// VanillaPotato is VanillaItems::POTATO().
func VanillaPotato() Item { return VanillaItem("potato") }

// VanillaPotion is VanillaItems::POTION().
func VanillaPotion() Item { return VanillaItem("potion") }

// VanillaPufferfish is VanillaItems::PUFFERFISH().
func VanillaPufferfish() Item { return VanillaItem("pufferfish") }

// VanillaPumpkinPie is VanillaItems::PUMPKIN_PIE().
func VanillaPumpkinPie() Item { return VanillaItem("pumpkin_pie") }

// VanillaPumpkinSeeds is VanillaItems::PUMPKIN_SEEDS().
func VanillaPumpkinSeeds() Item { return VanillaItem("pumpkin_seeds") }

// VanillaRabbitStew is VanillaItems::RABBIT_STEW().
func VanillaRabbitStew() Item { return VanillaItem("rabbit_stew") }

// VanillaRawBeef is VanillaItems::RAW_BEEF().
func VanillaRawBeef() Item { return VanillaItem("raw_beef") }

// VanillaRawChicken is VanillaItems::RAW_CHICKEN().
func VanillaRawChicken() Item { return VanillaItem("raw_chicken") }

// VanillaRawFish is VanillaItems::RAW_FISH().
func VanillaRawFish() Item { return VanillaItem("raw_fish") }

// VanillaRawMutton is VanillaItems::RAW_MUTTON().
func VanillaRawMutton() Item { return VanillaItem("raw_mutton") }

// VanillaRawPorkchop is VanillaItems::RAW_PORKCHOP().
func VanillaRawPorkchop() Item { return VanillaItem("raw_porkchop") }

// VanillaRawRabbit is VanillaItems::RAW_RABBIT().
func VanillaRawRabbit() Item { return VanillaItem("raw_rabbit") }

// VanillaRawSalmon is VanillaItems::RAW_SALMON().
func VanillaRawSalmon() Item { return VanillaItem("raw_salmon") }

// VanillaRedstoneDust is VanillaItems::REDSTONE_DUST().
func VanillaRedstoneDust() Item { return VanillaItem("redstone_dust") }

// VanillaRottenFlesh is VanillaItems::ROTTEN_FLESH().
func VanillaRottenFlesh() Item { return VanillaItem("rotten_flesh") }

// VanillaShears is VanillaItems::SHEARS().
func VanillaShears() Item { return VanillaItem("shears") }

// VanillaSnowball is VanillaItems::SNOWBALL().
func VanillaSnowball() Item { return VanillaItem("snowball") }

// VanillaSpiderEye is VanillaItems::SPIDER_EYE().
func VanillaSpiderEye() Item { return VanillaItem("spider_eye") }

// VanillaSplashPotion is VanillaItems::SPLASH_POTION().
func VanillaSplashPotion() Item { return VanillaItem("splash_potion") }

// VanillaSpyglass is VanillaItems::SPYGLASS().
func VanillaSpyglass() Item { return VanillaItem("spyglass") }

// VanillaSquidSpawnEgg is VanillaItems::SQUID_SPAWN_EGG().
func VanillaSquidSpawnEgg() Item { return VanillaItem("squid_spawn_egg") }

// VanillaSteak is VanillaItems::STEAK().
func VanillaSteak() Item { return VanillaItem("steak") }

// VanillaStick is VanillaItems::STICK().
func VanillaStick() Item { return VanillaItem("stick") }

// VanillaString is VanillaItems::STRING().
func VanillaString() Item { return VanillaItem("string") }

// VanillaSuspiciousStew is VanillaItems::SUSPICIOUS_STEW().
func VanillaSuspiciousStew() Item { return VanillaItem("suspicious_stew") }

// VanillaSweetBerries is VanillaItems::SWEET_BERRIES().
func VanillaSweetBerries() Item { return VanillaItem("sweet_berries") }

// VanillaTorchflowerSeeds is VanillaItems::TORCHFLOWER_SEEDS().
func VanillaTorchflowerSeeds() Item { return VanillaItem("torchflower_seeds") }

// VanillaTotem is VanillaItems::TOTEM().
func VanillaTotem() Item { return VanillaItem("totem") }

// VanillaTrident is VanillaItems::TRIDENT().
func VanillaTrident() Item { return VanillaItem("trident") }

// VanillaVillagerSpawnEgg is VanillaItems::VILLAGER_SPAWN_EGG().
func VanillaVillagerSpawnEgg() Item { return VanillaItem("villager_spawn_egg") }

// VanillaWheatSeeds is VanillaItems::WHEAT_SEEDS().
func VanillaWheatSeeds() Item { return VanillaItem("wheat_seeds") }

// VanillaWritableBook is VanillaItems::WRITABLE_BOOK().
func VanillaWritableBook() Item { return VanillaItem("writable_book") }

// VanillaWrittenBook is VanillaItems::WRITTEN_BOOK().
func VanillaWrittenBook() Item { return VanillaItem("written_book") }

// VanillaZombieSpawnEgg is VanillaItems::ZOMBIE_SPAWN_EGG().
func VanillaZombieSpawnEgg() Item { return VanillaItem("zombie_spawn_egg") }
