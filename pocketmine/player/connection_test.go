package player

import (
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
)

func TestChatBroadcastsWithTheStandardFormatter(t *testing.T) {
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)
	p.DoFirstSpawn()

	if !p.Chat("hello") {
		t.Fatal("Chat returned false")
	}
	if len(server.broadcast) != 2 { // the join message, then the chat message
		t.Fatalf("broadcast %d messages, want 2", len(server.broadcast))
	}
	tr, ok := server.broadcast[1].(*lang.Translatable)
	if !ok || tr.Text() != "chat.type.text" || tr.Parameter(0) != "Steve" || tr.Parameter(1) != "hello" {
		t.Errorf("broadcast %#v, want chat.type.text [Steve hello]", server.broadcast[0])
	}
	if len(session.messages) != 2 {
		t.Errorf("sender received %d messages, want its join and its own message", len(session.messages))
	}
}

func TestChatRunsCommandsAndLimitsMessagesPerTick(t *testing.T) {
	server := &fakeServer{}
	p, _ := newConnectedPlayer(t, server)

	p.Chat("/list")
	p.Chat("./tp") // "./" is stripped to "/", so this is a command too (Player::chat)
	if len(server.commands) != 2 || server.commands[0] != "list" || server.commands[1] != "tp" {
		t.Errorf("dispatched commands %v, want [list tp]", server.commands)
	}
	if len(server.broadcast) != 0 {
		t.Errorf("broadcast %v, want nothing (both were commands)", server.broadcast)
	}

	// Player::$messageCounter allows 2 messages per tick.
	if p.Chat("third") {
		t.Error("a third message in the same tick was accepted")
	}
	p.OnUpdate(p.LastUpdate + 1)
	if !p.Chat("next tick") {
		t.Error("a message after the next tick was rejected")
	}
}

func TestChatIgnoresBlankAndOverlongParts(t *testing.T) {
	server := &fakeServer{}
	p, _ := newConnectedPlayer(t, server)
	p.Chat("   ")
	p.Chat(strings.Repeat("a", MaxChatCharLength+1))
	if len(server.broadcast) != 0 {
		t.Errorf("broadcast %v, want nothing", server.broadcast)
	}
}

func TestSelectHotbarSlot(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	if !p.SelectHotbarSlot(4) || p.GetInventory().GetHeldItemIndex() != 4 {
		t.Errorf("SelectHotbarSlot(4) didn't select slot 4 (held %d)", p.GetInventory().GetHeldItemIndex())
	}
	if p.SelectHotbarSlot(9) {
		t.Error("SelectHotbarSlot(9) accepted a slot outside the hotbar")
	}
}

func TestToggleFlightNeedsAllowFlight(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	if p.ToggleFlight(true) || p.IsFlying() {
		t.Error("a survival player without allowFlight could start flying")
	}
	p.SetAllowFlight(true)
	if !p.ToggleFlight(true) || !p.IsFlying() {
		t.Error("a player with allowFlight couldn't start flying")
	}
}

func TestToggleSneakRejectsANoOpPressChange(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	if !p.ToggleSneak(true, true) || !p.IsSneaking() {
		t.Fatal("ToggleSneak(true, true) didn't start sneaking")
	}
	// Same sneak state, different press state: PHP cancels the event (mismatch).
	if p.ToggleSneak(true, false) {
		t.Error("ToggleSneak(true, false) while sneaking returned true, want false")
	}
}

func TestSaveDataRoundTrip(t *testing.T) {
	w := newTestWorld(t)
	p := newTestPlayerIn(t, w, math.NewVector3(10.5, 70, -3.5))
	p.SetGamemode(GameModeCreative)
	p.firstPlayed = 1234
	p.SetHealth(7)
	tag := p.GetSaveData()

	if got := tag.GetStringOr(TagLevel, ""); string(got) != w.GetFolderName() {
		t.Errorf("Level = %q, want %q", got, w.GetFolderName())
	}

	loaded, _ := newPlayerFor(t, &fakeServer{gamemode: GameModeSurvival}, w, p.GetLocation().Vector3, tag)
	if loaded.GetGamemode() != GameModeCreative {
		t.Errorf("game mode = %v, want the saved creative mode", loaded.GetGamemode())
	}
	if first, _ := loaded.GetFirstPlayed(); first != 1234 {
		t.Errorf("firstPlayed = %d, want 1234", first)
	}
	if loaded.GetHealth() != 7 {
		t.Errorf("health = %v, want 7", loaded.GetHealth())
	}

	forced, _ := newPlayerFor(t, &fakeServer{gamemode: GameModeSurvival, force: true}, w, p.GetLocation().Vector3, tag)
	if forced.GetGamemode() != GameModeSurvival {
		t.Errorf("with force-gamemode the game mode = %v, want the server's survival", forced.GetGamemode())
	}
}

func TestDoFirstSpawnAndOnPostDisconnect(t *testing.T) {
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)
	other, otherSession := newConnectedPlayer(t, server)
	other.DoFirstSpawn()

	p.DoFirstSpawn()
	if !p.IsSpawned() || p.NoDamageTicks != 60 {
		t.Errorf("after DoFirstSpawn: spawned=%v noDamageTicks=%d, want true/60", p.IsSpawned(), p.NoDamageTicks)
	}
	joined, ok := server.broadcast[0].(*lang.Translatable)
	if !ok || !strings.HasSuffix(joined.Text(), "multiplayer.player.joined") {
		t.Errorf("join message %v, want multiplayer.player.joined", server.broadcast[0])
	}

	before := len(otherSession.messages)
	session.disconnected = true
	p.OnPostDisconnect("client disconnect", nil)
	if len(otherSession.messages) != before+1 {
		t.Error("the other player didn't get the leave message")
	}
	if _, ok := server.saved["Steve"]; !ok {
		t.Error("player data wasn't saved on disconnect")
	}
	if len(server.removed) != 1 || server.removed[0] != p {
		t.Error("the player wasn't removed from the online list")
	}
	if !p.IsFlaggedForDespawn() || p.IsSpawned() {
		t.Error("the player wasn't despawned")
	}
}

func TestSetUsingItemUsesTheServerTick(t *testing.T) {
	server := &fakeServer{tick: 100}
	p, _ := newConnectedPlayer(t, server)
	p.SetUsingItem(true)
	server.tick = 130
	if !p.IsUsingItem() || p.GetItemUseDuration() != 30 {
		t.Errorf("using=%v duration=%d, want true/30", p.IsUsingItem(), p.GetItemUseDuration())
	}
	p.SetUsingItem(false)
	if p.IsUsingItem() || p.GetItemUseDuration() != -1 {
		t.Error("SetUsingItem(false) didn't stop item use")
	}
}

func TestChangedMetadataReachesThePlayerItself(t *testing.T) {
	// Entity::entityBaseTick calls $this->sendData(null, ...), which Player overrides to include
	// itself; the base Entity method must dispatch to that override.
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)
	p.DoFirstSpawn()
	p.OnUpdate(p.LastUpdate + 1)
	session.packets = nil

	p.SetSneaking(true)
	p.OnUpdate(p.LastUpdate + 1)

	for _, pk := range session.packets {
		if data, ok := pk.(*packet.SetActorData); ok && data.EntityRuntimeID == uint64(p.GetID()) {
			return
		}
	}
	t.Error("the player didn't receive its own SetActorData after its metadata changed")
}

func TestKnockBackIsSentToThePlayerItself(t *testing.T) {
	// Living::knockBack calls $this->setMotion(), which Player overrides to send SetActorMotion to
	// its own client.
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)
	session.packets = nil

	p.KnockBack(1, 0, 0.4, 0.4)

	for _, pk := range session.packets {
		if m, ok := pk.(*packet.SetActorMotion); ok && m.EntityRuntimeID == uint64(p.GetID()) {
			return
		}
	}
	t.Error("the knocked-back player didn't receive SetActorMotion")
}

func TestDoChunkRequestsFollowsNextChunkOrderRun(t *testing.T) {
	// Player::doChunkRequests: chunks are reordered (and the view area centre synced) only when
	// nextChunkOrderRun is due - right after the view distance changes, not on every tick.
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)
	p.SetViewDistance(4)

	p.DoChunkRequests()
	if len(session.startedChunks) == 0 {
		t.Fatal("no chunks after setting the view distance")
	}
	if session.viewAreaSyncs != 1 {
		t.Fatalf("view area synced %d times after the first reorder, want 1", session.viewAreaSyncs)
	}
	for range 10 {
		p.DoChunkRequests()
	}
	if session.viewAreaSyncs != 1 {
		t.Errorf("view area synced %d times while standing still, want still 1", session.viewAreaSyncs)
	}
}
