package server

import (
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/network/mcpe"
)

// windowPackets returns the ContainerOpen/ContainerClose packets sent since the last call.
func windowPackets(sender *recordingSender, from *int) (opens []*packet.ContainerOpen, closes []*packet.ContainerClose) {
	for _, pk := range sender.packets[*from:] {
		switch pk := pk.(type) {
		case *packet.ContainerOpen:
			opens = append(opens, pk)
		case *packet.ContainerClose:
			closes = append(closes, pk)
		}
	}
	*from = len(sender.packets)
	return opens, closes
}

func TestMainInventoryOpensReliably(t *testing.T) {
	s := newTestServer(t)
	session, sender := newTestSession(s, 1)
	session.Login(newTestPlayerInfo(t, "Alex").WithoutXboxData(), false, false)
	m := session.GetInventoryManager()
	seen := len(sender.packets)

	// With some latency the client asks twice for one key press: the inventory must open once and
	// stay open (the repeat used to close it again straight away).
	m.OnClientOpenMainInventory()
	m.OnClientOpenMainInventory()
	opens, closes := windowPackets(sender, &seen)
	if len(opens) != 1 || len(closes) != 0 {
		t.Fatalf("two open requests sent %d ContainerOpen and %d ContainerClose, want 1 and 0", len(opens), len(closes))
	}
	windowID := int(opens[0].WindowID)

	// The player closes it and opens it again, many times.
	for i := 0; i < 50; i++ {
		m.OnClientRemoveWindow(windowID)
		m.OnClientOpenMainInventory()
		opens, _ = windowPackets(sender, &seen)
		if len(opens) != 1 {
			t.Fatalf("open #%d sent %d ContainerOpen, want 1", i+2, len(opens))
		}
		windowID = int(opens[0].WindowID)
	}

	// The server closes the window, and the client never acknowledges it (it no longer shows the
	// window). Opening the inventory must not wait for that ack forever.
	m.OnCurrentWindowRemove()
	m.OnClientOpenMainInventory()
	if opens, _ = windowPackets(sender, &seen); len(opens) != 0 {
		t.Fatal("the inventory opened before the close was acknowledged or timed out")
	}
	old := mcpe.WindowCloseAckTimeout
	mcpe.WindowCloseAckTimeout = 0
	defer func() { mcpe.WindowCloseAckTimeout = old }()
	time.Sleep(time.Millisecond)
	m.OnClientOpenMainInventory()
	if opens, _ = windowPackets(sender, &seen); len(opens) != 1 {
		t.Fatalf("after the close ack timed out, opening sent %d ContainerOpen, want 1", len(opens))
	}
}
