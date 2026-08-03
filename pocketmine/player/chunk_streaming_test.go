package player

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestOrderChunksIsANoOpWhileViewDistanceIsUnset(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.OrderChunks()
	if len(p.GetUsedChunks()) != 0 || len(p.loadQueue) != 0 {
		t.Error("OrderChunks() with viewDistance unset (-1) did something - want a no-op")
	}
}

func TestOrderChunksThenRequestChunksLoadsChunksAroundThePlayer(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(8, 70, 8)) // chunk (0,0)
	p.SetViewDistance(2)

	p.OrderChunks()
	if len(p.loadQueue) == 0 {
		t.Fatal("OrderChunks() did not queue any chunks to load")
	}

	sent := p.RequestChunks()
	if len(sent) == 0 {
		t.Fatal("RequestChunks() returned no newly-ready chunks")
	}
	if len(p.loadQueue) != 0 {
		t.Error("RequestChunks() did not drain the load queue")
	}

	for _, c := range sent {
		status, ok := p.GetUsedChunkStatus(c[0], c[1])
		if !ok || status != UsedChunkStatusRequestedSending {
			t.Errorf("chunk %v status = %v (ok=%v), want UsedChunkStatusRequestedSending", c, status, ok)
		}
		if !p.IsUsingChunk(c[0], c[1]) {
			t.Errorf("IsUsingChunk%v = false after RequestChunks", c)
		}
		if !p.GetWorld().IsChunkInUse(c[0], c[1]) {
			t.Errorf("World.IsChunkInUse%v = false - player should be registered as a chunk loader", c)
		}
	}
}

func TestMarkChunkSentTransitionsRequestedSendingToSent(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(8, 70, 8))
	p.SetViewDistance(2)
	p.OrderChunks()
	sent := p.RequestChunks()
	if len(sent) == 0 {
		t.Fatal("no chunks returned by RequestChunks")
	}

	c := sent[0]
	p.MarkChunkSent(c[0], c[1])

	if !p.HasReceivedChunk(c[0], c[1]) {
		t.Error("HasReceivedChunk() = false after MarkChunkSent")
	}
	status, _ := p.GetUsedChunkStatus(c[0], c[1])
	if status != UsedChunkStatusSent {
		t.Errorf("status after MarkChunkSent = %v, want UsedChunkStatusSent", status)
	}
}

func TestMarkChunkSentIsANoOpForAChunkNotAwaitingSending(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.MarkChunkSent(0, 0) // never requested at all
	if p.HasReceivedChunk(0, 0) {
		t.Error("MarkChunkSent marked an untracked chunk as sent")
	}
}

func TestOrderChunksUnloadsChunksThatFallOutsideTheNewRadius(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(8, 70, 8))
	p.SetViewDistance(6)
	p.OrderChunks()
	p.RequestChunks()

	before := len(p.GetUsedChunks())
	if before == 0 {
		t.Fatal("no chunks were loaded to begin with")
	}

	p.SetViewDistance(1)
	p.OrderChunks()

	if len(p.GetUsedChunks()) >= before {
		t.Errorf("GetUsedChunks() count after shrinking view distance = %d, want fewer than %d", len(p.GetUsedChunks()), before)
	}
}

func TestOnChunkUnloadedTearsDownAUsedChunk(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(8, 70, 8))
	p.SetViewDistance(2)
	p.OrderChunks()
	sent := p.RequestChunks()
	if len(sent) == 0 {
		t.Fatal("no chunks returned by RequestChunks")
	}
	c := sent[0]

	p.OnChunkUnloaded(c[0], c[1], nil)

	if p.IsUsingChunk(c[0], c[1]) {
		t.Error("IsUsingChunk still true after OnChunkUnloaded")
	}
	if p.GetWorld().IsChunkInUse(c[0], c[1]) {
		t.Error("World still considers the chunk in use after OnChunkUnloaded")
	}
}

func TestOnChunkChangedResetsASentChunkBackToNeeded(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(8, 70, 8))
	p.SetViewDistance(2)
	p.OrderChunks()
	sent := p.RequestChunks()
	if len(sent) == 0 {
		t.Fatal("no chunks returned by RequestChunks")
	}
	c := sent[0]
	p.MarkChunkSent(c[0], c[1])

	p.OnChunkChanged(c[0], c[1], nil)

	status, ok := p.GetUsedChunkStatus(c[0], c[1])
	if !ok || status != UsedChunkStatusNeeded {
		t.Errorf("status after OnChunkChanged = %v (ok=%v), want UsedChunkStatusNeeded", status, ok)
	}
}
