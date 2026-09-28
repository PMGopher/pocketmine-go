package item

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// itemFromStringBcMap is resources/item_from_string_bc_map.json from PocketMine-MP: legacy item
// names (and numeric IDs) to their legacy string IDs.
//
//go:embed item_from_string_bc_map.json
var itemFromStringBcMap []byte

// LegacyStringToItemParserError is a port of pocketmine\item\LegacyStringToItemParserException.
type LegacyStringToItemParserError struct {
	Message string
	Cause   error
}

func (e *LegacyStringToItemParserError) Error() string { return e.Message }
func (e *LegacyStringToItemParserError) Unwrap() error { return e.Cause }

// LegacyItemDataFunc upgrades a legacy string ID + meta to current item data and deserializes it:
// GlobalItemDataHandlers::getUpgrader()->upgradeItemTypeDataString() followed by
// getDeserializer()->deserializeStack(). It's set by world/format/io, which owns those handlers
// (they import this package).
var LegacyItemDataFunc func(legacyID string, meta int) (Item, error)

// LegacyStringToItemParser is a port of pocketmine\item\LegacyStringToItemParser: parses the
// old-style `id:meta` strings (e.g. `diamond_pickaxe:5`, `minecraft:string`, `351:4`).
//
// PHP's constructor takes the ItemDataUpgrader and ItemDeserializer; here they're the single
// resolve function (see LegacyItemDataFunc for why).
type LegacyStringToItemParser struct {
	resolve func(legacyID string, meta int) (Item, error)
	// map is alias => legacy string ID.
	mappings map[string]string
}

var (
	legacyStringToItemParser     *LegacyStringToItemParser
	legacyStringToItemParserOnce sync.Once
)

// GetLegacyStringToItemParser is LegacyStringToItemParser::getInstance.
func GetLegacyStringToItemParser() *LegacyStringToItemParser {
	legacyStringToItemParserOnce.Do(func() {
		legacyStringToItemParser = makeLegacyStringToItemParser()
	})
	return legacyStringToItemParser
}

// makeLegacyStringToItemParser is a port of LegacyStringToItemParser::make.
func makeLegacyStringToItemParser() *LegacyStringToItemParser {
	result := NewLegacyStringToItemParser(func(legacyID string, meta int) (Item, error) {
		if LegacyItemDataFunc == nil {
			return nil, fmt.Errorf("item data handlers aren't loaded (world/format/io isn't imported)")
		}
		return LegacyItemDataFunc(legacyID, meta)
	})

	var mappings map[string]any
	if err := json.Unmarshal(itemFromStringBcMap, &mappings); err != nil {
		panic("Invalid mappings format, expected array")
	}
	for name, id := range mappings {
		idString, ok := id.(string)
		if !ok {
			panic("Invalid mappings format, expected string values")
		}
		result.AddMapping(name, idString)
	}
	return result
}

// NewLegacyStringToItemParser is LegacyStringToItemParser::__construct.
func NewLegacyStringToItemParser(resolve func(legacyID string, meta int) (Item, error)) *LegacyStringToItemParser {
	return &LegacyStringToItemParser{resolve: resolve, mappings: map[string]string{}}
}

// AddMapping is a port of LegacyStringToItemParser::addMapping.
func (p *LegacyStringToItemParser) AddMapping(alias, id string) { p.mappings[alias] = id }

// GetMappings is a port of LegacyStringToItemParser::getMappings.
func (p *LegacyStringToItemParser) GetMappings() map[string]string {
	result := make(map[string]string, len(p.mappings))
	for k, v := range p.mappings {
		result[k] = v
	}
	return result
}

// Parse is a port of LegacyStringToItemParser::parse. Accepted formats include
// `diamond_pickaxe:5`, `minecraft:string` and `351:4` (lapis lazuli ID:meta).
func (p *LegacyStringToItemParser) Parse(input string) (Item, error) {
	key := p.reprocess(input)
	//TODO: this should be limited to 2 parts, but 3 preserves old behaviour when given a string like 351:4:1
	b := strings.SplitN(key, ":", 3)

	meta := 0
	if len(b) > 1 {
		m, ok := phpNumericInt(b[1])
		if !ok {
			return nil, &LegacyStringToItemParserError{Message: "Unable to parse \"" + b[1] + "\" from \"" + input + "\" as a valid meta value"}
		}
		meta = m
	}

	lower := strings.ToLower(b[0])
	if lower == "0" || lower == "air" {
		//item deserializer doesn't recognize air items since they aren't supposed to exist
		return VanillaAir(), nil
	}

	legacyID, ok := p.mappings[lower]
	if !ok {
		return nil, &LegacyStringToItemParserError{Message: "Unable to resolve \"" + input + "\" to a valid item"}
	}
	it, err := p.resolve(legacyID, meta)
	if err != nil {
		return nil, &LegacyStringToItemParserError{Message: err.Error(), Cause: err}
	}
	return it, nil
}

// phpNumericInt is `is_numeric($s) ? (int) $s` for the meta part: integers and decimal/exponent
// numbers (truncated like PHP's int cast), with optional leading whitespace.
func phpNumericInt(s string) (int, bool) {
	t := strings.TrimLeft(s, " \t\n\r\v\f")
	t = strings.TrimRight(t, " \t\n\r\v\f")
	if t == "" {
		return 0, false
	}
	if i, err := strconv.Atoi(t); err == nil {
		return i, true
	}
	if strings.ContainsAny(t, "xXbBoO_") {
		return 0, false
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, false
	}
	return int(f), true
}

// reprocess is a port of LegacyStringToItemParser::reprocess.
func (p *LegacyStringToItemParser) reprocess(input string) string {
	return strings.NewReplacer(" ", "_", "minecraft:", "").Replace(strings.TrimSpace(input))
}
