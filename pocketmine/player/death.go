package player

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// Kill is a port of Player::kill (Living::kill): players that haven't spawned yet can't die.
func (p *Player) Kill() {
	if !p.spawned {
		return
	}
	p.Human.Kill()
}

// OnDeath is a port of Player::onDeath.
func (p *Player) OnDeath() {
	//Crafting grid must always be evacuated even if keep-inventory is true. This dumps the contents into the
	//main inventory and drops the rest on the ground.
	p.RemoveCurrentWindow()

	pos := p.GetPosition()
	p.SetDeathPosition(&pos, nil)

	drops := p.GetDrops()
	eventDrops := make([]entityevent.Item, len(drops))
	for i, d := range drops {
		eventDrops[i] = d
	}
	ev := playerevent.NewPlayerDeathEvent(p, eventDrops, p.GetXpDropAmount(), nil, p.GetLastDamageCause())
	event.Call(ev)

	loc := p.GetLocation()
	if !ev.GetKeepInventory() {
		for _, d := range ev.GetDrops() {
			if it, ok := d.(item.Item); ok {
				p.GetWorld().DropItem(loc.Vector3, it, nil, 10)
			}
		}

		clearInventory := func(inv inventory.Inventory) {
			kept := map[int]item.Item{}
			for slot, it := range inv.GetContents(false) {
				if it.KeepOnDeath() {
					kept[slot] = it
				}
			}
			inv.SetContents(kept)
		}
		p.GetInventory().SetHeldItemIndex(0)
		clearInventory(p.GetInventory())
		clearInventory(p.GetArmorInventory())
		clearInventory(p.GetOffHandInventory())
	}

	if !ev.GetKeepXp() {
		p.GetWorld().DropExperience(loc.Vector3, ev.GetXpDropAmount())
		level, progress := 0, 0.0
		p.GetXpManager().SetXpAndProgress(&level, &progress)
	}

	if msg := ev.GetDeathMessage(); !isEmptyMessage(msg) {
		p.server.BroadcastMessage(msg, nil)
	}

	p.StartDeathAnimation()

	p.GetNetworkSession().OnServerDeath(ev.GetDeathScreenMessage())
}

// OnDeathUpdate is a port of Player::onDeathUpdate: players are never flagged for despawn.
func (p *Player) OnDeathUpdate(tickDiff int) bool {
	p.Human.OnDeathUpdate(tickDiff)
	return false //never flag players for despawn
}

// Respawn is a port of Player::respawn.
func (p *Player) Respawn() {
	if p.server.IsHardcore() {
		if p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectBan(lang.KnownTranslationFactory.PocketmineDisconnectBanHardcore()), nil, nil) { //this allows plugins to prevent the ban by cancelling PlayerKickEvent
			p.server.GetNameBans().AddBan(p.GetName(), "Died in hardcore mode", nil, "")
		}
		return
	}

	p.actuallyRespawn()
}

// respawnAnchor is block.RespawnAnchor as the respawn needs it.
type respawnAnchor interface {
	block.Behavior
	GetCharges() int
	SetCharges(charges int)
}

// actuallyRespawn is a port of Player::actuallyRespawn. World::requestSafeSpawn is synchronous in
// this port, so the respawn completes immediately.
func (p *Player) actuallyRespawn() {
	if p.respawnLocked {
		return
	}
	p.respawnLocked = true

	p.logger.Debug("Waiting for safe respawn position to be located")
	spawn := p.GetSpawn()
	safeSpawn := spawn.World.GetSafeSpawn(spawn.Vector3)

	if !p.IsConnected() {
		return
	}
	p.logger.Debug("Respawn position located, completing respawn")
	ev := playerevent.NewPlayerRespawnEvent(p, entityevent.Position{Vector3: safeSpawn, World: spawn.World})
	spawnPosition := ev.GetRespawnPosition()
	spawnWorld := spawnPosition.World.(*world.World)
	if anchor, ok := spawnWorld.GetBlock(spawnPosition.Vector3).(respawnAnchor); ok {
		if anchor.GetCharges() > 0 {
			anchor.SetCharges(anchor.GetCharges() - 1)
			_ = spawnWorld.SetBlock(block.NewPosition(spawnPosition.X, spawnPosition.Y, spawnPosition.Z, spawnWorld), anchor)
			spawnWorld.AddSound(spawnPosition.Vector3, sound.RespawnAnchorDepleteSound{})
		} else if defaultWorld := p.server.GetWorldManager().GetDefaultWorld(); defaultWorld != nil {
			defaultSpawn := defaultWorld.GetSpawnLocation()
			p.SetSpawn(&defaultSpawn, defaultWorld)
			ev.SetRespawnPosition(entityevent.Position{Vector3: defaultSpawn, World: defaultWorld})
			p.SendMessage(lang.KnownTranslationFactory.TileRespawnAnchorNotValid().Prefix(utils.Gray))
		}
	}
	event.Call(ev)

	respawnPosition := ev.GetRespawnPosition()
	realSpawn := respawnPosition.Add(0.5, 0, 0.5)
	p.TeleportTo(realSpawn, respawnPosition.World.(*world.World), nil, nil)

	p.SetSprinting(false)
	p.SetSneaking(false)
	p.SetFlying(false)

	p.ExtinguishWithCause(entityevent.ExtinguishCauseRespawn)
	p.SetAirSupplyTicks(p.GetMaxAirSupplyTicks())
	p.DeadTicks = 0
	p.NoDamageTicks = 60

	p.GetEffects().Clear()
	p.SetHealth(float64(p.GetMaxHealth()))

	for _, attr := range p.GetAttributeMap().GetAll() {
		if attr.GetID() == entity.AttributeExperience || attr.GetID() == entity.AttributeExperienceLevel { //we have already reset both of those if needed when the player died
			continue
		}
		attr.ResetToDefault()
	}

	p.SpawnToAll()
	p.ScheduleUpdate()

	p.GetNetworkSession().OnServerRespawn()
	p.respawnLocked = false
}

var _ = math.Vector3{}
