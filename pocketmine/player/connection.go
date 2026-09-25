package player

import (
	"fmt"
	"reflect"

	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/form"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// DoFirstSpawn is a port of Player::doFirstSpawn: called by the network system when the pre-spawn
// sequence is completed (e.g. after sending spawn chunks). This fires join events and broadcasts
// join messages to other online players.
func (p *Player) DoFirstSpawn() {
	if p.spawned {
		return
	}
	p.spawned = true
	p.recheckBroadcastPermissions()
	p.GetPermissionRecalculationCallbacks().Add(ptrTo(permission.PermissionRecalculationCallback(func(changedPermissionsOldValues map[string]bool) {
		_, admin := changedPermissionsOldValues[BroadcastChannelAdministrative]
		_, users := changedPermissionsOldValues[BroadcastChannelUsers]
		if admin || users {
			p.recheckBroadcastPermissions()
		}
	})))

	ev := playerevent.NewPlayerJoinEvent(p, lang.KnownTranslationFactory.MultiplayerPlayerJoined(p.GetDisplayName()).Prefix(utils.Yellow))
	event.Call(ev)
	if msg := ev.GetJoinMessage(); msg != "" && msg != nil {
		p.server.BroadcastMessage(msg, nil)
	}

	p.NoDamageTicks = 60

	p.SpawnToAll()

	if p.GetHealth() <= 0 {
		p.logger.Debug("Quit while dead, forcing respawn")
		p.actuallyRespawn()
	}
}

// SendTitle is a port of Player::sendTitle: adds a title text to the user's screen, with an
// optional subtitle. Durations are in ticks; -1 uses the client's defaults.
func (p *Player) SendTitle(title, subtitle string, fadeIn, stay, fadeOut int) {
	p.SetTitleDuration(fadeIn, stay, fadeOut)
	if subtitle != "" {
		p.SendSubTitle(subtitle)
	}
	p.GetNetworkSession().OnTitle(title)
}

// SendSubTitle sets the subtitle message, without sending a title.
func (p *Player) SendSubTitle(subtitle string) { p.GetNetworkSession().OnSubTitle(subtitle) }

// SendActionBarMessage adds small text to the user's screen.
func (p *Player) SendActionBarMessage(message string) { p.GetNetworkSession().OnActionBar(message) }

// RemoveTitles removes the title from the client's screen.
func (p *Player) RemoveTitles() { p.GetNetworkSession().OnClearTitle() }

// ResetTitles resets the title duration settings to defaults and removes any existing titles.
func (p *Player) ResetTitles() { p.GetNetworkSession().OnResetTitleOptions() }

// SetTitleDuration sets the title duration (fade-in, stay and fade-out time in ticks).
func (p *Player) SetTitleDuration(fadeIn, stay, fadeOut int) {
	if fadeIn >= 0 && stay >= 0 && fadeOut >= 0 {
		p.GetNetworkSession().OnTitleDuration(fadeIn, stay, fadeOut)
	}
}

// SendMessage sends a direct chat message to a player (a string or *lang.Translatable).
func (p *Player) SendMessage(message any) { p.GetNetworkSession().OnChatMessage(message) }

// SendJukeboxPopup is a port of Player::sendJukeboxPopup.
func (p *Player) SendJukeboxPopup(message any) { p.GetNetworkSession().OnJukeboxPopup(message) }

// SendPopup sends a popup message to the player.
func (p *Player) SendPopup(message string) { p.GetNetworkSession().OnPopup(message) }

// SendTip is a port of Player::sendTip.
func (p *Player) SendTip(message string) { p.GetNetworkSession().OnTip(message) }

// SendToastNotification sends a toast message to the player, or queue to send it if a toast
// message is already shown.
func (p *Player) SendToastNotification(title, body string) {
	p.GetNetworkSession().OnToastNotification(title, body)
}

// SendForm sends a Form to the player, or queue to send it if a form is already open.
func (p *Player) SendForm(f form.Form) {
	id := p.formIDCounter
	p.formIDCounter++
	if p.GetNetworkSession().OnFormSent(id, f) {
		p.forms[id] = f
	}
}

// OnFormSubmit is a port of Player::onFormSubmit.
func (p *Player) OnFormSubmit(formID int, responseData any) bool {
	f, ok := p.forms[formID]
	if !ok {
		p.logger.Debug(fmt.Sprintf("Got unexpected response for form %d", formID))
		return false
	}
	defer delete(p.forms, formID)

	if err := f.HandleResponse(p, responseData); err != nil {
		if _, ok := err.(*form.FormValidationError); ok {
			p.logger.Critical(fmt.Sprintf("Failed to validate form %s: %s", reflect.TypeOf(f), err.Error()))
		} else {
			p.logger.Error(fmt.Sprintf("Error handling form %s response: %s", reflect.TypeOf(f), err.Error()))
		}
	}
	return true
}

// HasPendingForm returns whether the server is waiting for a response for a form with the given
// ID.
func (p *Player) HasPendingForm(formID int) bool {
	_, ok := p.forms[formID]
	return ok
}

// CloseAllForms closes the current viewing form and forms in queue.
func (p *Player) CloseAllForms() { p.GetNetworkSession().OnCloseAllForms() }

// Transfer is a port of Player::transfer: transfers a player to another server. message (nil for
// the default) is shown in the console when closing the player.
func (p *Player) Transfer(address string, port int, message any) bool {
	if message == nil {
		message = lang.KnownTranslationFactory.PocketmineDisconnectTransfer()
	}
	ev := playerevent.NewPlayerTransferEvent(p, address, port, message)
	event.Call(ev)
	if !ev.IsCancelled() {
		p.GetNetworkSession().Transfer(ev.GetAddress(), ev.GetPort(), ev.GetMessage())
		return true
	}
	return false
}

// isEmptyMessage reports whether a Translatable|string message is "".
func isEmptyMessage(message any) bool {
	s, ok := message.(string)
	return message == nil || (ok && s == "")
}

// Kick is a port of Player::kick: kicks a player from the server. reason is shown in the server
// log; quitMessage (nil for the default) is broadcast to online players; disconnectScreenMessage
// (nil to use the reason) is shown on the player's disconnection screen.
func (p *Player) Kick(reason, quitMessage, disconnectScreenMessage any) bool {
	if reason == nil {
		reason = ""
	}
	if quitMessage == nil {
		quitMessage = p.GetLeaveMessage()
	}
	ev := playerevent.NewPlayerKickEvent(p, reason, quitMessage, disconnectScreenMessage)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	reason = ev.GetDisconnectReason()
	if isEmptyMessage(reason) {
		reason = lang.KnownTranslationFactory.DisconnectionScreenNoReason()
	}
	disconnectScreenMessage = ev.GetDisconnectScreenMessage()
	if disconnectScreenMessage == nil {
		disconnectScreenMessage = reason
	}
	if isEmptyMessage(disconnectScreenMessage) {
		disconnectScreenMessage = lang.KnownTranslationFactory.DisconnectionScreenNoReason()
	}
	p.Disconnect(reason, ev.GetQuitMessage(), disconnectScreenMessage)
	return true
}

// Disconnect is a port of Player::disconnect: removes the player from the server. This cannot be
// cancelled. Prefer Kick instead. quitMessage nil uses the default; disconnectScreenMessage nil
// uses the reason.
func (p *Player) Disconnect(reason, quitMessage, disconnectScreenMessage any) {
	if !p.IsConnected() {
		return
	}
	if disconnectScreenMessage == nil {
		disconnectScreenMessage = reason
	}
	p.GetNetworkSession().OnPlayerDestroyed(reason, disconnectScreenMessage)
	p.OnPostDisconnect(reason, quitMessage)
}

// OnPostDisconnect is a port of Player::onPostDisconnect: executes post-disconnect actions and
// cleanups. quitMessage nil uses the default.
func (p *Player) OnPostDisconnect(reason, quitMessage any) {
	if p.IsConnected() {
		panic("Player is still connected")
	}

	//prevent the player receiving their own disconnect message
	p.server.UnsubscribeFromAllBroadcastChannels(p)

	p.RemoveCurrentWindow()

	if quitMessage == nil {
		quitMessage = p.GetLeaveMessage()
	}
	ev := playerevent.NewPlayerQuitEvent(p, quitMessage, reason)
	event.Call(ev)
	if msg := ev.GetQuitMessage(); !isEmptyMessage(msg) {
		p.server.BroadcastMessage(msg, nil)
	}
	p.Save()

	p.spawned = false

	p.StopSleep()
	p.setBlockBreakHandler(nil)
	p.DespawnFromAll()

	p.server.RemoveOnlinePlayer(p)

	for _, player := range p.server.GetOnlinePlayers() {
		if !player.CanSee(p) {
			player.ShowPlayer(p)
		}
	}
	p.hiddenPlayers = nil

	if loc := p.GetLocation(); loc.IsValid() {
		for index := range p.usedChunks {
			p.unloadChunk(index[0], index[1], nil)
		}
	}
	if len(p.usedChunks) != 0 {
		panic("Previous loop should have cleared this array")
	}
	p.loadQueue = map[[2]int]bool{}
	p.loadQueueOrder = nil

	p.RemoveCurrentWindow()
	p.removePermanentInventories()

	p.GetPermissionRecalculationCallbacks().Clear()

	p.FlagForDespawn()
}

// OnDispose is a port of Player::onDispose.
func (p *Player) OnDispose() {
	p.Disconnect("Player destroyed", nil, nil)
	p.cursorInventory.RemoveAllViewers()
	if grid, ok := p.craftingGrid.(interface{ RemoveAllViewers() }); ok {
		grid.RemoveAllViewers()
	}
	p.Human.OnDispose()
}

// DestroyCycles is a port of Player::destroyCycles.
func (p *Player) DestroyCycles() {
	p.networkSession = nil
	p.spawnPosition = nil
	p.deathPosition = nil
	p.blockBreakHandler = nil
	p.Permissible.Close()
	p.Human.DestroyCycles()
}

func ptrTo[T any](v T) *T { return &v }
