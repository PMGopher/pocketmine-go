package mcpe

import (
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestPacketTraceCollapsesRepeatsAndKeepsTheLatest(t *testing.T) {
	var trace packetTrace
	trace.record("->", &packet.LevelChunk{})
	trace.record("->", &packet.LevelChunk{})
	trace.record("<-", &packet.SubChunkRequest{})
	got := trace.String()
	if !strings.Contains(got, "-> LevelChunk x2") || !strings.Contains(got, "<- SubChunkRequest") {
		t.Errorf("trace = %q", got)
	}

	for i := 0; i < packetTraceSize*2; i++ {
		if i%2 == 0 {
			trace.record("->", &packet.Text{})
		} else {
			trace.record("<-", &packet.PlayerAuthInput{})
		}
	}
	if len(trace.entries) != packetTraceSize {
		t.Errorf("trace kept %d entries, want %d", len(trace.entries), packetTraceSize)
	}
	if strings.Contains(trace.String(), "LevelChunk") {
		t.Error("the oldest entries weren't dropped")
	}
}
