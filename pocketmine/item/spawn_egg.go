package item

// SpawnEgg is a port of the abstract pocketmine\item\SpawnEgg. PHP registers one anonymous
// subclass per mob whose createEntity constructs that mob; here the mob is identified by
// EntityTypeName and CreateSpawnEggEntityFunc constructs it.
type SpawnEgg struct {
	ItemBase

	// EntityTypeName is the PHP entity class the egg spawns ("Zombie", "Squid", "Villager").
	EntityTypeName string
}

func NewSpawnEgg(identifier ItemIdentifier, name, entityTypeName string) *SpawnEgg {
	s := &SpawnEgg{EntityTypeName: entityTypeName}
	s.Init(s, identifier, name)
	return s
}

func (s *SpawnEgg) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}
