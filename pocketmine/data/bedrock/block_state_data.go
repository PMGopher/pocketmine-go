package bedrock

import (
	"fmt"
	"sort"
	"strings"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/nbt"
)

// BlockStateData tag names, a port of BlockStateData::TAG_*.
const (
	BlockStateTagName    = "name"
	BlockStateTagStates  = "states"
	BlockStateTagVersion = "version"
)

// CurrentBlockStateVersion is a port of BlockStateData::CURRENT_VERSION
// (WorldDataVersions::BLOCK_STATES): the Bedrock version of the most recent backwards-incompatible
// change to blockstates, 1.21.60.33. It matches the newest schema in
// data/bedrock/block/upgrade/schema.
const CurrentBlockStateVersion int32 = (1 << 24) | (21 << 16) | (60 << 8) | 33

// NewCurrentBlockStateData is a port of BlockStateData::current.
func NewCurrentBlockStateData(name string, states map[string]any) BlockStateData {
	if states == nil {
		states = map[string]any{}
	}
	return BlockStateData{Name: name, States: states, Version: CurrentBlockStateVersion}
}

// GetState is a port of BlockStateData::getState: the value is an int32 (TAG_Int), uint8
// (TAG_Byte) or string (TAG_String).
func (d BlockStateData) GetState(name string) (any, bool) {
	v, ok := d.States[name]
	return v, ok
}

// GetVersionAsString is a port of BlockStateData::getVersionAsString.
func (d BlockStateData) GetVersionAsString() string {
	v := uint32(d.Version)
	return fmt.Sprintf("%d.%d.%d.%d", (v>>24)&0xff, (v>>16)&0xff, (v>>8)&0xff, v&0xff)
}

// BlockStateDeserializeError is a port of pocketmine\data\bedrock\block\BlockStateDeserializeException.
type BlockStateDeserializeError struct{ Message string }

func (e *BlockStateDeserializeError) Error() string { return e.Message }

// BlockStateDataFromNbt is a port of BlockStateData::fromNbt.
func BlockStateDataFromNbt(tag *nbt.CompoundTag) (BlockStateData, error) {
	name, err := tag.GetString(BlockStateTagName)
	if err != nil {
		return BlockStateData{}, &BlockStateDeserializeError{err.Error()}
	}
	statesTag, ok, err := tag.GetCompoundTag(BlockStateTagStates)
	if err != nil {
		return BlockStateData{}, &BlockStateDeserializeError{err.Error()}
	}
	if !ok {
		return BlockStateData{}, &BlockStateDeserializeError{fmt.Sprintf("Missing tag \"%s\"", BlockStateTagStates)}
	}
	version := tag.GetIntOr(BlockStateTagVersion, 0)
	if t, ok := tag.GetTag(BlockStateTagVersion); ok {
		if _, isInt := t.(nbt.IntTag); !isInt {
			return BlockStateData{}, &BlockStateDeserializeError{fmt.Sprintf("Expected a tag of type TAG_Int for \"%s\"", BlockStateTagVersion)}
		}
	}

	var extra []string
	for key := range tag.All() {
		switch key {
		case BlockStateTagName, BlockStateTagStates, BlockStateTagVersion, pocketmine.TagWorldDataVersion:
		default:
			extra = append(extra, key)
		}
	}
	if len(extra) != 0 {
		sort.Strings(extra)
		return BlockStateData{}, &BlockStateDeserializeError{"Unexpected extra keys: " + strings.Join(extra, ", ")}
	}

	states, err := BlockStatesFromNbt(statesTag)
	if err != nil {
		return BlockStateData{}, &BlockStateDeserializeError{err.Error()}
	}
	return BlockStateData{Name: string(name), States: states, Version: int32(version)}, nil
}

// BlockStatesFromNbt converts a blockstate's "states" compound into BlockStateData.States form
// (TAG_Int -> int32, TAG_Byte -> uint8, TAG_String -> string).
func BlockStatesFromNbt(statesTag *nbt.CompoundTag) (map[string]any, error) {
	states := make(map[string]any, statesTag.Count())
	for propName, propTag := range statesTag.All() {
		v, err := StateValueFromTag(propTag)
		if err != nil {
			return nil, fmt.Errorf("state property %q: %w", propName, err)
		}
		states[propName] = v
	}
	return states, nil
}

// StateValueFromTag converts one state property tag to its BlockStateData.States value.
func StateValueFromTag(t nbt.Tag) (any, error) {
	switch v := t.(type) {
	case nbt.IntTag:
		return int32(v), nil
	case nbt.ByteTag:
		return uint8(v), nil
	case nbt.StringTag:
		return string(v), nil
	}
	return nil, fmt.Errorf("unsupported NBT tag type %T", t)
}

// StateValueToTag is the reverse of StateValueFromTag.
func StateValueToTag(v any) (nbt.Tag, error) {
	switch v := v.(type) {
	case int32:
		return nbt.IntTag(v), nil
	case uint8:
		return nbt.ByteTag(int8(v)), nil
	case string:
		return nbt.StringTag(v), nil
	}
	return nil, fmt.Errorf("unsupported block state value type %T", v)
}

// ToVanillaNbt is a port of BlockStateData::toVanillaNbt: the blockstate exactly as vanilla
// Bedrock encodes it. States are written in name order so the output is deterministic.
func (d BlockStateData) ToVanillaNbt() *nbt.CompoundTag {
	names := make([]string, 0, len(d.States))
	for k := range d.States {
		names = append(names, k)
	}
	sort.Strings(names)
	states := nbt.NewCompoundTag()
	for _, k := range names {
		t, err := StateValueToTag(d.States[k])
		if err != nil {
			panic(fmt.Sprintf("block state %s property %q: %v", d.Name, k, err))
		}
		states.SetTag(k, t)
	}
	tag := nbt.NewCompoundTag()
	tag.SetString(BlockStateTagName, nbt.StringTag(d.Name))
	tag.SetInt(BlockStateTagVersion, nbt.IntTag(d.Version))
	tag.SetTag(BlockStateTagStates, states)
	return tag
}

// ToNbt is a port of BlockStateData::toNbt: the vanilla encoding plus PocketMine-MP's world data
// version, used for anything saved to disk.
func (d BlockStateData) ToNbt() *nbt.CompoundTag {
	tag := d.ToVanillaNbt()
	tag.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))
	return tag
}

// Equals is a port of BlockStateData::equals (the version isn't compared, like PHP).
func (d BlockStateData) Equals(that BlockStateData) bool {
	if d.Name != that.Name || len(d.States) != len(that.States) {
		return false
	}
	for k, v := range d.States {
		w, ok := that.States[k]
		if !ok || !StateValuesEqual(v, w) {
			return false
		}
	}
	return true
}

// StateValuesEqual is Tag::equals for two state values: same tag type and same value.
func StateValuesEqual(a, b any) bool {
	switch a := a.(type) {
	case int32:
		b, ok := b.(int32)
		return ok && a == b
	case uint8:
		b, ok := b.(uint8)
		return ok && a == b
	case string:
		b, ok := b.(string)
		return ok && a == b
	}
	return false
}
