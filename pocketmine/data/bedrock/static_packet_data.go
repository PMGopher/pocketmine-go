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

// entityIdentifiersData is assets/entity_identifiers.nbt: the network NBT
// StaticPacketCache::getAvailableActorIdentifiers sends as-is. pmmp/BedrockData's file is for
// 1.26.30 (136 entries with 1.26.30 runtime IDs); a 1.26.51 client that plays on Dragonfly left this
// server right after loading while it sent that list, so this is the list Dragonfly sends to
// 1.26.50 clients (MIT, see assets/LICENSE-dragonfly), captured from the wire.
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

// creativeContentData is an encoded CreativeContent packet payload for Bedrock 1.26.50, the source
// of the vanilla creative items (convert's CreativeInventory loader). PHP loads them from
// BedrockData's creative files, which have no 1.26.50 release yet, so this is the packet Dragonfly
// sends (MIT, see assets/LICENSE-dragonfly), captured from the wire. It uses the item network IDs
// of required_item_list.json, which comes from the same source.
//
//go:embed assets/creative_content.bin
var creativeContentData []byte

var (
	biomeDefinitionsOnce sync.Once
	biomeDefinitions     *packet.BiomeDefinitionList

	creativeContentOnce sync.Once
	creativeContent     *packet.CreativeContent
)

// CreativeContent is the captured 1.26.50 creative inventory packet (see creativeContentData).
func CreativeContent() *packet.CreativeContent {
	creativeContentOnce.Do(func() {
		pk := &packet.CreativeContent{}
		pk.Marshal(protocol.NewReader(bytes.NewReader(creativeContentData), 0, false))
		creativeContent = pk
	})
	return creativeContent
}

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
