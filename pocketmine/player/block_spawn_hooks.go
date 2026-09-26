package player

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world"
)

// init sets the block package's spawn hooks, which RespawnAnchor::onInteract uses.
func init() {
	block.PlayerSpawnFunc = func(p block.Player) (block.Position, bool) {
		pl, ok := p.(*Player)
		if !ok {
			return block.Position{}, false
		}
		spawn := pl.GetSpawn()
		if spawn.World == nil {
			return block.Position{}, false
		}
		return block.NewPosition(spawn.X, spawn.Y, spawn.Z, spawn.World), true
	}
	block.SetPlayerSpawnFunc = func(p block.Player, pos block.Position) {
		pl, ok := p.(*Player)
		if !ok {
			return
		}
		w, _ := pos.GetWorld()
		ww, _ := w.(*world.World)
		v := pos.Vector3
		pl.SetSpawn(&v, ww)
	}
}
