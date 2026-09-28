package generator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/world/format"
)

// FlatGeneratorOptions is a port of pocketmine\world\generator\FlatGeneratorOptions: a parsed flat
// world preset ("2;bedrock,59xstone,3xdirt,grass;1;decoration"). Structure holds the layers from
// y=0 upwards.
type FlatGeneratorOptions struct {
	Structure    []FlatLayer
	BiomeID      int32
	ExtraOptions map[string]map[string]string // an option without parameters maps to an empty map
}

// InvalidGeneratorOptionsError is a port of InvalidGeneratorOptionsException.
type InvalidGeneratorOptionsError struct{ Message string }

func (e *InvalidGeneratorOptionsError) Error() string { return e.Message }

var (
	flatLayerPattern  = regexp.MustCompile(`^(?:(\d+)[x|*])?(.+)$`)
	flatOptionPattern = regexp.MustCompile(`(([0-9a-z_]{1,})\(?([0-9a-z_ =:]{0,})\)?),?`)
)

// ParseFlatLayers is a port of FlatGeneratorOptions::parseLayers: "bedrock,2xdirt,grass".
func ParseFlatLayers(layers string) ([]FlatLayer, error) {
	var result []FlatLayer
	split := strings.SplitN(layers, ",", format.MaxSubChunks*format.SubChunkEdgeLength) // World::Y_MAX - World::Y_MIN
	for _, line := range split {
		line = strings.TrimSpace(line)
		matches := flatLayerPattern.FindStringSubmatch(line)
		if matches == nil {
			return nil, &InvalidGeneratorOptionsError{Message: fmt.Sprintf("Invalid preset layer %q", line)}
		}
		count := 1
		if matches[1] != "" {
			count, _ = strconv.Atoi(matches[1])
		}
		it, err := item.GetLegacyStringToItemParser().Parse(matches[2])
		if err != nil {
			return nil, &InvalidGeneratorOptionsError{Message: fmt.Sprintf("Invalid preset layer %q: %v", line, err)}
		}
		var b block.Behavior = block.VanillaAir() // Item::getBlock() is air for a non-block item
		if ib, ok := it.(interface{ GetBlock() block.Behavior }); ok {
			b = ib.GetBlock()
		}
		result = append(result, FlatLayer{Block: b, Height: count})
	}
	return result, nil
}

// ParseFlatPreset is a port of FlatGeneratorOptions::parsePreset.
func ParseFlatPreset(presetString string) (*FlatGeneratorOptions, error) {
	preset := strings.SplitN(presetString, ";", 4)
	part := func(i int) (string, bool) {
		if i < len(preset) {
			return preset[i], true
		}
		return "", false
	}
	blocks, _ := part(1)
	biomeID := int32(VanillaFlatBiomeID) // BiomeIds::PLAINS
	if s, ok := part(2); ok {
		v, _ := strconv.Atoi(strings.TrimSpace(s)) // PHP's (int) cast
		biomeID = int32(v)
	}
	optionsString, _ := part(3)
	structure, err := ParseFlatLayers(blocks)
	if err != nil {
		return nil, err
	}

	options := map[string]map[string]string{}
	//TODO: more error checking (as in PHP)
	for _, m := range flatOptionPattern.FindAllStringSubmatch(optionsString, -1) {
		params := map[string]string{}
		if m[3] != "" {
			for _, k := range strings.Split(m[3], " ") {
				//TODO: this should be limited to 2 parts, but 3 preserves old behaviour when given
				//e.g. treecount=20=1 (as in PHP)
				kv := strings.SplitN(k, "=", 3)
				if len(kv) > 1 {
					params[kv[0]] = kv[1]
				}
			}
		}
		options[m[2]] = params
	}
	return &FlatGeneratorOptions{Structure: structure, BiomeID: biomeID, ExtraOptions: options}, nil
}
