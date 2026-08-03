package entity

import (
	"fmt"

	"pocketmine-go/pocketmine/binaryutils"
)

// acceptedSkinSizes mirrors Skin::ACCEPTED_SKIN_SIZES.
var acceptedSkinSizes = map[int]bool{
	64 * 32 * 4:   true,
	64 * 64 * 4:   true,
	128 * 128 * 4: true,
}

// capeDataSize mirrors Skin's own hardcoded "must be exactly 8192 bytes" cape data length.
const capeDataSize = 8192

// Skin is a port of pocketmine\entity\Skin: raw skin/cape/geometry bytes, validated at
// construction. Not ported: the geometryData "un-pretty-print" JSON minification real PHP's
// constructor applies (CommentedJsonDecoder-decode then re-encode, purely to shrink the packet
// this data eventually goes out in) - a bandwidth optimisation with no effect on correctness, and
// this port has nowhere that sends this data over the network yet regardless (see
// entity.Human's own doc comment on why Skin isn't wired into Human/Player yet either).
type Skin struct {
	skinID       string
	skinData     []byte
	capeData     []byte
	geometryName string
	geometryData []byte
}

// NewSkin is a port of Skin::__construct, including its real validation (skin ID/geometry name/
// geometry data length limits, valid skin data dimensions, exact cape data size).
func NewSkin(skinID string, skinData, capeData []byte, geometryName string, geometryData []byte) (*Skin, error) {
	if len(skinID) > binaryutils.Int16Max {
		return nil, fmt.Errorf("skin: Skin ID must be at most %d bytes, but have %d bytes", binaryutils.Int16Max, len(skinID))
	}
	if len(geometryName) > binaryutils.Int16Max {
		return nil, fmt.Errorf("skin: Geometry name must be at most %d bytes, but have %d bytes", binaryutils.Int16Max, len(geometryName))
	}
	if len(geometryData) > binaryutils.Int32Max {
		return nil, fmt.Errorf("skin: Geometry data must be at most %d bytes, but have %d bytes", binaryutils.Int32Max, len(geometryData))
	}
	if skinID == "" {
		return nil, fmt.Errorf("skin: Skin ID must not be empty")
	}
	if !acceptedSkinSizes[len(skinData)] {
		return nil, fmt.Errorf("skin: invalid skin data size %d bytes", len(skinData))
	}
	if len(capeData) != 0 && len(capeData) != capeDataSize {
		return nil, fmt.Errorf("skin: invalid cape data size %d bytes (must be exactly %d bytes)", len(capeData), capeDataSize)
	}

	return &Skin{
		skinID:       skinID,
		skinData:     skinData,
		capeData:     capeData,
		geometryName: geometryName,
		geometryData: geometryData,
	}, nil
}

func (s *Skin) GetSkinID() string       { return s.skinID }
func (s *Skin) GetSkinData() []byte     { return s.skinData }
func (s *Skin) GetCapeData() []byte     { return s.capeData }
func (s *Skin) GetGeometryName() string { return s.geometryName }
func (s *Skin) GetGeometryData() []byte { return s.geometryData }
