package blockconvert

import (
	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// This file and vanilla_block_mappings_*.go are a port of
// pocketmine\data\bedrock\block\convert\VanillaBlockMappings: the serializer and deserializer of
// every vanilla block.

// b is Blocks::NAME() (VanillaBlocks).
func b(name string) block.Behavior { return block.VanillaBlock(name) }

// InitVanillaBlockMappings is a port of VanillaBlockMappings::init.
func InitVanillaBlockMappings(reg *BlockSerializerDeserializerRegistrar) {
	commonProperties := GetCommonProperties()
	registerSimpleIdOnlyMappings(reg)
	registerColoredMappings(reg, commonProperties)
	registerCandleMappings(reg, commonProperties)
	registerLeavesMappings(reg)
	registerSaplingMappings(reg)
	registerPlantMappings(reg, commonProperties)
	registerCoralMappings(reg, commonProperties)
	registerCopperMappings(reg, commonProperties)
	registerFlattenedEnumMappings(reg, commonProperties)
	registerFlattenedBoolMappings(reg, commonProperties)
	registerStoneLikeSlabMappings(reg)
	registerStoneLikeStairMappings(reg)
	registerStoneLikeWallMappings(reg, commonProperties)

	registerWoodMappings(reg, commonProperties)
	registerTorchMappings(reg, commonProperties)
	registerChemistryMappings(reg, commonProperties)
	register1to1CustomMappings(reg, commonProperties)

	registerSplitMappings(reg, commonProperties)
}

type idBlock struct {
	block block.Behavior
	id    string
}

func registerColoredMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	reg.MapColored(b("stained_hardened_glass"), "minecraft:hard_", "_stained_glass")
	reg.MapFlattenedId(NewFlattenedIdModel(b("stained_hardened_glass_pane")).IdComponents("minecraft:hard_", commonProperties.DyeColorIdInfix, "_stained_glass_pane").Properties(ConnectionProperties...)) // 1.26.50: connection flags

	reg.MapColored(b("carpet"), "minecraft:", "_carpet")
	reg.MapColored(b("concrete"), "minecraft:", "_concrete")
	reg.MapColored(b("concrete_powder"), "minecraft:", "_concrete_powder")
	reg.MapColored(b("dyed_shulker_box"), "minecraft:", "_shulker_box")
	reg.MapColored(b("stained_clay"), "minecraft:", "_terracotta")
	reg.MapColored(b("stained_glass"), "minecraft:", "_stained_glass")
	reg.MapFlattenedId(NewFlattenedIdModel(b("stained_glass_pane")).IdComponents("minecraft:", commonProperties.DyeColorIdInfix, "_stained_glass_pane").Properties(ConnectionProperties...)) // 1.26.50: connection flags
	reg.MapColored(b("wool"), "minecraft:", "_wool")

	reg.MapFlattenedId(NewFlattenedIdModel(b("glazed_terracotta")).
		IdComponents(
			"minecraft:",
			NewValueFromStringProperty("color", GetValueMappings().DyeColorWithSilver,
				func(b colored) blockutils.DyeColor { return b.GetColor() },
				func(b colored, v blockutils.DyeColor) { b.SetColor(v) },
			),
			"_glazed_terracotta",
		).
		Properties(commonProperties.HorizontalFacingClassic),
	)
}

type candleLike interface {
	GetCount() int
	SetCount(count int)
}

func registerCandleMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	candleProperties := []Property{
		commonProperties.Lit,
		NewIntPropertyWithOffset(ids.CANDLES, 0, 3, func(b candleLike) int { return b.GetCount() }, func(b candleLike, v int) { b.SetCount(v) }, 1),
	}
	cakeWithCandleProperties := []Property{commonProperties.Lit}
	reg.MapModel(NewModel(b("candle"), ids.CANDLE).Properties(candleProperties...))
	reg.MapModel(NewModel(b("cake_with_candle"), ids.CANDLE_CAKE).Properties(cakeWithCandleProperties...))

	reg.MapFlattenedId(NewFlattenedIdModel(b("dyed_candle")).
		IdComponents(
			"minecraft:",
			commonProperties.DyeColorIdInfix,
			"_candle",
		).
		Properties(candleProperties...),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("cake_with_dyed_candle")).
		IdComponents(
			"minecraft:",
			commonProperties.DyeColorIdInfix,
			"_candle_cake",
		).
		Properties(cakeWithCandleProperties...),
	)
}

func registerLeavesMappings(reg *BlockSerializerDeserializerRegistrar) {
	properties := []Property{
		NewBoolProperty(ids.PERSISTENT_BIT, func(b *block.Leaves) bool { return b.IsNoDecay() }, func(b *block.Leaves, v bool) { b.SetNoDecay(v) }),
		NewBoolProperty(ids.UPDATE_BIT, func(b *block.Leaves) bool { return b.IsCheckDecay() }, func(b *block.Leaves, v bool) { b.SetCheckDecay(v) }),
	}
	for _, e := range []idBlock{
		{b("acacia_leaves"), ids.ACACIA_LEAVES},
		{b("azalea_leaves"), ids.AZALEA_LEAVES},
		{b("flowering_azalea_leaves"), ids.AZALEA_LEAVES_FLOWERED},
		{b("birch_leaves"), ids.BIRCH_LEAVES},
		{b("cherry_leaves"), ids.CHERRY_LEAVES},
		{b("dark_oak_leaves"), ids.DARK_OAK_LEAVES},
		{b("jungle_leaves"), ids.JUNGLE_LEAVES},
		{b("mangrove_leaves"), ids.MANGROVE_LEAVES},
		{b("oak_leaves"), ids.OAK_LEAVES},
		{b("pale_oak_leaves"), ids.PALE_OAK_LEAVES},
		{b("spruce_leaves"), ids.SPRUCE_LEAVES},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(properties...))
	}
}

func registerSaplingMappings(reg *BlockSerializerDeserializerRegistrar) {
	properties := []Property{
		NewBoolProperty(ids.AGE_BIT, func(b *block.Sapling) bool { return b.IsReady() }, func(b *block.Sapling, v bool) { b.SetReady(v) }),
	}
	for _, e := range []idBlock{
		{b("acacia_sapling"), ids.ACACIA_SAPLING},
		{b("birch_sapling"), ids.BIRCH_SAPLING},
		{b("dark_oak_sapling"), ids.DARK_OAK_SAPLING},
		{b("jungle_sapling"), ids.JUNGLE_SAPLING},
		{b("oak_sapling"), ids.OAK_SAPLING},
		{b("spruce_sapling"), ids.SPRUCE_SAPLING},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(properties...))
	}
}

type mushroomBlockLike interface {
	GetMushroomBlockType() blockutils.MushroomBlockType
	SetMushroomBlockType(t blockutils.MushroomBlockType)
}

func registerPlantMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	reg.MapModel(NewModel(b("beetroots"), ids.BEETROOT).Properties(commonProperties.CropAgeMax7))
	reg.MapModel(NewModel(b("carrots"), ids.CARROTS).Properties(commonProperties.CropAgeMax7))
	reg.MapModel(NewModel(b("potatoes"), ids.POTATOES).Properties(commonProperties.CropAgeMax7))
	reg.MapModel(NewModel(b("wheat"), ids.WHEAT).Properties(commonProperties.CropAgeMax7))

	reg.MapModel(NewModel(b("melon_stem"), ids.MELON_STEM).Properties(commonProperties.StemProperties...))
	reg.MapModel(NewModel(b("pumpkin_stem"), ids.PUMPKIN_STEM).Properties(commonProperties.StemProperties...))

	for _, e := range []idBlock{
		{b("double_tallgrass"), ids.TALL_GRASS},
		{b("large_fern"), ids.LARGE_FERN},
		{b("lilac"), ids.LILAC},
		{b("peony"), ids.PEONY},
		{b("rose_bush"), ids.ROSE_BUSH},
		{b("sunflower"), ids.SUNFLOWER},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.DoublePlantHalf))
	}

	for _, e := range []idBlock{
		{b("brown_mushroom_block"), ids.BROWN_MUSHROOM_BLOCK},
		{b("red_mushroom_block"), ids.RED_MUSHROOM_BLOCK},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(
			NewValueFromIntProperty(ids.HUGE_MUSHROOM_BITS, GetValueMappings().MushroomBlockType,
				func(b mushroomBlockLike) blockutils.MushroomBlockType { return b.GetMushroomBlockType() },
				func(b mushroomBlockLike, v blockutils.MushroomBlockType) { b.SetMushroomBlockType(v) },
			),
		))
	}

	reg.MapModel(NewModel(b("glow_lichen"), ids.GLOW_LICHEN).Properties(commonProperties.MultiFacingFlags))
	reg.MapModel(NewModel(b("resin_clump"), ids.RESIN_CLUMP).Properties(commonProperties.MultiFacingFlags))

	reg.MapModel(NewModel(b("vines"), ids.VINE).Properties(
		NewValueSetFromIntProperty(
			ids.VINE_DIRECTION_BITS,
			newValueMap(
				p(math.North, ids.VINE_FLAG_NORTH),
				p(math.South, ids.VINE_FLAG_SOUTH),
				p(math.West, ids.VINE_FLAG_WEST),
				p(math.East, ids.VINE_FLAG_EAST),
			),
			func(b *block.Vine) []math.Facing { return b.GetFaces() },
			func(b *block.Vine, v []math.Facing) { b.SetFaces(v) },
		),
	))

	reg.MapModel(NewModel(b("sweet_berry_bush"), ids.SWEET_BERRY_BUSH).Properties(
		//TODO: berry bush only wants 0-3, but it can be bigger in MCPE due to misuse of GROWTH state which goes up to 7
		NewIntProperty(ids.GROWTH, 0, 7, func(b *block.SweetBerryBush) int { return b.GetAge() }, func(b *block.SweetBerryBush, v int) { b.SetAge(min(v, block.SweetBerryBushStageMature)) }),
	))
	reg.MapModel(NewModel(b("torchflower_crop"), ids.TORCHFLOWER_CROP).Properties(
		//TODO: this property can have values 0-7, but only 0-1 are valid
		NewIntProperty(ids.GROWTH, 0, 7,
			func(b *block.TorchflowerCrop) int {
				if b.IsReady() {
					return 1
				}
				return 0
			},
			func(b *block.TorchflowerCrop, v int) { b.SetReady(v != 0) },
		),
	))
}

func registerCoralMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	reg.MapFlattenedId(NewFlattenedIdModel(b("coral")).IdComponents(concat(commonProperties.CoralIdPrefixes, "_coral")...))
	reg.MapFlattenedId(NewFlattenedIdModel(b("coral_block")).IdComponents(concat(commonProperties.CoralIdPrefixes, "_coral_block")...))
	reg.MapFlattenedId(NewFlattenedIdModel(b("coral_fan")).
		IdComponents(concat(commonProperties.CoralIdPrefixes, "_coral_fan")...).
		Properties(
			NewValueFromIntProperty(ids.CORAL_FAN_DIRECTION, GetValueMappings().CoralAxis, func(b *block.FloorCoralFan) math.Axis { return b.GetAxis() }, func(b *block.FloorCoralFan, v math.Axis) { b.SetAxis(v) }),
		),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("wall_coral_fan")).
		IdComponents(concat(commonProperties.CoralIdPrefixes, "_coral_wall_fan")...).
		Properties(
			NewValueFromIntProperty(ids.CORAL_DIRECTION, GetValueMappings().HorizontalFacingCoral, hfGet, hfSet),
		),
	)
}

func registerCopperMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_bulb")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_bulb")...).
		Properties(
			commonProperties.Lit,
			NewBoolProperty(ids.POWERED_BIT, func(b poweredByRedstone) bool { return b.IsPowered() }, func(b poweredByRedstone, v bool) { b.SetPowered(v) }),
		),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper")).
		IdComponents(concat(
			commonProperties.CopperIdPrefixes,
			"copper",
			//HACK: the non-waxed, non-oxidised variant has a _block suffix, but none of the others do
			NewBoolFromStringProperty("bruhhhh", "", "_block", func(b copperMaterial) bool {
				return !b.IsWaxed() && b.GetOxidation() == blockutils.CopperOxidationNone
			}, nil),
		)...),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("chiseled_copper")).IdComponents(concat(commonProperties.CopperIdPrefixes, "chiseled_copper")...))
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_grate")).IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_grate")...))
	reg.MapFlattenedId(NewFlattenedIdModel(b("cut_copper")).IdComponents(concat(commonProperties.CopperIdPrefixes, "cut_copper")...))
	reg.MapFlattenedId(NewFlattenedIdModel(b("cut_copper_stairs")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "cut_copper_stairs")...).
		Properties(commonProperties.StairProperties...),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_trapdoor")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_trapdoor")...).
		Properties(commonProperties.TrapdoorProperties...),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_door")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_door")...).
		Properties(commonProperties.DoorProperties...),
	)

	reg.MapFlattenedId(NewFlattenedIdModel(b("cut_copper_slab")).
		IdComponents(concat(
			commonProperties.CopperIdPrefixes,
			commonProperties.SlabIdInfix,
			"cut_copper_slab",
		)...).
		Properties(commonProperties.SlabPositionProperty),
	)

	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_bars")).IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_bars")...).Properties(ConnectionProperties...)) // 1.26.50: connection flags
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_chain")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_chain")...).
		Properties(commonProperties.PillarAxis),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("copper_lantern")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "copper_lantern")...).
		Properties(
			NewBoolProperty(ids.HANGING, func(b hangingHolder) bool { return b.IsHanging() }, func(b hangingHolder, v bool) { b.SetHanging(v) }),
		),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("lightning_rod")).
		IdComponents(concat(commonProperties.CopperIdPrefixes, "lightning_rod")...).
		Properties(
			commonProperties.AnyFacingClassic,
			NewDummyProperty(ids.POWERED_BIT, false), //TODO
		),
	)
}

// flattenedCaveVinesVariant is a port of property\FlattenedCaveVinesVariant.
type flattenedCaveVinesVariant string

const (
	caveVinesNoBerries       flattenedCaveVinesVariant = ""
	caveVinesHeadWithBerries flattenedCaveVinesVariant = "_head_with_berries"
	caveVinesBodyWithBerries flattenedCaveVinesVariant = "_body_with_berries"
)

func registerFlattenedEnumMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	//A
	reg.MapFlattenedId(NewFlattenedIdModel(b("anvil")).
		IdComponents(
			NewValueFromStringProperty("id", newValueMap(
				p(0, ids.ANVIL),
				p(1, ids.CHIPPED_ANVIL),
				p(2, ids.DAMAGED_ANVIL),
			), func(b *block.Anvil) int { return b.GetDamage() }, func(b *block.Anvil, v int) { b.SetDamage(v) }),
		).
		Properties(commonProperties.HorizontalFacingCardinal),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("amethyst_cluster")).
		IdComponents(
			NewValueFromStringProperty("id", newValueMap(
				p(block.AmethystClusterStageSmallBud, ids.SMALL_AMETHYST_BUD),
				p(block.AmethystClusterStageMediumBud, ids.MEDIUM_AMETHYST_BUD),
				p(block.AmethystClusterStageLargeBud, ids.LARGE_AMETHYST_BUD),
				p(block.AmethystClusterStageCluster, ids.AMETHYST_CLUSTER),
			), func(b *block.AmethystCluster) int { return b.GetStage() }, func(b *block.AmethystCluster, v int) { b.SetStage(v) }),
		).
		Properties(commonProperties.BlockFace),
	)

	//C

	//This one is a special offender :<
	//I have no idea why this only has 3 IDs - there are 4 in Java and 4 visually distinct states in Bedrock
	reg.MapFlattenedId(NewFlattenedIdModel(b("cave_vines")).
		IdComponents(
			"minecraft:cave_vines",
			NewValueFromStringProperty(
				"variant",
				newValueMap(
					p(caveVinesNoBerries, string(caveVinesNoBerries)),
					p(caveVinesHeadWithBerries, string(caveVinesHeadWithBerries)),
					p(caveVinesBodyWithBerries, string(caveVinesBodyWithBerries)),
				),
				func(b *block.CaveVines) flattenedCaveVinesVariant {
					if !b.HasBerries() {
						return caveVinesNoBerries
					}
					if b.IsHead() {
						return caveVinesHeadWithBerries
					}
					return caveVinesBodyWithBerries
				},
				func(b *block.CaveVines, v flattenedCaveVinesVariant) {
					switch v {
					case caveVinesHeadWithBerries:
						b.SetBerries(true)
						b.SetHead(true)
					case caveVinesBodyWithBerries:
						b.SetBerries(true)
						b.SetHead(false)
					case caveVinesNoBerries:
						b.SetBerries(false)
						b.SetHead(false) //assume this isn't a head, since we don't have enough information
					}
				},
			),
		).
		Properties(
			NewIntProperty(ids.GROWING_PLANT_AGE, 0, 25, func(b *block.CaveVines) int { return b.GetAge() }, func(b *block.CaveVines, v int) { b.SetAge(v) }),
		),
	)

	//D
	reg.MapFlattenedId(NewFlattenedIdModel(b("dirt")).
		IdComponents(
			NewValueFromStringProperty("id", newValueMap(
				p(blockutils.DirtTypeNormal, ids.DIRT),
				p(blockutils.DirtTypeCoarse, ids.COARSE_DIRT),
				p(blockutils.DirtTypeRooted, ids.DIRT_WITH_ROOTS),
			), func(b *block.Dirt) blockutils.DirtType { return b.GetDirtType() }, func(b *block.Dirt, v blockutils.DirtType) { b.SetDirtType(v) }),
		),
	)

	//F
	reg.MapFlattenedId(NewFlattenedIdModel(b("froglight")).
		IdComponents(
			NewValueFromStringProperty("id", GetValueMappings().FroglightType, func(b *block.Froglight) blockutils.FroglightType { return b.GetFroglightType() }, func(b *block.Froglight, v blockutils.FroglightType) { b.SetFroglightType(v) }),
		).
		Properties(commonProperties.PillarAxis),
	)

	//L
	reg.MapFlattenedId(NewFlattenedIdModel(b("light")).
		IdComponents(
			"minecraft:light_block_",
			//this is a bit shit but it's easier than adapting IntProperty to support flattening :D
			NewValueFromStringProperty(
				"light_level",
				intStringMap(0, 15),
				func(b *block.Light) int { return b.GetLightLevel() },
				func(b *block.Light, v int) { b.SetLightLevel(v) },
			),
		),
	)

	//M
	reg.MapFlattenedId(NewFlattenedIdModel(b("mob_head")).
		IdComponents(
			NewValueFromStringProperty("id", GetValueMappings().MobHeadType, func(b *block.MobHead) blockutils.MobHeadType { return b.GetMobHeadType() }, func(b *block.MobHead, v blockutils.MobHeadType) { b.SetMobHeadType(v) }),
		).
		Properties(
			NewValueFromIntProperty(ids.FACING_DIRECTION, GetValueMappings().FacingExceptDown, func(b *block.MobHead) math.Facing { return b.GetFacing() }, func(b *block.MobHead, v math.Facing) { b.SetFacing(v) }),
		),
	)

	for _, e := range []struct {
		block    block.Behavior
		idSuffix string
	}{
		{b("lava"), "lava"},
		{b("water"), "water"},
	} {
		reg.MapFlattenedId(NewFlattenedIdModel(e.block).
			IdComponents(concat(commonProperties.LiquidIdPrefixes, e.idSuffix)...).
			Properties(commonProperties.LiquidData),
		)
	}
}

func registerFlattenedBoolMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	for _, e := range []struct {
		block    block.Behavior
		idSuffix string
	}{
		{b("blast_furnace"), "blast_furnace"},
		{b("furnace"), "furnace"},
		{b("smoker"), "smoker"},
	} {
		reg.MapFlattenedId(NewFlattenedIdModel(e.block).
			IdComponents(concat(commonProperties.FurnaceIdPrefixes, e.idSuffix)...).
			Properties(commonProperties.HorizontalFacingCardinal),
		)
	}

	for _, e := range []struct {
		block    block.Behavior
		idSuffix string
	}{
		{b("redstone_lamp"), "redstone_lamp"},
		{b("redstone_ore"), "redstone_ore"},
		{b("deepslate_redstone_ore"), "deepslate_redstone_ore"},
	} {
		reg.MapFlattenedId(NewFlattenedIdModel(e.block).IdComponents("minecraft:", commonProperties.LitIdInfix, e.idSuffix))
	}

	reg.MapFlattenedId(NewFlattenedIdModel(b("daylight_sensor")).
		IdComponents(
			"minecraft:daylight_detector",
			NewBoolFromStringProperty("inverted", "", "_inverted", func(b *block.DaylightSensor) bool { return b.IsInverted() }, func(b *block.DaylightSensor, v bool) { b.SetInverted(v) }),
		).
		Properties(commonProperties.AnalogRedstoneSignal),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("redstone_repeater")).
		IdComponents(
			"minecraft:",
			NewBoolFromStringProperty("powered", "un", "", func(b *block.RedstoneRepeater) bool { return b.IsPowered() }, func(b *block.RedstoneRepeater, v bool) { b.SetPowered(v) }),
			"powered_repeater",
		).
		Properties(
			commonProperties.HorizontalFacingCardinal,
			NewIntPropertyWithOffset(ids.REPEATER_DELAY, 0, 3, func(b *block.RedstoneRepeater) int { return b.GetDelay() }, func(b *block.RedstoneRepeater, v int) { b.SetDelay(v) }, 1),
		),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("redstone_comparator")).
		IdComponents(
			"minecraft:",
			//this property also appears in the state, so we ignore it in the ID
			//this is baked here purely to keep minecraft happy
			NewBoolFromStringProperty("dummy_powered", "un", "", func(b *block.RedstoneComparator) bool { return b.IsPowered() }, nil),
			"powered_comparator",
		).
		Properties(
			commonProperties.HorizontalFacingCardinal,
			NewBoolProperty(ids.OUTPUT_LIT_BIT, func(b *block.RedstoneComparator) bool { return b.IsPowered() }, func(b *block.RedstoneComparator, v bool) { b.SetPowered(v) }),
			NewBoolProperty(ids.OUTPUT_SUBTRACT_BIT, func(b *block.RedstoneComparator) bool { return b.IsSubtractMode() }, func(b *block.RedstoneComparator, v bool) { b.SetSubtractMode(v) }),
		),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("redstone_torch")).
		IdComponents(
			"minecraft:",
			NewBoolFromStringProperty("lit", "unlit_", "", func(b *block.RedstoneTorch) bool { return b.IsLit() }, func(b *block.RedstoneTorch, v bool) { b.SetLit(v) }),
			"redstone_torch",
		).
		Properties(commonProperties.TorchFacing),
	)
	reg.MapFlattenedId(NewFlattenedIdModel(b("sponge")).IdComponents(
		"minecraft:",
		NewBoolFromStringProperty("wet", "", "wet_", func(b *block.Sponge) bool { return b.IsWet() }, func(b *block.Sponge, v bool) { b.SetWet(v) }),
		"sponge",
	))
	reg.MapFlattenedId(NewFlattenedIdModel(b("tnt")).
		IdComponents(
			"minecraft:",
			NewBoolFromStringProperty("underwater", "", "underwater_", func(b *block.TNTBlock) bool { return b.DoesWorkUnderwater() }, func(b *block.TNTBlock, v bool) { b.SetWorksUnderwater(v) }),
			"tnt",
		).
		Properties(
			NewBoolProperty(ids.EXPLODE_BIT, func(b *block.TNTBlock) bool { return b.IsUnstable() }, func(b *block.TNTBlock, v bool) { b.SetUnstable(v) }),
		),
	)
}

func registerStoneLikeWallMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	for _, e := range []idBlock{
		{b("andesite_wall"), ids.ANDESITE_WALL},
		{b("blackstone_wall"), ids.BLACKSTONE_WALL},
		{b("brick_wall"), ids.BRICK_WALL},
		{b("cobbled_deepslate_wall"), ids.COBBLED_DEEPSLATE_WALL},
		{b("cobblestone_wall"), ids.COBBLESTONE_WALL},
		{b("deepslate_brick_wall"), ids.DEEPSLATE_BRICK_WALL},
		{b("deepslate_tile_wall"), ids.DEEPSLATE_TILE_WALL},
		{b("diorite_wall"), ids.DIORITE_WALL},
		{b("end_stone_brick_wall"), ids.END_STONE_BRICK_WALL},
		{b("granite_wall"), ids.GRANITE_WALL},
		{b("mossy_cobblestone_wall"), ids.MOSSY_COBBLESTONE_WALL},
		{b("mossy_stone_brick_wall"), ids.MOSSY_STONE_BRICK_WALL},
		{b("mud_brick_wall"), ids.MUD_BRICK_WALL},
		{b("nether_brick_wall"), ids.NETHER_BRICK_WALL},
		{b("polished_blackstone_brick_wall"), ids.POLISHED_BLACKSTONE_BRICK_WALL},
		{b("polished_blackstone_wall"), ids.POLISHED_BLACKSTONE_WALL},
		{b("polished_deepslate_wall"), ids.POLISHED_DEEPSLATE_WALL},
		{b("polished_tuff_wall"), ids.POLISHED_TUFF_WALL},
		{b("prismarine_wall"), ids.PRISMARINE_WALL},
		{b("red_nether_brick_wall"), ids.RED_NETHER_BRICK_WALL},
		{b("red_sandstone_wall"), ids.RED_SANDSTONE_WALL},
		{b("resin_brick_wall"), ids.RESIN_BRICK_WALL},
		{b("sandstone_wall"), ids.SANDSTONE_WALL},
		{b("stone_brick_wall"), ids.STONE_BRICK_WALL},
		{b("tuff_brick_wall"), ids.TUFF_BRICK_WALL},
		{b("tuff_wall"), ids.TUFF_WALL},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.WallProperties...))
	}
}
