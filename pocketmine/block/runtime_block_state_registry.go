package block

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

// baseLightFilter is LightUpdate::BASE_LIGHT_FILTER (world/light.BaseLightFilter; this package
// can't import world).
const baseLightFilter = 1

// GenerateStatePermutations is a port of Block::generateStatePermutations: every valid state of
// blk's type, found by decoding every possible state data value and keeping those that decode.
func GenerateStatePermutations(blk Behavior) []Behavior {
	base := blockOf(blk)
	//TODO: this bruteforce approach to discovering all valid states is very inefficient for larger state data sizes
	//at some point we'll need to find a better way to do this
	bits := base.requiredBlockItemStateBits + base.requiredBlockOnlyStateBits
	if bits > InternalStateDataBits {
		panic(fmt.Sprintf("Block state data cannot use more than %d bits", InternalStateDataBits))
	}
	var result []Behavior
	for blockItemStateData := 0; blockItemStateData < 1<<base.requiredBlockItemStateBits; blockItemStateData++ {
		withType := blk.Clone()
		if !decodeState(func() error { return withType.DecodeBlockItemState(blockItemStateData) }) {
			continue //invalid property combination, leave it
		}
		if encoded, err := blockOf(withType).encodeBlockItemState(); err != nil || encoded != blockItemStateData {
			panic(fmt.Sprintf("%T.DecodeBlockItemState() accepts invalid inputs (returned %d for input %d)", blk, encoded, blockItemStateData))
		}

		for blockOnlyStateData := 0; blockOnlyStateData < 1<<base.requiredBlockOnlyStateBits; blockOnlyStateData++ {
			withState := withType.Clone()
			if !decodeState(func() error { return withState.DecodeBlockOnlyState(blockOnlyStateData) }) {
				continue //invalid property combination, leave it
			}
			if encoded, err := blockOf(withState).encodeBlockOnlyState(); err != nil || encoded != blockOnlyStateData {
				panic(fmt.Sprintf("%T.DecodeBlockOnlyState() accepts invalid inputs (returned %d for input %d)", blk, encoded, blockOnlyStateData))
			}
			result = append(result, withState)
		}
	}
	return result
}

// decodeState runs a decode, reporting false for InvalidSerializedRuntimeDataException (which
// the runtime describers raise as a panic).
func decodeState(decode func() error) (ok bool) {
	defer func() {
		if p := recover(); p != nil {
			if err, isErr := p.(error); isErr {
				var invalid *runtime.InvalidSerializedRuntimeDataError
				if errors.As(err, &invalid) {
					ok = false
					return
				}
			}
			panic(p)
		}
	}()
	return decode() == nil
}

// blockBase is implemented by every concrete block through its embedded Block.
type blockBase interface {
	base() *Block
}

func (b *Block) base() *Block { return b }

func blockOf(blk Behavior) *Block { return blk.(blockBase).base() }

// Collision info values, RuntimeBlockStateRegistry::COLLISION_*.
const (
	CollisionCustom      = 0
	CollisionCube        = 1
	CollisionNone        = 2
	CollisionMayOverflow = 3
)

// RuntimeBlockStateRegistry is a port of pocketmine\block\RuntimeBlockStateRegistry: every known
// block state, by state ID, with the per-state tables the world (lighting, explosions, collisions)
// looks up.
type RuntimeBlockStateRegistry struct {
	mu sync.RWMutex

	fullList  map[int]Behavior
	typeIndex map[int]Behavior

	Light                map[int]int
	LightFilter          map[int]int
	BlocksDirectSkyLight map[int]bool
	BlastResistance      map[int]float64
	CollisionInfo        map[int]int
}

var (
	runtimeBlockStateRegistry     *RuntimeBlockStateRegistry
	runtimeBlockStateRegistryOnce sync.Once
)

// GetRuntimeBlockStateRegistry is RuntimeBlockStateRegistry::getInstance(): every vanilla block
// is registered on first use.
func GetRuntimeBlockStateRegistry() *RuntimeBlockStateRegistry {
	runtimeBlockStateRegistryOnce.Do(func() {
		r := &RuntimeBlockStateRegistry{
			fullList:             map[int]Behavior{},
			typeIndex:            map[int]Behavior{},
			Light:                map[int]int{},
			LightFilter:          map[int]int{},
			BlocksDirectSkyLight: map[int]bool{},
			BlastResistance:      map[int]float64{},
			CollisionInfo:        map[int]int{},
		}
		for _, name := range GetVanillaBlockNames() {
			r.Register(VanillaBlock(name))
		}
		runtimeBlockStateRegistry = r
	})
	return runtimeBlockStateRegistry
}

// Register is a port of RuntimeBlockStateRegistry::register.
func (r *RuntimeBlockStateRegistry) Register(blk Behavior) {
	r.mu.Lock()
	defer r.mu.Unlock()
	typeID := blk.GetTypeId()
	if _, ok := r.typeIndex[typeID]; ok {
		panic(fmt.Sprintf("Block ID %d is already used by another block", typeID))
	}
	r.typeIndex[typeID] = blk.Clone()

	for _, v := range GenerateStatePermutations(blk) {
		r.fillStaticArrays(v.GetStateId(), v)
	}
}

// calculateCollisionInfo is a port of RuntimeBlockStateRegistry::calculateCollisionInfo, minus the
// "overrides getModelPositionOffset()/readStateFromWorld()" check (Go has no reflection over which
// type declared a method); blocks whose boxes leave the cell are still COLLISION_MAY_OVERFLOW.
func calculateCollisionInfo(blk Behavior) int {
	boxes := blk.RecalculateCollisionBoxes()
	if len(boxes) == 0 {
		return CollisionNone
	}
	if len(boxes) == 1 &&
		boxes[0].MinX == 0 && boxes[0].MinY == 0 && boxes[0].MinZ == 0 &&
		boxes[0].MaxX == 1 && boxes[0].MaxY == 1 && boxes[0].MaxZ == 1 {
		return CollisionCube
	}
	for _, box := range boxes {
		if box.MinX < 0 || box.MaxX > 1 || box.MinY < 0 || box.MaxY > 1 || box.MinZ < 0 || box.MaxZ > 1 {
			return CollisionMayOverflow
		}
	}
	return CollisionCustom
}

// fillStaticArrays is a port of RuntimeBlockStateRegistry::fillStaticArrays.
func (r *RuntimeBlockStateRegistry) fillStaticArrays(index int, blk Behavior) {
	if index != blk.GetStateId() {
		panic("Cannot fill static arrays for an invalid blockstate")
	}
	r.fullList[index] = blk
	r.BlastResistance[index] = blk.GetBreakInfo().GetBlastResistance()
	r.Light[index] = blk.GetLightLevel()
	r.LightFilter[index] = min(15, blk.GetLightFilter()+baseLightFilter)
	if blk.BlocksDirectSkyLight() {
		r.BlocksDirectSkyLight[index] = true
	}
	r.CollisionInfo[index] = calculateCollisionInfo(blk)
}

// FromStateId is a port of RuntimeBlockStateRegistry::fromStateId: a new instance of the block with
// that state ID, or an UnknownBlock.
func (r *RuntimeBlockStateRegistry) FromStateId(stateID int) Behavior {
	if stateID < 0 {
		panic("Block state ID cannot be negative")
	}
	r.mu.RLock()
	blk, ok := r.fullList[stateID]
	r.mu.RUnlock()
	if ok { //hot
		return blk.Clone()
	}
	typeID := stateID >> InternalStateDataBits
	stateData := (stateID ^ typeID<<InternalStateDataBits) & InternalStateDataMask
	id, _ := NewBlockIdentifier(typeID, nil)
	return NewUnknownBlock(id, NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), stateData)
}

// HasStateId is a port of RuntimeBlockStateRegistry::hasStateId.
func (r *RuntimeBlockStateRegistry) HasStateId(stateID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.fullList[stateID]
	return ok
}

// GetAllKnownStates is a port of RuntimeBlockStateRegistry::getAllKnownStates (sorted by state ID).
func (r *RuntimeBlockStateRegistry) GetAllKnownStates() []Behavior {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]int, 0, len(r.fullList))
	for id := range r.fullList {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	result := make([]Behavior, len(ids))
	for i, id := range ids {
		result[i] = r.fullList[id]
	}
	return result
}
