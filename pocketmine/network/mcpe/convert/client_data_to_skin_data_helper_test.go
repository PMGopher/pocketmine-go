package convert

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestClientDataToSkinData(t *testing.T) {
	b64 := func(s []byte) string { return base64.StdEncoding.EncodeToString(s) }
	cd := login.ClientData{
		SkinID:              "skin-id",
		SkinData:            b64(make([]byte, 64*64*4)),
		SkinImageWidth:      64,
		SkinImageHeight:     64,
		SkinResourcePatch:   b64([]byte(`{"geometry":{"default":"geometry.humanoid.custom"}}`)),
		SkinGeometryVersion: b64([]byte("1.26.50")),
		ArmSize:             "slim",
		SkinColour:          "#b37b62",
		PersonaPieces:       []login.PersonaPiece{{PieceType: "persona_eyes", PackID: "00000000-0000-0000-0000-000000000001"}},
		PieceTintColours:    []login.PersonaPieceTintColour{{PieceType: "persona_eyes", Colours: [4]string{"#ffa12722", "#0", "#0", "#0"}}},
	}
	skin, err := ClientDataToSkinData(cd)
	if err != nil {
		t.Fatal(err)
	}
	if skin.SkinID != "skin-id" || len(skin.SkinData) != 64*64*4 || string(skin.GeometryDataEngineVersion) != "1.26.50" {
		t.Errorf("skin id/data/engine version not decoded: %q %d %q", skin.SkinID, len(skin.SkinData), skin.GeometryDataEngineVersion)
	}
	if skin.ArmSize != protocol.ArmSizeSlim {
		t.Errorf("arm size = %d, want slim", skin.ArmSize)
	}
	if skin.SkinColour != (color.RGBA{R: 0xb3, G: 0x7b, B: 0x62, A: 0xff}) {
		t.Errorf("skin colour = %v", skin.SkinColour)
	}
	if skin.PersonaPieces[0].PieceType != protocol.PieceTypeEyes {
		t.Errorf("piece type = %d, want eyes", skin.PersonaPieces[0].PieceType)
	}
	if skin.PieceTintColours[0].Colours[0] != (color.RGBA{A: 0xff, R: 0xa1, G: 0x27, B: 0x22}) || skin.PieceTintColours[0].Colours[1] != (color.RGBA{}) {
		t.Errorf("tint colours = %v", skin.PieceTintColours[0].Colours)
	}
	if !skin.Trusted || !skin.PrimaryUser || skin.FullID == "" {
		t.Error("SkinData defaults (verified, primary user, full skin ID) not set")
	}

	cd.SkinData = "not base64!"
	if _, err := ClientDataToSkinData(cd); err == nil {
		t.Error("invalid base64 skin data was accepted")
	}
}

func TestCoreGameModeToProtocol(t *testing.T) {
	// TypeConverter::coreGameModeToProtocol: spectator is sent as creative.
	for core, want := range map[int]int32{0: 0, 1: 1, 2: 2, 3: 1} {
		if got := CoreGameModeToProtocol(core); got != want {
			t.Errorf("CoreGameModeToProtocol(%d) = %d, want %d", core, got, want)
		}
	}
}
