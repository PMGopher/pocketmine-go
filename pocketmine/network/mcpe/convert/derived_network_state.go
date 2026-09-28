package convert

import (
	"fmt"
	"sort"
	"strings"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// Since 1.26.50 the client takes the connections of fences, glass panes, bars and tripwire and the
// corner shape of stairs from the block state (minecraft:connection_* and minecraft:corner), like
// Dragonfly sends them. PocketMine-MP 5.44.4 (1.26.30) has no such properties: the client worked
// them out itself. Like PocketMine-MP, this port keeps them out of the internal block state (they
// depend on the neighbours, worked out by ReadStateFromWorld), so the network state of those
// blocks is completed from the block read from the world: NetworkIDForBlock, used for sub-chunks
// and block updates, and DependsOnNeighbours for re-sending a block when a neighbour changes.

// connectionsGetter is block.Fence and block.Thin.
type connectionsGetter interface {
	GetConnections() map[math.Facing]bool
}

type stairShapeGetter interface {
	GetShape() blockutils.StairShape
}

type sideGetter interface {
	GetSide(side math.Facing, step int) block.Behavior
}

// derivedNetworkProperties returns the 1.26.50 properties of b worked out from its neighbours (b
// must have been read from the world), or nil if b has none.
func derivedNetworkProperties(b block.Behavior) map[string]any {
	switch v := b.(type) {
	case connectionsGetter:
		return connectionProperties(v.GetConnections())
	case stairShapeGetter:
		corner := ids.MC_CORNER_NONE
		switch v.GetShape() {
		case blockutils.StairShapeInnerLeft:
			corner = ids.MC_CORNER_INNER_LEFT
		case blockutils.StairShapeInnerRight:
			corner = ids.MC_CORNER_INNER_RIGHT
		case blockutils.StairShapeOuterLeft:
			corner = ids.MC_CORNER_OUTER_LEFT
		case blockutils.StairShapeOuterRight:
			corner = ids.MC_CORNER_OUTER_RIGHT
		}
		return map[string]any{ids.MC_CORNER: corner}
	case *block.Tripwire:
		// Tripwire connects to other tripwire (as Dragonfly works it out; PocketMine-MP's Tripwire
		// has no connections).
		connections := map[math.Facing]bool{}
		if sides, ok := b.(sideGetter); ok && b.GetPosition().IsValid() {
			for _, facing := range math.HorizontalFacing {
				if _, ok := sides.GetSide(facing, 1).(*block.Tripwire); ok {
					connections[facing] = true
				}
			}
		}
		return connectionProperties(connections)
	}
	return nil
}

func connectionProperties(connections map[math.Facing]bool) map[string]any {
	return map[string]any{
		ids.MC_CONNECTION_NORTH: boolByte(connections[math.North]),
		ids.MC_CONNECTION_EAST:  boolByte(connections[math.East]),
		ids.MC_CONNECTION_SOUTH: boolByte(connections[math.South]),
		ids.MC_CONNECTION_WEST:  boolByte(connections[math.West]),
	}
}

// boolByte is how the palette stores boolean properties (byte tags).
func boolByte(b bool) any {
	if b {
		return uint8(1)
	}
	return uint8(0)
}

// DependsOnNeighbours reports whether the network state of internalStateID has 1.26.50
// properties worked out from the neighbours (see NetworkIDForBlock).
func (t *BlockTranslator) DependsOnNeighbours(internalStateID int) bool {
	t.mu.RLock()
	derived, ok := t.derivedCache[internalStateID]
	t.mu.RUnlock()
	if ok {
		return derived
	}
	states := t.InternalIDToNetworkStateData(internalStateID).States
	_, derived = states[ids.MC_CORNER]
	if _, ok := states[ids.MC_CONNECTION_NORTH]; ok {
		derived = true
	}
	t.mu.Lock()
	t.derivedCache[internalStateID] = derived
	t.mu.Unlock()
	return derived
}

// NetworkIDForBlock is InternalIDToNetworkID for a block read from the world
// (World::getBlockAt): for fences, panes, bars, tripwire and stairs, the network state also gets
// the connections or corner the 1.26.50 client needs.
func (t *BlockTranslator) NetworkIDForBlock(b block.Behavior) int32 {
	stateID := b.GetStateId()
	networkID := t.InternalIDToNetworkID(stateID)
	if !t.DependsOnNeighbours(stateID) {
		return networkID
	}
	props := derivedNetworkProperties(b)
	if len(props) == 0 {
		return networkID
	}

	key := derivedKey(networkID, props)
	t.mu.RLock()
	id, ok := t.derivedIDCache[key]
	t.mu.RUnlock()
	if ok {
		return id
	}

	data := bedrock.BlockStates()[networkID]
	states := make(map[string]any, len(data.States))
	for k, v := range data.States {
		states[k] = v
	}
	for k, v := range props {
		states[k] = normalizeLike(data.States[k], v)
	}
	id, found := bedrock.RuntimeIDFor(data.Name, states)
	if !found {
		id = networkID
	}
	t.mu.Lock()
	t.derivedIDCache[key] = id
	t.mu.Unlock()
	return id
}

// normalizeLike converts v to the Go type the palette uses for the property (the vendored palette
// decodes byte tags as uint8 and string tags as string).
func normalizeLike(existing, v any) any {
	switch existing.(type) {
	case uint8:
		switch b := v.(type) {
		case uint8:
			return b
		case bool:
			if b {
				return uint8(1)
			}
			return uint8(0)
		}
	case bool:
		if b, ok := v.(uint8); ok {
			return b != 0
		}
	}
	return v
}

func derivedKey(networkID int32, props map[string]any) string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d", networkID)
	for _, k := range keys {
		fmt.Fprintf(&sb, ";%s=%v", k, props[k])
	}
	return sb.String()
}
