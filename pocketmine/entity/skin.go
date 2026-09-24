package entity

import (
	"encoding/json"
	"fmt"

	"github.com/df-mc/jsonc"

	"pocketmine-go/pocketmine/binaryutils"
)

// acceptedSkinSizes mirrors Skin::ACCEPTED_SKIN_SIZES.
var acceptedSkinSizes = []int{
	64 * 32 * 4,
	64 * 64 * 4,
	128 * 128 * 4,
}

// capeDataSize mirrors Skin's "must be exactly 8192 bytes" cape data length.
const capeDataSize = 8192

// InvalidSkinError is a port of pocketmine\entity\InvalidSkinException.
type InvalidSkinError struct{ Message string }

func (e *InvalidSkinError) Error() string { return e.Message }

// Skin is a port of pocketmine\entity\Skin: raw skin/cape/geometry data, validated at
// construction.
type Skin struct {
	skinID       string
	skinData     []byte
	capeData     []byte
	geometryName string
	geometryData []byte
}

func checkSkinFieldLength(value []byte, name string, maxLength int) error {
	if len(value) > maxLength {
		return &InvalidSkinError{Message: fmt.Sprintf("%s must be at most %d bytes, but have %d bytes", name, maxLength, len(value))}
	}
	return nil
}

// NewSkin is a port of Skin::__construct (PHP's defaults for the last three parameters are empty).
// The geometry data, if any, must be valid JSON (comments allowed, like PHP's
// CommentedJsonDecoder).
func NewSkin(skinID string, skinData, capeData []byte, geometryName string, geometryData []byte) (*Skin, error) {
	if err := checkSkinFieldLength([]byte(skinID), "Skin ID", binaryutils.Int16Max); err != nil {
		return nil, err
	}
	if err := checkSkinFieldLength([]byte(geometryName), "Geometry name", binaryutils.Int16Max); err != nil {
		return nil, err
	}
	if err := checkSkinFieldLength(geometryData, "Geometry data", binaryutils.Int32Max); err != nil {
		return nil, err
	}

	if skinID == "" {
		return nil, &InvalidSkinError{Message: "Skin ID must not be empty"}
	}
	accepted := false
	for _, size := range acceptedSkinSizes {
		if len(skinData) == size {
			accepted = true
			break
		}
	}
	if !accepted {
		return nil, &InvalidSkinError{Message: fmt.Sprintf("Invalid skin data size %d bytes (allowed sizes: %d, %d, %d)", len(skinData), acceptedSkinSizes[0], acceptedSkinSizes[1], acceptedSkinSizes[2])}
	}

	if len(capeData) != 0 && len(capeData) != capeDataSize {
		return nil, &InvalidSkinError{Message: fmt.Sprintf("Invalid cape data size %d bytes (must be exactly %d bytes)", len(capeData), capeDataSize)}
	}

	if len(geometryData) != 0 {
		var decoded any
		if err := json.Unmarshal(jsonc.ToJSON(geometryData), &decoded); err != nil {
			return nil, &InvalidSkinError{Message: "Invalid geometry data: " + err.Error()}
		}
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
