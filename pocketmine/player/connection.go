package player

import (
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/utils"
)

// Translation keys used for join/leave messages (KnownTranslationKeys; the generated
// KnownTranslationFactory isn't ported, see player/chat's chatTypeTextKey).
const (
	translationPlayerJoined = "multiplayer.player.joined"
	translationPlayerLeft   = "multiplayer.player.left"
)

// DoFirstSpawn is a port of Player::doFirstSpawn: the player appears in the world once the client
// has finished loading. Not ported: broadcast-channel permission rechecks (players aren't
// permissibles yet), the cancellable PlayerJoinEvent, and the "quit while dead" respawn (respawning
// isn't ported).
func (p *Player) DoFirstSpawn() {
	if p.spawned {
		return
	}

	joinMessage := lang.NewTranslatable(translationPlayerJoined, []any{p.GetDisplayName()}).Prefix(utils.Yellow)
	if p.server != nil {
		p.server.BroadcastMessage(joinMessage, nil)
	}

	p.NoDamageTicks = 60

	p.SetSpawned(true) // $this->spawned = true; $this->spawnToAll();
}

// GetLeaveMessage is a port of Player::getLeaveMessage.
func (p *Player) GetLeaveMessage() any {
	if p.spawned {
		return lang.NewTranslatable(translationPlayerLeft, []any{p.GetDisplayName()}).Prefix(utils.Yellow)
	}
	return ""
}

// OnPostDisconnect is a port of Player::onPostDisconnect: everything that happens to the player
// after its connection is gone. Not ported: window/inventory cleanup (no windows yet), the
// cancellable PlayerQuitEvent, sleeping and hidden players.
func (p *Player) OnPostDisconnect() {
	quitMessage := p.GetLeaveMessage()
	if p.server != nil && quitMessage != "" {
		//prevent the player receiving their own disconnect message
		// (PHP: unsubscribeFromAllBroadcastChannels)
		var recipients []*Player
		for _, other := range p.server.GetOnlinePlayers() {
			if other != p {
				recipients = append(recipients, other)
			}
		}
		p.server.BroadcastMessage(quitMessage, recipients)
	}
	p.Save()

	p.spawned = false

	p.blockBreakHandler = nil
	p.DespawnFromAll()

	if p.server != nil {
		p.server.RemoveOnlinePlayer(p)
	}

	if p.GetWorld() != nil {
		for pos := range p.usedChunks {
			p.unloadChunk(pos[0], pos[1])
		}
	}
	p.loadQueue = map[[2]int]bool{}
	p.loadQueueOrder = nil

	p.FlagForDespawn()
}

// Save is a port of Player::save.
func (p *Player) Save() {
	if p.server != nil {
		p.server.SaveOfflinePlayerData(p.username, p.GetSaveData())
	}
}
