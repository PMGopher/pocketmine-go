package raklib

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// The same account logging in from two game instances must get two pending logins: they used to
// share one key (UUID + name), so the second login overwrote the first's data and whichever
// finished first deleted it for the other.
func TestLoginKeyIsPerLoginAttempt(t *testing.T) {
	identity := login.IdentityData{Identity: "8c9d5c3e-8f1f-4a9b-9d2a-0a1b2c3d4e5f", DisplayName: "Steve"}
	a := loginKey(identity, login.ClientData{ClientRandomID: 1})
	b := loginKey(identity, login.ClientData{ClientRandomID: 2})
	if a == b {
		t.Error("two logins of the same account share a pending login key")
	}
	if a != loginKey(identity, login.ClientData{ClientRandomID: 1}) {
		t.Error("the same login attempt must always get the same key")
	}
}
