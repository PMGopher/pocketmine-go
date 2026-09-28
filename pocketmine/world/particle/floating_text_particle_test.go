package particle

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/math"
)

// FloatingTextParticle::encode: the first encode adds the fake entity with the title and text as
// its name tag, later ones remove it first, and an invisible particle is only removed.
func TestFloatingTextParticleEncode(t *testing.T) {
	old := NextEntityRuntimeIDFunc
	NextEntityRuntimeIDFunc = func() int { return 42 }
	t.Cleanup(func() { NextEntityRuntimeIDFunc = old })

	p := NewFloatingTextParticle("world", "hello")
	pks := p.Encode(math.NewVector3(1, 2, 3), nil)
	if len(pks) != 1 {
		t.Fatalf("first encode: %d packets, want AddActor", len(pks))
	}
	add, ok := pks[0].(*packet.AddActor)
	if !ok || add.EntityRuntimeID != 42 || add.EntityMetadata[protocol.EntityDataKeyName] != "hello\nworld" {
		t.Fatalf("first encode = %#v", pks[0])
	}

	pks = p.Encode(math.NewVector3(1, 2, 3), nil)
	if len(pks) != 2 {
		t.Fatalf("second encode: %d packets, want RemoveActor + AddActor", len(pks))
	}
	if rm, ok := pks[0].(*packet.RemoveActor); !ok || rm.EntityUniqueID != 42 {
		t.Errorf("second encode starts with %#v, want RemoveActor 42", pks[0])
	}

	p.SetInvisible(true)
	if pks = p.Encode(math.NewVector3(1, 2, 3), nil); len(pks) != 1 {
		t.Errorf("invisible encode: %d packets, want only RemoveActor", len(pks))
	}
}
