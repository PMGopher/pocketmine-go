package item

import (
	"fmt"
	"sort"
	"sync"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/utils"
)

// StringToItemParser is a port of pocketmine\item\StringToItemParser: parses item names from
// strings, for the /give command (and others).
type StringToItemParser struct {
	*utils.StringToTParser[Item]
	// reverseMap is state ID => aliases resolving to it.
	reverseMap map[int]map[string]bool
}

var (
	stringToItemParser     *StringToItemParser
	stringToItemParserOnce sync.Once
)

// GetStringToItemParser is StringToItemParser::getInstance.
func GetStringToItemParser() *StringToItemParser {
	stringToItemParserOnce.Do(func() {
		stringToItemParser = makeStringToItemParser()
	})
	return stringToItemParser
}

// makeStringToItemParser is a port of StringToItemParser::make.
func makeStringToItemParser() *StringToItemParser {
	result := &StringToItemParser{StringToTParser: utils.NewStringToTParser[Item](), reverseMap: map[int]map[string]bool{}}

	registerStringToItemParserDynamicBlocks(result)
	registerStringToItemParserBlocks(result)
	registerStringToItemParserDynamicItems(result)
	registerStringToItemParserItems(result)

	return result
}

type enumCase[T any] struct {
	name  string
	value T
}

var (
	dyeColorCases = []enumCase[blockutils.DyeColor]{
		{"white", blockutils.DyeColorWhite},
		{"orange", blockutils.DyeColorOrange},
		{"magenta", blockutils.DyeColorMagenta},
		{"light_blue", blockutils.DyeColorLightBlue},
		{"yellow", blockutils.DyeColorYellow},
		{"lime", blockutils.DyeColorLime},
		{"pink", blockutils.DyeColorPink},
		{"gray", blockutils.DyeColorGray},
		{"light_gray", blockutils.DyeColorLightGray},
		{"cyan", blockutils.DyeColorCyan},
		{"purple", blockutils.DyeColorPurple},
		{"blue", blockutils.DyeColorBlue},
		{"brown", blockutils.DyeColorBrown},
		{"green", blockutils.DyeColorGreen},
		{"red", blockutils.DyeColorRed},
		{"black", blockutils.DyeColorBlack},
	}
	coralTypeCases = []enumCase[blockutils.CoralType]{
		{"tube", blockutils.CoralTypeTube},
		{"brain", blockutils.CoralTypeBrain},
		{"bubble", blockutils.CoralTypeBubble},
		{"fire", blockutils.CoralTypeFire},
		{"horn", blockutils.CoralTypeHorn},
	}
	copperOxidationCases = []enumCase[blockutils.CopperOxidation]{
		{"none", blockutils.CopperOxidationNone},
		{"exposed", blockutils.CopperOxidationExposed},
		{"weathered", blockutils.CopperOxidationWeathered},
		{"oxidized", blockutils.CopperOxidationOxidized},
	}
	froglightTypeCases = []enumCase[blockutils.FroglightType]{
		{"ochre", blockutils.FroglightTypeOchre},
		{"pearlescent", blockutils.FroglightTypePearlescent},
		{"verdant", blockutils.FroglightTypeVerdant},
	}
	goatHornTypeCases = []enumCase[GoatHornType]{
		{"ponder", GoatHornTypePonder},
		{"sing", GoatHornTypeSing},
		{"seek", GoatHornTypeSeek},
		{"feel", GoatHornTypeFeel},
		{"admire", GoatHornTypeAdmire},
		{"call", GoatHornTypeCall},
		{"yearn", GoatHornTypeYearn},
		{"dream", GoatHornTypeDream},
	}
	suspiciousStewTypeCases = []enumCase[SuspiciousStewType]{
		{"poppy", SuspiciousStewTypePoppy},
		{"cornflower", SuspiciousStewTypeCornflower},
		{"tulip", SuspiciousStewTypeTulip},
		{"azure_bluet", SuspiciousStewTypeAzureBluet},
		{"lily_of_the_valley", SuspiciousStewTypeLilyOfTheValley},
		{"dandelion", SuspiciousStewTypeDandelion},
		{"blue_orchid", SuspiciousStewTypeBlueOrchid},
		{"allium", SuspiciousStewTypeAllium},
		{"oxeye_daisy", SuspiciousStewTypeOxeyeDaisy},
		{"wither_rose", SuspiciousStewTypeWitherRose},
	}
	potionTypeCases = []enumCase[PotionType]{
		{"water", PotionTypeWater},
		{"mundane", PotionTypeMundane},
		{"long_mundane", PotionTypeLongMundane},
		{"thick", PotionTypeThick},
		{"awkward", PotionTypeAwkward},
		{"night_vision", PotionTypeNightVision},
		{"long_night_vision", PotionTypeLongNightVision},
		{"invisibility", PotionTypeInvisibility},
		{"long_invisibility", PotionTypeLongInvisibility},
		{"leaping", PotionTypeLeaping},
		{"long_leaping", PotionTypeLongLeaping},
		{"strong_leaping", PotionTypeStrongLeaping},
		{"fire_resistance", PotionTypeFireResistance},
		{"long_fire_resistance", PotionTypeLongFireResistance},
		{"swiftness", PotionTypeSwiftness},
		{"long_swiftness", PotionTypeLongSwiftness},
		{"strong_swiftness", PotionTypeStrongSwiftness},
		{"slowness", PotionTypeSlowness},
		{"long_slowness", PotionTypeLongSlowness},
		{"water_breathing", PotionTypeWaterBreathing},
		{"long_water_breathing", PotionTypeLongWaterBreathing},
		{"healing", PotionTypeHealing},
		{"strong_healing", PotionTypeStrongHealing},
		{"harming", PotionTypeHarming},
		{"strong_harming", PotionTypeStrongHarming},
		{"poison", PotionTypePoison},
		{"long_poison", PotionTypeLongPoison},
		{"strong_poison", PotionTypeStrongPoison},
		{"regeneration", PotionTypeRegeneration},
		{"long_regeneration", PotionTypeLongRegeneration},
		{"strong_regeneration", PotionTypeStrongRegeneration},
		{"strength", PotionTypeStrength},
		{"long_strength", PotionTypeLongStrength},
		{"strong_strength", PotionTypeStrongStrength},
		{"weakness", PotionTypeWeakness},
		{"long_weakness", PotionTypeLongWeakness},
		{"wither", PotionTypeWither},
		{"turtle_master", PotionTypeTurtleMaster},
		{"long_turtle_master", PotionTypeLongTurtleMaster},
		{"strong_turtle_master", PotionTypeStrongTurtleMaster},
		{"slow_falling", PotionTypeSlowFalling},
		{"long_slow_falling", PotionTypeLongSlowFalling},
		{"strong_slowness", PotionTypeStrongSlowness},
	}
)

// registerStringToItemParserDynamicBlocks is a port of StringToItemParser::registerDynamicBlocks.
func registerStringToItemParserDynamicBlocks(p *StringToItemParser) {
	type colored interface{ SetColor(blockutils.DyeColor) }
	for _, c := range dyeColorCases {
		color := c.value
		register := func(name, blockName string) {
			p.mustRegisterBlock(c.name+"_"+name, func(string) block.Behavior {
				b := block.VanillaBlock(blockName)
				b.(colored).SetColor(color)
				return b
			})
		}
		//wall and floor banner are the same item
		register("banner", "banner")
		register("bed", "bed")
		register("candle", "dyed_candle")
		register("carpet", "carpet")
		register("concrete", "concrete")
		register("concrete_powder", "concrete_powder")
		register("glazed_terracotta", "glazed_terracotta")
		register("stained_clay", "stained_clay")
		register("stained_glass", "stained_glass")
		register("stained_glass_pane", "stained_glass_pane")
		register("stained_hardened_glass", "stained_hardened_glass")
		register("stained_hardened_glass_pane", "stained_hardened_glass_pane")
		register("wool", "wool")
		register("shulker_box", "dyed_shulker_box")
	}

	type coral interface{ SetCoralType(blockutils.CoralType) }
	for _, c := range coralTypeCases {
		coralType := c.value
		register := func(name, blockName string) {
			p.mustRegisterBlock(c.name+"_"+name, func(string) block.Behavior {
				b := block.VanillaBlock(blockName)
				b.(coral).SetCoralType(coralType)
				return b
			})
		}
		register("coral", "coral")
		register("coral_block", "coral_block")
		//wall and floor coral fans are the same item
		register("coral_fan", "coral_fan")
	}
	for i := block.LightMinLightLevel; i <= block.LightMaxLightLevel; i++ {
		level := i
		light := func(string) block.Behavior {
			b := block.VanillaBlock("light")
			b.(interface{ SetLightLevel(int) }).SetLightLevel(level)
			return b
		}
		//helper aliases, since we don't support passing data values in /give
		p.mustRegisterBlock(fmt.Sprintf("light_%d", i), light)
		p.mustRegisterBlock(fmt.Sprintf("light_block_%d", i), light)
	}

	type copper interface {
		SetOxidation(blockutils.CopperOxidation)
		SetWaxed(bool)
	}
	for _, c := range copperOxidationCases {
		oxidation := c.value
		oxPrefix := ""
		if oxidation != blockutils.CopperOxidationNone {
			oxPrefix = c.name + "_"
		}
		for _, w := range []struct {
			prefix string
			waxed  bool
		}{{"", false}, {"waxed_", true}} {
			waxed := w.waxed
			register := func(name, blockName string) {
				p.mustRegisterBlock(w.prefix+oxPrefix+name, func(string) block.Behavior {
					b := block.VanillaBlock(blockName)
					b.(copper).SetOxidation(oxidation)
					b.(copper).SetWaxed(waxed)
					return b
				})
			}
			register("copper_block", "copper")
			register("chiseled_copper", "chiseled_copper")
			register("copper_grate", "copper_grate")
			register("cut_copper_block", "cut_copper")
			register("cut_copper_stairs", "cut_copper_stairs")
			register("cut_copper_slab", "cut_copper_slab")
			register("copper_bulb", "copper_bulb")
			register("copper_door", "copper_door")
			register("copper_trapdoor", "copper_trapdoor")
			register("copper_bars", "copper_bars")
			register("copper_chain", "copper_chain")
			register("copper_lantern", "copper_lantern")
			register("lightning_rod", "lightning_rod")
		}
	}

	for _, c := range froglightTypeCases {
		froglightType := c.value
		p.mustRegisterBlock(c.name+"_froglight", func(string) block.Behavior {
			b := block.VanillaBlock("froglight")
			b.(interface {
				SetFroglightType(blockutils.FroglightType)
			}).SetFroglightType(froglightType)
			return b
		})
	}
}

// registerStringToItemParserDynamicItems is a port of StringToItemParser::registerDynamicItems.
func registerStringToItemParserDynamicItems(p *StringToItemParser) {
	for _, c := range dyeColorCases {
		color := c.value
		p.mustRegister(c.name+"_dye", func(string) Item {
			it := VanillaItem("dye")
			it.(*Dye).SetColor(color)
			return it
		})
	}

	for _, c := range goatHornTypeCases {
		hornType := c.value
		p.mustRegister(c.name+"_goat_horn", func(string) Item {
			it := VanillaItem("goat_horn")
			it.(*GoatHorn).SetHornType(hornType)
			return it
		})
	}

	for _, c := range suspiciousStewTypeCases {
		stewType := c.value
		p.mustRegister(c.name+"_suspicious_stew", func(string) Item {
			it := VanillaItem("suspicious_stew")
			it.(*SuspiciousStew).SetType(stewType)
			return it
		})
	}

	type potion interface{ SetType(PotionType) }
	for _, c := range potionTypeCases {
		potionType := c.value
		register := func(name, itemName string) {
			p.mustRegister(c.name+"_"+name, func(string) Item {
				it := VanillaItem(itemName)
				it.(potion).SetType(potionType)
				return it
			})
		}
		register("potion", "potion")
		register("splash_potion", "splash_potion")
		register("lingering_potion", "lingering_potion")
	}
}

// Register is a port of StringToItemParser::register: it also records the alias in the reverse map.
func (p *StringToItemParser) Register(alias string, callback func(input string) Item) error {
	if err := p.StringToTParser.Register(alias, callback); err != nil {
		return err
	}
	p.addReverse(callback(alias).GetStateId(), alias)
	return nil
}

// mustRegister is Register for the built-in aliases, where PHP's InvalidArgumentException on a
// duplicate would be a bug in the tables below.
func (p *StringToItemParser) mustRegister(alias string, callback func(input string) Item) {
	if err := p.Register(alias, callback); err != nil {
		panic(err)
	}
}

// Override is a port of StringToItemParser::override.
func (p *StringToItemParser) Override(alias string, callback func(input string) Item) {
	if oldItem, ok := p.Parse(alias); ok {
		oldStateID := oldItem.GetStateId()
		delete(p.reverseMap[oldStateID], alias)
		if len(p.reverseMap[oldStateID]) == 0 {
			delete(p.reverseMap, oldStateID)
		}
	}
	p.StringToTParser.Override(alias, callback)
	p.addReverse(callback(alias).GetStateId(), alias)
}

func (p *StringToItemParser) addReverse(stateID int, alias string) {
	if p.reverseMap[stateID] == nil {
		p.reverseMap[stateID] = map[string]bool{}
	}
	p.reverseMap[stateID][alias] = true
}

// RegisterBlock is a port of StringToItemParser::registerBlock.
func (p *StringToItemParser) RegisterBlock(alias string, callback func(input string) block.Behavior) error {
	return p.Register(alias, func(input string) Item { return blockAsItem(callback(input)) })
}

func (p *StringToItemParser) mustRegisterBlock(alias string, callback func(input string) block.Behavior) {
	if err := p.RegisterBlock(alias, callback); err != nil {
		panic(err)
	}
}

// blockAsItem is Block::asItem for a block from the registry.
func blockAsItem(b block.Behavior) Item {
	it, err := b.(interface{ AsItem() (block.Item, error) }).AsItem()
	if err != nil {
		panic(err)
	}
	return it.(Item)
}

// Parse is a port of StringToItemParser::parse: ok is false (PHP: null) if no alias matches.
func (p *StringToItemParser) Parse(input string) (Item, bool) {
	return p.StringToTParser.Parse(input)
}

// LookupAliases is a port of StringToItemParser::lookupAliases: the registered aliases that
// resolve to the given item. Unlike PHP's insertion-ordered array, the result is sorted.
func (p *StringToItemParser) LookupAliases(it Item) []string {
	aliases := make([]string, 0, len(p.reverseMap[it.GetStateId()]))
	for alias := range p.reverseMap[it.GetStateId()] {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases
}

// LookupBlockAliases is a port of StringToItemParser::lookupBlockAliases.
func (p *StringToItemParser) LookupBlockAliases(b block.Behavior) []string {
	return p.LookupAliases(blockAsItem(b))
}

// The tables below are StringToItemParser::registerBlocks and registerItems, converted line by line
// from the PHP source.

func registerStringToItemParserBlocks(p *StringToItemParser) {
	p.mustRegisterBlock("acacia_button", func(string) block.Behavior { return block.VanillaBlock("acacia_button") })
	p.mustRegisterBlock("acacia_door", func(string) block.Behavior { return block.VanillaBlock("acacia_door") })
	p.mustRegisterBlock("acacia_door_block", func(string) block.Behavior { return block.VanillaBlock("acacia_door") })
	p.mustRegisterBlock("acacia_fence", func(string) block.Behavior { return block.VanillaBlock("acacia_fence") })
	p.mustRegisterBlock("acacia_fence_gate", func(string) block.Behavior { return block.VanillaBlock("acacia_fence_gate") })
	p.mustRegisterBlock("acacia_leaves", func(string) block.Behavior { return block.VanillaBlock("acacia_leaves") })
	p.mustRegisterBlock("acacia_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("acacia_planks", func(string) block.Behavior { return block.VanillaBlock("acacia_planks") })
	p.mustRegisterBlock("acacia_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("acacia_pressure_plate") })
	p.mustRegisterBlock("acacia_sapling", func(string) block.Behavior { return block.VanillaBlock("acacia_sapling") })
	p.mustRegisterBlock("acacia_sign", func(string) block.Behavior { return block.VanillaBlock("acacia_sign") })
	p.mustRegisterBlock("acacia_slab", func(string) block.Behavior { return block.VanillaBlock("acacia_slab") })
	p.mustRegisterBlock("acacia_stairs", func(string) block.Behavior { return block.VanillaBlock("acacia_stairs") })
	p.mustRegisterBlock("acacia_standing_sign", func(string) block.Behavior { return block.VanillaBlock("acacia_sign") })
	p.mustRegisterBlock("acacia_trapdoor", func(string) block.Behavior { return block.VanillaBlock("acacia_trapdoor") })
	p.mustRegisterBlock("acacia_wall_sign", func(string) block.Behavior { return block.VanillaBlock("acacia_wall_sign") })
	p.mustRegisterBlock("acacia_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("acacia_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("acacia_stairs") })
	p.mustRegisterBlock("acacia_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("acacia_stairs") })
	p.mustRegisterBlock("activator_rail", func(string) block.Behavior { return block.VanillaBlock("activator_rail") })
	p.mustRegisterBlock("active_redstone_lamp", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_lamp")
			v.(interface{ SetPowered(bool) }).SetPowered(true)
			return v
		}()
	})
	p.mustRegisterBlock("air", func(string) block.Behavior { return block.VanillaBlock("air") })
	p.mustRegisterBlock("all_sided_mushroom_stem", func(string) block.Behavior { return block.VanillaBlock("all_sided_mushroom_stem") })
	p.mustRegisterBlock("allium", func(string) block.Behavior { return block.VanillaBlock("allium") })
	p.mustRegisterBlock("amethyst_block", func(string) block.Behavior { return block.VanillaBlock("amethyst") })
	p.mustRegisterBlock("amethyst_cluster", func(string) block.Behavior { return block.VanillaBlock("amethyst_cluster") })
	p.mustRegisterBlock("ancient_debris", func(string) block.Behavior { return block.VanillaBlock("ancient_debris") })
	p.mustRegisterBlock("andesite", func(string) block.Behavior { return block.VanillaBlock("andesite") })
	p.mustRegisterBlock("andesite_slab", func(string) block.Behavior { return block.VanillaBlock("andesite_slab") })
	p.mustRegisterBlock("andesite_stairs", func(string) block.Behavior { return block.VanillaBlock("andesite_stairs") })
	p.mustRegisterBlock("andesite_wall", func(string) block.Behavior { return block.VanillaBlock("andesite_wall") })
	p.mustRegisterBlock("anvil", func(string) block.Behavior { return block.VanillaBlock("anvil") })
	p.mustRegisterBlock("ateupd_block", func(string) block.Behavior { return block.VanillaBlock("info_update2") })
	p.mustRegisterBlock("azalea", func(string) block.Behavior { return block.VanillaBlock("azalea") })
	p.mustRegisterBlock("azalea_leaves", func(string) block.Behavior { return block.VanillaBlock("azalea_leaves") })
	p.mustRegisterBlock("azure_bluet", func(string) block.Behavior { return block.VanillaBlock("azure_bluet") })
	p.mustRegisterBlock("bamboo", func(string) block.Behavior { return block.VanillaBlock("bamboo") })
	p.mustRegisterBlock("bamboo_block", func(string) block.Behavior { return block.VanillaBlock("bamboo_block") })
	p.mustRegisterBlock("bamboo_button", func(string) block.Behavior { return block.VanillaBlock("bamboo_button") })
	p.mustRegisterBlock("bamboo_door", func(string) block.Behavior { return block.VanillaBlock("bamboo_door") })
	p.mustRegisterBlock("bamboo_fence", func(string) block.Behavior { return block.VanillaBlock("bamboo_fence") })
	p.mustRegisterBlock("bamboo_fence_gate", func(string) block.Behavior { return block.VanillaBlock("bamboo_fence_gate") })
	p.mustRegisterBlock("bamboo_mosaic", func(string) block.Behavior { return block.VanillaBlock("bamboo_mosaic") })
	p.mustRegisterBlock("bamboo_mosaic_slab", func(string) block.Behavior { return block.VanillaBlock("bamboo_mosaic_slab") })
	p.mustRegisterBlock("bamboo_mosaic_stairs", func(string) block.Behavior { return block.VanillaBlock("bamboo_mosaic_stairs") })
	p.mustRegisterBlock("bamboo_planks", func(string) block.Behavior { return block.VanillaBlock("bamboo_planks") })
	p.mustRegisterBlock("bamboo_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("bamboo_pressure_plate") })
	p.mustRegisterBlock("bamboo_sapling", func(string) block.Behavior { return block.VanillaBlock("bamboo_sapling") })
	p.mustRegisterBlock("bamboo_sign", func(string) block.Behavior { return block.VanillaBlock("bamboo_sign") })
	p.mustRegisterBlock("bamboo_slab", func(string) block.Behavior { return block.VanillaBlock("bamboo_slab") })
	p.mustRegisterBlock("bamboo_stairs", func(string) block.Behavior { return block.VanillaBlock("bamboo_stairs") })
	p.mustRegisterBlock("bamboo_trapdoor", func(string) block.Behavior { return block.VanillaBlock("bamboo_trapdoor") })
	p.mustRegisterBlock("banner", func(string) block.Behavior { return block.VanillaBlock("banner") })
	p.mustRegisterBlock("barrel", func(string) block.Behavior { return block.VanillaBlock("barrel") })
	p.mustRegisterBlock("barrier", func(string) block.Behavior { return block.VanillaBlock("barrier") })
	p.mustRegisterBlock("basalt", func(string) block.Behavior { return block.VanillaBlock("basalt") })
	p.mustRegisterBlock("beacon", func(string) block.Behavior { return block.VanillaBlock("beacon") })
	p.mustRegisterBlock("bed", func(string) block.Behavior { return block.VanillaBlock("bed") })
	p.mustRegisterBlock("bed_block", func(string) block.Behavior { return block.VanillaBlock("bed") })
	p.mustRegisterBlock("bedrock", func(string) block.Behavior { return block.VanillaBlock("bedrock") })
	p.mustRegisterBlock("beetroot_block", func(string) block.Behavior { return block.VanillaBlock("beetroots") })
	p.mustRegisterBlock("beetroots", func(string) block.Behavior { return block.VanillaBlock("beetroots") })
	p.mustRegisterBlock("bell", func(string) block.Behavior { return block.VanillaBlock("bell") })
	p.mustRegisterBlock("big_dripleaf", func(string) block.Behavior { return block.VanillaBlock("big_dripleaf_head") })
	p.mustRegisterBlock("birch_button", func(string) block.Behavior { return block.VanillaBlock("birch_button") })
	p.mustRegisterBlock("birch_door", func(string) block.Behavior { return block.VanillaBlock("birch_door") })
	p.mustRegisterBlock("birch_door_block", func(string) block.Behavior { return block.VanillaBlock("birch_door") })
	p.mustRegisterBlock("birch_fence", func(string) block.Behavior { return block.VanillaBlock("birch_fence") })
	p.mustRegisterBlock("birch_fence_gate", func(string) block.Behavior { return block.VanillaBlock("birch_fence_gate") })
	p.mustRegisterBlock("birch_leaves", func(string) block.Behavior { return block.VanillaBlock("birch_leaves") })
	p.mustRegisterBlock("birch_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("birch_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("birch_planks", func(string) block.Behavior { return block.VanillaBlock("birch_planks") })
	p.mustRegisterBlock("birch_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("birch_pressure_plate") })
	p.mustRegisterBlock("birch_sapling", func(string) block.Behavior { return block.VanillaBlock("birch_sapling") })
	p.mustRegisterBlock("birch_sign", func(string) block.Behavior { return block.VanillaBlock("birch_sign") })
	p.mustRegisterBlock("birch_slab", func(string) block.Behavior { return block.VanillaBlock("birch_slab") })
	p.mustRegisterBlock("birch_stairs", func(string) block.Behavior { return block.VanillaBlock("birch_stairs") })
	p.mustRegisterBlock("birch_standing_sign", func(string) block.Behavior { return block.VanillaBlock("birch_sign") })
	p.mustRegisterBlock("birch_trapdoor", func(string) block.Behavior { return block.VanillaBlock("birch_trapdoor") })
	p.mustRegisterBlock("birch_wall_sign", func(string) block.Behavior { return block.VanillaBlock("birch_wall_sign") })
	p.mustRegisterBlock("birch_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("birch_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("birch_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("birch_stairs") })
	p.mustRegisterBlock("birch_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("birch_stairs") })
	p.mustRegisterBlock("blackstone", func(string) block.Behavior { return block.VanillaBlock("blackstone") })
	p.mustRegisterBlock("blackstone_slab", func(string) block.Behavior { return block.VanillaBlock("blackstone_slab") })
	p.mustRegisterBlock("blackstone_stairs", func(string) block.Behavior { return block.VanillaBlock("blackstone_stairs") })
	p.mustRegisterBlock("blackstone_wall", func(string) block.Behavior { return block.VanillaBlock("blackstone_wall") })
	p.mustRegisterBlock("blast_furnace", func(string) block.Behavior { return block.VanillaBlock("blast_furnace") })
	p.mustRegisterBlock("blue_ice", func(string) block.Behavior { return block.VanillaBlock("blue_ice") })
	p.mustRegisterBlock("blue_orchid", func(string) block.Behavior { return block.VanillaBlock("blue_orchid") })
	p.mustRegisterBlock("blue_torch", func(string) block.Behavior { return block.VanillaBlock("blue_torch") })
	p.mustRegisterBlock("bone_block", func(string) block.Behavior { return block.VanillaBlock("bone_block") })
	p.mustRegisterBlock("bookshelf", func(string) block.Behavior { return block.VanillaBlock("bookshelf") })
	p.mustRegisterBlock("brewing_stand", func(string) block.Behavior { return block.VanillaBlock("brewing_stand") })
	p.mustRegisterBlock("brewing_stand_block", func(string) block.Behavior { return block.VanillaBlock("brewing_stand") })
	p.mustRegisterBlock("brick_block", func(string) block.Behavior { return block.VanillaBlock("bricks") })
	p.mustRegisterBlock("brick_slab", func(string) block.Behavior { return block.VanillaBlock("brick_slab") })
	p.mustRegisterBlock("brick_stairs", func(string) block.Behavior { return block.VanillaBlock("brick_stairs") })
	p.mustRegisterBlock("brick_wall", func(string) block.Behavior { return block.VanillaBlock("brick_wall") })
	p.mustRegisterBlock("bricks", func(string) block.Behavior { return block.VanillaBlock("bricks") })
	p.mustRegisterBlock("bricks_block", func(string) block.Behavior { return block.VanillaBlock("bricks") })
	p.mustRegisterBlock("brown_mushroom", func(string) block.Behavior { return block.VanillaBlock("brown_mushroom") })
	p.mustRegisterBlock("brown_mushroom_block", func(string) block.Behavior { return block.VanillaBlock("brown_mushroom_block") })
	p.mustRegisterBlock("budding_amethyst", func(string) block.Behavior { return block.VanillaBlock("budding_amethyst") })
	p.mustRegisterBlock("burning_furnace", func(string) block.Behavior { return block.VanillaBlock("furnace") })
	p.mustRegisterBlock("bush", func(string) block.Behavior { return block.VanillaBlock("dead_bush") })
	p.mustRegisterBlock("cactus", func(string) block.Behavior { return block.VanillaBlock("cactus") })
	p.mustRegisterBlock("cactus_flower", func(string) block.Behavior { return block.VanillaBlock("cactus_flower") })
	p.mustRegisterBlock("cake", func(string) block.Behavior { return block.VanillaBlock("cake") })
	p.mustRegisterBlock("cake_block", func(string) block.Behavior { return block.VanillaBlock("cake") })
	p.mustRegisterBlock("calcite", func(string) block.Behavior { return block.VanillaBlock("calcite") })
	p.mustRegisterBlock("campfire", func(string) block.Behavior { return block.VanillaBlock("campfire") })
	p.mustRegisterBlock("candle", func(string) block.Behavior { return block.VanillaBlock("candle") })
	p.mustRegisterBlock("carpet", func(string) block.Behavior { return block.VanillaBlock("carpet") })
	p.mustRegisterBlock("carrot_block", func(string) block.Behavior { return block.VanillaBlock("carrots") })
	p.mustRegisterBlock("carrots", func(string) block.Behavior { return block.VanillaBlock("carrots") })
	p.mustRegisterBlock("cartography_table", func(string) block.Behavior { return block.VanillaBlock("cartography_table") })
	p.mustRegisterBlock("carved_pumpkin", func(string) block.Behavior { return block.VanillaBlock("carved_pumpkin") })
	p.mustRegisterBlock("cauldron", func(string) block.Behavior { return block.VanillaBlock("cauldron") })
	p.mustRegisterBlock("cave_vines", func(string) block.Behavior { return block.VanillaBlock("cave_vines") })
	p.mustRegisterBlock("chain", func(string) block.Behavior { return block.VanillaBlock("chain") })
	p.mustRegisterBlock("cherry_button", func(string) block.Behavior { return block.VanillaBlock("cherry_button") })
	p.mustRegisterBlock("cherry_door", func(string) block.Behavior { return block.VanillaBlock("cherry_door") })
	p.mustRegisterBlock("cherry_fence", func(string) block.Behavior { return block.VanillaBlock("cherry_fence") })
	p.mustRegisterBlock("cherry_fence_gate", func(string) block.Behavior { return block.VanillaBlock("cherry_fence_gate") })
	p.mustRegisterBlock("cherry_leaves", func(string) block.Behavior { return block.VanillaBlock("cherry_leaves") })
	p.mustRegisterBlock("cherry_log", func(string) block.Behavior { return block.VanillaBlock("cherry_log") })
	p.mustRegisterBlock("cherry_planks", func(string) block.Behavior { return block.VanillaBlock("cherry_planks") })
	p.mustRegisterBlock("cherry_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("cherry_pressure_plate") })
	p.mustRegisterBlock("cherry_sign", func(string) block.Behavior { return block.VanillaBlock("cherry_sign") })
	p.mustRegisterBlock("cherry_slab", func(string) block.Behavior { return block.VanillaBlock("cherry_slab") })
	p.mustRegisterBlock("cherry_stairs", func(string) block.Behavior { return block.VanillaBlock("cherry_stairs") })
	p.mustRegisterBlock("cherry_trapdoor", func(string) block.Behavior { return block.VanillaBlock("cherry_trapdoor") })
	p.mustRegisterBlock("cherry_wood", func(string) block.Behavior { return block.VanillaBlock("cherry_wood") })
	p.mustRegisterBlock("chemical_heat", func(string) block.Behavior { return block.VanillaBlock("chemical_heat") })
	p.mustRegisterBlock("chemistry_table", func(string) block.Behavior { return block.VanillaBlock("compound_creator") })
	p.mustRegisterBlock("chest", func(string) block.Behavior { return block.VanillaBlock("chest") })
	p.mustRegisterBlock("chipped_anvil", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("anvil")
			v.(interface{ SetDamage(int) }).SetDamage(1)
			return v
		}()
	})
	p.mustRegisterBlock("chiseled_bookshelf", func(string) block.Behavior { return block.VanillaBlock("chiseled_bookshelf") })
	p.mustRegisterBlock("chiseled_deepslate", func(string) block.Behavior { return block.VanillaBlock("chiseled_deepslate") })
	p.mustRegisterBlock("chiseled_nether_bricks", func(string) block.Behavior { return block.VanillaBlock("chiseled_nether_bricks") })
	p.mustRegisterBlock("chiseled_polished_blackstone", func(string) block.Behavior { return block.VanillaBlock("chiseled_polished_blackstone") })
	p.mustRegisterBlock("chiseled_quartz", func(string) block.Behavior { return block.VanillaBlock("chiseled_quartz") })
	p.mustRegisterBlock("chiseled_red_sandstone", func(string) block.Behavior { return block.VanillaBlock("chiseled_red_sandstone") })
	p.mustRegisterBlock("chiseled_resin_bricks", func(string) block.Behavior { return block.VanillaBlock("chiseled_resin_bricks") })
	p.mustRegisterBlock("chiseled_sandstone", func(string) block.Behavior { return block.VanillaBlock("chiseled_sandstone") })
	p.mustRegisterBlock("chiseled_stone_bricks", func(string) block.Behavior { return block.VanillaBlock("chiseled_stone_bricks") })
	p.mustRegisterBlock("chiseled_tuff", func(string) block.Behavior { return block.VanillaBlock("chiseled_tuff") })
	p.mustRegisterBlock("chiseled_tuff_bricks", func(string) block.Behavior { return block.VanillaBlock("chiseled_tuff_bricks") })
	p.mustRegisterBlock("chorus_flower", func(string) block.Behavior { return block.VanillaBlock("chorus_flower") })
	p.mustRegisterBlock("chorus_plant", func(string) block.Behavior { return block.VanillaBlock("chorus_plant") })
	p.mustRegisterBlock("clay_block", func(string) block.Behavior { return block.VanillaBlock("clay") })
	p.mustRegisterBlock("coal_block", func(string) block.Behavior { return block.VanillaBlock("coal") })
	p.mustRegisterBlock("coal_ore", func(string) block.Behavior { return block.VanillaBlock("coal_ore") })
	p.mustRegisterBlock("coarse_dirt", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dirt")
			v.(interface{ SetDirtType(blockutils.DirtType) }).SetDirtType(blockutils.DirtTypeCoarse)
			return v
		}()
	})
	p.mustRegisterBlock("cobble", func(string) block.Behavior { return block.VanillaBlock("cobblestone") })
	p.mustRegisterBlock("cobble_stairs", func(string) block.Behavior { return block.VanillaBlock("cobblestone_stairs") })
	p.mustRegisterBlock("cobble_wall", func(string) block.Behavior { return block.VanillaBlock("cobblestone_wall") })
	p.mustRegisterBlock("cobbled_deepslate", func(string) block.Behavior { return block.VanillaBlock("cobbled_deepslate") })
	p.mustRegisterBlock("cobbled_deepslate_slab", func(string) block.Behavior { return block.VanillaBlock("cobbled_deepslate_slab") })
	p.mustRegisterBlock("cobbled_deepslate_stairs", func(string) block.Behavior { return block.VanillaBlock("cobbled_deepslate_stairs") })
	p.mustRegisterBlock("cobbled_deepslate_wall", func(string) block.Behavior { return block.VanillaBlock("cobbled_deepslate_wall") })
	p.mustRegisterBlock("cobblestone", func(string) block.Behavior { return block.VanillaBlock("cobblestone") })
	p.mustRegisterBlock("cobblestone_slab", func(string) block.Behavior { return block.VanillaBlock("cobblestone_slab") })
	p.mustRegisterBlock("cobblestone_stairs", func(string) block.Behavior { return block.VanillaBlock("cobblestone_stairs") })
	p.mustRegisterBlock("cobblestone_wall", func(string) block.Behavior { return block.VanillaBlock("cobblestone_wall") })
	p.mustRegisterBlock("cobweb", func(string) block.Behavior { return block.VanillaBlock("cobweb") })
	p.mustRegisterBlock("cocoa", func(string) block.Behavior { return block.VanillaBlock("cocoa_pod") })
	p.mustRegisterBlock("cocoa_block", func(string) block.Behavior { return block.VanillaBlock("cocoa_pod") })
	p.mustRegisterBlock("cocoa_pod", func(string) block.Behavior { return block.VanillaBlock("cocoa_pod") })
	p.mustRegisterBlock("cocoa_pods", func(string) block.Behavior { return block.VanillaBlock("cocoa_pod") })
	p.mustRegisterBlock("colored_torch_bp", func(string) block.Behavior { return block.VanillaBlock("blue_torch") })
	p.mustRegisterBlock("colored_torch_rg", func(string) block.Behavior { return block.VanillaBlock("red_torch") })
	p.mustRegisterBlock("comparator", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("comparator_block", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("compound_creator", func(string) block.Behavior { return block.VanillaBlock("compound_creator") })
	p.mustRegisterBlock("concrete", func(string) block.Behavior { return block.VanillaBlock("concrete") })
	p.mustRegisterBlock("concrete_powder", func(string) block.Behavior { return block.VanillaBlock("concrete_powder") })
	p.mustRegisterBlock("concretepowder", func(string) block.Behavior { return block.VanillaBlock("concrete_powder") })
	p.mustRegisterBlock("copper_ore", func(string) block.Behavior { return block.VanillaBlock("copper_ore") })
	p.mustRegisterBlock("copper_torch", func(string) block.Behavior { return block.VanillaBlock("copper_torch") })
	p.mustRegisterBlock("coral", func(string) block.Behavior { return block.VanillaBlock("coral") })
	p.mustRegisterBlock("coral_block", func(string) block.Behavior { return block.VanillaBlock("coral_block") })
	p.mustRegisterBlock("coral_fan", func(string) block.Behavior { return block.VanillaBlock("coral_fan") })
	p.mustRegisterBlock("coral_fan_dead", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("coral_fan")
			v.(interface{ SetCoralType(blockutils.CoralType) }).SetCoralType(blockutils.CoralTypeTube)
			v.(interface{ SetDead(bool) }).SetDead(true)
			return v
		}()
	})
	p.mustRegisterBlock("coral_fan_hang", func(string) block.Behavior { return block.VanillaBlock("wall_coral_fan") })
	p.mustRegisterBlock("coral_fan_hang2", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("wall_coral_fan")
			v.(interface{ SetCoralType(blockutils.CoralType) }).SetCoralType(blockutils.CoralTypeBubble)
			return v
		}()
	})
	p.mustRegisterBlock("coral_fan_hang3", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("wall_coral_fan")
			v.(interface{ SetCoralType(blockutils.CoralType) }).SetCoralType(blockutils.CoralTypeHorn)
			return v
		}()
	})
	p.mustRegisterBlock("cornflower", func(string) block.Behavior { return block.VanillaBlock("cornflower") })
	p.mustRegisterBlock("cracked_deepslate_bricks", func(string) block.Behavior { return block.VanillaBlock("cracked_deepslate_bricks") })
	p.mustRegisterBlock("cracked_deepslate_tiles", func(string) block.Behavior { return block.VanillaBlock("cracked_deepslate_tiles") })
	p.mustRegisterBlock("cracked_nether_bricks", func(string) block.Behavior { return block.VanillaBlock("cracked_nether_bricks") })
	p.mustRegisterBlock("cracked_polished_blackstone_bricks", func(string) block.Behavior { return block.VanillaBlock("cracked_polished_blackstone_bricks") })
	p.mustRegisterBlock("cracked_stone_bricks", func(string) block.Behavior { return block.VanillaBlock("cracked_stone_bricks") })
	p.mustRegisterBlock("crafting_table", func(string) block.Behavior { return block.VanillaBlock("crafting_table") })
	p.mustRegisterBlock("creeper_head", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeCreeper)
			return v
		}()
	})
	p.mustRegisterBlock("crimson_button", func(string) block.Behavior { return block.VanillaBlock("crimson_button") })
	p.mustRegisterBlock("crimson_door", func(string) block.Behavior { return block.VanillaBlock("crimson_door") })
	p.mustRegisterBlock("crimson_fence", func(string) block.Behavior { return block.VanillaBlock("crimson_fence") })
	p.mustRegisterBlock("crimson_fence_gate", func(string) block.Behavior { return block.VanillaBlock("crimson_fence_gate") })
	p.mustRegisterBlock("crimson_fungus", func(string) block.Behavior { return block.VanillaBlock("crimson_fungus") })
	p.mustRegisterBlock("crimson_hyphae", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("crimson_hyphae")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("crimson_nylium", func(string) block.Behavior { return block.VanillaBlock("crimson_nylium") })
	p.mustRegisterBlock("crimson_planks", func(string) block.Behavior { return block.VanillaBlock("crimson_planks") })
	p.mustRegisterBlock("crimson_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("crimson_pressure_plate") })
	p.mustRegisterBlock("crimson_roots", func(string) block.Behavior { return block.VanillaBlock("crimson_roots") })
	p.mustRegisterBlock("crimson_sign", func(string) block.Behavior { return block.VanillaBlock("crimson_sign") })
	p.mustRegisterBlock("crimson_slab", func(string) block.Behavior { return block.VanillaBlock("crimson_slab") })
	p.mustRegisterBlock("crimson_stairs", func(string) block.Behavior { return block.VanillaBlock("crimson_stairs") })
	p.mustRegisterBlock("crimson_stem", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("crimson_stem")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("crimson_trapdoor", func(string) block.Behavior { return block.VanillaBlock("crimson_trapdoor") })
	p.mustRegisterBlock("crying_obsidian", func(string) block.Behavior { return block.VanillaBlock("crying_obsidian") })
	p.mustRegisterBlock("cut_red_sandstone", func(string) block.Behavior { return block.VanillaBlock("cut_red_sandstone") })
	p.mustRegisterBlock("cut_red_sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("cut_red_sandstone_slab") })
	p.mustRegisterBlock("cut_sandstone", func(string) block.Behavior { return block.VanillaBlock("cut_sandstone") })
	p.mustRegisterBlock("cut_sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("cut_sandstone_slab") })
	p.mustRegisterBlock("damaged_anvil", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("anvil")
			v.(interface{ SetDamage(int) }).SetDamage(2)
			return v
		}()
	})
	p.mustRegisterBlock("dandelion", func(string) block.Behavior { return block.VanillaBlock("dandelion") })
	p.mustRegisterBlock("dark_oak_button", func(string) block.Behavior { return block.VanillaBlock("dark_oak_button") })
	p.mustRegisterBlock("dark_oak_door", func(string) block.Behavior { return block.VanillaBlock("dark_oak_door") })
	p.mustRegisterBlock("dark_oak_door_block", func(string) block.Behavior { return block.VanillaBlock("dark_oak_door") })
	p.mustRegisterBlock("dark_oak_fence", func(string) block.Behavior { return block.VanillaBlock("dark_oak_fence") })
	p.mustRegisterBlock("dark_oak_fence_gate", func(string) block.Behavior { return block.VanillaBlock("dark_oak_fence_gate") })
	p.mustRegisterBlock("dark_oak_leaves", func(string) block.Behavior { return block.VanillaBlock("dark_oak_leaves") })
	p.mustRegisterBlock("dark_oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dark_oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("dark_oak_planks", func(string) block.Behavior { return block.VanillaBlock("dark_oak_planks") })
	p.mustRegisterBlock("dark_oak_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("dark_oak_pressure_plate") })
	p.mustRegisterBlock("dark_oak_sapling", func(string) block.Behavior { return block.VanillaBlock("dark_oak_sapling") })
	p.mustRegisterBlock("dark_oak_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_sign") })
	p.mustRegisterBlock("dark_oak_slab", func(string) block.Behavior { return block.VanillaBlock("dark_oak_slab") })
	p.mustRegisterBlock("dark_oak_stairs", func(string) block.Behavior { return block.VanillaBlock("dark_oak_stairs") })
	p.mustRegisterBlock("dark_oak_standing_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_sign") })
	p.mustRegisterBlock("dark_oak_trapdoor", func(string) block.Behavior { return block.VanillaBlock("dark_oak_trapdoor") })
	p.mustRegisterBlock("dark_oak_wall_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_wall_sign") })
	p.mustRegisterBlock("dark_oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dark_oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("dark_oak_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("dark_oak_stairs") })
	p.mustRegisterBlock("dark_oak_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("dark_oak_stairs") })
	p.mustRegisterBlock("dark_prismarine", func(string) block.Behavior { return block.VanillaBlock("dark_prismarine") })
	p.mustRegisterBlock("dark_prismarine_slab", func(string) block.Behavior { return block.VanillaBlock("dark_prismarine_slab") })
	p.mustRegisterBlock("dark_prismarine_stairs", func(string) block.Behavior { return block.VanillaBlock("dark_prismarine_stairs") })
	p.mustRegisterBlock("darkoak_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_sign") })
	p.mustRegisterBlock("darkoak_standing_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_sign") })
	p.mustRegisterBlock("darkoak_wall_sign", func(string) block.Behavior { return block.VanillaBlock("dark_oak_wall_sign") })
	p.mustRegisterBlock("daylight_detector", func(string) block.Behavior { return block.VanillaBlock("daylight_sensor") })
	p.mustRegisterBlock("daylight_detector_inverted", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("daylight_sensor")
			v.(interface{ SetInverted(bool) }).SetInverted(true)
			return v
		}()
	})
	p.mustRegisterBlock("daylight_sensor", func(string) block.Behavior { return block.VanillaBlock("daylight_sensor") })
	p.mustRegisterBlock("daylight_sensor_inverted", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("daylight_sensor")
			v.(interface{ SetInverted(bool) }).SetInverted(true)
			return v
		}()
	})
	p.mustRegisterBlock("dead_bush", func(string) block.Behavior { return block.VanillaBlock("dead_bush") })
	p.mustRegisterBlock("deadbush", func(string) block.Behavior { return block.VanillaBlock("dead_bush") })
	p.mustRegisterBlock("deepslate", func(string) block.Behavior { return block.VanillaBlock("deepslate") })
	p.mustRegisterBlock("deepslate_bricks", func(string) block.Behavior { return block.VanillaBlock("deepslate_bricks") })
	p.mustRegisterBlock("deepslate_brick_slab", func(string) block.Behavior { return block.VanillaBlock("deepslate_brick_slab") })
	p.mustRegisterBlock("deepslate_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("deepslate_brick_stairs") })
	p.mustRegisterBlock("deepslate_brick_wall", func(string) block.Behavior { return block.VanillaBlock("deepslate_brick_wall") })
	p.mustRegisterBlock("deepslate_tiles", func(string) block.Behavior { return block.VanillaBlock("deepslate_tiles") })
	p.mustRegisterBlock("deepslate_tile_slab", func(string) block.Behavior { return block.VanillaBlock("deepslate_tile_slab") })
	p.mustRegisterBlock("deepslate_tile_stairs", func(string) block.Behavior { return block.VanillaBlock("deepslate_tile_stairs") })
	p.mustRegisterBlock("deepslate_tile_wall", func(string) block.Behavior { return block.VanillaBlock("deepslate_tile_wall") })
	p.mustRegisterBlock("deepslate_coal_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_coal_ore") })
	p.mustRegisterBlock("deepslate_copper_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_copper_ore") })
	p.mustRegisterBlock("deepslate_diamond_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_diamond_ore") })
	p.mustRegisterBlock("deepslate_emerald_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_emerald_ore") })
	p.mustRegisterBlock("deepslate_gold_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_gold_ore") })
	p.mustRegisterBlock("deepslate_iron_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_iron_ore") })
	p.mustRegisterBlock("deepslate_lapis_lazuli_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_lapis_lazuli_ore") })
	p.mustRegisterBlock("deepslate_redstone_ore", func(string) block.Behavior { return block.VanillaBlock("deepslate_redstone_ore") })
	p.mustRegisterBlock("detector_rail", func(string) block.Behavior { return block.VanillaBlock("detector_rail") })
	p.mustRegisterBlock("diamond_block", func(string) block.Behavior { return block.VanillaBlock("diamond") })
	p.mustRegisterBlock("diamond_ore", func(string) block.Behavior { return block.VanillaBlock("diamond_ore") })
	p.mustRegisterBlock("diorite", func(string) block.Behavior { return block.VanillaBlock("diorite") })
	p.mustRegisterBlock("diorite_slab", func(string) block.Behavior { return block.VanillaBlock("diorite_slab") })
	p.mustRegisterBlock("diorite_stairs", func(string) block.Behavior { return block.VanillaBlock("diorite_stairs") })
	p.mustRegisterBlock("diorite_wall", func(string) block.Behavior { return block.VanillaBlock("diorite_wall") })
	p.mustRegisterBlock("dirt", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dirt")
			v.(interface{ SetDirtType(blockutils.DirtType) }).SetDirtType(blockutils.DirtTypeNormal)
			return v
		}()
	})
	p.mustRegisterBlock("dirt_with_roots", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dirt")
			v.(interface{ SetDirtType(blockutils.DirtType) }).SetDirtType(blockutils.DirtTypeRooted)
			return v
		}()
	})
	p.mustRegisterBlock("door_block", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("double_plant", func(string) block.Behavior { return block.VanillaBlock("sunflower") })
	p.mustRegisterBlock("double_red_sandstone_slab", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("red_sandstone_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_slab", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("stone_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_slabs", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("stone_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_stone_slab", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("stone_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_stone_slab2", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("red_sandstone_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_stone_slab3", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("end_stone_brick_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_stone_slab4", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mossy_stone_brick_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_tallgrass", func(string) block.Behavior { return block.VanillaBlock("double_tallgrass") })
	p.mustRegisterBlock("double_wood_slab", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_wood_slabs", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_wooden_slab", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("double_wooden_slabs", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_slab")
			v.(interface{ SetSlabType(blockutils.SlabType) }).SetSlabType(blockutils.SlabTypeDouble)
			return v
		}()
	})
	p.mustRegisterBlock("dragon_egg", func(string) block.Behavior { return block.VanillaBlock("dragon_egg") })
	p.mustRegisterBlock("dragon_head", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeDragon)
			return v
		}()
	})
	p.mustRegisterBlock("dried_kelp_block", func(string) block.Behavior { return block.VanillaBlock("dried_kelp") })
	p.mustRegisterBlock("dyed_shulker_box", func(string) block.Behavior { return block.VanillaBlock("dyed_shulker_box") })
	p.mustRegisterBlock("element_0", func(string) block.Behavior { return block.VanillaBlock("element_zero") })
	p.mustRegisterBlock("element_1", func(string) block.Behavior { return block.VanillaBlock("element_hydrogen") })
	p.mustRegisterBlock("element_10", func(string) block.Behavior { return block.VanillaBlock("element_neon") })
	p.mustRegisterBlock("element_100", func(string) block.Behavior { return block.VanillaBlock("element_fermium") })
	p.mustRegisterBlock("element_101", func(string) block.Behavior { return block.VanillaBlock("element_mendelevium") })
	p.mustRegisterBlock("element_102", func(string) block.Behavior { return block.VanillaBlock("element_nobelium") })
	p.mustRegisterBlock("element_103", func(string) block.Behavior { return block.VanillaBlock("element_lawrencium") })
	p.mustRegisterBlock("element_104", func(string) block.Behavior { return block.VanillaBlock("element_rutherfordium") })
	p.mustRegisterBlock("element_105", func(string) block.Behavior { return block.VanillaBlock("element_dubnium") })
	p.mustRegisterBlock("element_106", func(string) block.Behavior { return block.VanillaBlock("element_seaborgium") })
	p.mustRegisterBlock("element_107", func(string) block.Behavior { return block.VanillaBlock("element_bohrium") })
	p.mustRegisterBlock("element_108", func(string) block.Behavior { return block.VanillaBlock("element_hassium") })
	p.mustRegisterBlock("element_109", func(string) block.Behavior { return block.VanillaBlock("element_meitnerium") })
	p.mustRegisterBlock("element_11", func(string) block.Behavior { return block.VanillaBlock("element_sodium") })
	p.mustRegisterBlock("element_110", func(string) block.Behavior { return block.VanillaBlock("element_darmstadtium") })
	p.mustRegisterBlock("element_111", func(string) block.Behavior { return block.VanillaBlock("element_roentgenium") })
	p.mustRegisterBlock("element_112", func(string) block.Behavior { return block.VanillaBlock("element_copernicium") })
	p.mustRegisterBlock("element_113", func(string) block.Behavior { return block.VanillaBlock("element_nihonium") })
	p.mustRegisterBlock("element_114", func(string) block.Behavior { return block.VanillaBlock("element_flerovium") })
	p.mustRegisterBlock("element_115", func(string) block.Behavior { return block.VanillaBlock("element_moscovium") })
	p.mustRegisterBlock("element_116", func(string) block.Behavior { return block.VanillaBlock("element_livermorium") })
	p.mustRegisterBlock("element_117", func(string) block.Behavior { return block.VanillaBlock("element_tennessine") })
	p.mustRegisterBlock("element_118", func(string) block.Behavior { return block.VanillaBlock("element_oganesson") })
	p.mustRegisterBlock("element_12", func(string) block.Behavior { return block.VanillaBlock("element_magnesium") })
	p.mustRegisterBlock("element_13", func(string) block.Behavior { return block.VanillaBlock("element_aluminum") })
	p.mustRegisterBlock("element_14", func(string) block.Behavior { return block.VanillaBlock("element_silicon") })
	p.mustRegisterBlock("element_15", func(string) block.Behavior { return block.VanillaBlock("element_phosphorus") })
	p.mustRegisterBlock("element_16", func(string) block.Behavior { return block.VanillaBlock("element_sulfur") })
	p.mustRegisterBlock("element_17", func(string) block.Behavior { return block.VanillaBlock("element_chlorine") })
	p.mustRegisterBlock("element_18", func(string) block.Behavior { return block.VanillaBlock("element_argon") })
	p.mustRegisterBlock("element_19", func(string) block.Behavior { return block.VanillaBlock("element_potassium") })
	p.mustRegisterBlock("element_2", func(string) block.Behavior { return block.VanillaBlock("element_helium") })
	p.mustRegisterBlock("element_20", func(string) block.Behavior { return block.VanillaBlock("element_calcium") })
	p.mustRegisterBlock("element_21", func(string) block.Behavior { return block.VanillaBlock("element_scandium") })
	p.mustRegisterBlock("element_22", func(string) block.Behavior { return block.VanillaBlock("element_titanium") })
	p.mustRegisterBlock("element_23", func(string) block.Behavior { return block.VanillaBlock("element_vanadium") })
	p.mustRegisterBlock("element_24", func(string) block.Behavior { return block.VanillaBlock("element_chromium") })
	p.mustRegisterBlock("element_25", func(string) block.Behavior { return block.VanillaBlock("element_manganese") })
	p.mustRegisterBlock("element_26", func(string) block.Behavior { return block.VanillaBlock("element_iron") })
	p.mustRegisterBlock("element_27", func(string) block.Behavior { return block.VanillaBlock("element_cobalt") })
	p.mustRegisterBlock("element_28", func(string) block.Behavior { return block.VanillaBlock("element_nickel") })
	p.mustRegisterBlock("element_29", func(string) block.Behavior { return block.VanillaBlock("element_copper") })
	p.mustRegisterBlock("element_3", func(string) block.Behavior { return block.VanillaBlock("element_lithium") })
	p.mustRegisterBlock("element_30", func(string) block.Behavior { return block.VanillaBlock("element_zinc") })
	p.mustRegisterBlock("element_31", func(string) block.Behavior { return block.VanillaBlock("element_gallium") })
	p.mustRegisterBlock("element_32", func(string) block.Behavior { return block.VanillaBlock("element_germanium") })
	p.mustRegisterBlock("element_33", func(string) block.Behavior { return block.VanillaBlock("element_arsenic") })
	p.mustRegisterBlock("element_34", func(string) block.Behavior { return block.VanillaBlock("element_selenium") })
	p.mustRegisterBlock("element_35", func(string) block.Behavior { return block.VanillaBlock("element_bromine") })
	p.mustRegisterBlock("element_36", func(string) block.Behavior { return block.VanillaBlock("element_krypton") })
	p.mustRegisterBlock("element_37", func(string) block.Behavior { return block.VanillaBlock("element_rubidium") })
	p.mustRegisterBlock("element_38", func(string) block.Behavior { return block.VanillaBlock("element_strontium") })
	p.mustRegisterBlock("element_39", func(string) block.Behavior { return block.VanillaBlock("element_yttrium") })
	p.mustRegisterBlock("element_4", func(string) block.Behavior { return block.VanillaBlock("element_beryllium") })
	p.mustRegisterBlock("element_40", func(string) block.Behavior { return block.VanillaBlock("element_zirconium") })
	p.mustRegisterBlock("element_41", func(string) block.Behavior { return block.VanillaBlock("element_niobium") })
	p.mustRegisterBlock("element_42", func(string) block.Behavior { return block.VanillaBlock("element_molybdenum") })
	p.mustRegisterBlock("element_43", func(string) block.Behavior { return block.VanillaBlock("element_technetium") })
	p.mustRegisterBlock("element_44", func(string) block.Behavior { return block.VanillaBlock("element_ruthenium") })
	p.mustRegisterBlock("element_45", func(string) block.Behavior { return block.VanillaBlock("element_rhodium") })
	p.mustRegisterBlock("element_46", func(string) block.Behavior { return block.VanillaBlock("element_palladium") })
	p.mustRegisterBlock("element_47", func(string) block.Behavior { return block.VanillaBlock("element_silver") })
	p.mustRegisterBlock("element_48", func(string) block.Behavior { return block.VanillaBlock("element_cadmium") })
	p.mustRegisterBlock("element_49", func(string) block.Behavior { return block.VanillaBlock("element_indium") })
	p.mustRegisterBlock("element_5", func(string) block.Behavior { return block.VanillaBlock("element_boron") })
	p.mustRegisterBlock("element_50", func(string) block.Behavior { return block.VanillaBlock("element_tin") })
	p.mustRegisterBlock("element_51", func(string) block.Behavior { return block.VanillaBlock("element_antimony") })
	p.mustRegisterBlock("element_52", func(string) block.Behavior { return block.VanillaBlock("element_tellurium") })
	p.mustRegisterBlock("element_53", func(string) block.Behavior { return block.VanillaBlock("element_iodine") })
	p.mustRegisterBlock("element_54", func(string) block.Behavior { return block.VanillaBlock("element_xenon") })
	p.mustRegisterBlock("element_55", func(string) block.Behavior { return block.VanillaBlock("element_cesium") })
	p.mustRegisterBlock("element_56", func(string) block.Behavior { return block.VanillaBlock("element_barium") })
	p.mustRegisterBlock("element_57", func(string) block.Behavior { return block.VanillaBlock("element_lanthanum") })
	p.mustRegisterBlock("element_58", func(string) block.Behavior { return block.VanillaBlock("element_cerium") })
	p.mustRegisterBlock("element_59", func(string) block.Behavior { return block.VanillaBlock("element_praseodymium") })
	p.mustRegisterBlock("element_6", func(string) block.Behavior { return block.VanillaBlock("element_carbon") })
	p.mustRegisterBlock("element_60", func(string) block.Behavior { return block.VanillaBlock("element_neodymium") })
	p.mustRegisterBlock("element_61", func(string) block.Behavior { return block.VanillaBlock("element_promethium") })
	p.mustRegisterBlock("element_62", func(string) block.Behavior { return block.VanillaBlock("element_samarium") })
	p.mustRegisterBlock("element_63", func(string) block.Behavior { return block.VanillaBlock("element_europium") })
	p.mustRegisterBlock("element_64", func(string) block.Behavior { return block.VanillaBlock("element_gadolinium") })
	p.mustRegisterBlock("element_65", func(string) block.Behavior { return block.VanillaBlock("element_terbium") })
	p.mustRegisterBlock("element_66", func(string) block.Behavior { return block.VanillaBlock("element_dysprosium") })
	p.mustRegisterBlock("element_67", func(string) block.Behavior { return block.VanillaBlock("element_holmium") })
	p.mustRegisterBlock("element_68", func(string) block.Behavior { return block.VanillaBlock("element_erbium") })
	p.mustRegisterBlock("element_69", func(string) block.Behavior { return block.VanillaBlock("element_thulium") })
	p.mustRegisterBlock("element_7", func(string) block.Behavior { return block.VanillaBlock("element_nitrogen") })
	p.mustRegisterBlock("element_70", func(string) block.Behavior { return block.VanillaBlock("element_ytterbium") })
	p.mustRegisterBlock("element_71", func(string) block.Behavior { return block.VanillaBlock("element_lutetium") })
	p.mustRegisterBlock("element_72", func(string) block.Behavior { return block.VanillaBlock("element_hafnium") })
	p.mustRegisterBlock("element_73", func(string) block.Behavior { return block.VanillaBlock("element_tantalum") })
	p.mustRegisterBlock("element_74", func(string) block.Behavior { return block.VanillaBlock("element_tungsten") })
	p.mustRegisterBlock("element_75", func(string) block.Behavior { return block.VanillaBlock("element_rhenium") })
	p.mustRegisterBlock("element_76", func(string) block.Behavior { return block.VanillaBlock("element_osmium") })
	p.mustRegisterBlock("element_77", func(string) block.Behavior { return block.VanillaBlock("element_iridium") })
	p.mustRegisterBlock("element_78", func(string) block.Behavior { return block.VanillaBlock("element_platinum") })
	p.mustRegisterBlock("element_79", func(string) block.Behavior { return block.VanillaBlock("element_gold") })
	p.mustRegisterBlock("element_8", func(string) block.Behavior { return block.VanillaBlock("element_oxygen") })
	p.mustRegisterBlock("element_80", func(string) block.Behavior { return block.VanillaBlock("element_mercury") })
	p.mustRegisterBlock("element_81", func(string) block.Behavior { return block.VanillaBlock("element_thallium") })
	p.mustRegisterBlock("element_82", func(string) block.Behavior { return block.VanillaBlock("element_lead") })
	p.mustRegisterBlock("element_83", func(string) block.Behavior { return block.VanillaBlock("element_bismuth") })
	p.mustRegisterBlock("element_84", func(string) block.Behavior { return block.VanillaBlock("element_polonium") })
	p.mustRegisterBlock("element_85", func(string) block.Behavior { return block.VanillaBlock("element_astatine") })
	p.mustRegisterBlock("element_86", func(string) block.Behavior { return block.VanillaBlock("element_radon") })
	p.mustRegisterBlock("element_87", func(string) block.Behavior { return block.VanillaBlock("element_francium") })
	p.mustRegisterBlock("element_88", func(string) block.Behavior { return block.VanillaBlock("element_radium") })
	p.mustRegisterBlock("element_89", func(string) block.Behavior { return block.VanillaBlock("element_actinium") })
	p.mustRegisterBlock("element_9", func(string) block.Behavior { return block.VanillaBlock("element_fluorine") })
	p.mustRegisterBlock("element_90", func(string) block.Behavior { return block.VanillaBlock("element_thorium") })
	p.mustRegisterBlock("element_91", func(string) block.Behavior { return block.VanillaBlock("element_protactinium") })
	p.mustRegisterBlock("element_92", func(string) block.Behavior { return block.VanillaBlock("element_uranium") })
	p.mustRegisterBlock("element_93", func(string) block.Behavior { return block.VanillaBlock("element_neptunium") })
	p.mustRegisterBlock("element_94", func(string) block.Behavior { return block.VanillaBlock("element_plutonium") })
	p.mustRegisterBlock("element_95", func(string) block.Behavior { return block.VanillaBlock("element_americium") })
	p.mustRegisterBlock("element_96", func(string) block.Behavior { return block.VanillaBlock("element_curium") })
	p.mustRegisterBlock("element_97", func(string) block.Behavior { return block.VanillaBlock("element_berkelium") })
	p.mustRegisterBlock("element_98", func(string) block.Behavior { return block.VanillaBlock("element_californium") })
	p.mustRegisterBlock("element_99", func(string) block.Behavior { return block.VanillaBlock("element_einsteinium") })
	p.mustRegisterBlock("element_actinium", func(string) block.Behavior { return block.VanillaBlock("element_actinium") })
	p.mustRegisterBlock("element_aluminum", func(string) block.Behavior { return block.VanillaBlock("element_aluminum") })
	p.mustRegisterBlock("element_americium", func(string) block.Behavior { return block.VanillaBlock("element_americium") })
	p.mustRegisterBlock("element_antimony", func(string) block.Behavior { return block.VanillaBlock("element_antimony") })
	p.mustRegisterBlock("element_argon", func(string) block.Behavior { return block.VanillaBlock("element_argon") })
	p.mustRegisterBlock("element_arsenic", func(string) block.Behavior { return block.VanillaBlock("element_arsenic") })
	p.mustRegisterBlock("element_astatine", func(string) block.Behavior { return block.VanillaBlock("element_astatine") })
	p.mustRegisterBlock("element_barium", func(string) block.Behavior { return block.VanillaBlock("element_barium") })
	p.mustRegisterBlock("element_berkelium", func(string) block.Behavior { return block.VanillaBlock("element_berkelium") })
	p.mustRegisterBlock("element_beryllium", func(string) block.Behavior { return block.VanillaBlock("element_beryllium") })
	p.mustRegisterBlock("element_bismuth", func(string) block.Behavior { return block.VanillaBlock("element_bismuth") })
	p.mustRegisterBlock("element_bohrium", func(string) block.Behavior { return block.VanillaBlock("element_bohrium") })
	p.mustRegisterBlock("element_boron", func(string) block.Behavior { return block.VanillaBlock("element_boron") })
	p.mustRegisterBlock("element_bromine", func(string) block.Behavior { return block.VanillaBlock("element_bromine") })
	p.mustRegisterBlock("element_cadmium", func(string) block.Behavior { return block.VanillaBlock("element_cadmium") })
	p.mustRegisterBlock("element_calcium", func(string) block.Behavior { return block.VanillaBlock("element_calcium") })
	p.mustRegisterBlock("element_californium", func(string) block.Behavior { return block.VanillaBlock("element_californium") })
	p.mustRegisterBlock("element_carbon", func(string) block.Behavior { return block.VanillaBlock("element_carbon") })
	p.mustRegisterBlock("element_cerium", func(string) block.Behavior { return block.VanillaBlock("element_cerium") })
	p.mustRegisterBlock("element_cesium", func(string) block.Behavior { return block.VanillaBlock("element_cesium") })
	p.mustRegisterBlock("element_chlorine", func(string) block.Behavior { return block.VanillaBlock("element_chlorine") })
	p.mustRegisterBlock("element_chromium", func(string) block.Behavior { return block.VanillaBlock("element_chromium") })
	p.mustRegisterBlock("element_cobalt", func(string) block.Behavior { return block.VanillaBlock("element_cobalt") })
	p.mustRegisterBlock("element_constructor", func(string) block.Behavior { return block.VanillaBlock("element_constructor") })
	p.mustRegisterBlock("element_copernicium", func(string) block.Behavior { return block.VanillaBlock("element_copernicium") })
	p.mustRegisterBlock("element_copper", func(string) block.Behavior { return block.VanillaBlock("element_copper") })
	p.mustRegisterBlock("element_curium", func(string) block.Behavior { return block.VanillaBlock("element_curium") })
	p.mustRegisterBlock("element_darmstadtium", func(string) block.Behavior { return block.VanillaBlock("element_darmstadtium") })
	p.mustRegisterBlock("element_dubnium", func(string) block.Behavior { return block.VanillaBlock("element_dubnium") })
	p.mustRegisterBlock("element_dysprosium", func(string) block.Behavior { return block.VanillaBlock("element_dysprosium") })
	p.mustRegisterBlock("element_einsteinium", func(string) block.Behavior { return block.VanillaBlock("element_einsteinium") })
	p.mustRegisterBlock("element_erbium", func(string) block.Behavior { return block.VanillaBlock("element_erbium") })
	p.mustRegisterBlock("element_europium", func(string) block.Behavior { return block.VanillaBlock("element_europium") })
	p.mustRegisterBlock("element_fermium", func(string) block.Behavior { return block.VanillaBlock("element_fermium") })
	p.mustRegisterBlock("element_flerovium", func(string) block.Behavior { return block.VanillaBlock("element_flerovium") })
	p.mustRegisterBlock("element_fluorine", func(string) block.Behavior { return block.VanillaBlock("element_fluorine") })
	p.mustRegisterBlock("element_francium", func(string) block.Behavior { return block.VanillaBlock("element_francium") })
	p.mustRegisterBlock("element_gadolinium", func(string) block.Behavior { return block.VanillaBlock("element_gadolinium") })
	p.mustRegisterBlock("element_gallium", func(string) block.Behavior { return block.VanillaBlock("element_gallium") })
	p.mustRegisterBlock("element_germanium", func(string) block.Behavior { return block.VanillaBlock("element_germanium") })
	p.mustRegisterBlock("element_gold", func(string) block.Behavior { return block.VanillaBlock("element_gold") })
	p.mustRegisterBlock("element_hafnium", func(string) block.Behavior { return block.VanillaBlock("element_hafnium") })
	p.mustRegisterBlock("element_hassium", func(string) block.Behavior { return block.VanillaBlock("element_hassium") })
	p.mustRegisterBlock("element_helium", func(string) block.Behavior { return block.VanillaBlock("element_helium") })
	p.mustRegisterBlock("element_holmium", func(string) block.Behavior { return block.VanillaBlock("element_holmium") })
	p.mustRegisterBlock("element_hydrogen", func(string) block.Behavior { return block.VanillaBlock("element_hydrogen") })
	p.mustRegisterBlock("element_indium", func(string) block.Behavior { return block.VanillaBlock("element_indium") })
	p.mustRegisterBlock("element_iodine", func(string) block.Behavior { return block.VanillaBlock("element_iodine") })
	p.mustRegisterBlock("element_iridium", func(string) block.Behavior { return block.VanillaBlock("element_iridium") })
	p.mustRegisterBlock("element_iron", func(string) block.Behavior { return block.VanillaBlock("element_iron") })
	p.mustRegisterBlock("element_krypton", func(string) block.Behavior { return block.VanillaBlock("element_krypton") })
	p.mustRegisterBlock("element_lanthanum", func(string) block.Behavior { return block.VanillaBlock("element_lanthanum") })
	p.mustRegisterBlock("element_lawrencium", func(string) block.Behavior { return block.VanillaBlock("element_lawrencium") })
	p.mustRegisterBlock("element_lead", func(string) block.Behavior { return block.VanillaBlock("element_lead") })
	p.mustRegisterBlock("element_lithium", func(string) block.Behavior { return block.VanillaBlock("element_lithium") })
	p.mustRegisterBlock("element_livermorium", func(string) block.Behavior { return block.VanillaBlock("element_livermorium") })
	p.mustRegisterBlock("element_lutetium", func(string) block.Behavior { return block.VanillaBlock("element_lutetium") })
	p.mustRegisterBlock("element_magnesium", func(string) block.Behavior { return block.VanillaBlock("element_magnesium") })
	p.mustRegisterBlock("element_manganese", func(string) block.Behavior { return block.VanillaBlock("element_manganese") })
	p.mustRegisterBlock("element_meitnerium", func(string) block.Behavior { return block.VanillaBlock("element_meitnerium") })
	p.mustRegisterBlock("element_mendelevium", func(string) block.Behavior { return block.VanillaBlock("element_mendelevium") })
	p.mustRegisterBlock("element_mercury", func(string) block.Behavior { return block.VanillaBlock("element_mercury") })
	p.mustRegisterBlock("element_molybdenum", func(string) block.Behavior { return block.VanillaBlock("element_molybdenum") })
	p.mustRegisterBlock("element_moscovium", func(string) block.Behavior { return block.VanillaBlock("element_moscovium") })
	p.mustRegisterBlock("element_neodymium", func(string) block.Behavior { return block.VanillaBlock("element_neodymium") })
	p.mustRegisterBlock("element_neon", func(string) block.Behavior { return block.VanillaBlock("element_neon") })
	p.mustRegisterBlock("element_neptunium", func(string) block.Behavior { return block.VanillaBlock("element_neptunium") })
	p.mustRegisterBlock("element_nickel", func(string) block.Behavior { return block.VanillaBlock("element_nickel") })
	p.mustRegisterBlock("element_nihonium", func(string) block.Behavior { return block.VanillaBlock("element_nihonium") })
	p.mustRegisterBlock("element_niobium", func(string) block.Behavior { return block.VanillaBlock("element_niobium") })
	p.mustRegisterBlock("element_nitrogen", func(string) block.Behavior { return block.VanillaBlock("element_nitrogen") })
	p.mustRegisterBlock("element_nobelium", func(string) block.Behavior { return block.VanillaBlock("element_nobelium") })
	p.mustRegisterBlock("element_oganesson", func(string) block.Behavior { return block.VanillaBlock("element_oganesson") })
	p.mustRegisterBlock("element_osmium", func(string) block.Behavior { return block.VanillaBlock("element_osmium") })
	p.mustRegisterBlock("element_oxygen", func(string) block.Behavior { return block.VanillaBlock("element_oxygen") })
	p.mustRegisterBlock("element_palladium", func(string) block.Behavior { return block.VanillaBlock("element_palladium") })
	p.mustRegisterBlock("element_phosphorus", func(string) block.Behavior { return block.VanillaBlock("element_phosphorus") })
	p.mustRegisterBlock("element_platinum", func(string) block.Behavior { return block.VanillaBlock("element_platinum") })
	p.mustRegisterBlock("element_plutonium", func(string) block.Behavior { return block.VanillaBlock("element_plutonium") })
	p.mustRegisterBlock("element_polonium", func(string) block.Behavior { return block.VanillaBlock("element_polonium") })
	p.mustRegisterBlock("element_potassium", func(string) block.Behavior { return block.VanillaBlock("element_potassium") })
	p.mustRegisterBlock("element_praseodymium", func(string) block.Behavior { return block.VanillaBlock("element_praseodymium") })
	p.mustRegisterBlock("element_promethium", func(string) block.Behavior { return block.VanillaBlock("element_promethium") })
	p.mustRegisterBlock("element_protactinium", func(string) block.Behavior { return block.VanillaBlock("element_protactinium") })
	p.mustRegisterBlock("element_radium", func(string) block.Behavior { return block.VanillaBlock("element_radium") })
	p.mustRegisterBlock("element_radon", func(string) block.Behavior { return block.VanillaBlock("element_radon") })
	p.mustRegisterBlock("element_rhenium", func(string) block.Behavior { return block.VanillaBlock("element_rhenium") })
	p.mustRegisterBlock("element_rhodium", func(string) block.Behavior { return block.VanillaBlock("element_rhodium") })
	p.mustRegisterBlock("element_roentgenium", func(string) block.Behavior { return block.VanillaBlock("element_roentgenium") })
	p.mustRegisterBlock("element_rubidium", func(string) block.Behavior { return block.VanillaBlock("element_rubidium") })
	p.mustRegisterBlock("element_ruthenium", func(string) block.Behavior { return block.VanillaBlock("element_ruthenium") })
	p.mustRegisterBlock("element_rutherfordium", func(string) block.Behavior { return block.VanillaBlock("element_rutherfordium") })
	p.mustRegisterBlock("element_samarium", func(string) block.Behavior { return block.VanillaBlock("element_samarium") })
	p.mustRegisterBlock("element_scandium", func(string) block.Behavior { return block.VanillaBlock("element_scandium") })
	p.mustRegisterBlock("element_seaborgium", func(string) block.Behavior { return block.VanillaBlock("element_seaborgium") })
	p.mustRegisterBlock("element_selenium", func(string) block.Behavior { return block.VanillaBlock("element_selenium") })
	p.mustRegisterBlock("element_silicon", func(string) block.Behavior { return block.VanillaBlock("element_silicon") })
	p.mustRegisterBlock("element_silver", func(string) block.Behavior { return block.VanillaBlock("element_silver") })
	p.mustRegisterBlock("element_sodium", func(string) block.Behavior { return block.VanillaBlock("element_sodium") })
	p.mustRegisterBlock("element_strontium", func(string) block.Behavior { return block.VanillaBlock("element_strontium") })
	p.mustRegisterBlock("element_sulfur", func(string) block.Behavior { return block.VanillaBlock("element_sulfur") })
	p.mustRegisterBlock("element_tantalum", func(string) block.Behavior { return block.VanillaBlock("element_tantalum") })
	p.mustRegisterBlock("element_technetium", func(string) block.Behavior { return block.VanillaBlock("element_technetium") })
	p.mustRegisterBlock("element_tellurium", func(string) block.Behavior { return block.VanillaBlock("element_tellurium") })
	p.mustRegisterBlock("element_tennessine", func(string) block.Behavior { return block.VanillaBlock("element_tennessine") })
	p.mustRegisterBlock("element_terbium", func(string) block.Behavior { return block.VanillaBlock("element_terbium") })
	p.mustRegisterBlock("element_thallium", func(string) block.Behavior { return block.VanillaBlock("element_thallium") })
	p.mustRegisterBlock("element_thorium", func(string) block.Behavior { return block.VanillaBlock("element_thorium") })
	p.mustRegisterBlock("element_thulium", func(string) block.Behavior { return block.VanillaBlock("element_thulium") })
	p.mustRegisterBlock("element_tin", func(string) block.Behavior { return block.VanillaBlock("element_tin") })
	p.mustRegisterBlock("element_titanium", func(string) block.Behavior { return block.VanillaBlock("element_titanium") })
	p.mustRegisterBlock("element_tungsten", func(string) block.Behavior { return block.VanillaBlock("element_tungsten") })
	p.mustRegisterBlock("element_uranium", func(string) block.Behavior { return block.VanillaBlock("element_uranium") })
	p.mustRegisterBlock("element_vanadium", func(string) block.Behavior { return block.VanillaBlock("element_vanadium") })
	p.mustRegisterBlock("element_xenon", func(string) block.Behavior { return block.VanillaBlock("element_xenon") })
	p.mustRegisterBlock("element_ytterbium", func(string) block.Behavior { return block.VanillaBlock("element_ytterbium") })
	p.mustRegisterBlock("element_yttrium", func(string) block.Behavior { return block.VanillaBlock("element_yttrium") })
	p.mustRegisterBlock("element_zero", func(string) block.Behavior { return block.VanillaBlock("element_zero") })
	p.mustRegisterBlock("element_zinc", func(string) block.Behavior { return block.VanillaBlock("element_zinc") })
	p.mustRegisterBlock("element_zirconium", func(string) block.Behavior { return block.VanillaBlock("element_zirconium") })
	p.mustRegisterBlock("emerald_block", func(string) block.Behavior { return block.VanillaBlock("emerald") })
	p.mustRegisterBlock("emerald_ore", func(string) block.Behavior { return block.VanillaBlock("emerald_ore") })
	p.mustRegisterBlock("enchant_table", func(string) block.Behavior { return block.VanillaBlock("enchanting_table") })
	p.mustRegisterBlock("enchanting_table", func(string) block.Behavior { return block.VanillaBlock("enchanting_table") })
	p.mustRegisterBlock("enchantment_table", func(string) block.Behavior { return block.VanillaBlock("enchanting_table") })
	p.mustRegisterBlock("end_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("end_stone_brick_stairs") })
	p.mustRegisterBlock("end_bricks", func(string) block.Behavior { return block.VanillaBlock("end_stone_bricks") })
	p.mustRegisterBlock("end_portal_frame", func(string) block.Behavior { return block.VanillaBlock("end_portal_frame") })
	p.mustRegisterBlock("end_rod", func(string) block.Behavior { return block.VanillaBlock("end_rod") })
	p.mustRegisterBlock("end_stone", func(string) block.Behavior { return block.VanillaBlock("end_stone") })
	p.mustRegisterBlock("end_stone_brick_slab", func(string) block.Behavior { return block.VanillaBlock("end_stone_brick_slab") })
	p.mustRegisterBlock("end_stone_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("end_stone_brick_stairs") })
	p.mustRegisterBlock("end_stone_brick_wall", func(string) block.Behavior { return block.VanillaBlock("end_stone_brick_wall") })
	p.mustRegisterBlock("end_stone_bricks", func(string) block.Behavior { return block.VanillaBlock("end_stone_bricks") })
	p.mustRegisterBlock("ender_chest", func(string) block.Behavior { return block.VanillaBlock("ender_chest") })
	p.mustRegisterBlock("fake_wooden_slab", func(string) block.Behavior { return block.VanillaBlock("fake_wooden_slab") })
	p.mustRegisterBlock("farmland", func(string) block.Behavior { return block.VanillaBlock("farmland") })
	p.mustRegisterBlock("fence", func(string) block.Behavior { return block.VanillaBlock("oak_fence") })
	p.mustRegisterBlock("fence_gate", func(string) block.Behavior { return block.VanillaBlock("oak_fence_gate") })
	p.mustRegisterBlock("fence_gate_acacia", func(string) block.Behavior { return block.VanillaBlock("acacia_fence_gate") })
	p.mustRegisterBlock("fence_gate_birch", func(string) block.Behavior { return block.VanillaBlock("birch_fence_gate") })
	p.mustRegisterBlock("fence_gate_dark_oak", func(string) block.Behavior { return block.VanillaBlock("dark_oak_fence_gate") })
	p.mustRegisterBlock("fence_gate_jungle", func(string) block.Behavior { return block.VanillaBlock("jungle_fence_gate") })
	p.mustRegisterBlock("fence_gate_spruce", func(string) block.Behavior { return block.VanillaBlock("spruce_fence_gate") })
	p.mustRegisterBlock("fern", func(string) block.Behavior { return block.VanillaBlock("fern") })
	p.mustRegisterBlock("fire", func(string) block.Behavior { return block.VanillaBlock("fire") })
	p.mustRegisterBlock("fletching_table", func(string) block.Behavior { return block.VanillaBlock("fletching_table") })
	p.mustRegisterBlock("flower_pot", func(string) block.Behavior { return block.VanillaBlock("flower_pot") })
	p.mustRegisterBlock("flower_pot_block", func(string) block.Behavior { return block.VanillaBlock("flower_pot") })
	p.mustRegisterBlock("flowering_azalea", func(string) block.Behavior { return block.VanillaBlock("flowering_azalea") })
	p.mustRegisterBlock("flowering_azalea_leaves", func(string) block.Behavior { return block.VanillaBlock("flowering_azalea_leaves") })
	p.mustRegisterBlock("flowing_lava", func(string) block.Behavior { return block.VanillaBlock("lava") })
	p.mustRegisterBlock("flowing_water", func(string) block.Behavior { return block.VanillaBlock("water") })
	p.mustRegisterBlock("frame", func(string) block.Behavior { return block.VanillaBlock("item_frame") })
	p.mustRegisterBlock("frame_block", func(string) block.Behavior { return block.VanillaBlock("item_frame") })
	p.mustRegisterBlock("frosted_ice", func(string) block.Behavior { return block.VanillaBlock("frosted_ice") })
	p.mustRegisterBlock("furnace", func(string) block.Behavior { return block.VanillaBlock("furnace") })
	p.mustRegisterBlock("gilded_blackstone", func(string) block.Behavior { return block.VanillaBlock("gilded_blackstone") })
	p.mustRegisterBlock("glass", func(string) block.Behavior { return block.VanillaBlock("glass") })
	p.mustRegisterBlock("glass_pane", func(string) block.Behavior { return block.VanillaBlock("glass_pane") })
	p.mustRegisterBlock("glass_panel", func(string) block.Behavior { return block.VanillaBlock("glass_pane") })
	p.mustRegisterBlock("glazed_terracotta", func(string) block.Behavior { return block.VanillaBlock("glazed_terracotta") })
	p.mustRegisterBlock("glow_frame", func(string) block.Behavior { return block.VanillaBlock("glowing_item_frame") })
	p.mustRegisterBlock("glow_item_frame", func(string) block.Behavior { return block.VanillaBlock("glowing_item_frame") })
	p.mustRegisterBlock("glowing_item_frame", func(string) block.Behavior { return block.VanillaBlock("glowing_item_frame") })
	p.mustRegisterBlock("glowing_obsidian", func(string) block.Behavior { return block.VanillaBlock("glowing_obsidian") })
	p.mustRegisterBlock("glowing_redstone_ore", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_ore")
			v.(interface{ SetLit(bool) }).SetLit(true)
			return v
		}()
	})
	p.mustRegisterBlock("glowingobsidian", func(string) block.Behavior { return block.VanillaBlock("glowing_obsidian") })
	p.mustRegisterBlock("glowstone", func(string) block.Behavior { return block.VanillaBlock("glowstone") })
	p.mustRegisterBlock("glowstone_block", func(string) block.Behavior { return block.VanillaBlock("glowstone") })
	p.mustRegisterBlock("glow_lichen", func(string) block.Behavior { return block.VanillaBlock("glow_lichen") })
	p.mustRegisterBlock("gold", func(string) block.Behavior { return block.VanillaBlock("gold") })
	p.mustRegisterBlock("gold_block", func(string) block.Behavior { return block.VanillaBlock("gold") })
	p.mustRegisterBlock("gold_ore", func(string) block.Behavior { return block.VanillaBlock("gold_ore") })
	p.mustRegisterBlock("gold_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_light") })
	p.mustRegisterBlock("golden_rail", func(string) block.Behavior { return block.VanillaBlock("powered_rail") })
	p.mustRegisterBlock("granite", func(string) block.Behavior { return block.VanillaBlock("granite") })
	p.mustRegisterBlock("granite_slab", func(string) block.Behavior { return block.VanillaBlock("granite_slab") })
	p.mustRegisterBlock("granite_stairs", func(string) block.Behavior { return block.VanillaBlock("granite_stairs") })
	p.mustRegisterBlock("granite_wall", func(string) block.Behavior { return block.VanillaBlock("granite_wall") })
	p.mustRegisterBlock("grass", func(string) block.Behavior { return block.VanillaBlock("grass") })
	p.mustRegisterBlock("grass_path", func(string) block.Behavior { return block.VanillaBlock("grass_path") })
	p.mustRegisterBlock("gravel", func(string) block.Behavior { return block.VanillaBlock("gravel") })
	p.mustRegisterBlock("green_torch", func(string) block.Behavior { return block.VanillaBlock("green_torch") })
	p.mustRegisterBlock("hanging_roots", func(string) block.Behavior { return block.VanillaBlock("hanging_roots") })
	p.mustRegisterBlock("hard_glass", func(string) block.Behavior { return block.VanillaBlock("hardened_glass") })
	p.mustRegisterBlock("hard_glass_pane", func(string) block.Behavior { return block.VanillaBlock("hardened_glass_pane") })
	p.mustRegisterBlock("hard_stained_glass", func(string) block.Behavior { return block.VanillaBlock("stained_hardened_glass") })
	p.mustRegisterBlock("hard_stained_glass_pane", func(string) block.Behavior { return block.VanillaBlock("stained_hardened_glass_pane") })
	p.mustRegisterBlock("hardened_clay", func(string) block.Behavior { return block.VanillaBlock("hardened_clay") })
	p.mustRegisterBlock("hardened_glass", func(string) block.Behavior { return block.VanillaBlock("hardened_glass") })
	p.mustRegisterBlock("hardened_glass_pane", func(string) block.Behavior { return block.VanillaBlock("hardened_glass_pane") })
	p.mustRegisterBlock("hay_bale", func(string) block.Behavior { return block.VanillaBlock("hay_bale") })
	p.mustRegisterBlock("hay_block", func(string) block.Behavior { return block.VanillaBlock("hay_bale") })
	p.mustRegisterBlock("heavy_weighted_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_heavy") })
	p.mustRegisterBlock("honeycomb_block", func(string) block.Behavior { return block.VanillaBlock("honeycomb") })
	p.mustRegisterBlock("hopper", func(string) block.Behavior { return block.VanillaBlock("hopper") })
	p.mustRegisterBlock("hopper_block", func(string) block.Behavior { return block.VanillaBlock("hopper") })
	p.mustRegisterBlock("ice", func(string) block.Behavior { return block.VanillaBlock("ice") })
	p.mustRegisterBlock("inactive_redstone_lamp", func(string) block.Behavior { return block.VanillaBlock("redstone_lamp") })
	p.mustRegisterBlock("infested_chiseled_stone_brick", func(string) block.Behavior { return block.VanillaBlock("infested_chiseled_stone_brick") })
	p.mustRegisterBlock("infested_cobblestone", func(string) block.Behavior { return block.VanillaBlock("infested_cobblestone") })
	p.mustRegisterBlock("infested_cracked_stone_brick", func(string) block.Behavior { return block.VanillaBlock("infested_cracked_stone_brick") })
	p.mustRegisterBlock("infested_deepslate", func(string) block.Behavior { return block.VanillaBlock("infested_deepslate") })
	p.mustRegisterBlock("infested_mossy_stone_brick", func(string) block.Behavior { return block.VanillaBlock("infested_mossy_stone_brick") })
	p.mustRegisterBlock("infested_stone", func(string) block.Behavior { return block.VanillaBlock("infested_stone") })
	p.mustRegisterBlock("infested_stone_brick", func(string) block.Behavior { return block.VanillaBlock("infested_stone_brick") })
	p.mustRegisterBlock("info_reserved6", func(string) block.Behavior { return block.VanillaBlock("reserved6") })
	p.mustRegisterBlock("info_update", func(string) block.Behavior { return block.VanillaBlock("info_update") })
	p.mustRegisterBlock("info_update2", func(string) block.Behavior { return block.VanillaBlock("info_update2") })
	p.mustRegisterBlock("inverted_daylight_sensor", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("daylight_sensor")
			v.(interface{ SetInverted(bool) }).SetInverted(true)
			return v
		}()
	})
	p.mustRegisterBlock("invisible_bedrock", func(string) block.Behavior { return block.VanillaBlock("invisible_bedrock") })
	p.mustRegisterBlock("invisiblebedrock", func(string) block.Behavior { return block.VanillaBlock("invisible_bedrock") })
	p.mustRegisterBlock("iron", func(string) block.Behavior { return block.VanillaBlock("iron") })
	p.mustRegisterBlock("iron_bar", func(string) block.Behavior { return block.VanillaBlock("iron_bars") })
	p.mustRegisterBlock("iron_bars", func(string) block.Behavior { return block.VanillaBlock("iron_bars") })
	p.mustRegisterBlock("iron_block", func(string) block.Behavior { return block.VanillaBlock("iron") })
	p.mustRegisterBlock("iron_door", func(string) block.Behavior { return block.VanillaBlock("iron_door") })
	p.mustRegisterBlock("iron_door_block", func(string) block.Behavior { return block.VanillaBlock("iron_door") })
	p.mustRegisterBlock("iron_ore", func(string) block.Behavior { return block.VanillaBlock("iron_ore") })
	p.mustRegisterBlock("iron_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_heavy") })
	p.mustRegisterBlock("iron_trapdoor", func(string) block.Behavior { return block.VanillaBlock("iron_trapdoor") })
	p.mustRegisterBlock("item_frame", func(string) block.Behavior { return block.VanillaBlock("item_frame") })
	p.mustRegisterBlock("item_frame_block", func(string) block.Behavior { return block.VanillaBlock("item_frame") })
	p.mustRegisterBlock("jack_o_lantern", func(string) block.Behavior { return block.VanillaBlock("lit_pumpkin") })
	p.mustRegisterBlock("jukebox", func(string) block.Behavior { return block.VanillaBlock("jukebox") })
	p.mustRegisterBlock("jungle_button", func(string) block.Behavior { return block.VanillaBlock("jungle_button") })
	p.mustRegisterBlock("jungle_door", func(string) block.Behavior { return block.VanillaBlock("jungle_door") })
	p.mustRegisterBlock("jungle_door_block", func(string) block.Behavior { return block.VanillaBlock("jungle_door") })
	p.mustRegisterBlock("jungle_fence", func(string) block.Behavior { return block.VanillaBlock("jungle_fence") })
	p.mustRegisterBlock("jungle_fence_gate", func(string) block.Behavior { return block.VanillaBlock("jungle_fence_gate") })
	p.mustRegisterBlock("jungle_leaves", func(string) block.Behavior { return block.VanillaBlock("jungle_leaves") })
	p.mustRegisterBlock("jungle_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("jungle_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("jungle_planks", func(string) block.Behavior { return block.VanillaBlock("jungle_planks") })
	p.mustRegisterBlock("jungle_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("jungle_pressure_plate") })
	p.mustRegisterBlock("jungle_sapling", func(string) block.Behavior { return block.VanillaBlock("jungle_sapling") })
	p.mustRegisterBlock("jungle_sign", func(string) block.Behavior { return block.VanillaBlock("jungle_sign") })
	p.mustRegisterBlock("jungle_slab", func(string) block.Behavior { return block.VanillaBlock("jungle_slab") })
	p.mustRegisterBlock("jungle_stairs", func(string) block.Behavior { return block.VanillaBlock("jungle_stairs") })
	p.mustRegisterBlock("jungle_standing_sign", func(string) block.Behavior { return block.VanillaBlock("jungle_sign") })
	p.mustRegisterBlock("jungle_trapdoor", func(string) block.Behavior { return block.VanillaBlock("jungle_trapdoor") })
	p.mustRegisterBlock("jungle_wall_sign", func(string) block.Behavior { return block.VanillaBlock("jungle_wall_sign") })
	p.mustRegisterBlock("jungle_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("jungle_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("jungle_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("jungle_stairs") })
	p.mustRegisterBlock("jungle_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("jungle_stairs") })
	p.mustRegisterBlock("lab_table", func(string) block.Behavior { return block.VanillaBlock("lab_table") })
	p.mustRegisterBlock("ladder", func(string) block.Behavior { return block.VanillaBlock("ladder") })
	p.mustRegisterBlock("lantern", func(string) block.Behavior { return block.VanillaBlock("lantern") })
	p.mustRegisterBlock("lapis_block", func(string) block.Behavior { return block.VanillaBlock("lapis_lazuli") })
	p.mustRegisterBlock("lapis_lazuli_block", func(string) block.Behavior { return block.VanillaBlock("lapis_lazuli") })
	p.mustRegisterBlock("lapis_lazuli_ore", func(string) block.Behavior { return block.VanillaBlock("lapis_lazuli_ore") })
	p.mustRegisterBlock("lapis_ore", func(string) block.Behavior { return block.VanillaBlock("lapis_lazuli_ore") })
	p.mustRegisterBlock("large_amethyst_bud", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("amethyst_cluster")
			v.(interface{ SetStage(int) }).SetStage(block.AmethystClusterStageLargeBud)
			return v
		}()
	})
	p.mustRegisterBlock("large_fern", func(string) block.Behavior { return block.VanillaBlock("large_fern") })
	p.mustRegisterBlock("lava", func(string) block.Behavior { return block.VanillaBlock("lava") })
	p.mustRegisterBlock("leave", func(string) block.Behavior { return block.VanillaBlock("oak_leaves") })
	p.mustRegisterBlock("leave2", func(string) block.Behavior { return block.VanillaBlock("acacia_leaves") })
	p.mustRegisterBlock("leaves", func(string) block.Behavior { return block.VanillaBlock("oak_leaves") })
	p.mustRegisterBlock("leaves2", func(string) block.Behavior { return block.VanillaBlock("acacia_leaves") })
	p.mustRegisterBlock("lectern", func(string) block.Behavior { return block.VanillaBlock("lectern") })
	p.mustRegisterBlock("legacy_stonecutter", func(string) block.Behavior { return block.VanillaBlock("legacy_stonecutter") })
	p.mustRegisterBlock("lever", func(string) block.Behavior { return block.VanillaBlock("lever") })
	p.mustRegisterBlock("light", func(string) block.Behavior { return block.VanillaBlock("light") })
	p.mustRegisterBlock("light_block", func(string) block.Behavior { return block.VanillaBlock("light") })
	p.mustRegisterBlock("light_weighted_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_light") })
	p.mustRegisterBlock("lilac", func(string) block.Behavior { return block.VanillaBlock("lilac") })
	p.mustRegisterBlock("lily_of_the_valley", func(string) block.Behavior { return block.VanillaBlock("lily_of_the_valley") })
	p.mustRegisterBlock("lily_pad", func(string) block.Behavior { return block.VanillaBlock("lily_pad") })
	p.mustRegisterBlock("lit_blast_furnace", func(string) block.Behavior { return block.VanillaBlock("blast_furnace") })
	p.mustRegisterBlock("lit_furnace", func(string) block.Behavior { return block.VanillaBlock("furnace") })
	p.mustRegisterBlock("lit_pumpkin", func(string) block.Behavior { return block.VanillaBlock("lit_pumpkin") })
	p.mustRegisterBlock("lit_redstone_lamp", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_lamp")
			v.(interface{ SetPowered(bool) }).SetPowered(true)
			return v
		}()
	})
	p.mustRegisterBlock("lit_redstone_ore", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_ore")
			v.(interface{ SetLit(bool) }).SetLit(true)
			return v
		}()
	})
	p.mustRegisterBlock("lit_redstone_torch", func(string) block.Behavior { return block.VanillaBlock("redstone_torch") })
	p.mustRegisterBlock("lit_smoker", func(string) block.Behavior { return block.VanillaBlock("smoker") })
	p.mustRegisterBlock("log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("log2", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("loom", func(string) block.Behavior { return block.VanillaBlock("loom") })
	p.mustRegisterBlock("magma", func(string) block.Behavior { return block.VanillaBlock("magma") })
	p.mustRegisterBlock("mangrove_button", func(string) block.Behavior { return block.VanillaBlock("mangrove_button") })
	p.mustRegisterBlock("mangrove_door", func(string) block.Behavior { return block.VanillaBlock("mangrove_door") })
	p.mustRegisterBlock("mangrove_fence", func(string) block.Behavior { return block.VanillaBlock("mangrove_fence") })
	p.mustRegisterBlock("mangrove_fence_gate", func(string) block.Behavior { return block.VanillaBlock("mangrove_fence_gate") })
	p.mustRegisterBlock("mangrove_leaves", func(string) block.Behavior { return block.VanillaBlock("mangrove_leaves") })
	p.mustRegisterBlock("mangrove_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mangrove_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("mangrove_planks", func(string) block.Behavior { return block.VanillaBlock("mangrove_planks") })
	p.mustRegisterBlock("mangrove_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("mangrove_pressure_plate") })
	p.mustRegisterBlock("mangrove_roots", func(string) block.Behavior { return block.VanillaBlock("mangrove_roots") })
	p.mustRegisterBlock("mangrove_sign", func(string) block.Behavior { return block.VanillaBlock("mangrove_sign") })
	p.mustRegisterBlock("mangrove_slab", func(string) block.Behavior { return block.VanillaBlock("mangrove_slab") })
	p.mustRegisterBlock("mangrove_stairs", func(string) block.Behavior { return block.VanillaBlock("mangrove_stairs") })
	p.mustRegisterBlock("mangrove_trapdoor", func(string) block.Behavior { return block.VanillaBlock("mangrove_trapdoor") })
	p.mustRegisterBlock("mangrove_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mangrove_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("material_reducer", func(string) block.Behavior { return block.VanillaBlock("material_reducer") })
	p.mustRegisterBlock("medium_amethyst_bud", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("amethyst_cluster")
			v.(interface{ SetStage(int) }).SetStage(block.AmethystClusterStageMediumBud)
			return v
		}()
	})
	p.mustRegisterBlock("melon_block", func(string) block.Behavior { return block.VanillaBlock("melon") })
	p.mustRegisterBlock("melon_stem", func(string) block.Behavior { return block.VanillaBlock("melon_stem") })
	p.mustRegisterBlock("mob_head", func(string) block.Behavior { return block.VanillaBlock("mob_head") })
	p.mustRegisterBlock("mob_head_block", func(string) block.Behavior { return block.VanillaBlock("mob_head") })
	p.mustRegisterBlock("mob_spawner", func(string) block.Behavior { return block.VanillaBlock("monster_spawner") })
	p.mustRegisterBlock("monster_egg", func(string) block.Behavior { return block.VanillaBlock("infested_stone") })
	p.mustRegisterBlock("monster_egg_block", func(string) block.Behavior { return block.VanillaBlock("infested_stone") })
	p.mustRegisterBlock("monster_spawner", func(string) block.Behavior { return block.VanillaBlock("monster_spawner") })
	p.mustRegisterBlock("moss_stone", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone") })
	p.mustRegisterBlock("mossy_cobblestone", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone") })
	p.mustRegisterBlock("mossy_cobblestone_slab", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone_slab") })
	p.mustRegisterBlock("mossy_cobblestone_stairs", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone_stairs") })
	p.mustRegisterBlock("mossy_cobblestone_wall", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone_wall") })
	p.mustRegisterBlock("mossy_stone", func(string) block.Behavior { return block.VanillaBlock("mossy_cobblestone") })
	p.mustRegisterBlock("mossy_stone_brick_slab", func(string) block.Behavior { return block.VanillaBlock("mossy_stone_brick_slab") })
	p.mustRegisterBlock("mossy_stone_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("mossy_stone_brick_stairs") })
	p.mustRegisterBlock("mossy_stone_brick_wall", func(string) block.Behavior { return block.VanillaBlock("mossy_stone_brick_wall") })
	p.mustRegisterBlock("mossy_stone_bricks", func(string) block.Behavior { return block.VanillaBlock("mossy_stone_bricks") })
	p.mustRegisterBlock("mud", func(string) block.Behavior { return block.VanillaBlock("mud") })
	p.mustRegisterBlock("mud_bricks", func(string) block.Behavior { return block.VanillaBlock("mud_bricks") })
	p.mustRegisterBlock("mud_brick_slab", func(string) block.Behavior { return block.VanillaBlock("mud_brick_slab") })
	p.mustRegisterBlock("mud_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("mud_brick_stairs") })
	p.mustRegisterBlock("mud_brick_wall", func(string) block.Behavior { return block.VanillaBlock("mud_brick_wall") })
	p.mustRegisterBlock("muddy_mangrove_roots", func(string) block.Behavior { return block.VanillaBlock("muddy_mangrove_roots") })
	p.mustRegisterBlock("mushroom_stem", func(string) block.Behavior { return block.VanillaBlock("mushroom_stem") })
	p.mustRegisterBlock("mycelium", func(string) block.Behavior { return block.VanillaBlock("mycelium") })
	p.mustRegisterBlock("nether_brick_block", func(string) block.Behavior { return block.VanillaBlock("nether_bricks") })
	p.mustRegisterBlock("nether_brick_fence", func(string) block.Behavior { return block.VanillaBlock("nether_brick_fence") })
	p.mustRegisterBlock("nether_brick_slab", func(string) block.Behavior { return block.VanillaBlock("nether_brick_slab") })
	p.mustRegisterBlock("nether_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("nether_brick_stairs") })
	p.mustRegisterBlock("nether_brick_wall", func(string) block.Behavior { return block.VanillaBlock("nether_brick_wall") })
	p.mustRegisterBlock("nether_bricks", func(string) block.Behavior { return block.VanillaBlock("nether_bricks") })
	p.mustRegisterBlock("nether_bricks_stairs", func(string) block.Behavior { return block.VanillaBlock("nether_brick_stairs") })
	p.mustRegisterBlock("nether_gold_ore", func(string) block.Behavior { return block.VanillaBlock("nether_gold_ore") })
	p.mustRegisterBlock("nether_portal", func(string) block.Behavior { return block.VanillaBlock("nether_portal") })
	p.mustRegisterBlock("nether_quartz_ore", func(string) block.Behavior { return block.VanillaBlock("nether_quartz_ore") })
	p.mustRegisterBlock("nether_reactor", func(string) block.Behavior { return block.VanillaBlock("nether_reactor_core") })
	p.mustRegisterBlock("nether_reactor_core", func(string) block.Behavior { return block.VanillaBlock("nether_reactor_core") })
	p.mustRegisterBlock("nether_sprouts", func(string) block.Behavior { return block.VanillaBlock("nether_sprouts") })
	p.mustRegisterBlock("nether_wart", func(string) block.Behavior { return block.VanillaBlock("nether_wart") })
	p.mustRegisterBlock("nether_wart_block", func(string) block.Behavior { return block.VanillaBlock("nether_wart_block") })
	p.mustRegisterBlock("nether_wart_plant", func(string) block.Behavior { return block.VanillaBlock("nether_wart") })
	p.mustRegisterBlock("netherite_block", func(string) block.Behavior { return block.VanillaBlock("netherite") })
	p.mustRegisterBlock("netherrack", func(string) block.Behavior { return block.VanillaBlock("netherrack") })
	p.mustRegisterBlock("netherreactor", func(string) block.Behavior { return block.VanillaBlock("nether_reactor_core") })
	p.mustRegisterBlock("normal_stone_stairs", func(string) block.Behavior { return block.VanillaBlock("stone_stairs") })
	p.mustRegisterBlock("note_block", func(string) block.Behavior { return block.VanillaBlock("note_block") })
	p.mustRegisterBlock("noteblock", func(string) block.Behavior { return block.VanillaBlock("note_block") })
	p.mustRegisterBlock("oak_button", func(string) block.Behavior { return block.VanillaBlock("oak_button") })
	p.mustRegisterBlock("oak_door", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("oak_door_block", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("oak_fence", func(string) block.Behavior { return block.VanillaBlock("oak_fence") })
	p.mustRegisterBlock("oak_fence_gate", func(string) block.Behavior { return block.VanillaBlock("oak_fence_gate") })
	p.mustRegisterBlock("oak_leaves", func(string) block.Behavior { return block.VanillaBlock("oak_leaves") })
	p.mustRegisterBlock("oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("oak_planks", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("oak_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("oak_pressure_plate") })
	p.mustRegisterBlock("oak_sapling", func(string) block.Behavior { return block.VanillaBlock("oak_sapling") })
	p.mustRegisterBlock("oak_sign", func(string) block.Behavior { return block.VanillaBlock("oak_sign") })
	p.mustRegisterBlock("oak_slab", func(string) block.Behavior { return block.VanillaBlock("oak_slab") })
	p.mustRegisterBlock("oak_stairs", func(string) block.Behavior { return block.VanillaBlock("oak_stairs") })
	p.mustRegisterBlock("oak_standing_sign", func(string) block.Behavior { return block.VanillaBlock("oak_sign") })
	p.mustRegisterBlock("oak_trapdoor", func(string) block.Behavior { return block.VanillaBlock("oak_trapdoor") })
	p.mustRegisterBlock("oak_wall_sign", func(string) block.Behavior { return block.VanillaBlock("oak_wall_sign") })
	p.mustRegisterBlock("oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("oak_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("oak_stairs") })
	p.mustRegisterBlock("oak_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("oak_stairs") })
	p.mustRegisterBlock("obsidian", func(string) block.Behavior { return block.VanillaBlock("obsidian") })
	p.mustRegisterBlock("orange_tulip", func(string) block.Behavior { return block.VanillaBlock("orange_tulip") })
	p.mustRegisterBlock("oxeye_daisy", func(string) block.Behavior { return block.VanillaBlock("oxeye_daisy") })
	p.mustRegisterBlock("packed_ice", func(string) block.Behavior { return block.VanillaBlock("packed_ice") })
	p.mustRegisterBlock("packed_mud", func(string) block.Behavior { return block.VanillaBlock("packed_mud") })
	p.mustRegisterBlock("pale_oak_button", func(string) block.Behavior { return block.VanillaBlock("pale_oak_button") })
	p.mustRegisterBlock("pale_oak_door", func(string) block.Behavior { return block.VanillaBlock("pale_oak_door") })
	p.mustRegisterBlock("pale_oak_fence", func(string) block.Behavior { return block.VanillaBlock("pale_oak_fence") })
	p.mustRegisterBlock("pale_oak_fence_gate", func(string) block.Behavior { return block.VanillaBlock("pale_oak_fence_gate") })
	p.mustRegisterBlock("pale_oak_leaves", func(string) block.Behavior { return block.VanillaBlock("pale_oak_leaves") })
	p.mustRegisterBlock("pale_oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("pale_oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("pale_oak_planks", func(string) block.Behavior { return block.VanillaBlock("pale_oak_planks") })
	p.mustRegisterBlock("pale_oak_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("pale_oak_pressure_plate") })
	p.mustRegisterBlock("pale_oak_sign", func(string) block.Behavior { return block.VanillaBlock("pale_oak_sign") })
	p.mustRegisterBlock("pale_oak_slab", func(string) block.Behavior { return block.VanillaBlock("pale_oak_slab") })
	p.mustRegisterBlock("pale_oak_stairs", func(string) block.Behavior { return block.VanillaBlock("pale_oak_stairs") })
	p.mustRegisterBlock("pale_oak_trapdoor", func(string) block.Behavior { return block.VanillaBlock("pale_oak_trapdoor") })
	p.mustRegisterBlock("pale_oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("pale_oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("peony", func(string) block.Behavior { return block.VanillaBlock("peony") })
	p.mustRegisterBlock("pink_petals", func(string) block.Behavior { return block.VanillaBlock("pink_petals") })
	p.mustRegisterBlock("pink_tulip", func(string) block.Behavior { return block.VanillaBlock("pink_tulip") })
	p.mustRegisterBlock("piglin_head", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypePiglin)
			return v
		}()
	})
	p.mustRegisterBlock("pitcher_plant", func(string) block.Behavior { return block.VanillaBlock("pitcher_plant") })
	p.mustRegisterBlock("plank", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("planks", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("player_head", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypePlayer)
			return v
		}()
	})
	p.mustRegisterBlock("podzol", func(string) block.Behavior { return block.VanillaBlock("podzol") })
	p.mustRegisterBlock("polished_andesite", func(string) block.Behavior { return block.VanillaBlock("polished_andesite") })
	p.mustRegisterBlock("polished_andesite_slab", func(string) block.Behavior { return block.VanillaBlock("polished_andesite_slab") })
	p.mustRegisterBlock("polished_andesite_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_andesite_stairs") })
	p.mustRegisterBlock("polished_basalt", func(string) block.Behavior { return block.VanillaBlock("polished_basalt") })
	p.mustRegisterBlock("polished_blackstone", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone") })
	p.mustRegisterBlock("polished_blackstone_brick_slab", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_brick_slab") })
	p.mustRegisterBlock("polished_blackstone_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_brick_stairs") })
	p.mustRegisterBlock("polished_blackstone_brick_wall", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_brick_wall") })
	p.mustRegisterBlock("polished_blackstone_bricks", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_bricks") })
	p.mustRegisterBlock("polished_blackstone_button", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_button") })
	p.mustRegisterBlock("polished_blackstone_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_pressure_plate") })
	p.mustRegisterBlock("polished_blackstone_slab", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_slab") })
	p.mustRegisterBlock("polished_blackstone_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_stairs") })
	p.mustRegisterBlock("polished_blackstone_wall", func(string) block.Behavior { return block.VanillaBlock("polished_blackstone_wall") })
	p.mustRegisterBlock("polished_deepslate", func(string) block.Behavior { return block.VanillaBlock("polished_deepslate") })
	p.mustRegisterBlock("polished_deepslate_slab", func(string) block.Behavior { return block.VanillaBlock("polished_deepslate_slab") })
	p.mustRegisterBlock("polished_deepslate_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_deepslate_stairs") })
	p.mustRegisterBlock("polished_deepslate_wall", func(string) block.Behavior { return block.VanillaBlock("polished_deepslate_wall") })
	p.mustRegisterBlock("polished_diorite", func(string) block.Behavior { return block.VanillaBlock("polished_diorite") })
	p.mustRegisterBlock("polished_diorite_slab", func(string) block.Behavior { return block.VanillaBlock("polished_diorite_slab") })
	p.mustRegisterBlock("polished_diorite_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_diorite_stairs") })
	p.mustRegisterBlock("polished_granite", func(string) block.Behavior { return block.VanillaBlock("polished_granite") })
	p.mustRegisterBlock("polished_granite_slab", func(string) block.Behavior { return block.VanillaBlock("polished_granite_slab") })
	p.mustRegisterBlock("polished_granite_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_granite_stairs") })
	p.mustRegisterBlock("polished_tuff", func(string) block.Behavior { return block.VanillaBlock("polished_tuff") })
	p.mustRegisterBlock("polished_tuff_slab", func(string) block.Behavior { return block.VanillaBlock("polished_tuff_slab") })
	p.mustRegisterBlock("polished_tuff_stairs", func(string) block.Behavior { return block.VanillaBlock("polished_tuff_stairs") })
	p.mustRegisterBlock("polished_tuff_wall", func(string) block.Behavior { return block.VanillaBlock("polished_tuff_wall") })
	p.mustRegisterBlock("poppy", func(string) block.Behavior { return block.VanillaBlock("poppy") })
	p.mustRegisterBlock("portal", func(string) block.Behavior { return block.VanillaBlock("nether_portal") })
	p.mustRegisterBlock("portal_block", func(string) block.Behavior { return block.VanillaBlock("nether_portal") })
	p.mustRegisterBlock("potato_block", func(string) block.Behavior { return block.VanillaBlock("potatoes") })
	p.mustRegisterBlock("potatoes", func(string) block.Behavior { return block.VanillaBlock("potatoes") })
	p.mustRegisterBlock("powered_comparator", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("powered_comparator_block", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("powered_rail", func(string) block.Behavior { return block.VanillaBlock("powered_rail") })
	p.mustRegisterBlock("powered_repeater", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_repeater")
			v.(interface{ SetPowered(bool) }).SetPowered(true)
			return v
		}()
	})
	p.mustRegisterBlock("powered_repeater_block", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("redstone_repeater")
			v.(interface{ SetPowered(bool) }).SetPowered(true)
			return v
		}()
	})
	p.mustRegisterBlock("prismarine", func(string) block.Behavior { return block.VanillaBlock("prismarine") })
	p.mustRegisterBlock("prismarine_bricks", func(string) block.Behavior { return block.VanillaBlock("prismarine_bricks") })
	p.mustRegisterBlock("prismarine_bricks_slab", func(string) block.Behavior { return block.VanillaBlock("prismarine_bricks_slab") })
	p.mustRegisterBlock("prismarine_bricks_stairs", func(string) block.Behavior { return block.VanillaBlock("prismarine_bricks_stairs") })
	p.mustRegisterBlock("prismarine_slab", func(string) block.Behavior { return block.VanillaBlock("prismarine_slab") })
	p.mustRegisterBlock("prismarine_stairs", func(string) block.Behavior { return block.VanillaBlock("prismarine_stairs") })
	p.mustRegisterBlock("prismarine_wall", func(string) block.Behavior { return block.VanillaBlock("prismarine_wall") })
	p.mustRegisterBlock("pumpkin", func(string) block.Behavior { return block.VanillaBlock("pumpkin") })
	p.mustRegisterBlock("pumpkin_stem", func(string) block.Behavior { return block.VanillaBlock("pumpkin_stem") })
	p.mustRegisterBlock("purple_torch", func(string) block.Behavior { return block.VanillaBlock("purple_torch") })
	p.mustRegisterBlock("purpur", func(string) block.Behavior { return block.VanillaBlock("purpur") })
	p.mustRegisterBlock("purpur_block", func(string) block.Behavior { return block.VanillaBlock("purpur") })
	p.mustRegisterBlock("purpur_pillar", func(string) block.Behavior { return block.VanillaBlock("purpur_pillar") })
	p.mustRegisterBlock("purpur_slab", func(string) block.Behavior { return block.VanillaBlock("purpur_slab") })
	p.mustRegisterBlock("purpur_stairs", func(string) block.Behavior { return block.VanillaBlock("purpur_stairs") })
	p.mustRegisterBlock("quartz_block", func(string) block.Behavior { return block.VanillaBlock("quartz") })
	p.mustRegisterBlock("quartz_bricks", func(string) block.Behavior { return block.VanillaBlock("quartz_bricks") })
	p.mustRegisterBlock("quartz_ore", func(string) block.Behavior { return block.VanillaBlock("nether_quartz_ore") })
	p.mustRegisterBlock("quartz_pillar", func(string) block.Behavior { return block.VanillaBlock("quartz_pillar") })
	p.mustRegisterBlock("quartz_slab", func(string) block.Behavior { return block.VanillaBlock("quartz_slab") })
	p.mustRegisterBlock("quartz_stairs", func(string) block.Behavior { return block.VanillaBlock("quartz_stairs") })
	p.mustRegisterBlock("rail", func(string) block.Behavior { return block.VanillaBlock("rail") })
	p.mustRegisterBlock("raw_copper_block", func(string) block.Behavior { return block.VanillaBlock("raw_copper") })
	p.mustRegisterBlock("raw_gold_block", func(string) block.Behavior { return block.VanillaBlock("raw_gold") })
	p.mustRegisterBlock("raw_iron_block", func(string) block.Behavior { return block.VanillaBlock("raw_iron") })
	p.mustRegisterBlock("red_flower", func(string) block.Behavior { return block.VanillaBlock("poppy") })
	p.mustRegisterBlock("red_mushroom", func(string) block.Behavior { return block.VanillaBlock("red_mushroom") })
	p.mustRegisterBlock("red_mushroom_block", func(string) block.Behavior { return block.VanillaBlock("red_mushroom_block") })
	p.mustRegisterBlock("red_nether_brick", func(string) block.Behavior { return block.VanillaBlock("red_nether_bricks") })
	p.mustRegisterBlock("red_nether_brick_slab", func(string) block.Behavior { return block.VanillaBlock("red_nether_brick_slab") })
	p.mustRegisterBlock("red_nether_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("red_nether_brick_stairs") })
	p.mustRegisterBlock("red_nether_brick_wall", func(string) block.Behavior { return block.VanillaBlock("red_nether_brick_wall") })
	p.mustRegisterBlock("red_nether_bricks", func(string) block.Behavior { return block.VanillaBlock("red_nether_bricks") })
	p.mustRegisterBlock("red_sand", func(string) block.Behavior { return block.VanillaBlock("red_sand") })
	p.mustRegisterBlock("red_sandstone", func(string) block.Behavior { return block.VanillaBlock("red_sandstone") })
	p.mustRegisterBlock("red_sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("red_sandstone_slab") })
	p.mustRegisterBlock("red_sandstone_stairs", func(string) block.Behavior { return block.VanillaBlock("red_sandstone_stairs") })
	p.mustRegisterBlock("red_sandstone_wall", func(string) block.Behavior { return block.VanillaBlock("red_sandstone_wall") })
	p.mustRegisterBlock("red_torch", func(string) block.Behavior { return block.VanillaBlock("red_torch") })
	p.mustRegisterBlock("red_tulip", func(string) block.Behavior { return block.VanillaBlock("red_tulip") })
	p.mustRegisterBlock("redstone_block", func(string) block.Behavior { return block.VanillaBlock("redstone") })
	p.mustRegisterBlock("redstone_comparator", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("redstone_lamp", func(string) block.Behavior { return block.VanillaBlock("redstone_lamp") })
	p.mustRegisterBlock("redstone_ore", func(string) block.Behavior { return block.VanillaBlock("redstone_ore") })
	p.mustRegisterBlock("redstone_repeater", func(string) block.Behavior { return block.VanillaBlock("redstone_repeater") })
	p.mustRegisterBlock("redstone_torch", func(string) block.Behavior { return block.VanillaBlock("redstone_torch") })
	p.mustRegisterBlock("redstone_wire", func(string) block.Behavior { return block.VanillaBlock("redstone_wire") })
	p.mustRegisterBlock("reeds", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("reeds_block", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("reinforced_deepslate", func(string) block.Behavior { return block.VanillaBlock("reinforced_deepslate") })
	p.mustRegisterBlock("repeater", func(string) block.Behavior { return block.VanillaBlock("redstone_repeater") })
	p.mustRegisterBlock("repeater_block", func(string) block.Behavior { return block.VanillaBlock("redstone_repeater") })
	p.mustRegisterBlock("reserved6", func(string) block.Behavior { return block.VanillaBlock("reserved6") })
	p.mustRegisterBlock("resin", func(string) block.Behavior { return block.VanillaBlock("resin") })
	p.mustRegisterBlock("resin_block", func(string) block.Behavior { return block.VanillaBlock("resin") })
	p.mustRegisterBlock("resin_brick_slab", func(string) block.Behavior { return block.VanillaBlock("resin_brick_slab") })
	p.mustRegisterBlock("resin_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("resin_brick_stairs") })
	p.mustRegisterBlock("resin_brick_wall", func(string) block.Behavior { return block.VanillaBlock("resin_brick_wall") })
	p.mustRegisterBlock("resin_bricks", func(string) block.Behavior { return block.VanillaBlock("resin_bricks") })
	p.mustRegisterBlock("resin_clump", func(string) block.Behavior { return block.VanillaBlock("resin_clump") })
	p.mustRegisterBlock("respawn_anchor", func(string) block.Behavior { return block.VanillaBlock("respawn_anchor") })
	p.mustRegisterBlock("rooted_dirt", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dirt")
			v.(interface{ SetDirtType(blockutils.DirtType) }).SetDirtType(blockutils.DirtTypeRooted)
			return v
		}()
	})
	p.mustRegisterBlock("rose", func(string) block.Behavior { return block.VanillaBlock("poppy") })
	p.mustRegisterBlock("rose_bush", func(string) block.Behavior { return block.VanillaBlock("rose_bush") })
	p.mustRegisterBlock("sand", func(string) block.Behavior { return block.VanillaBlock("sand") })
	p.mustRegisterBlock("sandstone", func(string) block.Behavior { return block.VanillaBlock("sandstone") })
	p.mustRegisterBlock("sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("sandstone_slab") })
	p.mustRegisterBlock("sandstone_stairs", func(string) block.Behavior { return block.VanillaBlock("sandstone_stairs") })
	p.mustRegisterBlock("sandstone_wall", func(string) block.Behavior { return block.VanillaBlock("sandstone_wall") })
	p.mustRegisterBlock("sapling", func(string) block.Behavior { return block.VanillaBlock("oak_sapling") })
	p.mustRegisterBlock("sculk", func(string) block.Behavior { return block.VanillaBlock("sculk") })
	p.mustRegisterBlock("sea_lantern", func(string) block.Behavior { return block.VanillaBlock("sea_lantern") })
	p.mustRegisterBlock("sea_pickle", func(string) block.Behavior { return block.VanillaBlock("sea_pickle") })
	p.mustRegisterBlock("sealantern", func(string) block.Behavior { return block.VanillaBlock("sea_lantern") })
	p.mustRegisterBlock("shroomlight", func(string) block.Behavior { return block.VanillaBlock("shroomlight") })
	p.mustRegisterBlock("shulker_box", func(string) block.Behavior { return block.VanillaBlock("shulker_box") })
	p.mustRegisterBlock("sign", func(string) block.Behavior { return block.VanillaBlock("oak_sign") })
	p.mustRegisterBlock("sign_post", func(string) block.Behavior { return block.VanillaBlock("oak_sign") })
	p.mustRegisterBlock("skeleton_skull", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeSkeleton)
			return v
		}()
	})
	p.mustRegisterBlock("skull", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeSkeleton)
			return v
		}()
	})
	p.mustRegisterBlock("skull_block", func(string) block.Behavior { return block.VanillaBlock("mob_head") })
	p.mustRegisterBlock("slab", func(string) block.Behavior { return block.VanillaBlock("smooth_stone_slab") })
	p.mustRegisterBlock("slabs", func(string) block.Behavior { return block.VanillaBlock("smooth_stone_slab") })
	p.mustRegisterBlock("slime", func(string) block.Behavior { return block.VanillaBlock("slime") })
	p.mustRegisterBlock("slime_block", func(string) block.Behavior { return block.VanillaBlock("slime") })
	p.mustRegisterBlock("small_amethyst_bud", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("amethyst_cluster")
			v.(interface{ SetStage(int) }).SetStage(block.AmethystClusterStageSmallBud)
			return v
		}()
	})
	p.mustRegisterBlock("small_dripleaf", func(string) block.Behavior { return block.VanillaBlock("small_dripleaf") })
	p.mustRegisterBlock("smithing_table", func(string) block.Behavior { return block.VanillaBlock("smithing_table") })
	p.mustRegisterBlock("smoker", func(string) block.Behavior { return block.VanillaBlock("smoker") })
	p.mustRegisterBlock("smooth_basalt", func(string) block.Behavior { return block.VanillaBlock("smooth_basalt") })
	p.mustRegisterBlock("smooth_quartz", func(string) block.Behavior { return block.VanillaBlock("smooth_quartz") })
	p.mustRegisterBlock("smooth_quartz_slab", func(string) block.Behavior { return block.VanillaBlock("smooth_quartz_slab") })
	p.mustRegisterBlock("smooth_quartz_stairs", func(string) block.Behavior { return block.VanillaBlock("smooth_quartz_stairs") })
	p.mustRegisterBlock("smooth_red_sandstone", func(string) block.Behavior { return block.VanillaBlock("smooth_red_sandstone") })
	p.mustRegisterBlock("smooth_red_sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("smooth_red_sandstone_slab") })
	p.mustRegisterBlock("smooth_red_sandstone_stairs", func(string) block.Behavior { return block.VanillaBlock("smooth_red_sandstone_stairs") })
	p.mustRegisterBlock("smooth_sandstone", func(string) block.Behavior { return block.VanillaBlock("smooth_sandstone") })
	p.mustRegisterBlock("smooth_sandstone_slab", func(string) block.Behavior { return block.VanillaBlock("smooth_sandstone_slab") })
	p.mustRegisterBlock("smooth_sandstone_stairs", func(string) block.Behavior { return block.VanillaBlock("smooth_sandstone_stairs") })
	p.mustRegisterBlock("smooth_stone", func(string) block.Behavior { return block.VanillaBlock("smooth_stone") })
	p.mustRegisterBlock("smooth_stone_slab", func(string) block.Behavior { return block.VanillaBlock("smooth_stone_slab") })
	p.mustRegisterBlock("snow", func(string) block.Behavior { return block.VanillaBlock("snow") })
	p.mustRegisterBlock("snow_block", func(string) block.Behavior { return block.VanillaBlock("snow") })
	p.mustRegisterBlock("snow_layer", func(string) block.Behavior { return block.VanillaBlock("snow_layer") })
	p.mustRegisterBlock("soul_campfire", func(string) block.Behavior { return block.VanillaBlock("soul_campfire") })
	p.mustRegisterBlock("soul_lantern", func(string) block.Behavior { return block.VanillaBlock("soul_lantern") })
	p.mustRegisterBlock("soul_sand", func(string) block.Behavior { return block.VanillaBlock("soul_sand") })
	p.mustRegisterBlock("soul_soil", func(string) block.Behavior { return block.VanillaBlock("soul_soil") })
	p.mustRegisterBlock("soul_torch", func(string) block.Behavior { return block.VanillaBlock("soul_torch") })
	p.mustRegisterBlock("sponge", func(string) block.Behavior { return block.VanillaBlock("sponge") })
	p.mustRegisterBlock("spore_blossom", func(string) block.Behavior { return block.VanillaBlock("spore_blossom") })
	p.mustRegisterBlock("spruce_button", func(string) block.Behavior { return block.VanillaBlock("spruce_button") })
	p.mustRegisterBlock("spruce_door", func(string) block.Behavior { return block.VanillaBlock("spruce_door") })
	p.mustRegisterBlock("spruce_door_block", func(string) block.Behavior { return block.VanillaBlock("spruce_door") })
	p.mustRegisterBlock("spruce_fence", func(string) block.Behavior { return block.VanillaBlock("spruce_fence") })
	p.mustRegisterBlock("spruce_fence_gate", func(string) block.Behavior { return block.VanillaBlock("spruce_fence_gate") })
	p.mustRegisterBlock("spruce_leaves", func(string) block.Behavior { return block.VanillaBlock("spruce_leaves") })
	p.mustRegisterBlock("spruce_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("spruce_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("spruce_planks", func(string) block.Behavior { return block.VanillaBlock("spruce_planks") })
	p.mustRegisterBlock("spruce_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("spruce_pressure_plate") })
	p.mustRegisterBlock("spruce_sapling", func(string) block.Behavior { return block.VanillaBlock("spruce_sapling") })
	p.mustRegisterBlock("spruce_sign", func(string) block.Behavior { return block.VanillaBlock("spruce_sign") })
	p.mustRegisterBlock("spruce_slab", func(string) block.Behavior { return block.VanillaBlock("spruce_slab") })
	p.mustRegisterBlock("spruce_stairs", func(string) block.Behavior { return block.VanillaBlock("spruce_stairs") })
	p.mustRegisterBlock("spruce_standing_sign", func(string) block.Behavior { return block.VanillaBlock("spruce_sign") })
	p.mustRegisterBlock("spruce_trapdoor", func(string) block.Behavior { return block.VanillaBlock("spruce_trapdoor") })
	p.mustRegisterBlock("spruce_wall_sign", func(string) block.Behavior { return block.VanillaBlock("spruce_wall_sign") })
	p.mustRegisterBlock("spruce_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("spruce_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("spruce_wood_stairs", func(string) block.Behavior { return block.VanillaBlock("spruce_stairs") })
	p.mustRegisterBlock("spruce_wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("spruce_stairs") })
	p.mustRegisterBlock("stained_clay", func(string) block.Behavior { return block.VanillaBlock("stained_clay") })
	p.mustRegisterBlock("stained_glass", func(string) block.Behavior { return block.VanillaBlock("stained_glass") })
	p.mustRegisterBlock("stained_glass_pane", func(string) block.Behavior { return block.VanillaBlock("stained_glass_pane") })
	p.mustRegisterBlock("stained_hardened_clay", func(string) block.Behavior { return block.VanillaBlock("stained_clay") })
	p.mustRegisterBlock("stained_hardened_glass", func(string) block.Behavior { return block.VanillaBlock("stained_hardened_glass") })
	p.mustRegisterBlock("stained_hardened_glass_pane", func(string) block.Behavior { return block.VanillaBlock("stained_hardened_glass_pane") })
	p.mustRegisterBlock("standing_banner", func(string) block.Behavior { return block.VanillaBlock("banner") })
	p.mustRegisterBlock("standing_sign", func(string) block.Behavior { return block.VanillaBlock("oak_sign") })
	p.mustRegisterBlock("still_lava", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("lava")
			v.(interface{ SetStill(bool) }).SetStill(true)
			return v
		}()
	})
	p.mustRegisterBlock("still_water", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("water")
			v.(interface{ SetStill(bool) }).SetStill(true)
			return v
		}()
	})
	p.mustRegisterBlock("stone", func(string) block.Behavior { return block.VanillaBlock("stone") })
	p.mustRegisterBlock("stone_brick", func(string) block.Behavior { return block.VanillaBlock("stone_bricks") })
	p.mustRegisterBlock("stone_brick_slab", func(string) block.Behavior { return block.VanillaBlock("stone_brick_slab") })
	p.mustRegisterBlock("stone_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("stone_brick_stairs") })
	p.mustRegisterBlock("stone_brick_wall", func(string) block.Behavior { return block.VanillaBlock("stone_brick_wall") })
	p.mustRegisterBlock("stone_bricks", func(string) block.Behavior { return block.VanillaBlock("stone_bricks") })
	p.mustRegisterBlock("stone_button", func(string) block.Behavior { return block.VanillaBlock("stone_button") })
	p.mustRegisterBlock("stone_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("stone_pressure_plate") })
	p.mustRegisterBlock("stone_slab", func(string) block.Behavior { return block.VanillaBlock("stone_slab") })
	p.mustRegisterBlock("stone_slab2", func(string) block.Behavior { return block.VanillaBlock("red_sandstone_slab") })
	p.mustRegisterBlock("stone_slab3", func(string) block.Behavior { return block.VanillaBlock("end_stone_brick_slab") })
	p.mustRegisterBlock("stone_slab4", func(string) block.Behavior { return block.VanillaBlock("mossy_stone_brick_slab") })
	p.mustRegisterBlock("stone_stairs", func(string) block.Behavior { return block.VanillaBlock("stone_stairs") })
	p.mustRegisterBlock("stone_wall", func(string) block.Behavior { return block.VanillaBlock("cobblestone_wall") })
	p.mustRegisterBlock("stonebrick", func(string) block.Behavior { return block.VanillaBlock("stone_bricks") })
	p.mustRegisterBlock("stonecutter", func(string) block.Behavior { return block.VanillaBlock("stonecutter") })
	p.mustRegisterBlock("stonecutter_block", func(string) block.Behavior { return block.VanillaBlock("stonecutter") })
	p.mustRegisterBlock("stripped_acacia_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_acacia_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_bamboo_block", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("bamboo_block")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_birch_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("birch_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_birch_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("birch_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_cherry_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("cherry_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_cherry_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("cherry_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_crimson_hyphae", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("crimson_hyphae")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_crimson_stem", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("crimson_stem")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_dark_oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dark_oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_dark_oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("dark_oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_jungle_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("jungle_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_jungle_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("jungle_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_mangrove_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mangrove_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_mangrove_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mangrove_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_pale_oak_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("pale_oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_pale_oak_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("pale_oak_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_spruce_log", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("spruce_log")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_spruce_wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("spruce_wood")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_warped_hyphae", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("warped_hyphae")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("stripped_warped_stem", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("warped_stem")
			v.(interface{ SetStripped(bool) }).SetStripped(true)
			return v
		}()
	})
	p.mustRegisterBlock("structure_void", func(string) block.Behavior { return block.VanillaBlock("structure_void") })
	p.mustRegisterBlock("sugar_cane", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("sugar_canes", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("sugarcane", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("sugarcane_block", func(string) block.Behavior { return block.VanillaBlock("sugarcane") })
	p.mustRegisterBlock("sunflower", func(string) block.Behavior { return block.VanillaBlock("sunflower") })
	p.mustRegisterBlock("sweet_berry_bush", func(string) block.Behavior { return block.VanillaBlock("sweet_berry_bush") })
	p.mustRegisterBlock("tall_grass", func(string) block.Behavior { return block.VanillaBlock("tall_grass") })
	p.mustRegisterBlock("tallgrass", func(string) block.Behavior { return block.VanillaBlock("fern") })
	p.mustRegisterBlock("terracotta", func(string) block.Behavior { return block.VanillaBlock("stained_clay") })
	p.mustRegisterBlock("tinted_glass", func(string) block.Behavior { return block.VanillaBlock("tinted_glass") })
	p.mustRegisterBlock("tnt", func(string) block.Behavior { return block.VanillaBlock("tnt") })
	p.mustRegisterBlock("torch", func(string) block.Behavior { return block.VanillaBlock("torch") })
	p.mustRegisterBlock("torchflower", func(string) block.Behavior { return block.VanillaBlock("torchflower") })
	p.mustRegisterBlock("trapdoor", func(string) block.Behavior { return block.VanillaBlock("oak_trapdoor") })
	p.mustRegisterBlock("trapped_chest", func(string) block.Behavior { return block.VanillaBlock("trapped_chest") })
	p.mustRegisterBlock("trip_wire", func(string) block.Behavior { return block.VanillaBlock("tripwire") })
	p.mustRegisterBlock("tripwire", func(string) block.Behavior { return block.VanillaBlock("tripwire") })
	p.mustRegisterBlock("tripwire_hook", func(string) block.Behavior { return block.VanillaBlock("tripwire_hook") })
	p.mustRegisterBlock("trunk", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("trunk2", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("tuff", func(string) block.Behavior { return block.VanillaBlock("tuff") })
	p.mustRegisterBlock("tuff_bricks", func(string) block.Behavior { return block.VanillaBlock("tuff_bricks") })
	p.mustRegisterBlock("tuff_brick_slab", func(string) block.Behavior { return block.VanillaBlock("tuff_brick_slab") })
	p.mustRegisterBlock("tuff_brick_stairs", func(string) block.Behavior { return block.VanillaBlock("tuff_brick_stairs") })
	p.mustRegisterBlock("tuff_brick_wall", func(string) block.Behavior { return block.VanillaBlock("tuff_brick_wall") })
	p.mustRegisterBlock("tuff_slab", func(string) block.Behavior { return block.VanillaBlock("tuff_slab") })
	p.mustRegisterBlock("tuff_stairs", func(string) block.Behavior { return block.VanillaBlock("tuff_stairs") })
	p.mustRegisterBlock("tuff_wall", func(string) block.Behavior { return block.VanillaBlock("tuff_wall") })
	p.mustRegisterBlock("twisting_vines", func(string) block.Behavior { return block.VanillaBlock("twisting_vines") })
	p.mustRegisterBlock("underwater_tnt", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("tnt")
			v.(interface{ SetWorksUnderwater(bool) }).SetWorksUnderwater(true)
			return v
		}()
	})
	p.mustRegisterBlock("underwater_torch", func(string) block.Behavior { return block.VanillaBlock("underwater_torch") })
	p.mustRegisterBlock("undyed_shulker_box", func(string) block.Behavior { return block.VanillaBlock("shulker_box") })
	p.mustRegisterBlock("unlit_redstone_torch", func(string) block.Behavior { return block.VanillaBlock("redstone_torch") })
	p.mustRegisterBlock("unpowered_comparator", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("unpowered_comparator_block", func(string) block.Behavior { return block.VanillaBlock("redstone_comparator") })
	p.mustRegisterBlock("unpowered_repeater", func(string) block.Behavior { return block.VanillaBlock("redstone_repeater") })
	p.mustRegisterBlock("unpowered_repeater_block", func(string) block.Behavior { return block.VanillaBlock("redstone_repeater") })
	p.mustRegisterBlock("update_block", func(string) block.Behavior { return block.VanillaBlock("info_update") })
	p.mustRegisterBlock("vine", func(string) block.Behavior { return block.VanillaBlock("vines") })
	p.mustRegisterBlock("vines", func(string) block.Behavior { return block.VanillaBlock("vines") })
	p.mustRegisterBlock("wall_banner", func(string) block.Behavior { return block.VanillaBlock("wall_banner") })
	p.mustRegisterBlock("wall_coral_fan", func(string) block.Behavior { return block.VanillaBlock("wall_coral_fan") })
	p.mustRegisterBlock("wall_sign", func(string) block.Behavior { return block.VanillaBlock("oak_wall_sign") })
	p.mustRegisterBlock("warped_button", func(string) block.Behavior { return block.VanillaBlock("warped_button") })
	p.mustRegisterBlock("warped_door", func(string) block.Behavior { return block.VanillaBlock("warped_door") })
	p.mustRegisterBlock("warped_fence", func(string) block.Behavior { return block.VanillaBlock("warped_fence") })
	p.mustRegisterBlock("warped_fence_gate", func(string) block.Behavior { return block.VanillaBlock("warped_fence_gate") })
	p.mustRegisterBlock("warped_fungus", func(string) block.Behavior { return block.VanillaBlock("warped_fungus") })
	p.mustRegisterBlock("warped_hyphae", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("warped_hyphae")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("warped_nylium", func(string) block.Behavior { return block.VanillaBlock("warped_nylium") })
	p.mustRegisterBlock("warped_planks", func(string) block.Behavior { return block.VanillaBlock("warped_planks") })
	p.mustRegisterBlock("warped_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("warped_pressure_plate") })
	p.mustRegisterBlock("warped_roots", func(string) block.Behavior { return block.VanillaBlock("warped_roots") })
	p.mustRegisterBlock("warped_sign", func(string) block.Behavior { return block.VanillaBlock("warped_sign") })
	p.mustRegisterBlock("warped_slab", func(string) block.Behavior { return block.VanillaBlock("warped_slab") })
	p.mustRegisterBlock("warped_stairs", func(string) block.Behavior { return block.VanillaBlock("warped_stairs") })
	p.mustRegisterBlock("warped_stem", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("warped_stem")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("warped_trapdoor", func(string) block.Behavior { return block.VanillaBlock("warped_trapdoor") })
	p.mustRegisterBlock("warped_wart_block", func(string) block.Behavior { return block.VanillaBlock("warped_wart_block") })
	p.mustRegisterBlock("water", func(string) block.Behavior { return block.VanillaBlock("water") })
	p.mustRegisterBlock("water_lily", func(string) block.Behavior { return block.VanillaBlock("lily_pad") })
	p.mustRegisterBlock("waterlily", func(string) block.Behavior { return block.VanillaBlock("lily_pad") })
	p.mustRegisterBlock("web", func(string) block.Behavior { return block.VanillaBlock("cobweb") })
	p.mustRegisterBlock("weeping_vines", func(string) block.Behavior { return block.VanillaBlock("weeping_vines") })
	p.mustRegisterBlock("weighted_pressure_plate_heavy", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_heavy") })
	p.mustRegisterBlock("weighted_pressure_plate_light", func(string) block.Behavior { return block.VanillaBlock("weighted_pressure_plate_light") })
	p.mustRegisterBlock("wheat_block", func(string) block.Behavior { return block.VanillaBlock("wheat") })
	p.mustRegisterBlock("white_tulip", func(string) block.Behavior { return block.VanillaBlock("white_tulip") })
	p.mustRegisterBlock("wither_rose", func(string) block.Behavior { return block.VanillaBlock("wither_rose") })
	p.mustRegisterBlock("wither_skeleton_skull", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeWitherSkeleton)
			return v
		}()
	})
	p.mustRegisterBlock("wood", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("oak_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("wood2", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("acacia_log")
			v.(interface{ SetStripped(bool) }).SetStripped(false)
			return v
		}()
	})
	p.mustRegisterBlock("wood_door_block", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("wood_slab", func(string) block.Behavior { return block.VanillaBlock("oak_slab") })
	p.mustRegisterBlock("wood_slabs", func(string) block.Behavior { return block.VanillaBlock("oak_slab") })
	p.mustRegisterBlock("wood_stairs", func(string) block.Behavior { return block.VanillaBlock("oak_stairs") })
	p.mustRegisterBlock("wooden_button", func(string) block.Behavior { return block.VanillaBlock("oak_button") })
	p.mustRegisterBlock("wooden_door", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("wooden_door_block", func(string) block.Behavior { return block.VanillaBlock("oak_door") })
	p.mustRegisterBlock("wooden_plank", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("wooden_planks", func(string) block.Behavior { return block.VanillaBlock("oak_planks") })
	p.mustRegisterBlock("wooden_pressure_plate", func(string) block.Behavior { return block.VanillaBlock("oak_pressure_plate") })
	p.mustRegisterBlock("wooden_slab", func(string) block.Behavior { return block.VanillaBlock("oak_slab") })
	p.mustRegisterBlock("wooden_slabs", func(string) block.Behavior { return block.VanillaBlock("oak_slab") })
	p.mustRegisterBlock("wooden_stairs", func(string) block.Behavior { return block.VanillaBlock("oak_stairs") })
	p.mustRegisterBlock("wooden_trapdoor", func(string) block.Behavior { return block.VanillaBlock("oak_trapdoor") })
	p.mustRegisterBlock("wool", func(string) block.Behavior { return block.VanillaBlock("wool") })
	p.mustRegisterBlock("workbench", func(string) block.Behavior { return block.VanillaBlock("crafting_table") })
	p.mustRegisterBlock("yellow_flower", func(string) block.Behavior { return block.VanillaBlock("dandelion") })
	p.mustRegisterBlock("zombie_head", func(string) block.Behavior {
		return func() block.Behavior {
			v := block.VanillaBlock("mob_head")
			v.(interface{ SetMobHeadType(blockutils.MobHeadType) }).SetMobHeadType(blockutils.MobHeadTypeZombie)
			return v
		}()
	})
}

func registerStringToItemParserItems(p *StringToItemParser) {
	p.mustRegister("acacia_boat", func(string) Item { return VanillaItem("acacia_boat") })
	p.mustRegister("acacia_hanging_sign", func(string) Item { return VanillaItem("acacia_hanging_sign") })
	p.mustRegister("amethyst_shard", func(string) Item { return VanillaItem("amethyst_shard") })
	p.mustRegister("antidote", func(string) Item {
		return func() Item {
			v := VanillaItem("medicine")
			v.(interface{ SetType(MedicineType) }).SetType(MedicineTypeAntidote)
			return v
		}()
	})
	p.mustRegister("apple", func(string) Item { return VanillaItem("apple") })
	p.mustRegister("apple_enchanted", func(string) Item { return VanillaItem("enchanted_golden_apple") })
	p.mustRegister("appleenchanted", func(string) Item { return VanillaItem("enchanted_golden_apple") })
	p.mustRegister("arrow", func(string) Item { return VanillaItem("arrow") })
	p.mustRegister("baked_potato", func(string) Item { return VanillaItem("baked_potato") })
	p.mustRegister("baked_potatoes", func(string) Item { return VanillaItem("baked_potato") })
	p.mustRegister("bamboo_hanging_sign", func(string) Item { return VanillaItem("bamboo_hanging_sign") })
	p.mustRegister("beef", func(string) Item { return VanillaItem("raw_beef") })
	p.mustRegister("beetroot", func(string) Item { return VanillaItem("beetroot") })
	p.mustRegister("beetroot_seed", func(string) Item { return VanillaItem("beetroot_seeds") })
	p.mustRegister("beetroot_seeds", func(string) Item { return VanillaItem("beetroot_seeds") })
	p.mustRegister("beetroot_soup", func(string) Item { return VanillaItem("beetroot_soup") })
	p.mustRegister("birch_boat", func(string) Item { return VanillaItem("birch_boat") })
	p.mustRegister("birch_hanging_sign", func(string) Item { return VanillaItem("birch_hanging_sign") })
	p.mustRegister("blaze_powder", func(string) Item { return VanillaItem("blaze_powder") })
	p.mustRegister("blaze_rod", func(string) Item { return VanillaItem("blaze_rod") })
	p.mustRegister("bleach", func(string) Item { return VanillaItem("bleach") })
	p.mustRegister("boat", func(string) Item { return VanillaItem("oak_boat") })
	p.mustRegister("bone", func(string) Item { return VanillaItem("bone") })
	p.mustRegister("bone_meal", func(string) Item { return VanillaItem("bone_meal") })
	p.mustRegister("book", func(string) Item { return VanillaItem("book") })
	p.mustRegister("bottle_o_enchanting", func(string) Item { return VanillaItem("experience_bottle") })
	p.mustRegister("bow", func(string) Item { return VanillaItem("bow") })
	p.mustRegister("bowl", func(string) Item { return VanillaItem("bowl") })
	p.mustRegister("bread", func(string) Item { return VanillaItem("bread") })
	p.mustRegister("brick", func(string) Item { return VanillaItem("brick") })
	p.mustRegister("bucket", func(string) Item { return VanillaItem("bucket") })
	p.mustRegister("carrot", func(string) Item { return VanillaItem("carrot") })
	p.mustRegister("chain_boots", func(string) Item { return VanillaItem("chainmail_boots") })
	p.mustRegister("chain_chestplate", func(string) Item { return VanillaItem("chainmail_chestplate") })
	p.mustRegister("chain_helmet", func(string) Item { return VanillaItem("chainmail_helmet") })
	p.mustRegister("chain_leggings", func(string) Item { return VanillaItem("chainmail_leggings") })
	p.mustRegister("chainmail_boots", func(string) Item { return VanillaItem("chainmail_boots") })
	p.mustRegister("chainmail_chestplate", func(string) Item { return VanillaItem("chainmail_chestplate") })
	p.mustRegister("chainmail_helmet", func(string) Item { return VanillaItem("chainmail_helmet") })
	p.mustRegister("chainmail_leggings", func(string) Item { return VanillaItem("chainmail_leggings") })
	p.mustRegister("charcoal", func(string) Item { return VanillaItem("charcoal") })
	p.mustRegister("chemical_aluminium_oxide", func(string) Item { return VanillaItem("chemical_aluminium_oxide") })
	p.mustRegister("chemical_ammonia", func(string) Item { return VanillaItem("chemical_ammonia") })
	p.mustRegister("chemical_barium_sulphate", func(string) Item { return VanillaItem("chemical_barium_sulphate") })
	p.mustRegister("chemical_benzene", func(string) Item { return VanillaItem("chemical_benzene") })
	p.mustRegister("chemical_boron_trioxide", func(string) Item { return VanillaItem("chemical_boron_trioxide") })
	p.mustRegister("chemical_calcium_bromide", func(string) Item { return VanillaItem("chemical_calcium_bromide") })
	p.mustRegister("chemical_calcium_chloride", func(string) Item { return VanillaItem("chemical_calcium_chloride") })
	p.mustRegister("chemical_cerium_chloride", func(string) Item { return VanillaItem("chemical_cerium_chloride") })
	p.mustRegister("chemical_charcoal", func(string) Item { return VanillaItem("chemical_charcoal") })
	p.mustRegister("chemical_crude_oil", func(string) Item { return VanillaItem("chemical_crude_oil") })
	p.mustRegister("chemical_glue", func(string) Item { return VanillaItem("chemical_glue") })
	p.mustRegister("chemical_hydrogen_peroxide", func(string) Item { return VanillaItem("chemical_hydrogen_peroxide") })
	p.mustRegister("chemical_hypochlorite", func(string) Item { return VanillaItem("chemical_hypochlorite") })
	p.mustRegister("chemical_ink", func(string) Item { return VanillaItem("chemical_ink") })
	p.mustRegister("chemical_iron_sulphide", func(string) Item { return VanillaItem("chemical_iron_sulphide") })
	p.mustRegister("chemical_latex", func(string) Item { return VanillaItem("chemical_latex") })
	p.mustRegister("chemical_lithium_hydride", func(string) Item { return VanillaItem("chemical_lithium_hydride") })
	p.mustRegister("chemical_luminol", func(string) Item { return VanillaItem("chemical_luminol") })
	p.mustRegister("chemical_magnesium_nitrate", func(string) Item { return VanillaItem("chemical_magnesium_nitrate") })
	p.mustRegister("chemical_magnesium_oxide", func(string) Item { return VanillaItem("chemical_magnesium_oxide") })
	p.mustRegister("chemical_magnesium_salts", func(string) Item { return VanillaItem("chemical_magnesium_salts") })
	p.mustRegister("chemical_mercuric_chloride", func(string) Item { return VanillaItem("chemical_mercuric_chloride") })
	p.mustRegister("chemical_polyethylene", func(string) Item { return VanillaItem("chemical_polyethylene") })
	p.mustRegister("chemical_potassium_chloride", func(string) Item { return VanillaItem("chemical_potassium_chloride") })
	p.mustRegister("chemical_potassium_iodide", func(string) Item { return VanillaItem("chemical_potassium_iodide") })
	p.mustRegister("chemical_rubbish", func(string) Item { return VanillaItem("chemical_rubbish") })
	p.mustRegister("chemical_salt", func(string) Item { return VanillaItem("chemical_salt") })
	p.mustRegister("chemical_soap", func(string) Item { return VanillaItem("chemical_soap") })
	p.mustRegister("chemical_sodium_acetate", func(string) Item { return VanillaItem("chemical_sodium_acetate") })
	p.mustRegister("chemical_sodium_fluoride", func(string) Item { return VanillaItem("chemical_sodium_fluoride") })
	p.mustRegister("chemical_sodium_hydride", func(string) Item { return VanillaItem("chemical_sodium_hydride") })
	p.mustRegister("chemical_sodium_hydroxide", func(string) Item { return VanillaItem("chemical_sodium_hydroxide") })
	p.mustRegister("chemical_sodium_hypochlorite", func(string) Item { return VanillaItem("chemical_sodium_hypochlorite") })
	p.mustRegister("chemical_sodium_oxide", func(string) Item { return VanillaItem("chemical_sodium_oxide") })
	p.mustRegister("chemical_sugar", func(string) Item { return VanillaItem("chemical_sugar") })
	p.mustRegister("chemical_sulphate", func(string) Item { return VanillaItem("chemical_sulphate") })
	p.mustRegister("chemical_tungsten_chloride", func(string) Item { return VanillaItem("chemical_tungsten_chloride") })
	p.mustRegister("chemical_water", func(string) Item { return VanillaItem("chemical_water") })
	p.mustRegister("cherry_hanging_sign", func(string) Item { return VanillaItem("cherry_hanging_sign") })
	p.mustRegister("chicken", func(string) Item { return VanillaItem("raw_chicken") })
	p.mustRegister("chorus_fruit", func(string) Item { return VanillaItem("chorus_fruit") })
	p.mustRegister("chorus_fruit_popped", func(string) Item { return VanillaItem("popped_chorus_fruit") })
	p.mustRegister("clay", func(string) Item { return VanillaItem("clay") })
	p.mustRegister("clay_ball", func(string) Item { return VanillaItem("clay") })
	p.mustRegister("clock", func(string) Item { return VanillaItem("clock") })
	p.mustRegister("clown_fish", func(string) Item { return VanillaItem("clownfish") })
	p.mustRegister("clownfish", func(string) Item { return VanillaItem("clownfish") })
	p.mustRegister("coal", func(string) Item { return VanillaItem("coal") })
	p.mustRegister("coast_armor_trim_smithing_template", func(string) Item { return VanillaItem("coast_armor_trim_smithing_template") })
	p.mustRegister("cocoa_beans", func(string) Item { return VanillaItem("cocoa_beans") })
	p.mustRegister("cod", func(string) Item { return VanillaItem("raw_fish") })
	p.mustRegister("compass", func(string) Item { return VanillaItem("compass") })
	p.mustRegister("compound", func(string) Item { return VanillaItem("chemical_salt") })
	p.mustRegister("cooked_beef", func(string) Item { return VanillaItem("steak") })
	p.mustRegister("cooked_chicken", func(string) Item { return VanillaItem("cooked_chicken") })
	p.mustRegister("cooked_cod", func(string) Item { return VanillaItem("cooked_fish") })
	p.mustRegister("cooked_fish", func(string) Item { return VanillaItem("cooked_fish") })
	p.mustRegister("cooked_mutton", func(string) Item { return VanillaItem("cooked_mutton") })
	p.mustRegister("cooked_porkchop", func(string) Item { return VanillaItem("cooked_porkchop") })
	p.mustRegister("cooked_rabbit", func(string) Item { return VanillaItem("cooked_rabbit") })
	p.mustRegister("cooked_salmon", func(string) Item { return VanillaItem("cooked_salmon") })
	p.mustRegister("cookie", func(string) Item { return VanillaItem("cookie") })
	p.mustRegister("copper_axe", func(string) Item { return VanillaItem("copper_axe") })
	p.mustRegister("copper_boots", func(string) Item { return VanillaItem("copper_boots") })
	p.mustRegister("copper_chestplate", func(string) Item { return VanillaItem("copper_chestplate") })
	p.mustRegister("copper_helmet", func(string) Item { return VanillaItem("copper_helmet") })
	p.mustRegister("copper_hoe", func(string) Item { return VanillaItem("copper_hoe") })
	p.mustRegister("copper_ingot", func(string) Item { return VanillaItem("copper_ingot") })
	p.mustRegister("copper_leggings", func(string) Item { return VanillaItem("copper_leggings") })
	p.mustRegister("copper_nugget", func(string) Item { return VanillaItem("copper_nugget") })
	p.mustRegister("copper_pickaxe", func(string) Item { return VanillaItem("copper_pickaxe") })
	p.mustRegister("copper_shovel", func(string) Item { return VanillaItem("copper_shovel") })
	p.mustRegister("copper_sword", func(string) Item { return VanillaItem("copper_sword") })
	p.mustRegister("crimson_hanging_sign", func(string) Item { return VanillaItem("crimson_hanging_sign") })
	p.mustRegister("dark_oak_boat", func(string) Item { return VanillaItem("dark_oak_boat") })
	p.mustRegister("dark_oak_hanging_sign", func(string) Item { return VanillaItem("dark_oak_hanging_sign") })
	p.mustRegister("diamond", func(string) Item { return VanillaItem("diamond") })
	p.mustRegister("diamond_axe", func(string) Item { return VanillaItem("diamond_axe") })
	p.mustRegister("diamond_boots", func(string) Item { return VanillaItem("diamond_boots") })
	p.mustRegister("diamond_chestplate", func(string) Item { return VanillaItem("diamond_chestplate") })
	p.mustRegister("diamond_helmet", func(string) Item { return VanillaItem("diamond_helmet") })
	p.mustRegister("diamond_hoe", func(string) Item { return VanillaItem("diamond_hoe") })
	p.mustRegister("diamond_leggings", func(string) Item { return VanillaItem("diamond_leggings") })
	p.mustRegister("diamond_pickaxe", func(string) Item { return VanillaItem("diamond_pickaxe") })
	p.mustRegister("diamond_shovel", func(string) Item { return VanillaItem("diamond_shovel") })
	p.mustRegister("diamond_sword", func(string) Item { return VanillaItem("diamond_sword") })
	p.mustRegister("disc_fragment_5", func(string) Item { return VanillaItem("disc_fragment_5") })
	p.mustRegister("dragon_breath", func(string) Item { return VanillaItem("dragon_breath") })
	p.mustRegister("dried_kelp", func(string) Item { return VanillaItem("dried_kelp") })
	p.mustRegister("dune_armor_trim_smithing_template", func(string) Item { return VanillaItem("dune_armor_trim_smithing_template") })
	p.mustRegister("dye", func(string) Item { return VanillaItem("ink_sac") })
	p.mustRegister("echo_shard", func(string) Item { return VanillaItem("echo_shard") })
	p.mustRegister("egg", func(string) Item { return VanillaItem("egg") })
	p.mustRegister("elixir", func(string) Item {
		return func() Item {
			v := VanillaItem("medicine")
			v.(interface{ SetType(MedicineType) }).SetType(MedicineTypeElixir)
			return v
		}()
	})
	p.mustRegister("emerald", func(string) Item { return VanillaItem("emerald") })
	p.mustRegister("enchanted_book", func(string) Item { return VanillaItem("enchanted_book") })
	p.mustRegister("enchanted_golden_apple", func(string) Item { return VanillaItem("enchanted_golden_apple") })
	p.mustRegister("enchanting_bottle", func(string) Item { return VanillaItem("experience_bottle") })
	p.mustRegister("end_crystal", func(string) Item { return VanillaItem("end_crystal") })
	p.mustRegister("ender_pearl", func(string) Item { return VanillaItem("ender_pearl") })
	p.mustRegister("experience_bottle", func(string) Item { return VanillaItem("experience_bottle") })
	p.mustRegister("eye_armor_trim_smithing_template", func(string) Item { return VanillaItem("eye_armor_trim_smithing_template") })
	p.mustRegister("eye_drops", func(string) Item {
		return func() Item {
			v := VanillaItem("medicine")
			v.(interface{ SetType(MedicineType) }).SetType(MedicineTypeEyeDrops)
			return v
		}()
	})
	p.mustRegister("feather", func(string) Item { return VanillaItem("feather") })
	p.mustRegister("fermented_spider_eye", func(string) Item { return VanillaItem("fermented_spider_eye") })
	p.mustRegister("firework_rocket", func(string) Item { return VanillaItem("firework_rocket") })
	p.mustRegister("firework_star", func(string) Item { return VanillaItem("firework_star") })
	p.mustRegister("fireworks", func(string) Item { return VanillaItem("firework_rocket") })
	p.mustRegister("fire_charge", func(string) Item { return VanillaItem("fire_charge") })
	p.mustRegister("fish", func(string) Item { return VanillaItem("raw_fish") })
	p.mustRegister("fishing_rod", func(string) Item { return VanillaItem("fishing_rod") })
	p.mustRegister("flint", func(string) Item { return VanillaItem("flint") })
	p.mustRegister("flint_and_steel", func(string) Item { return VanillaItem("flint_and_steel") })
	p.mustRegister("flint_steel", func(string) Item { return VanillaItem("flint_and_steel") })
	p.mustRegister("ghast_tear", func(string) Item { return VanillaItem("ghast_tear") })
	p.mustRegister("glass_bottle", func(string) Item { return VanillaItem("glass_bottle") })
	p.mustRegister("glistering_melon", func(string) Item { return VanillaItem("glistering_melon") })
	p.mustRegister("glow_berries", func(string) Item { return VanillaItem("glow_berries") })
	p.mustRegister("glow_ink_sac", func(string) Item { return VanillaItem("glow_ink_sac") })
	p.mustRegister("glowstone_dust", func(string) Item { return VanillaItem("glowstone_dust") })
	p.mustRegister("goat_horn", func(string) Item { return VanillaItem("goat_horn") })
	p.mustRegister("gold_axe", func(string) Item { return VanillaItem("golden_axe") })
	p.mustRegister("gold_boots", func(string) Item { return VanillaItem("golden_boots") })
	p.mustRegister("gold_chestplate", func(string) Item { return VanillaItem("golden_chestplate") })
	p.mustRegister("gold_helmet", func(string) Item { return VanillaItem("golden_helmet") })
	p.mustRegister("gold_hoe", func(string) Item { return VanillaItem("golden_hoe") })
	p.mustRegister("gold_ingot", func(string) Item { return VanillaItem("gold_ingot") })
	p.mustRegister("gold_leggings", func(string) Item { return VanillaItem("golden_leggings") })
	p.mustRegister("gold_nugget", func(string) Item { return VanillaItem("gold_nugget") })
	p.mustRegister("gold_pickaxe", func(string) Item { return VanillaItem("golden_pickaxe") })
	p.mustRegister("gold_shovel", func(string) Item { return VanillaItem("golden_shovel") })
	p.mustRegister("gold_sword", func(string) Item { return VanillaItem("golden_sword") })
	p.mustRegister("golden_apple", func(string) Item { return VanillaItem("golden_apple") })
	p.mustRegister("golden_axe", func(string) Item { return VanillaItem("golden_axe") })
	p.mustRegister("golden_boots", func(string) Item { return VanillaItem("golden_boots") })
	p.mustRegister("golden_carrot", func(string) Item { return VanillaItem("golden_carrot") })
	p.mustRegister("golden_chestplate", func(string) Item { return VanillaItem("golden_chestplate") })
	p.mustRegister("golden_helmet", func(string) Item { return VanillaItem("golden_helmet") })
	p.mustRegister("golden_hoe", func(string) Item { return VanillaItem("golden_hoe") })
	p.mustRegister("golden_leggings", func(string) Item { return VanillaItem("golden_leggings") })
	p.mustRegister("golden_nugget", func(string) Item { return VanillaItem("gold_nugget") })
	p.mustRegister("golden_pickaxe", func(string) Item { return VanillaItem("golden_pickaxe") })
	p.mustRegister("golden_shovel", func(string) Item { return VanillaItem("golden_shovel") })
	p.mustRegister("golden_sword", func(string) Item { return VanillaItem("golden_sword") })
	p.mustRegister("gunpowder", func(string) Item { return VanillaItem("gunpowder") })
	p.mustRegister("heart_of_the_sea", func(string) Item { return VanillaItem("heart_of_the_sea") })
	p.mustRegister("honey_bottle", func(string) Item { return VanillaItem("honey_bottle") })
	p.mustRegister("host_armor_trim_smithing_template", func(string) Item { return VanillaItem("host_armor_trim_smithing_template") })
	p.mustRegister("honeycomb", func(string) Item { return VanillaItem("honeycomb") })
	p.mustRegister("ice_bomb", func(string) Item { return VanillaItem("ice_bomb") })
	p.mustRegister("ink_sac", func(string) Item { return VanillaItem("ink_sac") })
	p.mustRegister("iron_axe", func(string) Item { return VanillaItem("iron_axe") })
	p.mustRegister("iron_boots", func(string) Item { return VanillaItem("iron_boots") })
	p.mustRegister("iron_chestplate", func(string) Item { return VanillaItem("iron_chestplate") })
	p.mustRegister("iron_helmet", func(string) Item { return VanillaItem("iron_helmet") })
	p.mustRegister("iron_hoe", func(string) Item { return VanillaItem("iron_hoe") })
	p.mustRegister("iron_ingot", func(string) Item { return VanillaItem("iron_ingot") })
	p.mustRegister("iron_leggings", func(string) Item { return VanillaItem("iron_leggings") })
	p.mustRegister("iron_nugget", func(string) Item { return VanillaItem("iron_nugget") })
	p.mustRegister("iron_pickaxe", func(string) Item { return VanillaItem("iron_pickaxe") })
	p.mustRegister("iron_shovel", func(string) Item { return VanillaItem("iron_shovel") })
	p.mustRegister("iron_sword", func(string) Item { return VanillaItem("iron_sword") })
	p.mustRegister("jungle_boat", func(string) Item { return VanillaItem("jungle_boat") })
	p.mustRegister("jungle_hanging_sign", func(string) Item { return VanillaItem("jungle_hanging_sign") })
	p.mustRegister("lapis_lazuli", func(string) Item { return VanillaItem("lapis_lazuli") })
	p.mustRegister("lava_bucket", func(string) Item { return VanillaItem("lava_bucket") })
	p.mustRegister("leather", func(string) Item { return VanillaItem("leather") })
	p.mustRegister("leather_boots", func(string) Item { return VanillaItem("leather_boots") })
	p.mustRegister("leather_cap", func(string) Item { return VanillaItem("leather_cap") })
	p.mustRegister("leather_chestplate", func(string) Item { return VanillaItem("leather_tunic") })
	p.mustRegister("leather_helmet", func(string) Item { return VanillaItem("leather_cap") })
	p.mustRegister("leather_leggings", func(string) Item { return VanillaItem("leather_pants") })
	p.mustRegister("leather_pants", func(string) Item { return VanillaItem("leather_pants") })
	p.mustRegister("leather_tunic", func(string) Item { return VanillaItem("leather_tunic") })
	p.mustRegister("magma_cream", func(string) Item { return VanillaItem("magma_cream") })
	p.mustRegister("mangrove_hanging_sign", func(string) Item { return VanillaItem("mangrove_hanging_sign") })
	p.mustRegister("melon", func(string) Item { return VanillaItem("melon") })
	p.mustRegister("melon_seeds", func(string) Item { return VanillaItem("melon_seeds") })
	p.mustRegister("melon_slice", func(string) Item { return VanillaItem("melon") })
	p.mustRegister("milk_bucket", func(string) Item { return VanillaItem("milk_bucket") })
	p.mustRegister("minecart", func(string) Item { return VanillaItem("minecart") })
	p.mustRegister("mushroom_stew", func(string) Item { return VanillaItem("mushroom_stew") })
	p.mustRegister("mutton", func(string) Item { return VanillaItem("raw_mutton") })
	p.mustRegister("mutton_cooked", func(string) Item { return VanillaItem("cooked_mutton") })
	p.mustRegister("mutton_raw", func(string) Item { return VanillaItem("raw_mutton") })
	p.mustRegister("muttoncooked", func(string) Item { return VanillaItem("cooked_mutton") })
	p.mustRegister("muttonraw", func(string) Item { return VanillaItem("raw_mutton") })
	p.mustRegister("name_tag", func(string) Item { return VanillaItem("name_tag") })
	p.mustRegister("nautilus_shell", func(string) Item { return VanillaItem("nautilus_shell") })
	p.mustRegister("nether_brick", func(string) Item { return VanillaItem("nether_brick") })
	p.mustRegister("nether_quartz", func(string) Item { return VanillaItem("nether_quartz") })
	p.mustRegister("nether_star", func(string) Item { return VanillaItem("nether_star") })
	p.mustRegister("netherbrick", func(string) Item { return VanillaItem("nether_brick") })
	p.mustRegister("netherite_axe", func(string) Item { return VanillaItem("netherite_axe") })
	p.mustRegister("netherite_boots", func(string) Item { return VanillaItem("netherite_boots") })
	p.mustRegister("netherite_chestplate", func(string) Item { return VanillaItem("netherite_chestplate") })
	p.mustRegister("netherite_helmet", func(string) Item { return VanillaItem("netherite_helmet") })
	p.mustRegister("netherite_hoe", func(string) Item { return VanillaItem("netherite_hoe") })
	p.mustRegister("netherite_ingot", func(string) Item { return VanillaItem("netherite_ingot") })
	p.mustRegister("netherite_leggings", func(string) Item { return VanillaItem("netherite_leggings") })
	p.mustRegister("netherite_pickaxe", func(string) Item { return VanillaItem("netherite_pickaxe") })
	p.mustRegister("netherite_scrap", func(string) Item { return VanillaItem("netherite_scrap") })
	p.mustRegister("netherite_shovel", func(string) Item { return VanillaItem("netherite_shovel") })
	p.mustRegister("netherite_sword", func(string) Item { return VanillaItem("netherite_sword") })
	p.mustRegister("netherstar", func(string) Item { return VanillaItem("nether_star") })
	p.mustRegister("netherite_upgrade_smithing_template", func(string) Item { return VanillaItem("netherite_upgrade_smithing_template") })
	p.mustRegister("oak_boat", func(string) Item { return VanillaItem("oak_boat") })
	p.mustRegister("oak_hanging_sign", func(string) Item { return VanillaItem("oak_hanging_sign") })
	p.mustRegister("painting", func(string) Item { return VanillaItem("painting") })
	p.mustRegister("pale_oak_hanging_sign", func(string) Item { return VanillaItem("pale_oak_hanging_sign") })
	p.mustRegister("paper", func(string) Item { return VanillaItem("paper") })
	p.mustRegister("phantom_membrane", func(string) Item { return VanillaItem("phantom_membrane") })
	p.mustRegister("pitcher_pod", func(string) Item { return VanillaItem("pitcher_pod") })
	p.mustRegister("poisonous_potato", func(string) Item { return VanillaItem("poisonous_potato") })
	p.mustRegister("popped_chorus_fruit", func(string) Item { return VanillaItem("popped_chorus_fruit") })
	p.mustRegister("porkchop", func(string) Item { return VanillaItem("raw_porkchop") })
	p.mustRegister("potato", func(string) Item { return VanillaItem("potato") })
	p.mustRegister("potion", func(string) Item { return VanillaItem("potion") })
	p.mustRegister("prismarine_crystals", func(string) Item { return VanillaItem("prismarine_crystals") })
	p.mustRegister("prismarine_shard", func(string) Item { return VanillaItem("prismarine_shard") })
	p.mustRegister("puffer_fish", func(string) Item { return VanillaItem("pufferfish") })
	p.mustRegister("pufferfish", func(string) Item { return VanillaItem("pufferfish") })
	p.mustRegister("pumpkin_pie", func(string) Item { return VanillaItem("pumpkin_pie") })
	p.mustRegister("pumpkin_seeds", func(string) Item { return VanillaItem("pumpkin_seeds") })
	p.mustRegister("quartz", func(string) Item { return VanillaItem("nether_quartz") })
	p.mustRegister("rabbit", func(string) Item { return VanillaItem("raw_rabbit") })
	p.mustRegister("rabbit_foot", func(string) Item { return VanillaItem("rabbit_foot") })
	p.mustRegister("rabbit_hide", func(string) Item { return VanillaItem("rabbit_hide") })
	p.mustRegister("rabbit_stew", func(string) Item { return VanillaItem("rabbit_stew") })
	p.mustRegister("raiser_armor_trim_smithing_template", func(string) Item { return VanillaItem("raiser_armor_trim_smithing_template") })
	p.mustRegister("raw_beef", func(string) Item { return VanillaItem("raw_beef") })
	p.mustRegister("raw_cod", func(string) Item { return VanillaItem("raw_fish") })
	p.mustRegister("raw_copper", func(string) Item { return VanillaItem("raw_copper") })
	p.mustRegister("raw_chicken", func(string) Item { return VanillaItem("raw_chicken") })
	p.mustRegister("raw_fish", func(string) Item { return VanillaItem("raw_fish") })
	p.mustRegister("raw_gold", func(string) Item { return VanillaItem("raw_gold") })
	p.mustRegister("raw_iron", func(string) Item { return VanillaItem("raw_iron") })
	p.mustRegister("raw_mutton", func(string) Item { return VanillaItem("raw_mutton") })
	p.mustRegister("raw_porkchop", func(string) Item { return VanillaItem("raw_porkchop") })
	p.mustRegister("raw_rabbit", func(string) Item { return VanillaItem("raw_rabbit") })
	p.mustRegister("raw_salmon", func(string) Item { return VanillaItem("raw_salmon") })
	p.mustRegister("record_11", func(string) Item { return VanillaItem("record_11") })
	p.mustRegister("record_13", func(string) Item { return VanillaItem("record_13") })
	p.mustRegister("record_5", func(string) Item { return VanillaItem("record_5") })
	p.mustRegister("record_blocks", func(string) Item { return VanillaItem("record_blocks") })
	p.mustRegister("record_cat", func(string) Item { return VanillaItem("record_cat") })
	p.mustRegister("record_chirp", func(string) Item { return VanillaItem("record_chirp") })
	p.mustRegister("record_creator", func(string) Item { return VanillaItem("record_creator") })
	p.mustRegister("record_creator_music_box", func(string) Item { return VanillaItem("record_creator_music_box") })
	p.mustRegister("record_far", func(string) Item { return VanillaItem("record_far") })
	p.mustRegister("record_lava_chicken", func(string) Item { return VanillaItem("record_lava_chicken") })
	p.mustRegister("record_mall", func(string) Item { return VanillaItem("record_mall") })
	p.mustRegister("record_mellohi", func(string) Item { return VanillaItem("record_mellohi") })
	p.mustRegister("record_otherside", func(string) Item { return VanillaItem("record_otherside") })
	p.mustRegister("record_pigstep", func(string) Item { return VanillaItem("record_pigstep") })
	p.mustRegister("record_precipice", func(string) Item { return VanillaItem("record_precipice") })
	p.mustRegister("record_relic", func(string) Item { return VanillaItem("record_relic") })
	p.mustRegister("record_stal", func(string) Item { return VanillaItem("record_stal") })
	p.mustRegister("record_strad", func(string) Item { return VanillaItem("record_strad") })
	p.mustRegister("record_wait", func(string) Item { return VanillaItem("record_wait") })
	p.mustRegister("record_ward", func(string) Item { return VanillaItem("record_ward") })
	p.mustRegister("recovery_compass", func(string) Item { return VanillaItem("recovery_compass") })
	p.mustRegister("redstone", func(string) Item { return VanillaItem("redstone_dust") })
	p.mustRegister("redstone_dust", func(string) Item { return VanillaItem("redstone_dust") })
	p.mustRegister("resin_brick", func(string) Item { return VanillaItem("resin_brick") })
	p.mustRegister("rib_armor_trim_smithing_template", func(string) Item { return VanillaItem("rib_armor_trim_smithing_template") })
	p.mustRegister("rotten_flesh", func(string) Item { return VanillaItem("rotten_flesh") })
	p.mustRegister("salmon", func(string) Item { return VanillaItem("raw_salmon") })
	p.mustRegister("scute", func(string) Item { return VanillaItem("scute") })
	p.mustRegister("sentry_armor_trim_smithing_template", func(string) Item { return VanillaItem("sentry_armor_trim_smithing_template") })
	p.mustRegister("shaper_armor_trim_smithing_template", func(string) Item { return VanillaItem("shaper_armor_trim_smithing_template") })
	p.mustRegister("seeds", func(string) Item { return VanillaItem("wheat_seeds") })
	p.mustRegister("shears", func(string) Item { return VanillaItem("shears") })
	p.mustRegister("shulker_shell", func(string) Item { return VanillaItem("shulker_shell") })
	p.mustRegister("silence_armor_trim_smithing_template", func(string) Item { return VanillaItem("silence_armor_trim_smithing_template") })
	p.mustRegister("slime_ball", func(string) Item { return VanillaItem("slimeball") })
	p.mustRegister("snout_armor_trim_smithing_template", func(string) Item { return VanillaItem("snout_armor_trim_smithing_template") })
	p.mustRegister("slimeball", func(string) Item { return VanillaItem("slimeball") })
	p.mustRegister("snowball", func(string) Item { return VanillaItem("snowball") })
	p.mustRegister("speckled_melon", func(string) Item { return VanillaItem("glistering_melon") })
	p.mustRegister("spider_eye", func(string) Item { return VanillaItem("spider_eye") })
	p.mustRegister("spire_armor_trim_smithing_template", func(string) Item { return VanillaItem("spire_armor_trim_smithing_template") })
	p.mustRegister("splash_potion", func(string) Item { return VanillaItem("splash_potion") })
	p.mustRegister("spruce_boat", func(string) Item { return VanillaItem("spruce_boat") })
	p.mustRegister("spruce_hanging_sign", func(string) Item { return VanillaItem("spruce_hanging_sign") })
	p.mustRegister("spyglass", func(string) Item { return VanillaItem("spyglass") })
	p.mustRegister("squid_spawn_egg", func(string) Item { return VanillaItem("squid_spawn_egg") })
	p.mustRegister("steak", func(string) Item { return VanillaItem("steak") })
	p.mustRegister("stick", func(string) Item { return VanillaItem("stick") })
	p.mustRegister("sticks", func(string) Item { return VanillaItem("stick") })
	p.mustRegister("stone_axe", func(string) Item { return VanillaItem("stone_axe") })
	p.mustRegister("stone_hoe", func(string) Item { return VanillaItem("stone_hoe") })
	p.mustRegister("stone_pickaxe", func(string) Item { return VanillaItem("stone_pickaxe") })
	p.mustRegister("stone_shovel", func(string) Item { return VanillaItem("stone_shovel") })
	p.mustRegister("stone_sword", func(string) Item { return VanillaItem("stone_sword") })
	p.mustRegister("string", func(string) Item { return VanillaItem("string") })
	p.mustRegister("sugar", func(string) Item { return VanillaItem("sugar") })
	p.mustRegister("suspicious_stew", func(string) Item { return VanillaItem("suspicious_stew") })
	p.mustRegister("sweet_berries", func(string) Item { return VanillaItem("sweet_berries") })
	p.mustRegister("tonic", func(string) Item {
		return func() Item {
			v := VanillaItem("medicine")
			v.(interface{ SetType(MedicineType) }).SetType(MedicineTypeTonic)
			return v
		}()
	})
	p.mustRegister("torchflower_seeds", func(string) Item { return VanillaItem("torchflower_seeds") })
	p.mustRegister("tide_armor_trim_smithing_template", func(string) Item { return VanillaItem("tide_armor_trim_smithing_template") })
	p.mustRegister("totem", func(string) Item { return VanillaItem("totem") })
	p.mustRegister("trident", func(string) Item { return VanillaItem("trident") })
	p.mustRegister("turtle_helmet", func(string) Item { return VanillaItem("turtle_helmet") })
	p.mustRegister("vex_armor_trim_smithing_template", func(string) Item { return VanillaItem("vex_armor_trim_smithing_template") })
	p.mustRegister("turtle_shell_piece", func(string) Item { return VanillaItem("scute") })
	p.mustRegister("villager_spawn_egg", func(string) Item { return VanillaItem("villager_spawn_egg") })
	p.mustRegister("ward_armor_trim_smithing_template", func(string) Item { return VanillaItem("ward_armor_trim_smithing_template") })
	p.mustRegister("warped_hanging_sign", func(string) Item { return VanillaItem("warped_hanging_sign") })
	p.mustRegister("water_bucket", func(string) Item { return VanillaItem("water_bucket") })
	p.mustRegister("wayfinder_armor_trim_smithing_template", func(string) Item { return VanillaItem("wayfinder_armor_trim_smithing_template") })
	p.mustRegister("wheat", func(string) Item { return VanillaItem("wheat") })
	p.mustRegister("wheat_seeds", func(string) Item { return VanillaItem("wheat_seeds") })
	p.mustRegister("wild_armor_trim_smithing_template", func(string) Item { return VanillaItem("wild_armor_trim_smithing_template") })
	p.mustRegister("wooden_axe", func(string) Item { return VanillaItem("wooden_axe") })
	p.mustRegister("wooden_hoe", func(string) Item { return VanillaItem("wooden_hoe") })
	p.mustRegister("wooden_pickaxe", func(string) Item { return VanillaItem("wooden_pickaxe") })
	p.mustRegister("wooden_shovel", func(string) Item { return VanillaItem("wooden_shovel") })
	p.mustRegister("wooden_sword", func(string) Item { return VanillaItem("wooden_sword") })
	p.mustRegister("writable_book", func(string) Item { return VanillaItem("writable_book") })
	p.mustRegister("written_book", func(string) Item { return VanillaItem("written_book") })
	p.mustRegister("zombie_spawn_egg", func(string) Item { return VanillaItem("zombie_spawn_egg") })
}
