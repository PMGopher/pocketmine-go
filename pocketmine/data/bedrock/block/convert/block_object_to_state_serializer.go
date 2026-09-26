package blockconvert

import (
	"sync"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockSerializerFunc serializes a block of one type into its state data.
type BlockSerializerFunc func(blk block.Behavior) bedrock.BlockStateData

// BlockObjectToStateSerializer is a port of
// pocketmine\data\bedrock\block\convert\BlockObjectToStateSerializer.
type BlockObjectToStateSerializer struct {
	serializers map[int]BlockSerializerFunc

	cacheMu sync.RWMutex
	cache   map[int]bedrock.BlockStateData
}

func NewBlockObjectToStateSerializer() *BlockObjectToStateSerializer {
	return &BlockObjectToStateSerializer{serializers: map[int]BlockSerializerFunc{}, cache: map[int]bedrock.BlockStateData{}}
}

// Serialize is a port of BlockObjectToStateSerializer::serialize: the state data of an internal
// block state ID, cached.
func (s *BlockObjectToStateSerializer) Serialize(stateID int) (bedrock.BlockStateData, error) {
	//TODO: singleton usage not ideal
	//TODO: we may want to deduplicate cache entries to avoid wasting memory
	s.cacheMu.RLock()
	data, ok := s.cache[stateID]
	s.cacheMu.RUnlock()
	if ok {
		return data, nil
	}
	data, err := s.SerializeBlock(block.GetRuntimeBlockStateRegistry().FromStateId(stateID))
	if err != nil {
		return bedrock.BlockStateData{}, err
	}
	s.cacheMu.Lock()
	s.cache[stateID] = data
	s.cacheMu.Unlock()
	return data, nil
}

// IsRegistered is a port of BlockObjectToStateSerializer::isRegistered.
func (s *BlockObjectToStateSerializer) IsRegistered(blk block.Behavior) bool {
	_, ok := s.serializers[blk.GetTypeId()]
	return ok
}

// Map is a port of BlockObjectToStateSerializer::map.
func (s *BlockObjectToStateSerializer) Map(blk block.Behavior, serializer BlockSerializerFunc) {
	if _, ok := s.serializers[blk.GetTypeId()]; ok {
		panic("Block type ID " + itoa(blk.GetTypeId()) + " (" + blk.GetName() + ") already has a serializer registered")
	}
	s.serializers[blk.GetTypeId()] = serializer
}

// MapStatic is map() with a fixed BlockStateData (a state-independent serializer).
func (s *BlockObjectToStateSerializer) MapStatic(blk block.Behavior, data bedrock.BlockStateData) {
	s.Map(blk, func(block.Behavior) bedrock.BlockStateData { return data })
}

// MapSimple is a port of BlockObjectToStateSerializer::mapSimple.
func (s *BlockObjectToStateSerializer) MapSimple(blk block.Behavior, id string) {
	s.MapStatic(blk, NewBlockStateWriter(id).GetBlockStateData())
}

// MapSlab is a port of BlockObjectToStateSerializer::mapSlab.
func (s *BlockObjectToStateSerializer) MapSlab(blk block.Behavior, singleID, doubleID string) {
	s.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
		return encodeSlab(as[slabLike](b), singleID, doubleID).GetBlockStateData()
	})
}

// MapStairs is a port of BlockObjectToStateSerializer::mapStairs.
func (s *BlockObjectToStateSerializer) MapStairs(blk block.Behavior, id string) {
	s.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
		return encodeStairs(b, NewBlockStateWriter(id)).GetBlockStateData()
	})
}

// MapLog is a port of BlockObjectToStateSerializer::mapLog.
func (s *BlockObjectToStateSerializer) MapLog(blk block.Behavior, unstrippedID, strippedID string) {
	s.Map(blk, func(b block.Behavior) bedrock.BlockStateData {
		return encodeLog(b, unstrippedID, strippedID).GetBlockStateData()
	})
}

// SerializeBlock is a port of BlockObjectToStateSerializer::serializeBlock.
func (s *BlockObjectToStateSerializer) SerializeBlock(blockState block.Behavior) (data bedrock.BlockStateData, err error) {
	defer recoverError(&err)
	typeID := blockState.GetTypeId()
	locatedSerializer, ok := s.serializers[typeID]
	if !ok {
		return bedrock.BlockStateData{}, &BlockStateSerializeError{Message: "No serializer registered for " + typeName(blockState) + " with type ID " + itoa(typeID)}
	}
	return locatedSerializer(blockState), nil
}
