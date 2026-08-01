package block

import "pocketmine-go/pocketmine/entity"

// Compile-time proof that entity.Entity/entity.Living satisfy this package's local
// forward-compatible Entity/Living interfaces (declared in world.go long before the entity
// package existed) - the whole point of that pattern paying off.
var (
	_ Entity = (*entity.Entity)(nil)
	_ Living = (*entity.Living)(nil)
)
