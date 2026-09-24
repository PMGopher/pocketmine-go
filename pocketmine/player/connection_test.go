package player

import (
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// fakeServer records what Player asks of the server.
type fakeServer struct {
	online    []*Player
	broadcast []any
	commands  []string
	saved     map[string]*nbt.CompoundTag
	removed   []*Player
	tick      int64
}

func (s *fakeServer) BroadcastMessage(message any, recipients []*Player) int {
	s.broadcast = append(s.broadcast, message)
	if recipients == nil {
		recipients = s.online
	}
	for _, p := range recipients {
		p.SendMessage(message)
	}
	return len(recipients)
}
func (s *fakeServer) DispatchCommand(sender *Player, commandLine string) bool {
	s.commands = append(s.commands, commandLine)
	return true
}
func (s *fakeServer) GetTick() int64              { return s.tick }
func (s *fakeServer) GetOnlinePlayers() []*Player { return s.online }
func (s *fakeServer) RemoveOnlinePlayer(p *Player) {
	s.removed = append(s.removed, p)
	for i, o := range s.online {
		if o == p {
			s.online = append(s.online[:i], s.online[i+1:]...)
		}
	}
}
func (s *fakeServer) SaveOfflinePlayerData(name string, data *nbt.CompoundTag) {
	if s.saved == nil {
		s.saved = map[string]*nbt.CompoundTag{}
	}
	s.saved[name] = data
}

// fakeSession records the packets and chat messages sent to a player.
type fakeSession struct {
	packets  []packet.Packet
	messages []any
}

func (s *fakeSession) SendDataPacket(pk packet.Packet) { s.packets = append(s.packets, pk) }
func (s *fakeSession) OnChatMessage(message any)       { s.messages = append(s.messages, message) }

func newConnectedPlayer(t *testing.T, server *fakeServer) (*Player, *fakeSession) {
	t.Helper()
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	session := &fakeSession{}
	p.SetServer(server)
	p.SetNetworkSession(session)
	server.online = append(server.online, p)
	return p, session
}

func TestChatBroadcastsWithTheStandardFormatter(t *testing.T) {
	server := &fakeServer{}
	p, session := newConnectedPlayer(t, server)

	if !p.Chat("hello") {
		t.Fatal("Chat returned false")
	}
	if len(server.broadcast) != 1 {
		t.Fatalf("broadcast %d messages, want 1", len(server.broadcast))
	}
	tr, ok := server.broadcast[0].(*lang.Translatable)
	if !ok || tr.Text() != "chat.type.text" || tr.Parameter(0) != "Steve" || tr.Parameter(1) != "hello" {
		t.Errorf("broadcast %#v, want chat.type.text [Steve hello]", server.broadcast[0])
	}
	if len(session.messages) != 1 {
		t.Errorf("sender received %d messages, want its own message", len(session.messages))
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
	p.SetFirstPlayed(1234)
	p.SetHealth(7)
	tag := p.GetSaveData()

	if got := tag.GetStringOr(TagLevel, ""); string(got) != w.GetFolderName() {
		t.Errorf("Level = %q, want %q", got, w.GetFolderName())
	}

	loaded := NewPlayerFromData("Steve", p.GetUniqueID(), "xuid-1", p.GetLocation(), GameModeSurvival, false, newTestSkin(t), tag, nil)
	if loaded.GetGamemode() != GameModeCreative {
		t.Errorf("game mode = %v, want the saved creative mode", loaded.GetGamemode())
	}
	if first, _ := loaded.GetFirstPlayed(); first != 1234 {
		t.Errorf("firstPlayed = %d, want 1234", first)
	}
	if loaded.GetHealth() != 7 {
		t.Errorf("health = %v, want 7", loaded.GetHealth())
	}

	forced := NewPlayerFromData("Steve", p.GetUniqueID(), "xuid-1", p.GetLocation(), GameModeSurvival, true, newTestSkin(t), tag, nil)
	if forced.GetGamemode() != GameModeSurvival {
		t.Errorf("with force-gamemode the game mode = %v, want the server's survival", forced.GetGamemode())
	}
}

func TestDoFirstSpawnAndOnPostDisconnect(t *testing.T) {
	server := &fakeServer{}
	p, _ := newConnectedPlayer(t, server)
	other, otherSession := newConnectedPlayer(t, server)
	_ = other

	p.DoFirstSpawn()
	if !p.IsSpawned() || p.NoDamageTicks != 60 {
		t.Errorf("after DoFirstSpawn: spawned=%v noDamageTicks=%d, want true/60", p.IsSpawned(), p.NoDamageTicks)
	}
	joined, ok := server.broadcast[0].(*lang.Translatable)
	if !ok || !strings.HasSuffix(joined.Text(), "multiplayer.player.joined") {
		t.Errorf("join message %v, want multiplayer.player.joined", server.broadcast[0])
	}

	before := len(otherSession.messages)
	p.OnPostDisconnect()
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
