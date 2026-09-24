package bedrock

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// The data behind pocketmine\network\mcpe\cache\StaticPacketCache: the two packets
// PreSpawnPacketHandler sends every player right after StartGame and ItemRegistry.

// entityIdentifiersData is assets/entity_identifiers.nbt, vendored unchanged from pmmp/BedrockData
// tag 6.7.0+bedrock-1.26.30 - the same file StaticPacketCache::getAvailableActorIdentifiers sends
// as-is (network NBT).
//
//go:embed assets/entity_identifiers.nbt
var entityIdentifiersData []byte

// biomeDefinitionsData is assets/biome_definitions.bin: an encoded BiomeDefinitionList packet
// payload for Bedrock 1.26.50. PHP's StaticPacketCache::getBiomeDefs builds it from BedrockData's
// biome_definitions.json, which has no 1.26.50 release yet; this is the packet df-mc/dragonfly (MIT,
// see assets/LICENSE-dragonfly) sends to 1.26.50 clients, captured from the wire.
//
//go:embed assets/biome_definitions.bin
var biomeDefinitionsData []byte

var (
	biomeDefinitionsOnce sync.Once
	biomeDefinitions     *packet.BiomeDefinitionList
)

// AvailableActorIdentifiers is a port of StaticPacketCache::getAvailableActorIdentifiers.
func AvailableActorIdentifiers() *packet.AvailableActorIdentifiers {
	return &packet.AvailableActorIdentifiers{SerialisedEntityIdentifiers: entityIdentifiersData}
}

// BiomeDefinitionList is a port of StaticPacketCache::getBiomeDefs.
func BiomeDefinitionList() *packet.BiomeDefinitionList {
	biomeDefinitionsOnce.Do(func() {
		pk := &packet.BiomeDefinitionList{}
		pk.Marshal(protocol.NewReader(bytes.NewReader(biomeDefinitionsData), 0, false))
		biomeDefinitions = pk
	})
	return biomeDefinitions
}
