package player

// ResourcePack is the surface PlayerResourcePackOfferEvent needs from
// pocketmine\resourcepacks\ResourcePack.
type ResourcePack interface {
	GetPackId() string
}

// PlayerResourcePackOfferEvent is a port of pocketmine\event\player\PlayerResourcePackOfferEvent:
// called after a player authenticates and is being offered resource packs to download. This
// event should be used to decide which resource packs the client should be sent.
type PlayerResourcePackOfferEvent struct {
	playerInfo     PlayerInfo
	resourcePacks  []ResourcePack
	encryptionKeys map[string]string
	mustAccept     bool
}

// NewPlayerResourcePackOfferEvent creates the event; encryptionKeys maps pack UUID to key.
func NewPlayerResourcePackOfferEvent(playerInfo PlayerInfo, resourcePacks []ResourcePack, encryptionKeys map[string]string, mustAccept bool) *PlayerResourcePackOfferEvent {
	if encryptionKeys == nil {
		encryptionKeys = map[string]string{}
	}
	return &PlayerResourcePackOfferEvent{playerInfo: playerInfo, resourcePacks: resourcePacks, encryptionKeys: encryptionKeys, mustAccept: mustAccept}
}

func (e *PlayerResourcePackOfferEvent) GetPlayerInfo() PlayerInfo { return e.playerInfo }

// AddResourcePack adds a resource pack to the top of the stack. The resources in this pack will
// be applied over the top of any existing packs. An empty encryptionKey means none.
func (e *PlayerResourcePackOfferEvent) AddResourcePack(entry ResourcePack, encryptionKey string) {
	e.resourcePacks = append([]ResourcePack{entry}, e.resourcePacks...)
	if encryptionKey != "" {
		e.encryptionKeys[entry.GetPackId()] = encryptionKey
	}
}

// SetResourcePacks replaces the list of packs offered to the player. The first pack in the list
// is the one applied on top.
func (e *PlayerResourcePackOfferEvent) SetResourcePacks(resourcePacks []ResourcePack, encryptionKeys map[string]string) {
	e.resourcePacks = resourcePacks
	if encryptionKeys == nil {
		encryptionKeys = map[string]string{}
	}
	e.encryptionKeys = encryptionKeys
}

func (e *PlayerResourcePackOfferEvent) GetResourcePacks() []ResourcePack { return e.resourcePacks }

func (e *PlayerResourcePackOfferEvent) GetEncryptionKeys() map[string]string {
	return e.encryptionKeys
}

func (e *PlayerResourcePackOfferEvent) SetMustAccept(mustAccept bool) { e.mustAccept = mustAccept }

func (e *PlayerResourcePackOfferEvent) MustAccept() bool { return e.mustAccept }
