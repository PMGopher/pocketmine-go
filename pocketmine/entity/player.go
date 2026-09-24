package entity

import (
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/world"
)

// Player is the surface entity classes need from pocketmine\player\Player (pickups, collision
// callbacks, game-mode checks). pocketmine/player imports this package, so it's a local interface;
// *player.Player satisfies it.
type Player interface {
	world.Entity
	world.EntityViewer

	GetInventory() *inventory.PlayerInventory
	GetOffHandInventory() *inventory.PlayerOffHandInventory

	// HasFiniteResources returns whether the player's game mode consumes items (survival and
	// adventure).
	HasFiniteResources() bool
	IsCreative() bool
	IsSpectator() bool
}

// AsPlayer narrows an entity (or viewer) to a Player - PHP's `instanceof Player`.
func AsPlayer(v any) (Player, bool) {
	p, ok := v.(Player)
	return p, ok
}
