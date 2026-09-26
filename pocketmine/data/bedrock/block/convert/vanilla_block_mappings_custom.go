package blockconvert

import (
	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

func registerWoodMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	//buttons
	for _, e := range []idBlock{
		{b("acacia_button"), ids.ACACIA_BUTTON},
		{b("bamboo_button"), ids.BAMBOO_BUTTON},
		{b("birch_button"), ids.BIRCH_BUTTON},
		{b("cherry_button"), ids.CHERRY_BUTTON},
		{b("crimson_button"), ids.CRIMSON_BUTTON},
		{b("dark_oak_button"), ids.DARK_OAK_BUTTON},
		{b("jungle_button"), ids.JUNGLE_BUTTON},
		{b("mangrove_button"), ids.MANGROVE_BUTTON},
		{b("oak_button"), ids.WOODEN_BUTTON},
		{b("pale_oak_button"), ids.PALE_OAK_BUTTON},
		{b("spruce_button"), ids.SPRUCE_BUTTON},
		{b("warped_button"), ids.WARPED_BUTTON},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.ButtonProperties...))
	}

	//doors
	for _, e := range []idBlock{
		{b("acacia_door"), ids.ACACIA_DOOR},
		{b("bamboo_door"), ids.BAMBOO_DOOR},
		{b("birch_door"), ids.BIRCH_DOOR},
		{b("cherry_door"), ids.CHERRY_DOOR},
		{b("crimson_door"), ids.CRIMSON_DOOR},
		{b("dark_oak_door"), ids.DARK_OAK_DOOR},
		{b("jungle_door"), ids.JUNGLE_DOOR},
		{b("mangrove_door"), ids.MANGROVE_DOOR},
		{b("oak_door"), ids.WOODEN_DOOR},
		{b("pale_oak_door"), ids.PALE_OAK_DOOR},
		{b("spruce_door"), ids.SPRUCE_DOOR},
		{b("warped_door"), ids.WARPED_DOOR},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.DoorProperties...))
	}

	//fences
	for _, e := range []idBlock{
		{b("acacia_fence"), ids.ACACIA_FENCE},
		{b("bamboo_fence"), ids.BAMBOO_FENCE},
		{b("birch_fence"), ids.BIRCH_FENCE},
		{b("cherry_fence"), ids.CHERRY_FENCE},
		{b("dark_oak_fence"), ids.DARK_OAK_FENCE},
		{b("jungle_fence"), ids.JUNGLE_FENCE},
		{b("mangrove_fence"), ids.MANGROVE_FENCE},
		{b("oak_fence"), ids.OAK_FENCE},
		{b("pale_oak_fence"), ids.PALE_OAK_FENCE},
		{b("spruce_fence"), ids.SPRUCE_FENCE},
		{b("crimson_fence"), ids.CRIMSON_FENCE},
		{b("warped_fence"), ids.WARPED_FENCE},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(ConnectionProperties...)) // 1.26.50: connection flags
	}

	for _, e := range []idBlock{
		{b("acacia_fence_gate"), ids.ACACIA_FENCE_GATE},
		{b("bamboo_fence_gate"), ids.BAMBOO_FENCE_GATE},
		{b("birch_fence_gate"), ids.BIRCH_FENCE_GATE},
		{b("cherry_fence_gate"), ids.CHERRY_FENCE_GATE},
		{b("dark_oak_fence_gate"), ids.DARK_OAK_FENCE_GATE},
		{b("jungle_fence_gate"), ids.JUNGLE_FENCE_GATE},
		{b("mangrove_fence_gate"), ids.MANGROVE_FENCE_GATE},
		{b("oak_fence_gate"), ids.FENCE_GATE},
		{b("pale_oak_fence_gate"), ids.PALE_OAK_FENCE_GATE},
		{b("spruce_fence_gate"), ids.SPRUCE_FENCE_GATE},
		{b("crimson_fence_gate"), ids.CRIMSON_FENCE_GATE},
		{b("warped_fence_gate"), ids.WARPED_FENCE_GATE},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.FenceGateProperties...))
	}

	for _, e := range []idBlock{
		{b("acacia_sign"), ids.ACACIA_STANDING_SIGN},
		{b("bamboo_sign"), ids.BAMBOO_STANDING_SIGN},
		{b("birch_sign"), ids.BIRCH_STANDING_SIGN},
		{b("cherry_sign"), ids.CHERRY_STANDING_SIGN},
		{b("dark_oak_sign"), ids.DARKOAK_STANDING_SIGN},
		{b("jungle_sign"), ids.JUNGLE_STANDING_SIGN},
		{b("mangrove_sign"), ids.MANGROVE_STANDING_SIGN},
		{b("oak_sign"), ids.STANDING_SIGN},
		{b("pale_oak_sign"), ids.PALE_OAK_STANDING_SIGN},
		{b("spruce_sign"), ids.SPRUCE_STANDING_SIGN},
		{b("crimson_sign"), ids.CRIMSON_STANDING_SIGN},
		{b("warped_sign"), ids.WARPED_STANDING_SIGN},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.FloorSignLikeRotation))
	}

	//logs
	for _, e := range []idBlock{
		{b("acacia_log"), "acacia_log"},
		{b("birch_log"), "birch_log"},
		{b("cherry_log"), "cherry_log"},
		{b("dark_oak_log"), "dark_oak_log"},
		{b("jungle_log"), "jungle_log"},
		{b("mangrove_log"), "mangrove_log"},
		{b("oak_log"), "oak_log"},
		{b("pale_oak_log"), "pale_oak_log"},
		{b("spruce_log"), "spruce_log"},
		{b("crimson_stem"), "crimson_stem"},
		{b("warped_stem"), "warped_stem"}, //all-sided logs
		{b("acacia_wood"), "acacia_wood"},
		{b("birch_wood"), "birch_wood"},
		{b("cherry_wood"), "cherry_wood"},
		{b("dark_oak_wood"), "dark_oak_wood"},
		{b("jungle_wood"), "jungle_wood"},
		{b("mangrove_wood"), "mangrove_wood"},
		{b("oak_wood"), "oak_wood"},
		{b("pale_oak_wood"), "pale_oak_wood"},
		{b("spruce_wood"), "spruce_wood"},
		{b("crimson_hyphae"), "crimson_hyphae"},
		{b("warped_hyphae"), "warped_hyphae"}, //bamboo is a special cookie - its name differs and there's no all-sided variant
		{b("bamboo_block"), "bamboo_block"},
	} {
		reg.MapFlattenedId(NewFlattenedIdModel(e.block).
			IdComponents(concat(commonProperties.WoodIdPrefixes, e.id)...).
			Properties(commonProperties.PillarAxis),
		)
	}

	//planks
	for _, e := range []idBlock{
		{b("acacia_planks"), ids.ACACIA_PLANKS},
		{b("bamboo_planks"), ids.BAMBOO_PLANKS},
		{b("bamboo_mosaic"), ids.BAMBOO_MOSAIC}, //special bamboo variant block
		{b("birch_planks"), ids.BIRCH_PLANKS},
		{b("cherry_planks"), ids.CHERRY_PLANKS},
		{b("dark_oak_planks"), ids.DARK_OAK_PLANKS},
		{b("jungle_planks"), ids.JUNGLE_PLANKS},
		{b("mangrove_planks"), ids.MANGROVE_PLANKS},
		{b("oak_planks"), ids.OAK_PLANKS},
		{b("pale_oak_planks"), ids.PALE_OAK_PLANKS},
		{b("spruce_planks"), ids.SPRUCE_PLANKS},
		{b("crimson_planks"), ids.CRIMSON_PLANKS},
		{b("warped_planks"), ids.WARPED_PLANKS},
	} {
		reg.MapSimple(e.block, e.id)
	}

	//pressure plates
	for _, e := range []idBlock{
		{b("acacia_pressure_plate"), ids.ACACIA_PRESSURE_PLATE},
		{b("bamboo_pressure_plate"), ids.BAMBOO_PRESSURE_PLATE},
		{b("birch_pressure_plate"), ids.BIRCH_PRESSURE_PLATE},
		{b("cherry_pressure_plate"), ids.CHERRY_PRESSURE_PLATE},
		{b("dark_oak_pressure_plate"), ids.DARK_OAK_PRESSURE_PLATE},
		{b("jungle_pressure_plate"), ids.JUNGLE_PRESSURE_PLATE},
		{b("mangrove_pressure_plate"), ids.MANGROVE_PRESSURE_PLATE},
		{b("oak_pressure_plate"), ids.WOODEN_PRESSURE_PLATE},
		{b("pale_oak_pressure_plate"), ids.PALE_OAK_PRESSURE_PLATE},
		{b("spruce_pressure_plate"), ids.SPRUCE_PRESSURE_PLATE},
		{b("crimson_pressure_plate"), ids.CRIMSON_PRESSURE_PLATE},
		{b("warped_pressure_plate"), ids.WARPED_PRESSURE_PLATE},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.SimplePressurePlateProperties...))
	}

	//slabs
	for _, e := range []idBlock{
		{b("acacia_slab"), "acacia"},
		{b("bamboo_slab"), "bamboo"},
		{b("bamboo_mosaic_slab"), "bamboo_mosaic"}, //special bamboo variant block
		{b("birch_slab"), "birch"},
		{b("cherry_slab"), "cherry"},
		{b("dark_oak_slab"), "dark_oak"},
		{b("jungle_slab"), "jungle"},
		{b("mangrove_slab"), "mangrove"},
		{b("oak_slab"), "oak"},
		{b("pale_oak_slab"), "pale_oak"},
		{b("spruce_slab"), "spruce"},
		{b("crimson_slab"), "crimson"},
		{b("warped_slab"), "warped"},
	} {
		reg.MapSlab(e.block, e.id)
	}

	//stairs
	for _, e := range []idBlock{
		{b("acacia_stairs"), ids.ACACIA_STAIRS},
		{b("bamboo_stairs"), ids.BAMBOO_STAIRS},
		{b("bamboo_mosaic_stairs"), ids.BAMBOO_MOSAIC_STAIRS}, //special bamboo variant block
		{b("birch_stairs"), ids.BIRCH_STAIRS},
		{b("cherry_stairs"), ids.CHERRY_STAIRS},
		{b("dark_oak_stairs"), ids.DARK_OAK_STAIRS},
		{b("jungle_stairs"), ids.JUNGLE_STAIRS},
		{b("mangrove_stairs"), ids.MANGROVE_STAIRS},
		{b("oak_stairs"), ids.OAK_STAIRS},
		{b("pale_oak_stairs"), ids.PALE_OAK_STAIRS},
		{b("spruce_stairs"), ids.SPRUCE_STAIRS},
		{b("crimson_stairs"), ids.CRIMSON_STAIRS},
		{b("warped_stairs"), ids.WARPED_STAIRS},
	} {
		reg.MapStairs(e.block, e.id)
	}

	//trapdoors
	for _, e := range []idBlock{
		{b("acacia_trapdoor"), ids.ACACIA_TRAPDOOR},
		{b("bamboo_trapdoor"), ids.BAMBOO_TRAPDOOR},
		{b("birch_trapdoor"), ids.BIRCH_TRAPDOOR},
		{b("cherry_trapdoor"), ids.CHERRY_TRAPDOOR},
		{b("dark_oak_trapdoor"), ids.DARK_OAK_TRAPDOOR},
		{b("jungle_trapdoor"), ids.JUNGLE_TRAPDOOR},
		{b("mangrove_trapdoor"), ids.MANGROVE_TRAPDOOR},
		{b("oak_trapdoor"), ids.TRAPDOOR},
		{b("pale_oak_trapdoor"), ids.PALE_OAK_TRAPDOOR},
		{b("spruce_trapdoor"), ids.SPRUCE_TRAPDOOR},
		{b("crimson_trapdoor"), ids.CRIMSON_TRAPDOOR},
		{b("warped_trapdoor"), ids.WARPED_TRAPDOOR},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.TrapdoorProperties...))
	}

	//wall signs
	for _, e := range []idBlock{
		{b("acacia_wall_sign"), ids.ACACIA_WALL_SIGN},
		{b("bamboo_wall_sign"), ids.BAMBOO_WALL_SIGN},
		{b("birch_wall_sign"), ids.BIRCH_WALL_SIGN},
		{b("cherry_wall_sign"), ids.CHERRY_WALL_SIGN},
		{b("dark_oak_wall_sign"), ids.DARKOAK_WALL_SIGN},
		{b("jungle_wall_sign"), ids.JUNGLE_WALL_SIGN},
		{b("mangrove_wall_sign"), ids.MANGROVE_WALL_SIGN},
		{b("oak_wall_sign"), ids.WALL_SIGN},
		{b("pale_oak_wall_sign"), ids.PALE_OAK_WALL_SIGN},
		{b("spruce_wall_sign"), ids.SPRUCE_WALL_SIGN},
		{b("crimson_wall_sign"), ids.CRIMSON_WALL_SIGN},
		{b("warped_wall_sign"), ids.WARPED_WALL_SIGN},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.HorizontalFacingClassic))
	}
}

func registerTorchMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	for _, e := range []idBlock{
		{b("blue_torch"), ids.COLORED_TORCH_BLUE},
		{b("copper_torch"), ids.COPPER_TORCH},
		{b("green_torch"), ids.COLORED_TORCH_GREEN},
		{b("purple_torch"), ids.COLORED_TORCH_PURPLE},
		{b("red_torch"), ids.COLORED_TORCH_RED},
		{b("soul_torch"), ids.SOUL_TORCH},
		{b("torch"), ids.TORCH},
		{b("underwater_torch"), ids.UNDERWATER_TORCH},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.TorchFacing))
	}
}

func registerChemistryMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	for _, e := range []idBlock{
		{b("compound_creator"), ids.COMPOUND_CREATOR},
		{b("element_constructor"), ids.ELEMENT_CONSTRUCTOR},
		{b("lab_table"), ids.LAB_TABLE},
		{b("material_reducer"), ids.MATERIAL_REDUCER},
	} {
		reg.MapModel(NewModel(e.block, e.id).Properties(commonProperties.HorizontalFacingSWNEInverted))
	}
}

type railShape interface {
	GetShape() int
	SetShape(shape int)
}

func railShapeProperty(max int) Property {
	return NewIntProperty(ids.RAIL_DIRECTION, 0, max, func(b railShape) int { return b.GetShape() }, func(b railShape, v int) { b.SetShape(v) })
}

func ageProperty(name string, min, max int) Property {
	return NewIntProperty(name, min, max, func(b ageable) int { return b.GetAge() }, func(b ageable, v int) { b.SetAge(v) })
}

func register1to1CustomMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	vm := GetValueMappings()
	//TODO: some of these have repeated accessor refs, we might be able to deduplicate them
	//A
	reg.MapModel(NewModel(b("activator_rail"), ids.ACTIVATOR_RAIL).Properties(
		NewBoolProperty(ids.RAIL_DATA_BIT, func(b poweredByRedstone) bool { return b.IsPowered() }, func(b poweredByRedstone, v bool) { b.SetPowered(v) }),
		railShapeProperty(5),
	))

	//B
	reg.MapModel(NewModel(b("bamboo"), ids.BAMBOO).Properties(
		NewValueFromStringProperty(ids.BAMBOO_LEAF_SIZE, vm.BambooLeafSize, func(b *block.Bamboo) int { return b.GetLeafSize() }, func(b *block.Bamboo, v int) { b.SetLeafSize(v) }),
		NewBoolProperty(ids.AGE_BIT, func(b *block.Bamboo) bool { return b.IsReady() }, func(b *block.Bamboo, v bool) { b.SetReady(v) }),
		NewBoolFromStringProperty(ids.BAMBOO_STALK_THICKNESS, ids.BAMBOO_STALK_THICKNESS_THIN, ids.BAMBOO_STALK_THICKNESS_THICK, func(b *block.Bamboo) bool { return b.IsThick() }, func(b *block.Bamboo, v bool) { b.SetThick(v) }),
	))
	reg.MapModel(NewModel(b("bamboo_sapling"), ids.BAMBOO_SAPLING).Properties(
		NewBoolProperty(ids.AGE_BIT, func(b *block.BambooSapling) bool { return b.IsReady() }, func(b *block.BambooSapling, v bool) { b.SetReady(v) }),
	))
	reg.MapModel(NewModel(b("banner"), ids.STANDING_BANNER).Properties(commonProperties.FloorSignLikeRotation))
	reg.MapModel(NewModel(b("barrel"), ids.BARREL).Properties(
		commonProperties.AnyFacingClassic,
		NewBoolProperty(ids.OPEN_BIT, func(b *block.Barrel) bool { return b.IsOpen() }, func(b *block.Barrel, v bool) { b.SetOpen(v) }),
	))
	reg.MapModel(NewModel(b("basalt"), ids.BASALT).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("bed"), ids.BED).Properties(
		NewBoolProperty(ids.HEAD_PIECE_BIT, func(b *block.Bed) bool { return b.IsHeadPart() }, func(b *block.Bed, v bool) { b.SetHead(v) }),
		NewBoolProperty(ids.OCCUPIED_BIT, func(b *block.Bed) bool { return b.IsOccupied() }, func(b *block.Bed, v bool) { b.SetOccupied(v) }),
		commonProperties.HorizontalFacingSWNE,
	))
	reg.MapModel(NewModel(b("bedrock"), ids.BEDROCK).Properties(
		NewBoolProperty(ids.INFINIBURN_BIT, func(b *block.Bedrock) bool { return b.BurnsForever() }, func(b *block.Bedrock, v bool) { b.SetBurnsForever(v) }),
	))
	reg.MapModel(NewModel(b("bell"), ids.BELL).Properties(
		UnusedBoolProperty(ids.TOGGLE_BIT, false),
		NewValueFromStringProperty(ids.ATTACHMENT, vm.BellAttachmentType, func(b *block.Bell) blockutils.BellAttachmentType { return b.GetAttachmentType() }, func(b *block.Bell, v blockutils.BellAttachmentType) { b.SetAttachmentType(v) }),
		commonProperties.HorizontalFacingSWNE,
	))
	reg.MapModel(NewModel(b("bone_block"), ids.BONE_BLOCK).Properties(
		UnusedIntProperty(ids.DEPRECATED, 0),
		commonProperties.PillarAxis,
	))

	var brewingStandProperties []Property
	for _, slot := range []block.BrewingStandSlot{block.BrewingStandSlotEast, block.BrewingStandSlotSouthwest, block.BrewingStandSlotNorthwest} {
		name := map[block.BrewingStandSlot]string{
			block.BrewingStandSlotEast:      ids.BREWING_STAND_SLOT_A_BIT,
			block.BrewingStandSlotSouthwest: ids.BREWING_STAND_SLOT_B_BIT,
			block.BrewingStandSlotNorthwest: ids.BREWING_STAND_SLOT_C_BIT,
		}[slot]
		brewingStandProperties = append(brewingStandProperties, NewBoolProperty(name, func(b *block.BrewingStand) bool { return b.HasSlot(slot) }, func(b *block.BrewingStand, v bool) { b.SetSlot(slot, v) }))
	}
	reg.MapModel(NewModel(b("brewing_stand"), ids.BREWING_STAND).Properties(brewingStandProperties...))

	//C
	reg.MapModel(NewModel(b("cactus"), ids.CACTUS).Properties(ageProperty(ids.AGE, 0, 15)))
	reg.MapModel(NewModel(b("cake"), ids.CAKE).Properties(
		NewIntProperty(ids.BITE_COUNTER, 0, 6, func(b *block.Cake) int { return b.GetBites() }, func(b *block.Cake, v int) { b.SetBites(v) }),
	))
	reg.MapModel(NewModel(b("campfire"), ids.CAMPFIRE).Properties(commonProperties.CampfireProperties...))
	reg.MapModel(NewModel(b("carved_pumpkin"), ids.CARVED_PUMPKIN).Properties(
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("chain"), ids.IRON_CHAIN).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("chiseled_bookshelf"), ids.CHISELED_BOOKSHELF).Properties(
		commonProperties.HorizontalFacingSWNE,
		NewValueSetFromIntProperty(
			ids.BOOKS_STORED,
			newValueMap(
				//these are (currently) the same as the internal values, but it's best not to rely on those in case Mojang mess with the flags
				p(blockutils.ChiseledBookshelfSlotTopLeft, 1<<0),
				p(blockutils.ChiseledBookshelfSlotTopMiddle, 1<<1),
				p(blockutils.ChiseledBookshelfSlotTopRight, 1<<2),
				p(blockutils.ChiseledBookshelfSlotBottomLeft, 1<<3),
				p(blockutils.ChiseledBookshelfSlotBottomMiddle, 1<<4),
				p(blockutils.ChiseledBookshelfSlotBottomRight, 1<<5),
			),
			func(b *block.ChiseledBookshelf) []blockutils.ChiseledBookshelfSlot {
				var slots []blockutils.ChiseledBookshelfSlot
				for _, slot := range []blockutils.ChiseledBookshelfSlot{
					blockutils.ChiseledBookshelfSlotTopLeft, blockutils.ChiseledBookshelfSlotTopMiddle, blockutils.ChiseledBookshelfSlotTopRight,
					blockutils.ChiseledBookshelfSlotBottomLeft, blockutils.ChiseledBookshelfSlotBottomMiddle, blockutils.ChiseledBookshelfSlotBottomRight,
				} {
					if b.HasSlot(slot) {
						slots = append(slots, slot)
					}
				}
				return slots
			},
			func(b *block.ChiseledBookshelf, v []blockutils.ChiseledBookshelfSlot) {
				slots := map[blockutils.ChiseledBookshelfSlot]bool{}
				for _, slot := range v {
					slots[slot] = true
				}
				b.SetSlots(slots)
			},
		),
	))
	reg.MapModel(NewModel(b("chiseled_quartz"), ids.CHISELED_QUARTZ_BLOCK).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("chest"), ids.CHEST).Properties(commonProperties.HorizontalFacingCardinal))
	reg.MapModel(NewModel(b("chorus_flower"), ids.CHORUS_FLOWER).Properties(ageProperty(ids.AGE, block.ChorusFlowerMinAge, block.ChorusFlowerMaxAge)))
	reg.MapModel(NewModel(b("cocoa_pod"), ids.COCOA).Properties(
		ageProperty(ids.AGE, 0, 2),
		commonProperties.HorizontalFacingSWNEInverted,
	))

	//D
	reg.MapModel(NewModel(b("deepslate"), ids.DEEPSLATE).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("detector_rail"), ids.DETECTOR_RAIL).Properties(
		NewBoolProperty(ids.RAIL_DATA_BIT, func(b *block.DetectorRail) bool { return b.IsActivated() }, func(b *block.DetectorRail, v bool) { b.SetActivated(v) }),
		railShapeProperty(5), //TODO: shared with ActivatorRail
	))

	//E
	reg.MapModel(NewModel(b("ender_chest"), ids.ENDER_CHEST).Properties(commonProperties.HorizontalFacingCardinal))
	reg.MapModel(NewModel(b("end_portal_frame"), ids.END_PORTAL_FRAME).Properties(
		NewBoolProperty(ids.END_PORTAL_EYE_BIT, func(b *block.EndPortalFrame) bool { return b.HasEye() }, func(b *block.EndPortalFrame, v bool) { b.SetEye(v) }),
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("end_rod"), ids.END_ROD).Properties(
		NewValueFromIntProperty(ids.FACING_DIRECTION, vm.FacingEndRod, hfGet, hfSet),
	))

	//F
	reg.MapModel(NewModel(b("farmland"), ids.FARMLAND).Properties(
		NewIntProperty(ids.MOISTURIZED_AMOUNT, 0, 7, func(b *block.Farmland) int { return b.GetWetness() }, func(b *block.Farmland, v int) { b.SetWetness(v) }),
	))
	reg.MapModel(NewModel(b("fire"), ids.FIRE).Properties(ageProperty(ids.AGE, 0, 15)))
	reg.MapModel(NewModel(b("flower_pot"), ids.FLOWER_POT).Properties(
		UnusedBoolProperty(ids.UPDATE_BIT, false),
	))
	reg.MapModel(NewModel(b("frosted_ice"), ids.FROSTED_ICE).Properties(ageProperty(ids.AGE, 0, 3)))

	//G
	reg.MapModel(NewModel(b("glowing_item_frame"), ids.GLOW_FRAME).Properties(commonProperties.ItemFrameProperties...))

	//H
	reg.MapModel(NewModel(b("hay_bale"), ids.HAY_BLOCK).Properties(
		UnusedIntProperty(ids.DEPRECATED, 0),
		commonProperties.PillarAxis,
	))
	reg.MapModel(NewModel(b("hopper"), ids.HOPPER).Properties(
		//kinda weird this doesn't use powered_bit?
		NewBoolProperty(ids.TOGGLE_BIT, func(b poweredByRedstone) bool { return b.IsPowered() }, func(b poweredByRedstone, v bool) { b.SetPowered(v) }),
		NewValueFromIntProperty(ids.FACING_DIRECTION, vm.FacingExceptUp, hfGet, hfSet),
	))

	//I
	reg.MapModel(NewModel(b("infested_deepslate"), ids.INFESTED_DEEPSLATE).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("iron_door"), ids.IRON_DOOR).Properties(commonProperties.DoorProperties...))
	reg.MapModel(NewModel(b("iron_trapdoor"), ids.IRON_TRAPDOOR).Properties(commonProperties.TrapdoorProperties...))
	reg.MapModel(NewModel(b("item_frame"), ids.FRAME).Properties(commonProperties.ItemFrameProperties...))

	//L
	reg.MapModel(NewModel(b("ladder"), ids.LADDER).Properties(commonProperties.HorizontalFacingClassic))
	reg.MapModel(NewModel(b("lantern"), ids.LANTERN).Properties(
		NewBoolProperty(ids.HANGING, func(b hangingHolder) bool { return b.IsHanging() }, func(b hangingHolder, v bool) { b.SetHanging(v) }),
	))
	reg.MapModel(NewModel(b("lectern"), ids.LECTERN).Properties(
		NewBoolProperty(ids.POWERED_BIT, func(b *block.Lectern) bool { return b.IsProducingSignal() }, func(b *block.Lectern, v bool) { b.SetProducingSignal(v) }),
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("lever"), ids.LEVER).Properties(
		NewValueFromStringProperty(ids.LEVER_DIRECTION, vm.LeverFacing, func(b *block.Lever) blockutils.LeverFacing { return b.GetFacing() }, func(b *block.Lever, v blockutils.LeverFacing) { b.SetFacing(v) }),
		NewBoolProperty(ids.OPEN_BIT, func(b *block.Lever) bool { return b.IsActivated() }, func(b *block.Lever, v bool) { b.SetActivated(v) }),
	))
	reg.MapModel(NewModel(b("lit_pumpkin"), ids.LIT_PUMPKIN).Properties(commonProperties.HorizontalFacingCardinal))
	reg.MapModel(NewModel(b("loom"), ids.LOOM).Properties(commonProperties.HorizontalFacingSWNE))

	//M
	reg.MapModel(NewModel(b("muddy_mangrove_roots"), ids.MUDDY_MANGROVE_ROOTS).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("nether_wart"), ids.NETHER_WART).Properties(ageProperty(ids.AGE, 0, 3)))
	reg.MapModel(NewModel(b("nether_portal"), ids.PORTAL).Properties(
		NewValueFromStringProperty(ids.PORTAL_AXIS, vm.PortalAxis, func(b *block.NetherPortal) math.Axis { return b.GetAxis() }, func(b *block.NetherPortal, v math.Axis) { b.SetAxis(v) }),
	))

	//P
	reg.MapModel(NewModel(b("pink_petals"), ids.PINK_PETALS).Properties(
		//Pink petals only uses 0-3, but GROWTH state can go up to 7
		NewIntPropertyWithOffset(ids.GROWTH, 0, 7, func(b *block.PinkPetals) int { return b.GetCount() }, func(b *block.PinkPetals, v int) { b.SetCount(min(v, block.PinkPetalsMaxCount)) }, 1),
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("powered_rail"), ids.GOLDEN_RAIL).Properties(
		NewBoolProperty(ids.RAIL_DATA_BIT, func(b poweredByRedstone) bool { return b.IsPowered() }, func(b poweredByRedstone, v bool) { b.SetPowered(v) }), //TODO: shared with ActivatorRail
		railShapeProperty(5), //TODO: shared with ActivatorRail
	))
	reg.MapModel(NewModel(b("pitcher_plant"), ids.PITCHER_PLANT).Properties(
		NewBoolProperty(ids.UPPER_BLOCK_BIT, func(b topHalf) bool { return b.IsTop() }, func(b topHalf, v bool) { b.SetTop(v) }), //TODO: don't we have helpers for this?
	))
	reg.MapModel(NewModel(b("polished_basalt"), ids.POLISHED_BASALT).Properties(commonProperties.PillarAxis))
	reg.MapModel(NewModel(b("polished_blackstone_button"), ids.POLISHED_BLACKSTONE_BUTTON).Properties(commonProperties.ButtonProperties...))
	reg.MapModel(NewModel(b("polished_blackstone_pressure_plate"), ids.POLISHED_BLACKSTONE_PRESSURE_PLATE).Properties(commonProperties.SimplePressurePlateProperties...))
	reg.MapModel(NewModel(b("pumpkin"), ids.PUMPKIN).Properties(
		//not used, has no visible effect
		commonProperties.DummyCardinalDirection,
	))
	reg.MapModel(NewModel(b("purpur"), ids.PURPUR_BLOCK).Properties(
		commonProperties.DummyPillarAxis,
	))
	reg.MapModel(NewModel(b("purpur_pillar"), ids.PURPUR_PILLAR).Properties(commonProperties.PillarAxis))

	//Q
	reg.MapModel(NewModel(b("quartz"), ids.QUARTZ_BLOCK).Properties(
		commonProperties.DummyPillarAxis,
	))
	reg.MapModel(NewModel(b("quartz_pillar"), ids.QUARTZ_PILLAR).Properties(commonProperties.PillarAxis))

	//R
	reg.MapModel(NewModel(b("rail"), ids.RAIL).Properties(railShapeProperty(9)))
	reg.MapModel(NewModel(b("redstone_wire"), ids.REDSTONE_WIRE).Properties(commonProperties.AnalogRedstoneSignal))
	reg.MapModel(NewModel(b("respawn_anchor"), ids.RESPAWN_ANCHOR).Properties(
		NewIntProperty(ids.RESPAWN_ANCHOR_CHARGE, 0, 4, func(b *block.RespawnAnchor) int { return b.GetCharges() }, func(b *block.RespawnAnchor, v int) { b.SetCharges(v) }),
	))

	//S
	reg.MapModel(NewModel(b("sea_pickle"), ids.SEA_PICKLE).Properties(
		NewIntPropertyWithOffset(ids.CLUSTER_COUNT, 0, 3, func(b *block.SeaPickle) int { return b.GetCount() }, func(b *block.SeaPickle, v int) { b.SetCount(v) }, 1),
		NewInvertedBoolProperty(ids.DEAD_BIT, func(b *block.SeaPickle) bool { return b.IsUnderwater() }, func(b *block.SeaPickle, v bool) { b.SetUnderwater(v) }),
	))
	reg.MapModel(NewModel(b("small_dripleaf"), ids.SMALL_DRIPLEAF_BLOCK).Properties(
		NewBoolProperty(ids.UPPER_BLOCK_BIT, func(b *block.SmallDripleaf) bool { return b.IsTop() }, func(b *block.SmallDripleaf, v bool) { b.SetTop(v) }),
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("smooth_quartz"), ids.SMOOTH_QUARTZ).Properties(
		commonProperties.DummyPillarAxis,
	))
	reg.MapModel(NewModel(b("snow_layer"), ids.SNOW_LAYER).Properties(
		NewDummyProperty(ids.COVERED_BIT, false),
		NewIntPropertyWithOffset(ids.HEIGHT, 0, 7, func(b *block.SnowLayer) int { return b.GetLayers() }, func(b *block.SnowLayer, v int) { b.SetLayers(v) }, 1),
	))
	reg.MapModel(NewModel(b("soul_campfire"), ids.SOUL_CAMPFIRE).Properties(commonProperties.CampfireProperties...))
	reg.MapModel(NewModel(b("soul_fire"), ids.SOUL_FIRE).Properties(
		NewDummyProperty(ids.AGE, 0), //this is useless for soul fire, since it doesn't have the logic associated
	))
	reg.MapModel(NewModel(b("soul_lantern"), ids.SOUL_LANTERN).Properties(
		NewBoolProperty(ids.HANGING, func(b hangingHolder) bool { return b.IsHanging() }, func(b hangingHolder, v bool) { b.SetHanging(v) }), //TODO: repeated
	))
	reg.MapModel(NewModel(b("stone_button"), ids.STONE_BUTTON).Properties(commonProperties.ButtonProperties...))
	reg.MapModel(NewModel(b("stone_pressure_plate"), ids.STONE_PRESSURE_PLATE).Properties(commonProperties.SimplePressurePlateProperties...))
	reg.MapModel(NewModel(b("stonecutter"), ids.STONECUTTER_BLOCK).Properties(
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("sugarcane"), ids.REEDS).Properties(ageProperty(ids.AGE, 0, 15)))

	//T
	reg.MapModel(NewModel(b("trapped_chest"), ids.TRAPPED_CHEST).Properties(
		commonProperties.HorizontalFacingCardinal,
	))
	reg.MapModel(NewModel(b("tripwire"), ids.TRIP_WIRE).Properties(
		NewBoolProperty(ids.ATTACHED_BIT, func(b *block.Tripwire) bool { return b.IsConnected() }, func(b *block.Tripwire, v bool) { b.SetConnected(v) }),
		NewBoolProperty(ids.DISARMED_BIT, func(b *block.Tripwire) bool { return b.IsDisarmed() }, func(b *block.Tripwire, v bool) { b.SetDisarmed(v) }),
		NewBoolProperty(ids.SUSPENDED_BIT, func(b *block.Tripwire) bool { return b.IsSuspended() }, func(b *block.Tripwire, v bool) { b.SetSuspended(v) }),
		NewBoolProperty(ids.POWERED_BIT, func(b *block.Tripwire) bool { return b.IsTriggered() }, func(b *block.Tripwire, v bool) { b.SetTriggered(v) }),
		ConnectionProperties[0], ConnectionProperties[1], ConnectionProperties[2], ConnectionProperties[3], // 1.26.50: connection flags
	))
	reg.MapModel(NewModel(b("tripwire_hook"), ids.TRIPWIRE_HOOK).Properties(
		NewBoolProperty(ids.ATTACHED_BIT, func(b *block.TripwireHook) bool { return b.IsConnected() }, func(b *block.TripwireHook, v bool) { b.SetConnected(v) }),
		NewBoolProperty(ids.POWERED_BIT, func(b *block.TripwireHook) bool { return b.IsPowered() }, func(b *block.TripwireHook, v bool) { b.SetPowered(v) }),
		commonProperties.HorizontalFacingSWNE,
	))

	reg.MapModel(NewModel(b("twisting_vines"), ids.TWISTING_VINES).Properties(ageProperty(ids.TWISTING_VINES_AGE, 0, 25)))

	//W
	reg.MapModel(NewModel(b("wall_banner"), ids.WALL_BANNER).Properties(commonProperties.HorizontalFacingClassic))
	reg.MapModel(NewModel(b("weeping_vines"), ids.WEEPING_VINES).Properties(ageProperty(ids.WEEPING_VINES_AGE, 0, 25)))
	reg.MapModel(NewModel(b("weighted_pressure_plate_heavy"), ids.HEAVY_WEIGHTED_PRESSURE_PLATE).Properties(commonProperties.AnalogRedstoneSignal))
	reg.MapModel(NewModel(b("weighted_pressure_plate_light"), ids.LIGHT_WEIGHTED_PRESSURE_PLATE).Properties(commonProperties.AnalogRedstoneSignal))
}

// mapAsymmetricSerializer is a port of VanillaBlockMappings::mapAsymmetricSerializer.
func mapAsymmetricSerializer(reg *BlockSerializerDeserializerRegistrar, model *Model) {
	id := model.GetID()
	properties := model.GetProperties()
	reg.Serializer.Map(model.GetBlock(), func(blk block.Behavior) bedrock.BlockStateData {
		writer := NewBlockStateWriter(id)
		for _, property := range properties {
			property.Serialize(blk, writer)
		}
		return writer.GetBlockStateData()
	})
}

// deserializeAsymmetric is a port of VanillaBlockMappings::deserializeAsymmetric.
func deserializeAsymmetric(model *Model, in *BlockStateReader) block.Behavior {
	blk := model.GetBlock().Clone()
	for _, property := range model.GetProperties() {
		property.Deserialize(blk, in)
	}
	return blk
}

// registerSplitMappings is a port of VanillaBlockMappings::registerSplitMappings: all mappings
// that still use the split form of serializer/deserializer registration. This is typically only
// used by blocks with one ID but multiple PM types (split by property). These currently can't be
// registered in a unified way, and due to their small number it may not be worth the effort to
// implement a unified way to deal with them.
func registerSplitMappings(reg *BlockSerializerDeserializerRegistrar, commonProperties *CommonProperties) {
	//big dripleaf - split into head / stem variants, as stems don't have tilt or leaf state
	bigDripleafHeadModel := NewModel(b("big_dripleaf_head"), ids.BIG_DRIPLEAF).Properties(
		commonProperties.HorizontalFacingCardinal,
		NewValueFromStringProperty(ids.BIG_DRIPLEAF_TILT, GetValueMappings().DripleafState, func(b *block.BigDripleafHead) blockutils.DripleafState { return b.GetLeafState() }, func(b *block.BigDripleafHead, v blockutils.DripleafState) { b.SetLeafState(v) }),
		NewDummyProperty(ids.BIG_DRIPLEAF_HEAD, true),
	)
	bigDripleafStemModel := NewModel(b("big_dripleaf_stem"), ids.BIG_DRIPLEAF).Properties(
		commonProperties.HorizontalFacingCardinal,
		NewDummyProperty(ids.BIG_DRIPLEAF_TILT, ids.BIG_DRIPLEAF_TILT_NONE),
		NewDummyProperty(ids.BIG_DRIPLEAF_HEAD, false),
	)
	mapAsymmetricSerializer(reg, bigDripleafHeadModel)
	mapAsymmetricSerializer(reg, bigDripleafStemModel)
	reg.Deserializer.Map(ids.BIG_DRIPLEAF, func(in *BlockStateReader) block.Behavior {
		if in.ReadBool(ids.BIG_DRIPLEAF_HEAD) {
			return deserializeAsymmetric(bigDripleafHeadModel, in)
		}
		return deserializeAsymmetric(bigDripleafStemModel, in)
	})

	fillLevelProp := NewIntProperty(ids.FILL_LEVEL, 1, 6, func(b fillLevelHolder) int { return b.GetFillLevel() }, func(b fillLevelHolder, v int) { b.SetFillLevel(v) })

	//this pretends to be a water cauldron on disk and stores its real information in the block actor data, therefore only a serializer is needed
	mapAsymmetricSerializer(reg, NewModel(b("potion_cauldron"), ids.CAULDRON).Properties(fillLevelProp, NewDummyProperty(ids.CAULDRON_LIQUID, ids.CAULDRON_LIQUID_WATER)))

	lavaCauldronModel := NewModel(b("lava_cauldron"), ids.CAULDRON).Properties(
		fillLevelProp,
		NewDummyProperty(ids.CAULDRON_LIQUID, ids.CAULDRON_LIQUID_LAVA),
	)
	waterCauldronModel := NewModel(b("water_cauldron"), ids.CAULDRON).Properties(
		fillLevelProp,
		NewDummyProperty(ids.CAULDRON_LIQUID, ids.CAULDRON_LIQUID_WATER),
	)
	emptyCauldronModel := NewModel(b("cauldron"), ids.CAULDRON).Properties(
		NewDummyProperty(ids.FILL_LEVEL, 0),
		NewDummyProperty(ids.CAULDRON_LIQUID, ids.CAULDRON_LIQUID_WATER),
	)
	mapAsymmetricSerializer(reg, lavaCauldronModel)
	mapAsymmetricSerializer(reg, waterCauldronModel)
	mapAsymmetricSerializer(reg, emptyCauldronModel)
	reg.Deserializer.Map(ids.CAULDRON, func(in *BlockStateReader) block.Behavior {
		if in.ReadInt(ids.FILL_LEVEL) == 0 {
			return deserializeAsymmetric(emptyCauldronModel, in)
		}
		switch liquid := in.ReadString(ids.CAULDRON_LIQUID); liquid {
		case ids.CAULDRON_LIQUID_WATER:
			return deserializeAsymmetric(waterCauldronModel, in)
		case ids.CAULDRON_LIQUID_LAVA:
			return deserializeAsymmetric(lavaCauldronModel, in)
		case ids.CAULDRON_LIQUID_POWDER_SNOW:
			panic(&UnsupportedBlockStateError{BlockStateDeserializeError{Message: "Powder snow is not supported yet"}})
		default:
			panic(in.BadValueError(ids.CAULDRON_LIQUID, liquid, ""))
		}
	})

	//mushroom stems, split for consistency with all-sided logs vs normal logs
	allSidedMushroomStemModel := NewModel(b("all_sided_mushroom_stem"), ids.MUSHROOM_STEM).Properties(NewDummyProperty(ids.HUGE_MUSHROOM_BITS, ids.MUSHROOM_BLOCK_ALL_STEM))
	mushroomStemModel := NewModel(b("mushroom_stem"), ids.MUSHROOM_STEM).Properties(NewDummyProperty(ids.HUGE_MUSHROOM_BITS, ids.MUSHROOM_BLOCK_STEM))
	mapAsymmetricSerializer(reg, allSidedMushroomStemModel)
	mapAsymmetricSerializer(reg, mushroomStemModel)
	reg.Deserializer.Map(ids.MUSHROOM_STEM, func(in *BlockStateReader) block.Behavior {
		switch in.ReadInt(ids.HUGE_MUSHROOM_BITS) {
		case ids.MUSHROOM_BLOCK_ALL_STEM:
			return deserializeAsymmetric(allSidedMushroomStemModel, in)
		case ids.MUSHROOM_BLOCK_STEM:
			return deserializeAsymmetric(mushroomStemModel, in)
		default:
			panic(deserializeError("This state does not exist"))
		}
	})

	//pitcher crop, split into single and double variants as double has different properties and behaviour
	//this will probably be the most annoying to unify
	pitcherCropModel := NewModel(b("pitcher_crop"), ids.PITCHER_CROP).Properties(
		NewIntProperty(ids.GROWTH, 0, block.PitcherCropMaxAge, func(b *block.PitcherCrop) int { return b.GetAge() }, func(b *block.PitcherCrop, v int) { b.SetAge(v) }),
		NewDummyProperty(ids.UPPER_BLOCK_BIT, false),
	)
	doublePitcherCropAgeOffset := block.PitcherCropMaxAge + 1
	doublePitcherCropModel := NewModel(b("double_pitcher_crop"), ids.PITCHER_CROP).Properties(
		NewIntPropertyWithOffset(
			ids.GROWTH,
			doublePitcherCropAgeOffset, //TODO: it would be a bit less awkward if the bounds applied _after_ applying the offset, instead of before
			7,
			func(b *block.DoublePitcherCrop) int { return b.GetAge() },
			func(b *block.DoublePitcherCrop, v int) { b.SetAge(min(v, block.DoublePitcherCropMaxAge)) }, //state may give up to 7, but only up to 4 is valid
			-doublePitcherCropAgeOffset,
		),
		NewBoolProperty(ids.UPPER_BLOCK_BIT, func(b *block.DoublePitcherCrop) bool { return b.IsTop() }, func(b *block.DoublePitcherCrop, v bool) { b.SetTop(v) }),
	)
	mapAsymmetricSerializer(reg, pitcherCropModel)
	mapAsymmetricSerializer(reg, doublePitcherCropModel)
	reg.Deserializer.Map(ids.PITCHER_CROP, func(in *BlockStateReader) block.Behavior {
		if in.ReadInt(ids.GROWTH) <= block.PitcherCropMaxAge {
			if in.ReadBool(ids.UPPER_BLOCK_BIT) {
				//top pitcher crop with age 0-2 is an invalid state, only the bottom half should exist in this case
				return b("air")
			}
			return deserializeAsymmetric(pitcherCropModel, in)
		}
		return deserializeAsymmetric(doublePitcherCropModel, in)
	})

	//these only exist within PM (mapped from tile properties) as they don't support the same properties as a
	//normal banner, therefore no deserializer is needed
	mapAsymmetricSerializer(reg, NewModel(b("ominous_banner"), ids.STANDING_BANNER).Properties(commonProperties.FloorSignLikeRotation))
	mapAsymmetricSerializer(reg, NewModel(b("ominous_wall_banner"), ids.WALL_BANNER).Properties(commonProperties.HorizontalFacingClassic))

	for _, e := range []struct {
		id, wood string
	}{
		{ids.ACACIA_HANGING_SIGN, "acacia"},
		{ids.BAMBOO_HANGING_SIGN, "bamboo"},
		{ids.BIRCH_HANGING_SIGN, "birch"},
		{ids.CHERRY_HANGING_SIGN, "cherry"},
		{ids.CRIMSON_HANGING_SIGN, "crimson"},
		{ids.DARK_OAK_HANGING_SIGN, "dark_oak"},
		{ids.JUNGLE_HANGING_SIGN, "jungle"},
		{ids.MANGROVE_HANGING_SIGN, "mangrove"},
		{ids.OAK_HANGING_SIGN, "oak"},
		{ids.PALE_OAK_HANGING_SIGN, "pale_oak"},
		{ids.SPRUCE_HANGING_SIGN, "spruce"},
		{ids.WARPED_HANGING_SIGN, "warped"},
	} {
		//attached_bit          - true for ceiling center signs, false for ceiling edges signs and wall signs
		//hanging               - true for all ceiling signs, false for wall signs
		//facing_direction      - used for ceiling edges signs and wall signs
		//ground_sign_direction - used by ceiling center signs only
		centerModel := NewModel(b(e.wood+"_ceiling_center_hanging_sign"), e.id).Properties(
			commonProperties.FloorSignLikeRotation,
			NewDummyProperty(ids.ATTACHED_BIT, true),
			NewDummyProperty(ids.HANGING, true),
			NewDummyProperty(ids.FACING_DIRECTION, 2),
		)
		edgesModel := NewModel(b(e.wood+"_ceiling_edges_hanging_sign"), e.id).Properties(
			NewDummyProperty(ids.GROUND_SIGN_DIRECTION, 0),
			NewDummyProperty(ids.ATTACHED_BIT, false),
			NewDummyProperty(ids.HANGING, true),
			commonProperties.HorizontalFacingClassic,
		)
		wallModel := NewModel(b(e.wood+"_wall_hanging_sign"), e.id).Properties(
			NewDummyProperty(ids.GROUND_SIGN_DIRECTION, 0),
			NewDummyProperty(ids.ATTACHED_BIT, false),
			NewDummyProperty(ids.HANGING, false),
			commonProperties.HorizontalFacingClassic,
		)
		mapAsymmetricSerializer(reg, centerModel)
		mapAsymmetricSerializer(reg, edgesModel)
		mapAsymmetricSerializer(reg, wallModel)
		reg.Deserializer.Map(e.id, func(in *BlockStateReader) block.Behavior {
			if in.ReadBool(ids.HANGING) {
				if in.ReadBool(ids.ATTACHED_BIT) {
					return deserializeAsymmetric(centerModel, in)
				}
				return deserializeAsymmetric(edgesModel, in)
			}
			return deserializeAsymmetric(wallModel, in)
		})
	}
}

type fillLevelHolder interface {
	GetFillLevel() int
	SetFillLevel(fillLevel int)
}
