package item

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"sync"
)

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

var vanillaAppleOnce sync.Once

func VanillaApple() Item {
	vanillaAppleOnce.Do(func() {
		vanillaApple = NewApple(NewItemIdentifier(APPLE), "Apple")
	})
	return vanillaApple.Clone()
}

var vanillaBakedPotatoOnce sync.Once

func VanillaBakedPotato() Item {
	vanillaBakedPotatoOnce.Do(func() {
		vanillaBakedPotato = NewBakedPotato(NewItemIdentifier(BAKED_POTATO), "Baked Potato")
	})
	return vanillaBakedPotato.Clone()
}

var vanillaBeetrootSeedsOnce sync.Once

func VanillaBeetrootSeeds() Item {
	vanillaBeetrootSeedsOnce.Do(func() {
		vanillaBeetrootSeeds = NewBeetrootSeeds(NewItemIdentifier(BEETROOT_SEEDS), "Beetroot Seeds")
	})
	return vanillaBeetrootSeeds.Clone()
}

var vanillaBeetrootSoupOnce sync.Once

func VanillaBeetrootSoup() Item {
	vanillaBeetrootSoupOnce.Do(func() {
		vanillaBeetrootSoup = NewBeetrootSoup(NewItemIdentifier(BEETROOT_SOUP), "Beetroot Soup")
	})
	return vanillaBeetrootSoup.Clone()
}

var vanillaBlazeRodOnce sync.Once

func VanillaBlazeRod() Item {
	vanillaBlazeRodOnce.Do(func() {
		vanillaBlazeRod = NewBlazeRod(NewItemIdentifier(BLAZE_ROD), "Blaze Rod")
	})
	return vanillaBlazeRod.Clone()
}

// VanillaBoneMeal is a port of VanillaItems::BONE_MEAL() - real PHP reuses the Fertilizer class
// under the "Bone Meal" display name.
var vanillaBoneMealOnce sync.Once

func VanillaBoneMeal() Item {
	vanillaBoneMealOnce.Do(func() {
		vanillaBoneMeal = NewFertilizer(NewItemIdentifier(BONE_MEAL), "Bone Meal")
	})
	return vanillaBoneMeal.Clone()
}

var vanillaBookOnce sync.Once

func VanillaBook() Item {
	vanillaBookOnce.Do(func() {
		vanillaBook = NewBook(NewItemIdentifier(BOOK), "Book")
	})
	return vanillaBook.Clone()
}

var vanillaBowlOnce sync.Once

func VanillaBowl() Item {
	vanillaBowlOnce.Do(func() {
		vanillaBowl = NewBowl(NewItemIdentifier(BOWL), "Bowl")
	})
	return vanillaBowl.Clone()
}

var vanillaBreadOnce sync.Once

func VanillaBread() Item {
	vanillaBreadOnce.Do(func() {
		vanillaBread = NewBread(NewItemIdentifier(BREAD), "Bread")
	})
	return vanillaBread.Clone()
}

var vanillaBucketOnce sync.Once

func VanillaBucket() Item {
	vanillaBucketOnce.Do(func() {
		vanillaBucket = NewBucket(NewItemIdentifier(BUCKET), "Bucket")
	})
	return vanillaBucket.Clone()
}

var vanillaCarrotOnce sync.Once

func VanillaCarrot() Item {
	vanillaCarrotOnce.Do(func() {
		vanillaCarrot = NewCarrot(NewItemIdentifier(CARROT), "Carrot")
	})
	return vanillaCarrot.Clone()
}

// VanillaCharcoal is a port of VanillaItems::CHARCOAL() - real PHP reuses the Coal class under the
// "Charcoal" display name (a separate item type ID from VanillaCoal below).
var vanillaCharcoalOnce sync.Once

func VanillaCharcoal() Item {
	vanillaCharcoalOnce.Do(func() {
		vanillaCharcoal = NewCoal(NewItemIdentifier(CHARCOAL), "Charcoal")
	})
	return vanillaCharcoal.Clone()
}

var vanillaClockOnce sync.Once

func VanillaClock() Item {
	vanillaClockOnce.Do(func() {
		vanillaClock = NewClock(NewItemIdentifier(CLOCK), "Clock")
	})
	return vanillaClock.Clone()
}

var vanillaClownfishOnce sync.Once

func VanillaClownfish() Item {
	vanillaClownfishOnce.Do(func() {
		vanillaClownfish = NewClownfish(NewItemIdentifier(CLOWNFISH), "Clownfish")
	})
	return vanillaClownfish.Clone()
}

var vanillaCoalOnce sync.Once

func VanillaCoal() Item {
	vanillaCoalOnce.Do(func() {
		vanillaCoal = NewCoal(NewItemIdentifier(COAL), "Coal")
	})
	return vanillaCoal.Clone()
}

var vanillaCocoaBeansOnce sync.Once

func VanillaCocoaBeans() Item {
	vanillaCocoaBeansOnce.Do(func() {
		vanillaCocoaBeans = NewCocoaBeans(NewItemIdentifier(COCOA_BEANS), "Cocoa Beans")
	})
	return vanillaCocoaBeans.Clone()
}

var vanillaCompassOnce sync.Once

func VanillaCompass() Item {
	vanillaCompassOnce.Do(func() {
		vanillaCompass = NewCompass(NewItemIdentifier(COMPASS), "Compass")
	})
	return vanillaCompass.Clone()
}

var vanillaCookedChickenOnce sync.Once

func VanillaCookedChicken() Item {
	vanillaCookedChickenOnce.Do(func() {
		vanillaCookedChicken = NewCookedChicken(NewItemIdentifier(COOKED_CHICKEN), "Cooked Chicken")
	})
	return vanillaCookedChicken.Clone()
}

var vanillaCookedFishOnce sync.Once

func VanillaCookedFish() Item {
	vanillaCookedFishOnce.Do(func() {
		vanillaCookedFish = NewCookedFish(NewItemIdentifier(COOKED_FISH), "Cooked Fish")
	})
	return vanillaCookedFish.Clone()
}

var vanillaCookedMuttonOnce sync.Once

func VanillaCookedMutton() Item {
	vanillaCookedMuttonOnce.Do(func() {
		vanillaCookedMutton = NewCookedMutton(NewItemIdentifier(COOKED_MUTTON), "Cooked Mutton")
	})
	return vanillaCookedMutton.Clone()
}

var vanillaCookedPorkchopOnce sync.Once

func VanillaCookedPorkchop() Item {
	vanillaCookedPorkchopOnce.Do(func() {
		vanillaCookedPorkchop = NewCookedPorkchop(NewItemIdentifier(COOKED_PORKCHOP), "Cooked Porkchop")
	})
	return vanillaCookedPorkchop.Clone()
}

var vanillaCookedRabbitOnce sync.Once

func VanillaCookedRabbit() Item {
	vanillaCookedRabbitOnce.Do(func() {
		vanillaCookedRabbit = NewCookedRabbit(NewItemIdentifier(COOKED_RABBIT), "Cooked Rabbit")
	})
	return vanillaCookedRabbit.Clone()
}

var vanillaCookedSalmonOnce sync.Once

func VanillaCookedSalmon() Item {
	vanillaCookedSalmonOnce.Do(func() {
		vanillaCookedSalmon = NewCookedSalmon(NewItemIdentifier(COOKED_SALMON), "Cooked Salmon")
	})
	return vanillaCookedSalmon.Clone()
}

var vanillaCookieOnce sync.Once

func VanillaCookie() Item {
	vanillaCookieOnce.Do(func() {
		vanillaCookie = NewCookie(NewItemIdentifier(COOKIE), "Cookie")
	})
	return vanillaCookie.Clone()
}

var vanillaDriedKelpOnce sync.Once

func VanillaDriedKelp() Item {
	vanillaDriedKelpOnce.Do(func() {
		vanillaDriedKelp = NewDriedKelp(NewItemIdentifier(DRIED_KELP), "Dried Kelp")
	})
	return vanillaDriedKelp.Clone()
}

var vanillaDyeOnce sync.Once

func VanillaDye() Item {
	vanillaDyeOnce.Do(func() {
		vanillaDye = NewDye(NewItemIdentifier(DYE), "Dye")
	})
	return vanillaDye.Clone()
}

var vanillaEnchantedBookOnce sync.Once

func VanillaEnchantedBook() Item {
	vanillaEnchantedBookOnce.Do(func() {
		vanillaEnchantedBook = NewEnchantedBook(NewItemIdentifier(ENCHANTED_BOOK), "Enchanted Book")
	})
	return vanillaEnchantedBook.Clone()
}

var vanillaEnchantedGoldenAppleOnce sync.Once

func VanillaEnchantedGoldenApple() Item {
	vanillaEnchantedGoldenAppleOnce.Do(func() {
		vanillaEnchantedGoldenApple = NewGoldenAppleEnchanted(NewItemIdentifier(ENCHANTED_GOLDEN_APPLE), "Enchanted Golden Apple")
	})
	return vanillaEnchantedGoldenApple.Clone()
}

var vanillaFireworkStarOnce sync.Once

func VanillaFireworkStar() Item {
	vanillaFireworkStarOnce.Do(func() {
		vanillaFireworkStar = NewFireworkStar(NewItemIdentifier(FIREWORK_STAR), "Firework Star")
	})
	return vanillaFireworkStar.Clone()
}

var vanillaFishingRodOnce sync.Once

func VanillaFishingRod() Item {
	vanillaFishingRodOnce.Do(func() {
		vanillaFishingRod = NewFishingRod(NewItemIdentifier(FISHING_ROD), "Fishing Rod")
	})
	return vanillaFishingRod.Clone()
}

var vanillaFlintAndSteelOnce sync.Once

func VanillaFlintAndSteel() Item {
	vanillaFlintAndSteelOnce.Do(func() {
		vanillaFlintAndSteel = NewFlintSteel(NewItemIdentifier(FLINT_AND_STEEL), "Flint and Steel")
	})
	return vanillaFlintAndSteel.Clone()
}

var vanillaGlassBottleOnce sync.Once

func VanillaGlassBottle() Item {
	vanillaGlassBottleOnce.Do(func() {
		vanillaGlassBottle = NewGlassBottle(NewItemIdentifier(GLASS_BOTTLE), "Glass Bottle")
	})
	return vanillaGlassBottle.Clone()
}

var vanillaGlowBerriesOnce sync.Once

func VanillaGlowBerries() Item {
	vanillaGlowBerriesOnce.Do(func() {
		vanillaGlowBerries = NewGlowBerries(NewItemIdentifier(GLOW_BERRIES), "Glow Berries")
	})
	return vanillaGlowBerries.Clone()
}

var vanillaGoatHornOnce sync.Once

func VanillaGoatHorn() Item {
	vanillaGoatHornOnce.Do(func() {
		vanillaGoatHorn = NewGoatHorn(NewItemIdentifier(GOAT_HORN), "Goat Horn")
	})
	return vanillaGoatHorn.Clone()
}

var vanillaGoldenAppleOnce sync.Once

func VanillaGoldenApple() Item {
	vanillaGoldenAppleOnce.Do(func() {
		vanillaGoldenApple = NewGoldenApple(NewItemIdentifier(GOLDEN_APPLE), "Golden Apple")
	})
	return vanillaGoldenApple.Clone()
}

var vanillaGoldenCarrotOnce sync.Once

func VanillaGoldenCarrot() Item {
	vanillaGoldenCarrotOnce.Do(func() {
		vanillaGoldenCarrot = NewGoldenCarrot(NewItemIdentifier(GOLDEN_CARROT), "Golden Carrot")
	})
	return vanillaGoldenCarrot.Clone()
}

var vanillaHoneyBottleOnce sync.Once

func VanillaHoneyBottle() Item {
	vanillaHoneyBottleOnce.Do(func() {
		vanillaHoneyBottle = NewHoneyBottle(NewItemIdentifier(HONEY_BOTTLE), "Honey Bottle")
	})
	return vanillaHoneyBottle.Clone()
}

// VanillaLingeringPotion is a port of VanillaItems::LINGERING_POTION() - real PHP reuses the
// SplashPotion class with linger:true.
var vanillaLingeringPotionOnce sync.Once

func VanillaLingeringPotion() Item {
	vanillaLingeringPotionOnce.Do(func() {
		vanillaLingeringPotion = NewSplashPotion(NewItemIdentifier(LINGERING_POTION), "Lingering Potion", true)
	})
	return vanillaLingeringPotion.Clone()
}

var vanillaMedicineOnce sync.Once

func VanillaMedicine() Item {
	vanillaMedicineOnce.Do(func() {
		vanillaMedicine = NewMedicine(NewItemIdentifier(MEDICINE), "Medicine")
	})
	return vanillaMedicine.Clone()
}

var vanillaMelonSeedsOnce sync.Once

func VanillaMelonSeeds() Item {
	vanillaMelonSeedsOnce.Do(func() {
		vanillaMelonSeeds = NewMelonSeeds(NewItemIdentifier(MELON_SEEDS), "Melon Seeds")
	})
	return vanillaMelonSeeds.Clone()
}

var vanillaMilkBucketOnce sync.Once

func VanillaMilkBucket() Item {
	vanillaMilkBucketOnce.Do(func() {
		vanillaMilkBucket = NewMilkBucket(NewItemIdentifier(MILK_BUCKET), "Milk Bucket")
	})
	return vanillaMilkBucket.Clone()
}

var vanillaMinecartOnce sync.Once

func VanillaMinecart() Item {
	vanillaMinecartOnce.Do(func() {
		vanillaMinecart = NewMinecart(NewItemIdentifier(MINECART), "Minecart")
	})
	return vanillaMinecart.Clone()
}

var vanillaMushroomStewOnce sync.Once

func VanillaMushroomStew() Item {
	vanillaMushroomStewOnce.Do(func() {
		vanillaMushroomStew = NewMushroomStew(NewItemIdentifier(MUSHROOM_STEW), "Mushroom Stew")
	})
	return vanillaMushroomStew.Clone()
}

var vanillaPitcherPodOnce sync.Once

func VanillaPitcherPod() Item {
	vanillaPitcherPodOnce.Do(func() {
		vanillaPitcherPod = NewPitcherPod(NewItemIdentifier(PITCHER_POD), "Pitcher Pod")
	})
	return vanillaPitcherPod.Clone()
}

var vanillaPoisonousPotatoOnce sync.Once

func VanillaPoisonousPotato() Item {
	vanillaPoisonousPotatoOnce.Do(func() {
		vanillaPoisonousPotato = NewPoisonousPotato(NewItemIdentifier(POISONOUS_POTATO), "Poisonous Potato")
	})
	return vanillaPoisonousPotato.Clone()
}

var vanillaPotatoOnce sync.Once

func VanillaPotato() Item {
	vanillaPotatoOnce.Do(func() {
		vanillaPotato = NewPotato(NewItemIdentifier(POTATO), "Potato")
	})
	return vanillaPotato.Clone()
}

var vanillaPotionOnce sync.Once

func VanillaPotion() Item {
	vanillaPotionOnce.Do(func() {
		vanillaPotion = NewPotion(NewItemIdentifier(POTION), "Potion")
	})
	return vanillaPotion.Clone()
}

var vanillaPufferfishOnce sync.Once

func VanillaPufferfish() Item {
	vanillaPufferfishOnce.Do(func() {
		vanillaPufferfish = NewPufferfish(NewItemIdentifier(PUFFERFISH), "Pufferfish")
	})
	return vanillaPufferfish.Clone()
}

var vanillaPumpkinPieOnce sync.Once

func VanillaPumpkinPie() Item {
	vanillaPumpkinPieOnce.Do(func() {
		vanillaPumpkinPie = NewPumpkinPie(NewItemIdentifier(PUMPKIN_PIE), "Pumpkin Pie")
	})
	return vanillaPumpkinPie.Clone()
}

var vanillaPumpkinSeedsOnce sync.Once

func VanillaPumpkinSeeds() Item {
	vanillaPumpkinSeedsOnce.Do(func() {
		vanillaPumpkinSeeds = NewPumpkinSeeds(NewItemIdentifier(PUMPKIN_SEEDS), "Pumpkin Seeds")
	})
	return vanillaPumpkinSeeds.Clone()
}

var vanillaRabbitStewOnce sync.Once

func VanillaRabbitStew() Item {
	vanillaRabbitStewOnce.Do(func() {
		vanillaRabbitStew = NewRabbitStew(NewItemIdentifier(RABBIT_STEW), "Rabbit Stew")
	})
	return vanillaRabbitStew.Clone()
}

var vanillaRawBeefOnce sync.Once

func VanillaRawBeef() Item {
	vanillaRawBeefOnce.Do(func() {
		vanillaRawBeef = NewRawBeef(NewItemIdentifier(RAW_BEEF), "Raw Beef")
	})
	return vanillaRawBeef.Clone()
}

var vanillaRawChickenOnce sync.Once

func VanillaRawChicken() Item {
	vanillaRawChickenOnce.Do(func() {
		vanillaRawChicken = NewRawChicken(NewItemIdentifier(RAW_CHICKEN), "Raw Chicken")
	})
	return vanillaRawChicken.Clone()
}

var vanillaRawFishOnce sync.Once

func VanillaRawFish() Item {
	vanillaRawFishOnce.Do(func() {
		vanillaRawFish = NewRawFish(NewItemIdentifier(RAW_FISH), "Raw Fish")
	})
	return vanillaRawFish.Clone()
}

var vanillaRawMuttonOnce sync.Once

func VanillaRawMutton() Item {
	vanillaRawMuttonOnce.Do(func() {
		vanillaRawMutton = NewRawMutton(NewItemIdentifier(RAW_MUTTON), "Raw Mutton")
	})
	return vanillaRawMutton.Clone()
}

var vanillaRawPorkchopOnce sync.Once

func VanillaRawPorkchop() Item {
	vanillaRawPorkchopOnce.Do(func() {
		vanillaRawPorkchop = NewRawPorkchop(NewItemIdentifier(RAW_PORKCHOP), "Raw Porkchop")
	})
	return vanillaRawPorkchop.Clone()
}

var vanillaRawRabbitOnce sync.Once

func VanillaRawRabbit() Item {
	vanillaRawRabbitOnce.Do(func() {
		vanillaRawRabbit = NewRawRabbit(NewItemIdentifier(RAW_RABBIT), "Raw Rabbit")
	})
	return vanillaRawRabbit.Clone()
}

var vanillaRawSalmonOnce sync.Once

func VanillaRawSalmon() Item {
	vanillaRawSalmonOnce.Do(func() {
		vanillaRawSalmon = NewRawSalmon(NewItemIdentifier(RAW_SALMON), "Raw Salmon")
	})
	return vanillaRawSalmon.Clone()
}

// VanillaRedstoneDust is a port of VanillaItems::REDSTONE_DUST() - real PHP's Redstone class is
// displayed simply as "Redstone".
var vanillaRedstoneDustOnce sync.Once

func VanillaRedstoneDust() Item {
	vanillaRedstoneDustOnce.Do(func() {
		vanillaRedstoneDust = NewRedstone(NewItemIdentifier(REDSTONE_DUST), "Redstone")
	})
	return vanillaRedstoneDust.Clone()
}

var vanillaRottenFleshOnce sync.Once

func VanillaRottenFlesh() Item {
	vanillaRottenFleshOnce.Do(func() {
		vanillaRottenFlesh = NewRottenFlesh(NewItemIdentifier(ROTTEN_FLESH), "Rotten Flesh")
	})
	return vanillaRottenFlesh.Clone()
}

var vanillaShearsOnce sync.Once

func VanillaShears() Item {
	vanillaShearsOnce.Do(func() {
		vanillaShears = NewShears(NewItemIdentifier(SHEARS), "Shears")
	})
	return vanillaShears.Clone()
}

var vanillaSpiderEyeOnce sync.Once

func VanillaSpiderEye() Item {
	vanillaSpiderEyeOnce.Do(func() {
		vanillaSpiderEye = NewSpiderEye(NewItemIdentifier(SPIDER_EYE), "Spider Eye")
	})
	return vanillaSpiderEye.Clone()
}

var vanillaSplashPotionOnce sync.Once

func VanillaSplashPotion() Item {
	vanillaSplashPotionOnce.Do(func() {
		vanillaSplashPotion = NewSplashPotion(NewItemIdentifier(SPLASH_POTION), "Splash Potion", false)
	})
	return vanillaSplashPotion.Clone()
}

var vanillaSpyglassOnce sync.Once

func VanillaSpyglass() Item {
	vanillaSpyglassOnce.Do(func() {
		vanillaSpyglass = NewSpyglass(NewItemIdentifier(SPYGLASS), "Spyglass")
	})
	return vanillaSpyglass.Clone()
}

var vanillaSteakOnce sync.Once

func VanillaSteak() Item {
	vanillaSteakOnce.Do(func() {
		vanillaSteak = NewSteak(NewItemIdentifier(STEAK), "Steak")
	})
	return vanillaSteak.Clone()
}

var vanillaStickOnce sync.Once

func VanillaStick() Item {
	vanillaStickOnce.Do(func() {
		vanillaStick = NewStick(NewItemIdentifier(STICK), "Stick")
	})
	return vanillaStick.Clone()
}

var vanillaStringOnce sync.Once

func VanillaString() Item {
	vanillaStringOnce.Do(func() {
		vanillaString = NewStringItem(NewItemIdentifier(STRING), "String")
	})
	return vanillaString.Clone()
}

var vanillaSuspiciousStewOnce sync.Once

func VanillaSuspiciousStew() Item {
	vanillaSuspiciousStewOnce.Do(func() {
		vanillaSuspiciousStew = NewSuspiciousStew(NewItemIdentifier(SUSPICIOUS_STEW), "Suspicious Stew")
	})
	return vanillaSuspiciousStew.Clone()
}

var vanillaSweetBerriesOnce sync.Once

func VanillaSweetBerries() Item {
	vanillaSweetBerriesOnce.Do(func() {
		vanillaSweetBerries = NewSweetBerries(NewItemIdentifier(SWEET_BERRIES), "Sweet Berries")
	})
	return vanillaSweetBerries.Clone()
}

var vanillaTorchflowerSeedsOnce sync.Once

func VanillaTorchflowerSeeds() Item {
	vanillaTorchflowerSeedsOnce.Do(func() {
		vanillaTorchflowerSeeds = NewTorchflowerSeeds(NewItemIdentifier(TORCHFLOWER_SEEDS), "Torchflower Seeds")
	})
	return vanillaTorchflowerSeeds.Clone()
}

var vanillaTotemOnce sync.Once

func VanillaTotem() Item {
	vanillaTotemOnce.Do(func() {
		vanillaTotem = NewTotem(NewItemIdentifier(TOTEM), "Totem of Undying")
	})
	return vanillaTotem.Clone()
}

var vanillaTridentOnce sync.Once

func VanillaTrident() Item {
	vanillaTridentOnce.Do(func() {
		vanillaTrident = NewTrident(NewItemIdentifier(TRIDENT), "Trident")
	})
	return vanillaTrident.Clone()
}

var vanillaWheatSeedsOnce sync.Once

func VanillaWheatSeeds() Item {
	vanillaWheatSeedsOnce.Do(func() {
		vanillaWheatSeeds = NewWheatSeeds(NewItemIdentifier(WHEAT_SEEDS), "Wheat Seeds")
	})
	return vanillaWheatSeeds.Clone()
}

// VanillaWritableBook is a port of VanillaItems::WRITABLE_BOOK() - real PHP's display name is
// "Book & Quill", not a literal "Writable Book".
var vanillaWritableBookOnce sync.Once

func VanillaWritableBook() Item {
	vanillaWritableBookOnce.Do(func() {
		vanillaWritableBook = NewWritableBook(NewItemIdentifier(WRITABLE_BOOK), "Book & Quill")
	})
	return vanillaWritableBook.Clone()
}

var vanillaWrittenBookOnce sync.Once

func VanillaWrittenBook() Item {
	vanillaWrittenBookOnce.Do(func() {
		vanillaWrittenBook = NewWrittenBook(NewItemIdentifier(WRITTEN_BOOK), "Written Book")
	})
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
	vanillaRecordsMu.Lock()
	defer vanillaRecordsMu.Unlock()
	if vanillaRecords == nil {
		vanillaRecords = map[int]Item{}
	}
	if vanillaRecords[typeID] == nil {
		vanillaRecords[typeID] = NewRecord(NewItemIdentifier(typeID), recordType, name)
	}
	return vanillaRecords[typeID].Clone()
}

// vanillaRecordsMu guards vanillaRecords (the getters may be called from worker goroutines).
var vanillaRecordsMu sync.Mutex

// VanillaRecords returns every vanilla music disc item, in VanillaItemsInputs.php's own
// registration order.
func VanillaRecords() []Item {
	items := make([]Item, len(recordEntries))
	for i, e := range recordEntries {
		items[i] = vanillaRecord(e.typeID, e.recordType, e.name)
	}
	return items
}
