package defaults

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/server"
)

type discardSender struct{}

func (discardSender) WritePacket(packet.Packet) error { return nil }
func (discardSender) Close() error                    { return nil }

func newTestPlayer(t *testing.T, s *server.Server, name string) *player.Player {
	t.Helper()
	skin, err := entity.NewSkin("Standard_Custom", make([]byte, 64*64*4), nil, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	broadcaster := mcpe.NewStandardPacketBroadcaster()
	session := mcpe.NewNetworkSession(s, s.GetNetwork().GetSessionManager(), discardSender{}, nil, broadcaster, mcpe.NewStandardEntityEventBroadcaster(broadcaster), "127.0.0.1", 1)
	session.Login(player.NewXboxLivePlayerInfo("", name, uuid.NewString(), skin, "en_US", nil).WithoutXboxData(), false, false)
	return session.GetPlayer()
}

func lastMessage(r *recordingConsole) string {
	if len(r.messages) == 0 {
		return ""
	}
	return r.messages[len(r.messages)-1]
}

func TestGiveCommand(t *testing.T) {
	s, sender := newTestServer(t)
	p := newTestPlayer(t, s, "Alex")

	s.DispatchCommand(sender, "give Alex diamond_sword 1 {display:{Name:\"Blade\"}}", false)
	held := p.GetInventory().GetItem(0)
	if held.GetTypeId() != item.VanillaItem("diamond_sword").GetTypeId() || held.GetCount() != 1 || held.GetCustomName() != "Blade" {
		t.Fatalf("slot 0 = %v (name %q), want a Diamond Sword named Blade", held, held.GetCustomName())
	}

	// No count gives a full stack; legacy id:meta names still work.
	s.DispatchCommand(sender, "give Alex 351:4", false)
	lapis := p.GetInventory().GetItem(1)
	if lapis.GetTypeId() != item.VanillaItem("lapis_lazuli").GetTypeId() || lapis.GetCount() != 64 {
		t.Errorf("slot 1 = %v, want 64 lapis lazuli", lapis)
	}

	s.DispatchCommand(sender, "give Alex no_such_item", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "no_such_item") {
		t.Errorf("unknown item message = %q", msg)
	}
	s.DispatchCommand(sender, "give Alex stone 1 {bad", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "unexpected end of stream") {
		t.Errorf("tag error message = %q", msg)
	}
}

func TestClearCommand(t *testing.T) {
	s, sender := newTestServer(t)
	p := newTestPlayer(t, s, "Alex")
	inv := p.GetInventory()

	dirt := item.GetStringToItemParser()
	d, _ := dirt.Parse("dirt")
	d.SetCount(10)
	inv.SetItem(0, d.Clone())
	inv.SetItem(5, d.Clone())
	stone, _ := dirt.Parse("stone")
	inv.SetItem(1, stone)

	s.DispatchCommand(sender, "clear Alex dirt 0", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "20") {
		t.Errorf("/clear with count 0 = %q, want the 20 matching items counted", msg)
	}
	s.DispatchCommand(sender, "clear Alex dirt 15", false)
	if inv.GetItem(0).GetCount() != 0 || inv.GetItem(5).GetCount() != 5 {
		t.Errorf("after clearing 15 dirt: slot 0 = %v, slot 5 = %v", inv.GetItem(0), inv.GetItem(5))
	}
	s.DispatchCommand(sender, "clear Alex", false)
	if len(inv.GetContents(false)) != 0 {
		t.Errorf("inventory not empty after /clear: %v", inv.GetContents(false))
	}
}

func TestEnchantAndEffectCommands(t *testing.T) {
	s, sender := newTestServer(t)
	p := newTestPlayer(t, s, "Alex")

	p.GetInventory().SetItemInHand(item.VanillaBook())
	s.DispatchCommand(sender, "enchant Alex sharpness 3", false)
	held := p.GetInventory().GetItemInHand()
	if held.GetTypeId() != item.VanillaItem("enchanted_book").GetTypeId() || held.GetEnchantmentLevel(enchantment.VanillaSharpness()) != 3 {
		t.Errorf("held = %v, want an enchanted book with Sharpness III", held)
	}
	s.DispatchCommand(sender, "enchant Alex sharpness 9", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "9") {
		t.Errorf("too-high level message = %q", msg)
	}

	s.DispatchCommand(sender, "effect Alex speed 30 1", false)
	speed := p.GetEffects().Get(effect.VanillaSpeed())
	if speed == nil || speed.GetDuration() != 600 || speed.GetAmplifier() != 1 {
		t.Fatalf("speed = %v, want 600 ticks at amplifier 1", speed)
	}
	s.DispatchCommand(sender, "effect Alex speed 0", false)
	if p.GetEffects().Has(effect.VanillaSpeed()) {
		t.Error("duration 0 didn't remove the effect")
	}
	s.DispatchCommand(sender, "effect Alex haste infinite", false)
	if h := p.GetEffects().Get(effect.VanillaHaste()); h == nil || !h.IsInfinite() {
		t.Errorf("haste = %v, want infinite", h)
	}
	s.DispatchCommand(sender, "effect Alex clear", false)
	if len(p.GetEffects().All()) != 0 {
		t.Error("/effect clear left effects")
	}
}
