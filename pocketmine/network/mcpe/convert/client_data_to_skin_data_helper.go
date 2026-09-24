package convert

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// personaPieceTypes maps the login JSON's persona piece type names onto gophertunnel's
// protocol.PieceType* values (the network encoding since Bedrock 1.26.50).
var personaPieceTypes = map[string]uint32{
	"persona_skeleton":       protocol.PieceTypeSkeleton,
	"persona_body":           protocol.PieceTypeBody,
	"persona_skin":           protocol.PieceTypeSkin,
	"persona_bottom":         protocol.PieceTypeBottom,
	"persona_feet":           protocol.PieceTypeFeet,
	"persona_dress":          protocol.PieceTypeDress,
	"persona_top":            protocol.PieceTypeTop,
	"persona_high_pants":     protocol.PieceTypeHighPants,
	"persona_hands":          protocol.PieceTypeHands,
	"persona_outerwear":      protocol.PieceTypeOuterwear,
	"persona_facial_hair":    protocol.PieceTypeFacialHair,
	"persona_mouth":          protocol.PieceTypeMouth,
	"persona_eyes":           protocol.PieceTypeEyes,
	"persona_hair":           protocol.PieceTypeHair,
	"persona_hood":           protocol.PieceTypeHood,
	"persona_back":           protocol.PieceTypeBack,
	"persona_face_accessory": protocol.PieceTypeFaceAccessory,
	"persona_head":           protocol.PieceTypeHead,
	"persona_legs":           protocol.PieceTypeLegs,
	"persona_left_leg":       protocol.PieceTypeLeftLeg,
	"persona_right_leg":      protocol.PieceTypeRightLeg,
	"persona_arms":           protocol.PieceTypeArms,
	"persona_left_arm":       protocol.PieceTypeLeftArm,
	"persona_right_arm":      protocol.PieceTypeRightArm,
	"persona_capes":          protocol.PieceTypeCapes,
	"persona_classic_skin":   protocol.PieceTypeClassicSkin,
	"persona_emote":          protocol.PieceTypeEmote,
}

func personaPieceType(name string) uint32 {
	if t, ok := personaPieceTypes[name]; ok {
		return t
	}
	return protocol.PieceTypeUnknown
}

func armSize(name string) uint8 {
	if name == "slim" {
		return protocol.ArmSizeSlim
	}
	return protocol.ArmSizeWide
}

func parseUUIDOrNil(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// parseHexColour parses the login JSON's "#RRGGBB" (SkinColor) or "#AARRGGBB" (piece tint) colours.
// Anything else (e.g. the "#0" padding in tint colour lists) is transparent black.
func parseHexColour(s string) color.RGBA {
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	if err != nil {
		return color.RGBA{}
	}
	switch len(strings.TrimPrefix(s, "#")) {
	case 6:
		return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}
	case 8:
		return color.RGBA{A: uint8(v >> 24), R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v)}
	}
	return color.RGBA{}
}

// ClientDataToSkinData is a port of
// pocketmine\network\mcpe\protocol\types\login\clientdata\ClientDataToSkinDataHelper::fromClientData
// (pmmp/BedrockProtocol): the skin a client sent in its login data, as protocol skin data. The login
// JSON keeps persona piece types, tint colours and the arm size as strings, which gophertunnel's
// protocol.Skin has as typed values since 1.26.50 (see personaPieceType, parseHexColour, armSize).
func ClientDataToSkinData(cd login.ClientData) (protocol.Skin, error) {
	decode := func(value, field string) ([]byte, error) { // ClientDataToSkinDataHelper::safeB64Decode
		b, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 in %s: %w", field, err)
		}
		return b, nil
	}

	animations := make([]protocol.SkinAnimation, 0, len(cd.AnimatedImageData))
	for k, animation := range cd.AnimatedImageData {
		image, err := decode(animation.Image, fmt.Sprintf("AnimatedImageData.%d.Image", k))
		if err != nil {
			return protocol.Skin{}, err
		}
		animations = append(animations, protocol.SkinAnimation{
			ImageWidth:     uint32(animation.ImageWidth),
			ImageHeight:    uint32(animation.ImageHeight),
			ImageData:      image,
			AnimationType:  uint32(animation.Type),
			FrameCount:     float32(animation.Frames),
			ExpressionType: uint32(animation.AnimationExpression),
		})
	}

	fields := map[string]string{
		"SkinResourcePatch":             cd.SkinResourcePatch,
		"SkinData":                      cd.SkinData,
		"CapeData":                      cd.CapeData,
		"SkinGeometryData":              cd.SkinGeometry,
		"SkinGeometryDataEngineVersion": cd.SkinGeometryVersion, //yes, they actually base64'd the version!
		"SkinAnimationData":             cd.SkinAnimationData,
	}
	decoded := make(map[string][]byte, len(fields))
	for field, value := range fields {
		b, err := decode(value, field)
		if err != nil {
			return protocol.Skin{}, err
		}
		decoded[field] = b
	}

	pieces := make([]protocol.PersonaPiece, len(cd.PersonaPieces))
	for i, p := range cd.PersonaPieces {
		pieces[i] = protocol.PersonaPiece{
			PieceID:   p.PieceID,
			PieceType: personaPieceType(p.PieceType),
			PackID:    parseUUIDOrNil(p.PackID),
			Default:   p.Default,
			ProductID: p.ProductID,
		}
	}
	tints := make([]protocol.PersonaPieceTintColour, len(cd.PieceTintColours))
	for i, t := range cd.PieceTintColours {
		tint := protocol.PersonaPieceTintColour{PieceType: t.PieceType}
		for j, c := range t.Colours {
			tint.Colours[j] = parseHexColour(c)
		}
		tints[i] = tint
	}

	return protocol.Skin{
		SkinID:                    cd.SkinID,
		PlayFabID:                 "",
		SkinResourcePatch:         decoded["SkinResourcePatch"],
		SkinImageWidth:            uint32(cd.SkinImageWidth),
		SkinImageHeight:           uint32(cd.SkinImageHeight),
		SkinData:                  decoded["SkinData"],
		Animations:                animations,
		CapeImageWidth:            uint32(cd.CapeImageWidth),
		CapeImageHeight:           uint32(cd.CapeImageHeight),
		CapeData:                  decoded["CapeData"],
		SkinGeometry:              decoded["SkinGeometryData"],
		GeometryDataEngineVersion: decoded["SkinGeometryDataEngineVersion"],
		AnimationData:             decoded["SkinAnimationData"],
		CapeID:                    cd.CapeID,
		FullID:                    uuid.New().String(), // null -> SkinData generates a UUID
		ArmSize:                   armSize(cd.ArmSize),
		SkinColour:                parseHexColour(cd.SkinColour),
		PersonaPieces:             pieces,
		PieceTintColours:          tints,
		Trusted:                   true,
		PremiumSkin:               cd.PremiumSkin,
		PersonaSkin:               cd.PersonaSkin,
		PersonaCapeOnClassicSkin:  cd.CapeOnClassicSkin,
		PrimaryUser:               true, //assume this is true? there's no field for it ...
		OverrideAppearance:        cd.OverrideSkin,
	}, nil
}
