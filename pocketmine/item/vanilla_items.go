package item

import blockutils "pocketmine-go/pocketmine/block/utils"

// VanillaItems is a port of a slice of pocketmine\item\VanillaItems (itself generated from
// VanillaItemsInputs.php - porting the full generator, including every armor/tool-tier/spawn-egg/
// smithing-template entry, is its own follow-up undertaking, matching VanillaBlocks' own documented
// scope note). This covers every item type that already has a real concrete Go implementation in
// this package and a matching entry in VanillaItemsInputs.php - food, dyes, potions, records, and
// other simple items - not yet tools/armor (those need ToolTier/ArmorMaterial iteration tables this
// pass doesn't build out). Each getter matches VanillaBlocks' own shape: lazily construct a cached
// singleton on first use, then return a Clone() of it every call (items are mutable per-instance,
// so callers must never share the singleton itself).
//
// Every display name and item-type-ID mapping below is copied directly from the real
// VanillaItemsInputs.php registration for that key, not guessed.
var (
	vanillaApple                Item
	vanillaBakedPotato          Item
	vanillaBeetrootSeeds        Item
	vanillaBeetrootSoup         Item
	vanillaBlazeRod             Item
	vanillaBoneMeal             Item
	vanillaBook                 Item
	vanillaBowl                 Item
	vanillaBread                Item
	vanillaBucket               Item
	vanillaCarrot               Item
	vanillaCharcoal             Item
	vanillaClock                Item
	vanillaClownfish            Item
	vanillaCoal                 Item
	vanillaCocoaBeans           Item
	vanillaCompass              Item
	vanillaCookedChicken        Item
	vanillaCookedFish           Item
	vanillaCookedMutton         Item
	vanillaCookedPorkchop       Item
	vanillaCookedRabbit         Item
	vanillaCookedSalmon         Item
	vanillaCookie               Item
	vanillaDriedKelp            Item
	vanillaDye                  Item
	vanillaEnchantedBook        Item
	vanillaEnchantedGoldenApple Item
	vanillaFireworkStar         Item
	vanillaFishingRod           Item
	vanillaFlintAndSteel        Item
	vanillaGlassBottle          Item
	vanillaGlowBerries          Item
	vanillaGoatHorn             Item
	vanillaGoldenApple          Item
	vanillaGoldenCarrot         Item
	vanillaHoneyBottle          Item
	vanillaLingeringPotion      Item
	vanillaMedicine             Item
	vanillaMelonSeeds           Item
	vanillaMilkBucket           Item
	vanillaMinecart             Item
	vanillaMushroomStew         Item
	vanillaPitcherPod           Item
	vanillaPoisonousPotato      Item
	vanillaPotato               Item
	vanillaPotion               Item
	vanillaPufferfish           Item
	vanillaPumpkinPie           Item
	vanillaPumpkinSeeds         Item
	vanillaRabbitStew           Item
	vanillaRawBeef              Item
	vanillaRawChicken           Item
	vanillaRawFish              Item
	vanillaRawMutton            Item
	vanillaRawPorkchop          Item
	vanillaRawRabbit            Item
	vanillaRawSalmon            Item
	vanillaRedstoneDust         Item
	vanillaRottenFlesh          Item
	vanillaShears               Item
	vanillaSpiderEye            Item
	vanillaSplashPotion         Item
	vanillaSpyglass             Item
	vanillaSteak                Item
	vanillaStick                Item
	vanillaString               Item
	vanillaSuspiciousStew       Item
	vanillaSweetBerries         Item
	vanillaTorchflowerSeeds     Item
	vanillaTotem                Item
	vanillaTrident              Item
	vanillaWheatSeeds           Item
	vanillaWritableBook         Item
	vanillaWrittenBook          Item

	vanillaRecords map[int]Item
)

func VanillaApple() Item {
	if vanillaApple == nil {
		vanillaApple = NewApple(NewItemIdentifier(APPLE), "Apple")
	}
	return vanillaApple.Clone()
}

func VanillaBakedPotato() Item {
	if vanillaBakedPotato == nil {
		vanillaBakedPotato = NewBakedPotato(NewItemIdentifier(BAKED_POTATO), "Baked Potato")
	}
	return vanillaBakedPotato.Clone()
}

func VanillaBeetrootSeeds() Item {
	if vanillaBeetrootSeeds == nil {
		vanillaBeetrootSeeds = NewBeetrootSeeds(NewItemIdentifier(BEETROOT_SEEDS), "Beetroot Seeds")
	}
	return vanillaBeetrootSeeds.Clone()
}

func VanillaBeetrootSoup() Item {
	if vanillaBeetrootSoup == nil {
		vanillaBeetrootSoup = NewBeetrootSoup(NewItemIdentifier(BEETROOT_SOUP), "Beetroot Soup")
	}
	return vanillaBeetrootSoup.Clone()
}

func VanillaBlazeRod() Item {
	if vanillaBlazeRod == nil {
		vanillaBlazeRod = NewBlazeRod(NewItemIdentifier(BLAZE_ROD), "Blaze Rod")
	}
	return vanillaBlazeRod.Clone()
}

// VanillaBoneMeal is a port of VanillaItems::BONE_MEAL() - real PHP reuses the Fertilizer class
// under the "Bone Meal" display name.
func VanillaBoneMeal() Item {
	if vanillaBoneMeal == nil {
		vanillaBoneMeal = NewFertilizer(NewItemIdentifier(BONE_MEAL), "Bone Meal")
	}
	return vanillaBoneMeal.Clone()
}

func VanillaBook() Item {
	if vanillaBook == nil {
		vanillaBook = NewBook(NewItemIdentifier(BOOK), "Book")
	}
	return vanillaBook.Clone()
}

func VanillaBowl() Item {
	if vanillaBowl == nil {
		vanillaBowl = NewBowl(NewItemIdentifier(BOWL), "Bowl")
	}
	return vanillaBowl.Clone()
}

func VanillaBread() Item {
	if vanillaBread == nil {
		vanillaBread = NewBread(NewItemIdentifier(BREAD), "Bread")
	}
	return vanillaBread.Clone()
}

func VanillaBucket() Item {
	if vanillaBucket == nil {
		vanillaBucket = NewBucket(NewItemIdentifier(BUCKET), "Bucket")
	}
	return vanillaBucket.Clone()
}

func VanillaCarrot() Item {
	if vanillaCarrot == nil {
		vanillaCarrot = NewCarrot(NewItemIdentifier(CARROT), "Carrot")
	}
	return vanillaCarrot.Clone()
}

// VanillaCharcoal is a port of VanillaItems::CHARCOAL() - real PHP reuses the Coal class under the
// "Charcoal" display name (a separate item type ID from VanillaCoal below).
func VanillaCharcoal() Item {
	if vanillaCharcoal == nil {
		vanillaCharcoal = NewCoal(NewItemIdentifier(CHARCOAL), "Charcoal")
	}
	return vanillaCharcoal.Clone()
}

func VanillaClock() Item {
	if vanillaClock == nil {
		vanillaClock = NewClock(NewItemIdentifier(CLOCK), "Clock")
	}
	return vanillaClock.Clone()
}

func VanillaClownfish() Item {
	if vanillaClownfish == nil {
		vanillaClownfish = NewClownfish(NewItemIdentifier(CLOWNFISH), "Clownfish")
	}
	return vanillaClownfish.Clone()
}

func VanillaCoal() Item {
	if vanillaCoal == nil {
		vanillaCoal = NewCoal(NewItemIdentifier(COAL), "Coal")
	}
	return vanillaCoal.Clone()
}

func VanillaCocoaBeans() Item {
	if vanillaCocoaBeans == nil {
		vanillaCocoaBeans = NewCocoaBeans(NewItemIdentifier(COCOA_BEANS), "Cocoa Beans")
	}
	return vanillaCocoaBeans.Clone()
}

func VanillaCompass() Item {
	if vanillaCompass == nil {
		vanillaCompass = NewCompass(NewItemIdentifier(COMPASS), "Compass")
	}
	return vanillaCompass.Clone()
}

func VanillaCookedChicken() Item {
	if vanillaCookedChicken == nil {
		vanillaCookedChicken = NewCookedChicken(NewItemIdentifier(COOKED_CHICKEN), "Cooked Chicken")
	}
	return vanillaCookedChicken.Clone()
}

func VanillaCookedFish() Item {
	if vanillaCookedFish == nil {
		vanillaCookedFish = NewCookedFish(NewItemIdentifier(COOKED_FISH), "Cooked Fish")
	}
	return vanillaCookedFish.Clone()
}

func VanillaCookedMutton() Item {
	if vanillaCookedMutton == nil {
		vanillaCookedMutton = NewCookedMutton(NewItemIdentifier(COOKED_MUTTON), "Cooked Mutton")
	}
	return vanillaCookedMutton.Clone()
}

func VanillaCookedPorkchop() Item {
	if vanillaCookedPorkchop == nil {
		vanillaCookedPorkchop = NewCookedPorkchop(NewItemIdentifier(COOKED_PORKCHOP), "Cooked Porkchop")
	}
	return vanillaCookedPorkchop.Clone()
}

func VanillaCookedRabbit() Item {
	if vanillaCookedRabbit == nil {
		vanillaCookedRabbit = NewCookedRabbit(NewItemIdentifier(COOKED_RABBIT), "Cooked Rabbit")
	}
	return vanillaCookedRabbit.Clone()
}

func VanillaCookedSalmon() Item {
	if vanillaCookedSalmon == nil {
		vanillaCookedSalmon = NewCookedSalmon(NewItemIdentifier(COOKED_SALMON), "Cooked Salmon")
	}
	return vanillaCookedSalmon.Clone()
}

func VanillaCookie() Item {
	if vanillaCookie == nil {
		vanillaCookie = NewCookie(NewItemIdentifier(COOKIE), "Cookie")
	}
	return vanillaCookie.Clone()
}

func VanillaDriedKelp() Item {
	if vanillaDriedKelp == nil {
		vanillaDriedKelp = NewDriedKelp(NewItemIdentifier(DRIED_KELP), "Dried Kelp")
	}
	return vanillaDriedKelp.Clone()
}

func VanillaDye() Item {
	if vanillaDye == nil {
		vanillaDye = NewDye(NewItemIdentifier(DYE), "Dye")
	}
	return vanillaDye.Clone()
}

func VanillaEnchantedBook() Item {
	if vanillaEnchantedBook == nil {
		vanillaEnchantedBook = NewEnchantedBook(NewItemIdentifier(ENCHANTED_BOOK), "Enchanted Book")
	}
	return vanillaEnchantedBook.Clone()
}

func VanillaEnchantedGoldenApple() Item {
	if vanillaEnchantedGoldenApple == nil {
		vanillaEnchantedGoldenApple = NewGoldenAppleEnchanted(NewItemIdentifier(ENCHANTED_GOLDEN_APPLE), "Enchanted Golden Apple")
	}
	return vanillaEnchantedGoldenApple.Clone()
}

func VanillaFireworkStar() Item {
	if vanillaFireworkStar == nil {
		vanillaFireworkStar = NewFireworkStar(NewItemIdentifier(FIREWORK_STAR), "Firework Star")
	}
	return vanillaFireworkStar.Clone()
}

func VanillaFishingRod() Item {
	if vanillaFishingRod == nil {
		vanillaFishingRod = NewFishingRod(NewItemIdentifier(FISHING_ROD), "Fishing Rod")
	}
	return vanillaFishingRod.Clone()
}

func VanillaFlintAndSteel() Item {
	if vanillaFlintAndSteel == nil {
		vanillaFlintAndSteel = NewFlintSteel(NewItemIdentifier(FLINT_AND_STEEL), "Flint and Steel")
	}
	return vanillaFlintAndSteel.Clone()
}

func VanillaGlassBottle() Item {
	if vanillaGlassBottle == nil {
		vanillaGlassBottle = NewGlassBottle(NewItemIdentifier(GLASS_BOTTLE), "Glass Bottle")
	}
	return vanillaGlassBottle.Clone()
}

func VanillaGlowBerries() Item {
	if vanillaGlowBerries == nil {
		vanillaGlowBerries = NewGlowBerries(NewItemIdentifier(GLOW_BERRIES), "Glow Berries")
	}
	return vanillaGlowBerries.Clone()
}

func VanillaGoatHorn() Item {
	if vanillaGoatHorn == nil {
		vanillaGoatHorn = NewGoatHorn(NewItemIdentifier(GOAT_HORN), "Goat Horn")
	}
	return vanillaGoatHorn.Clone()
}

func VanillaGoldenApple() Item {
	if vanillaGoldenApple == nil {
		vanillaGoldenApple = NewGoldenApple(NewItemIdentifier(GOLDEN_APPLE), "Golden Apple")
	}
	return vanillaGoldenApple.Clone()
}

func VanillaGoldenCarrot() Item {
	if vanillaGoldenCarrot == nil {
		vanillaGoldenCarrot = NewGoldenCarrot(NewItemIdentifier(GOLDEN_CARROT), "Golden Carrot")
	}
	return vanillaGoldenCarrot.Clone()
}

func VanillaHoneyBottle() Item {
	if vanillaHoneyBottle == nil {
		vanillaHoneyBottle = NewHoneyBottle(NewItemIdentifier(HONEY_BOTTLE), "Honey Bottle")
	}
	return vanillaHoneyBottle.Clone()
}

// VanillaLingeringPotion is a port of VanillaItems::LINGERING_POTION() - real PHP reuses the
// SplashPotion class with linger:true.
func VanillaLingeringPotion() Item {
	if vanillaLingeringPotion == nil {
		vanillaLingeringPotion = NewSplashPotion(NewItemIdentifier(LINGERING_POTION), "Lingering Potion", true)
	}
	return vanillaLingeringPotion.Clone()
}

func VanillaMedicine() Item {
	if vanillaMedicine == nil {
		vanillaMedicine = NewMedicine(NewItemIdentifier(MEDICINE), "Medicine")
	}
	return vanillaMedicine.Clone()
}

func VanillaMelonSeeds() Item {
	if vanillaMelonSeeds == nil {
		vanillaMelonSeeds = NewMelonSeeds(NewItemIdentifier(MELON_SEEDS), "Melon Seeds")
	}
	return vanillaMelonSeeds.Clone()
}

func VanillaMilkBucket() Item {
	if vanillaMilkBucket == nil {
		vanillaMilkBucket = NewMilkBucket(NewItemIdentifier(MILK_BUCKET), "Milk Bucket")
	}
	return vanillaMilkBucket.Clone()
}

func VanillaMinecart() Item {
	if vanillaMinecart == nil {
		vanillaMinecart = NewMinecart(NewItemIdentifier(MINECART), "Minecart")
	}
	return vanillaMinecart.Clone()
}

func VanillaMushroomStew() Item {
	if vanillaMushroomStew == nil {
		vanillaMushroomStew = NewMushroomStew(NewItemIdentifier(MUSHROOM_STEW), "Mushroom Stew")
	}
	return vanillaMushroomStew.Clone()
}

func VanillaPitcherPod() Item {
	if vanillaPitcherPod == nil {
		vanillaPitcherPod = NewPitcherPod(NewItemIdentifier(PITCHER_POD), "Pitcher Pod")
	}
	return vanillaPitcherPod.Clone()
}

func VanillaPoisonousPotato() Item {
	if vanillaPoisonousPotato == nil {
		vanillaPoisonousPotato = NewPoisonousPotato(NewItemIdentifier(POISONOUS_POTATO), "Poisonous Potato")
	}
	return vanillaPoisonousPotato.Clone()
}

func VanillaPotato() Item {
	if vanillaPotato == nil {
		vanillaPotato = NewPotato(NewItemIdentifier(POTATO), "Potato")
	}
	return vanillaPotato.Clone()
}

func VanillaPotion() Item {
	if vanillaPotion == nil {
		vanillaPotion = NewPotion(NewItemIdentifier(POTION), "Potion")
	}
	return vanillaPotion.Clone()
}

func VanillaPufferfish() Item {
	if vanillaPufferfish == nil {
		vanillaPufferfish = NewPufferfish(NewItemIdentifier(PUFFERFISH), "Pufferfish")
	}
	return vanillaPufferfish.Clone()
}

func VanillaPumpkinPie() Item {
	if vanillaPumpkinPie == nil {
		vanillaPumpkinPie = NewPumpkinPie(NewItemIdentifier(PUMPKIN_PIE), "Pumpkin Pie")
	}
	return vanillaPumpkinPie.Clone()
}

func VanillaPumpkinSeeds() Item {
	if vanillaPumpkinSeeds == nil {
		vanillaPumpkinSeeds = NewPumpkinSeeds(NewItemIdentifier(PUMPKIN_SEEDS), "Pumpkin Seeds")
	}
	return vanillaPumpkinSeeds.Clone()
}

func VanillaRabbitStew() Item {
	if vanillaRabbitStew == nil {
		vanillaRabbitStew = NewRabbitStew(NewItemIdentifier(RABBIT_STEW), "Rabbit Stew")
	}
	return vanillaRabbitStew.Clone()
}

func VanillaRawBeef() Item {
	if vanillaRawBeef == nil {
		vanillaRawBeef = NewRawBeef(NewItemIdentifier(RAW_BEEF), "Raw Beef")
	}
	return vanillaRawBeef.Clone()
}

func VanillaRawChicken() Item {
	if vanillaRawChicken == nil {
		vanillaRawChicken = NewRawChicken(NewItemIdentifier(RAW_CHICKEN), "Raw Chicken")
	}
	return vanillaRawChicken.Clone()
}

func VanillaRawFish() Item {
	if vanillaRawFish == nil {
		vanillaRawFish = NewRawFish(NewItemIdentifier(RAW_FISH), "Raw Fish")
	}
	return vanillaRawFish.Clone()
}

func VanillaRawMutton() Item {
	if vanillaRawMutton == nil {
		vanillaRawMutton = NewRawMutton(NewItemIdentifier(RAW_MUTTON), "Raw Mutton")
	}
	return vanillaRawMutton.Clone()
}

func VanillaRawPorkchop() Item {
	if vanillaRawPorkchop == nil {
		vanillaRawPorkchop = NewRawPorkchop(NewItemIdentifier(RAW_PORKCHOP), "Raw Porkchop")
	}
	return vanillaRawPorkchop.Clone()
}

func VanillaRawRabbit() Item {
	if vanillaRawRabbit == nil {
		vanillaRawRabbit = NewRawRabbit(NewItemIdentifier(RAW_RABBIT), "Raw Rabbit")
	}
	return vanillaRawRabbit.Clone()
}

func VanillaRawSalmon() Item {
	if vanillaRawSalmon == nil {
		vanillaRawSalmon = NewRawSalmon(NewItemIdentifier(RAW_SALMON), "Raw Salmon")
	}
	return vanillaRawSalmon.Clone()
}

// VanillaRedstoneDust is a port of VanillaItems::REDSTONE_DUST() - real PHP's Redstone class is
// displayed simply as "Redstone".
func VanillaRedstoneDust() Item {
	if vanillaRedstoneDust == nil {
		vanillaRedstoneDust = NewRedstone(NewItemIdentifier(REDSTONE_DUST), "Redstone")
	}
	return vanillaRedstoneDust.Clone()
}

func VanillaRottenFlesh() Item {
	if vanillaRottenFlesh == nil {
		vanillaRottenFlesh = NewRottenFlesh(NewItemIdentifier(ROTTEN_FLESH), "Rotten Flesh")
	}
	return vanillaRottenFlesh.Clone()
}

func VanillaShears() Item {
	if vanillaShears == nil {
		vanillaShears = NewShears(NewItemIdentifier(SHEARS), "Shears")
	}
	return vanillaShears.Clone()
}

func VanillaSpiderEye() Item {
	if vanillaSpiderEye == nil {
		vanillaSpiderEye = NewSpiderEye(NewItemIdentifier(SPIDER_EYE), "Spider Eye")
	}
	return vanillaSpiderEye.Clone()
}

func VanillaSplashPotion() Item {
	if vanillaSplashPotion == nil {
		vanillaSplashPotion = NewSplashPotion(NewItemIdentifier(SPLASH_POTION), "Splash Potion", false)
	}
	return vanillaSplashPotion.Clone()
}

func VanillaSpyglass() Item {
	if vanillaSpyglass == nil {
		vanillaSpyglass = NewSpyglass(NewItemIdentifier(SPYGLASS), "Spyglass")
	}
	return vanillaSpyglass.Clone()
}

func VanillaSteak() Item {
	if vanillaSteak == nil {
		vanillaSteak = NewSteak(NewItemIdentifier(STEAK), "Steak")
	}
	return vanillaSteak.Clone()
}

func VanillaStick() Item {
	if vanillaStick == nil {
		vanillaStick = NewStick(NewItemIdentifier(STICK), "Stick")
	}
	return vanillaStick.Clone()
}

func VanillaString() Item {
	if vanillaString == nil {
		vanillaString = NewStringItem(NewItemIdentifier(STRING), "String")
	}
	return vanillaString.Clone()
}

func VanillaSuspiciousStew() Item {
	if vanillaSuspiciousStew == nil {
		vanillaSuspiciousStew = NewSuspiciousStew(NewItemIdentifier(SUSPICIOUS_STEW), "Suspicious Stew")
	}
	return vanillaSuspiciousStew.Clone()
}

func VanillaSweetBerries() Item {
	if vanillaSweetBerries == nil {
		vanillaSweetBerries = NewSweetBerries(NewItemIdentifier(SWEET_BERRIES), "Sweet Berries")
	}
	return vanillaSweetBerries.Clone()
}

func VanillaTorchflowerSeeds() Item {
	if vanillaTorchflowerSeeds == nil {
		vanillaTorchflowerSeeds = NewTorchflowerSeeds(NewItemIdentifier(TORCHFLOWER_SEEDS), "Torchflower Seeds")
	}
	return vanillaTorchflowerSeeds.Clone()
}

func VanillaTotem() Item {
	if vanillaTotem == nil {
		vanillaTotem = NewTotem(NewItemIdentifier(TOTEM), "Totem of Undying")
	}
	return vanillaTotem.Clone()
}

func VanillaTrident() Item {
	if vanillaTrident == nil {
		vanillaTrident = NewTrident(NewItemIdentifier(TRIDENT), "Trident")
	}
	return vanillaTrident.Clone()
}

func VanillaWheatSeeds() Item {
	if vanillaWheatSeeds == nil {
		vanillaWheatSeeds = NewWheatSeeds(NewItemIdentifier(WHEAT_SEEDS), "Wheat Seeds")
	}
	return vanillaWheatSeeds.Clone()
}

// VanillaWritableBook is a port of VanillaItems::WRITABLE_BOOK() - real PHP's display name is
// "Book & Quill", not a literal "Writable Book".
func VanillaWritableBook() Item {
	if vanillaWritableBook == nil {
		vanillaWritableBook = NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Book & Quill")
	}
	return vanillaWritableBook.Clone()
}

func VanillaWrittenBook() Item {
	if vanillaWrittenBook == nil {
		vanillaWrittenBook = NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	}
	return vanillaWrittenBook.Clone()
}

// recordEntry is one row of the real VanillaItemsInputs.php record registration table (key, Go
// RecordType constant, item type ID constant, display name) - copied directly, not guessed.
type recordEntry struct {
	typeID     int
	recordType blockutils.RecordType
	name       string
}

var recordEntries = []recordEntry{
	{RECORD_11, blockutils.RecordTypeDisk11, "Record 11"},
	{RECORD_13, blockutils.RecordTypeDisk13, "Record 13"},
	{RECORD_5, blockutils.RecordTypeDisk5, "Record 5"},
	{RECORD_BLOCKS, blockutils.RecordTypeDiskBlocks, "Record Blocks"},
	{RECORD_CAT, blockutils.RecordTypeDiskCat, "Record Cat"},
	{RECORD_CHIRP, blockutils.RecordTypeDiskChirp, "Record Chirp"},
	{RECORD_CREATOR, blockutils.RecordTypeDiskCreator, "Record Creator"},
	{RECORD_CREATOR_MUSIC_BOX, blockutils.RecordTypeDiskCreatorMusicBox, "Record Creator (Music Box)"},
	{RECORD_FAR, blockutils.RecordTypeDiskFar, "Record Far"},
	{RECORD_LAVA_CHICKEN, blockutils.RecordTypeDiskLavaChicken, "Record Lava Chicken"},
	{RECORD_MALL, blockutils.RecordTypeDiskMall, "Record Mall"},
	{RECORD_MELLOHI, blockutils.RecordTypeDiskMellohi, "Record Mellohi"},
	{RECORD_OTHERSIDE, blockutils.RecordTypeDiskOtherside, "Record Otherside"},
	{RECORD_PIGSTEP, blockutils.RecordTypeDiskPigstep, "Record Pigstep"},
	{RECORD_PRECIPICE, blockutils.RecordTypeDiskPrecipice, "Record Precipice"},
	{RECORD_RELIC, blockutils.RecordTypeDiskRelic, "Record Relic"},
	{RECORD_STAL, blockutils.RecordTypeDiskStal, "Record Stal"},
	{RECORD_STRAD, blockutils.RecordTypeDiskStrad, "Record Strad"},
	{RECORD_WAIT, blockutils.RecordTypeDiskWait, "Record Wait"},
	{RECORD_WARD, blockutils.RecordTypeDiskWard, "Record Ward"},
}

// vanillaRecord is a port of the several VanillaItems::RECORD_*() getters, sharing one lazily-
// initialized cache map keyed by item type ID (rather than 20 separate package-level vars, purely
// to keep this section shorter - functionally identical to every other getter in this file).
func vanillaRecord(typeID int, recordType blockutils.RecordType, name string) Item {
	if vanillaRecords == nil {
		vanillaRecords = map[int]Item{}
	}
	if vanillaRecords[typeID] == nil {
		vanillaRecords[typeID] = NewRecord(NewItemIdentifier(typeID), recordType, name)
	}
	return vanillaRecords[typeID].Clone()
}

// VanillaRecords returns every vanilla music disc item, in VanillaItemsInputs.php's own
// registration order.
func VanillaRecords() []Item {
	items := make([]Item, len(recordEntries))
	for i, e := range recordEntries {
		items[i] = vanillaRecord(e.typeID, e.recordType, e.name)
	}
	return items
}
