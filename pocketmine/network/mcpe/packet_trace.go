package mcpe

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// packetTrace keeps the last packets a session sent and received, with repeats collapsed, so that
// when a client closes the connection by itself (it doesn't say why) the log shows what it got
// right before. It has no PocketMine-MP counterpart; it's diagnostics for the protocol layer,
// which this port doesn't share with PocketMine-MP (see AGENTS.md §1).
type packetTrace struct {
	mu      sync.Mutex
	entries []traceEntry
}

type traceEntry struct {
	desc  string
	count int
}

// packetTraceSize is how many distinct (collapsed) entries are kept.
const packetTraceSize = 80

func (t *packetTrace) record(direction string, pk packet.Packet) {
	desc := direction + " " + strings.TrimPrefix(fmt.Sprintf("%T", pk), "*packet.")
	t.mu.Lock()
	defer t.mu.Unlock()
	if n := len(t.entries); n > 0 && t.entries[n-1].desc == desc {
		t.entries[n-1].count++
		return
	}
	t.entries = append(t.entries, traceEntry{desc: desc, count: 1})
	if len(t.entries) > packetTraceSize {
		t.entries = t.entries[len(t.entries)-packetTraceSize:]
	}
}

// String lists the trace, oldest first.
func (t *packetTrace) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var b strings.Builder
	for _, e := range t.entries {
		if e.count > 1 {
			fmt.Fprintf(&b, "\n  %s x%d", e.desc, e.count)
		} else {
			fmt.Fprintf(&b, "\n  %s", e.desc)
		}
	}
	return b.String()
}
