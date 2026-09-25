package player

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// newStreamingPlayer is a connected player at (8, 70, 8), chunk (0,0).
func newStreamingPlayer(t *testing.T) (*Player, *fakeSession) {
	t.Helper()
	return newPlayerFor(t, &fakeServer{}, newTestWorld(t), math.NewVector3(8, 70, 8), nil)
}

func TestOrderChunksIsANoOpWhileViewDistanceIsUnset(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.orderChunks()
	if len(p.loadQueue) != 0 || session.viewAreaSyncs != 0 {
		t.Error("orderChunks() with viewDistance unset (-1) did something - want a no-op")
	}
}

func TestDoChunkRequestsSendsChunksAroundThePlayer(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.SetViewDistance(2)

	for range 20 {
		p.DoChunkRequests()
	}
	if len(session.startedChunks) == 0 {
		t.Fatal("no chunks were sent")
	}
	if len(p.loadQueue) != 0 {
		t.Error("repeated DoChunkRequests() did not drain the load queue")
	}
	for _, c := range session.startedChunks {
		if !p.IsUsingChunk(c[0], c[1]) {
			t.Errorf("IsUsingChunk%v = false after sending it", c)
		}
		if status, _ := p.GetUsedChunkStatus(c[0], c[1]); status != UsedChunkStatusRequestedSending {
			t.Errorf("status of %v before the send completed = %v, want RequestedSending", c, status)
		}
	}

	session.completeChunkSends()
	c := session.startedChunks[0]
	if !p.HasReceivedChunk(c[0], c[1]) {
		t.Error("HasReceivedChunk() = false after the send completed")
	}
}

func TestRequestChunksSendsAtMostChunksPerTickNearestFirst(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.SetViewDistance(4)

	p.DoChunkRequests()
	if len(session.startedChunks) != p.chunksPerTick {
		t.Fatalf("sent %d chunks, want chunksPerTick (%d)", len(session.startedChunks), p.chunksPerTick)
	}
	if session.startedChunks[0] != [2]int{0, 0} {
		t.Errorf("first chunk sent = %v, want the player's own chunk (0,0)", session.startedChunks[0])
	}
	p.DoChunkRequests()
	if len(session.startedChunks) != 2*p.chunksPerTick {
		t.Errorf("sent %d chunks after two ticks, want %d", len(session.startedChunks), 2*p.chunksPerTick)
	}
}

func TestOrderChunksUnloadsChunksThatFallOutsideTheNewRadius(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.SetViewDistance(6)
	for range 100 {
		p.DoChunkRequests()
	}
	session.completeChunkSends()
	before := len(p.GetUsedChunks())

	p.SetViewDistance(2)
	p.DoChunkRequests()
	if len(p.GetUsedChunks()) >= before {
		t.Errorf("used chunks after shrinking view distance = %d, want fewer than %d", len(p.GetUsedChunks()), before)
	}
}

func TestOnChunkUnloadedTearsDownAUsedChunk(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.SetViewDistance(2)
	p.DoChunkRequests()
	c := session.startedChunks[0]

	p.OnChunkUnloaded(c[0], c[1], nil)

	if p.IsUsingChunk(c[0], c[1]) {
		t.Error("IsUsingChunk still true after OnChunkUnloaded")
	}
	if p.GetWorld().IsChunkInUse(c[0], c[1]) {
		t.Error("World still considers the chunk in use after OnChunkUnloaded")
	}
}

func TestOnChunkChangedResetsASentChunkBackToNeeded(t *testing.T) {
	p, session := newStreamingPlayer(t)
	p.SetViewDistance(2)
	p.DoChunkRequests()
	session.completeChunkSends()
	c := session.startedChunks[0]

	p.OnChunkChanged(c[0], c[1], nil)

	status, ok := p.GetUsedChunkStatus(c[0], c[1])
	if !ok || status != UsedChunkStatusNeeded {
		t.Errorf("status after OnChunkChanged = %v (ok=%v), want UsedChunkStatusNeeded", status, ok)
	}
}
