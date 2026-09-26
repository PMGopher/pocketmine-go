package blockconvert

import (
	"sync"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockDeserializerFunc creates a block from a state's properties.
type BlockDeserializerFunc func(in *BlockStateReader) block.Behavior

// BlockStateToObjectDeserializer is a port of
// pocketmine\data\bedrock\block\convert\BlockStateToObjectDeserializer.
type BlockStateToObjectDeserializer struct {
	deserializeFuncs map[string]BlockDeserializerFunc

	cacheMu     sync.RWMutex
	simpleCache map[string]int
}

func NewBlockStateToObjectDeserializer() *BlockStateToObjectDeserializer {
	return &BlockStateToObjectDeserializer{deserializeFuncs: map[string]BlockDeserializerFunc{}, simpleCache: map[string]int{}}
}

// Deserialize is a port of BlockStateToObjectDeserializer::deserialize: the internal state ID of
// the block with that state data.
func (d *BlockStateToObjectDeserializer) Deserialize(stateData bedrock.BlockStateData) (int, error) {
	if len(stateData.States) == 0 {
		//if a block has zero properties, we can keep a map of string ID -> internal blockstate ID
		d.cacheMu.RLock()
		stateID, ok := d.simpleCache[stateData.Name]
		d.cacheMu.RUnlock()
		if ok {
			return stateID, nil
		}
		stateID, err := d.deserializeToStateID(stateData)
		if err != nil {
			return 0, err
		}
		d.cacheMu.Lock()
		d.simpleCache[stateData.Name] = stateID
		d.cacheMu.Unlock()
		return stateID, nil
	}
	//we can't cache blocks that have properties - go ahead and deserialize the slow way
	return d.deserializeToStateID(stateData)
}

func (d *BlockStateToObjectDeserializer) deserializeToStateID(stateData bedrock.BlockStateData) (int, error) {
	blk, err := d.DeserializeBlock(stateData)
	if err != nil {
		return 0, err
	}
	stateID := blk.GetStateId()
	//plugin devs seem to keep missing this and causing core crashes, so we need to verify this at the earliest
	//available opportunity
	if !block.GetRuntimeBlockStateRegistry().HasStateId(stateID) {
		panic("State ID " + itoa(stateID) + " returned by deserializer for " + stateData.Name + " is not registered in RuntimeBlockStateRegistry")
	}
	return stateID, nil
}

// Map is a port of BlockStateToObjectDeserializer::map.
func (d *BlockStateToObjectDeserializer) Map(id string, c BlockDeserializerFunc) {
	d.deserializeFuncs[id] = c
	d.cacheMu.Lock()
	d.simpleCache = map[string]int{}
	d.cacheMu.Unlock()
}

// GetDeserializerForID is a port of BlockStateToObjectDeserializer::getDeserializerForId.
func (d *BlockStateToObjectDeserializer) GetDeserializerForID(id string) (BlockDeserializerFunc, bool) {
	c, ok := d.deserializeFuncs[id]
	return c, ok
}

// MapSimple is a port of BlockStateToObjectDeserializer::mapSimple.
func (d *BlockStateToObjectDeserializer) MapSimple(id string, getBlock func() block.Behavior) {
	d.Map(id, func(*BlockStateReader) block.Behavior { return getBlock() })
}

// MapSlab is a port of BlockStateToObjectDeserializer::mapSlab.
func (d *BlockStateToObjectDeserializer) MapSlab(singleID, doubleID string, getBlock func() block.Behavior) {
	d.Map(singleID, func(in *BlockStateReader) block.Behavior { return decodeSingleSlab(getBlock(), in) })
	d.Map(doubleID, func(in *BlockStateReader) block.Behavior { return decodeDoubleSlab(getBlock(), in) })
}

// MapStairs is a port of BlockStateToObjectDeserializer::mapStairs.
func (d *BlockStateToObjectDeserializer) MapStairs(id string, getBlock func() block.Behavior) {
	d.Map(id, func(in *BlockStateReader) block.Behavior { return decodeStairs(getBlock(), in) })
}

// MapLog is a port of BlockStateToObjectDeserializer::mapLog.
func (d *BlockStateToObjectDeserializer) MapLog(unstrippedID, strippedID string, getBlock func() block.Behavior) {
	d.Map(unstrippedID, func(in *BlockStateReader) block.Behavior { return decodeLog(getBlock(), false, in) })
	d.Map(strippedID, func(in *BlockStateReader) block.Behavior { return decodeLog(getBlock(), true, in) })
}

// DeserializeBlock is a port of BlockStateToObjectDeserializer::deserializeBlock.
func (d *BlockStateToObjectDeserializer) DeserializeBlock(blockStateData bedrock.BlockStateData) (blk block.Behavior, err error) {
	defer recoverError(&err)
	id := blockStateData.Name
	c, ok := d.deserializeFuncs[id]
	if !ok {
		return nil, &UnsupportedBlockStateError{BlockStateDeserializeError{Message: "Unknown block ID \"" + id + "\""}}
	}
	reader := NewBlockStateReader(blockStateData)
	blk = c(reader)
	reader.CheckUnreadProperties()
	return blk, nil
}
