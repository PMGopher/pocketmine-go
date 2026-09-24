package entity

// NeverSavedWithChunkEntity is a port of pocketmine\entity\NeverSavedWithChunkEntity: entities
// implementing it are never saved to disk with their chunk (and so needn't be registered with
// EntityFactory), e.g. players and fireworks. PHP's is an empty marker interface; Go needs a
// method to make the marker explicit.
type NeverSavedWithChunkEntity interface {
	NeverSavedWithChunk()
}
