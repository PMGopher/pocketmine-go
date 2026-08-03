package leveldb

import (
	"fmt"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
)

// blockStateToNBT builds the persistent "blockstate" compound real Bedrock world saves use for
// every block palette entry: {name: TAG_String, states: TAG_Compound, version: TAG_Int}. Matches
// BlockStateSerializer::serialize(...)->toNbt() by construction (bedrock.BlockStateData already
// has this exact Name/States/Version shape).
func blockStateToNBT(data bedrock.BlockStateData) (*nbt.CompoundTag, error) {
	states := nbt.NewCompoundTag()
	for name, value := range data.States {
		switch v := value.(type) {
		case int32:
			states.SetInt(name, nbt.IntTag(v))
		case uint8:
			states.SetByte(name, nbt.ByteTag(int8(v)))
		case string:
			states.SetString(name, nbt.StringTag(v))
		default:
			return nil, fmt.Errorf("leveldb: unsupported block state property type %T for %q", value, name)
		}
	}

	tag := nbt.NewCompoundTag()
	tag.SetString("name", nbt.StringTag(data.Name))
	tag.SetTag("states", states)
	tag.SetInt("version", nbt.IntTag(data.Version))
	return tag, nil
}

// nbtToBlockState is the reverse of blockStateToNBT.
func nbtToBlockState(tag *nbt.CompoundTag) (bedrock.BlockStateData, error) {
	name, err := tag.GetString("name")
	if err != nil {
		return bedrock.BlockStateData{}, fmt.Errorf("leveldb: blockstate compound missing \"name\": %w", err)
	}
	version, err := tag.GetInt("version")
	if err != nil {
		return bedrock.BlockStateData{}, fmt.Errorf("leveldb: blockstate compound missing \"version\": %w", err)
	}
	statesTag, ok, err := tag.GetCompoundTag("states")
	if err != nil {
		return bedrock.BlockStateData{}, fmt.Errorf("leveldb: blockstate compound has an invalid \"states\": %w", err)
	}

	states := map[string]any{}
	if ok {
		for propName, propTag := range statesTag.All() {
			switch v := propTag.(type) {
			case nbt.IntTag:
				states[propName] = int32(v)
			case nbt.ByteTag:
				states[propName] = uint8(v)
			case nbt.StringTag:
				states[propName] = string(v)
			default:
				return bedrock.BlockStateData{}, fmt.Errorf("leveldb: unsupported NBT tag type %T for state property %q", propTag, propName)
			}
		}
	}

	return bedrock.BlockStateData{Name: string(name), States: states, Version: int32(version)}, nil
}
