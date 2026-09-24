package entity

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SkinToNetwork is a port of pocketmine\network\mcpe\convert\LegacySkinAdapter::toSkinData: a Skin
// as the protocol's skin data. It lives here rather than in network/mcpe/convert because it needs
// the Skin type, and this package already imports convert.
func SkinToNetwork(skin *Skin) protocol.Skin {
	capeData := skin.GetCapeData()
	var capeWidth, capeHeight uint32
	if len(capeData) != 0 {
		capeWidth, capeHeight = 64, 32
	}
	geometryName := skin.GetGeometryName()
	if geometryName == "" {
		geometryName = "geometry.humanoid.custom"
	}
	resourcePatch, err := json.Marshal(map[string]any{"geometry": map[string]any{"default": geometryName}})
	if err != nil {
		panic(err)
	}
	width, height := legacySkinImageSize(len(skin.GetSkinData()))
	return protocol.Skin{
		SkinID:            skin.GetSkinID(),
		PlayFabID:         "", //TODO: playfab ID
		SkinResourcePatch: resourcePatch,
		SkinImageWidth:    width,
		SkinImageHeight:   height,
		SkinData:          skin.GetSkinData(),
		CapeImageWidth:    capeWidth,
		CapeImageHeight:   capeHeight,
		CapeData:          capeData,
		SkinGeometry:      skin.GetGeometryData(),
		// The rest are SkinData::__construct's defaults, which toSkinData doesn't override.
		GeometryDataEngineVersion: []byte(protocol.CurrentVersion), // ProtocolInfo::MINECRAFT_VERSION_NETWORK
		FullID:                    uuid.New().String(),             // $fullSkinId ?? Uuid::uuid4()
		ArmSize:                   protocol.ArmSizeWide,
		Trusted:                   true, // isVerified
		PrimaryUser:               true,
		OverrideAppearance:        true,
	}
}

// legacySkinImageSize is SkinImage::fromLegacy's size table.
func legacySkinImageSize(dataLength int) (width, height uint32) {
	switch dataLength {
	case 64 * 32 * 4:
		return 64, 32
	case 64 * 64 * 4:
		return 64, 64
	case 128 * 128 * 4:
		return 128, 128
	}
	panic(fmt.Sprintf("Unknown size %d", dataLength))
}

// SkinFromNetwork is a port of LegacySkinAdapter::fromSkinData.
func SkinFromNetwork(data protocol.Skin) (*Skin, error) {
	if data.PersonaSkin {
		skinData := make([]byte, 4096*4)
		for i := 0; i < 4096; i++ {
			copy(skinData[i*4:], []byte{0x80, 0x80, 0x80, 0xff})
		}
		return NewSkin("Standard_Custom", skinData, nil, "", nil)
	}

	capeData := data.CapeData
	if data.PersonaCapeOnClassicSkin {
		capeData = nil
	}

	var resourcePatch struct {
		Geometry struct {
			Default string `json:"default"`
		} `json:"geometry"`
	}
	if err := json.Unmarshal(data.SkinResourcePatch, &resourcePatch); err != nil || resourcePatch.Geometry.Default == "" {
		return nil, &InvalidSkinError{Message: "Missing geometry name field"}
	}

	return NewSkin(data.SkinID, data.SkinData, capeData, resourcePatch.Geometry.Default, data.SkinGeometry)
}
