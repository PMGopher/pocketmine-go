package mcpe

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// playerListSkin is the skin sent in PlayerList for a player, built from the skin its client sent
// at login the way Dragonfly does (server.go parseSkin + session_list.go skinToProtocol): the
// client's own skin image, geometry and resource patch, persona flag, PlayFab ID and full skin ID,
// without persona pieces or tint colours.
//
// It has no PocketMine-MP counterpart: PHP sends LegacySkinAdapter::toSkinData($player->getSkin()),
// which for a persona skin is a replacement 64x64 skin with no geometry data. A 1.26.51 client
// (which plays fine on Dragonfly) disconnected right after loading, when it renders its own
// PlayerList entry, while this server sent it that converted skin.
func playerListSkin(cd login.ClientData) protocol.Skin {
	decode := func(s string) []byte {
		b, _ := base64.StdEncoding.DecodeString(s)
		return b
	}

	animations := make([]protocol.SkinAnimation, 0, len(cd.AnimatedImageData))
	for _, animation := range cd.AnimatedImageData {
		animations = append(animations, protocol.SkinAnimation{
			ImageWidth:     uint32(animation.ImageWidth),
			ImageHeight:    uint32(animation.ImageHeight),
			ImageData:      decode(animation.Image),
			AnimationType:  uint32(animation.Type),
			FrameCount:     float32(animation.Frames),
			ExpressionType: uint32(animation.AnimationExpression),
		})
	}

	fullID := cd.SkinID
	if fullID == "" {
		fullID = uuid.NewString()
	}
	model := decode(cd.SkinGeometry)
	if len(model) == 0 {
		model = []byte("{}")
	}
	return protocol.Skin{
		PlayFabID:                 cd.PlayFabID,
		SkinID:                    uuid.NewString(),
		SkinResourcePatch:         skinResourcePatch(decode(cd.SkinResourcePatch)),
		SkinImageWidth:            uint32(cd.SkinImageWidth),
		SkinImageHeight:           uint32(cd.SkinImageHeight),
		SkinData:                  decode(cd.SkinData),
		CapeImageWidth:            uint32(cd.CapeImageWidth),
		CapeImageHeight:           uint32(cd.CapeImageHeight),
		CapeData:                  decode(cd.CapeData),
		SkinGeometry:              model,
		PersonaSkin:               cd.PersonaSkin,
		CapeID:                    uuid.NewString(),
		FullID:                    fullID,
		Animations:                animations,
		Trusted:                   true,
		OverrideAppearance:        true,
		GeometryDataEngineVersion: []byte(protocol.CurrentVersion),
	}
}

// skinResourcePatch re-encodes a client's resource patch the way Dragonfly does
// (skin.DecodeModelConfig + ModelConfig.Encode): only geometry.default and geometry.animated_face
// are kept, as compact JSON. Everything else a persona skin puts in its patch is dropped.
func skinResourcePatch(patch []byte) []byte {
	var container struct {
		Geometry struct {
			Default      string `json:"default"`
			AnimatedFace string `json:"animated_face,omitempty"`
		} `json:"geometry"`
	}
	_ = json.Unmarshal(patch, &container)
	b, _ := json.Marshal(container)
	return b
}
