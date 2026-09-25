package mcpe

import (
	"encoding/base64"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestPlayerListSkinKeepsTheClientsSkin(t *testing.T) {
	b64 := func(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
	cd := login.ClientData{
		SkinID:            "persona-skin",
		PlayFabID:         "abc123",
		PersonaSkin:       true,
		SkinImageWidth:    256,
		SkinImageHeight:   256,
		SkinData:          b64(make([]byte, 256*256*4)),
		SkinResourcePatch: b64([]byte("{\n  \"geometry\" : {\n    \"default\" : \"geometry.persona_1\"\n  },\n  \"persona_reset_resource_definitions\" : true\n}\n")),
		SkinGeometry:      "",
	}
	skin := playerListSkin(cd)
	if !skin.PersonaSkin || skin.SkinImageWidth != 256 || len(skin.SkinData) != 256*256*4 {
		t.Errorf("persona=%v width=%d data=%d, want the client's own persona skin", skin.PersonaSkin, skin.SkinImageWidth, len(skin.SkinData))
	}
	if string(skin.SkinGeometry) != "{}" {
		t.Errorf("geometry = %q, want {} for a skin without geometry data", skin.SkinGeometry)
	}
	if skin.PlayFabID != "abc123" || skin.FullID != "persona-skin" || string(skin.GeometryDataEngineVersion) != protocol.CurrentVersion {
		t.Errorf("playFab=%q full=%q engine=%q", skin.PlayFabID, skin.FullID, skin.GeometryDataEngineVersion)
	}
	if string(skin.SkinResourcePatch) != `{"geometry":{"default":"geometry.persona_1"}}` {
		t.Errorf("resource patch = %s, want Dragonfly's compact geometry-only patch", skin.SkinResourcePatch)
	}
	if len(skin.PersonaPieces) != 0 || len(skin.PieceTintColours) != 0 {
		t.Error("persona pieces/tints are sent (Dragonfly doesn't send them)")
	}
}
