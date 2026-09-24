package entity

// EntitySizeInfo is a port of pocketmine\entity\EntitySizeInfo.
type EntitySizeInfo struct {
	height    float64
	width     float64
	eyeHeight float64
}

// NewEntitySizeInfo is a port of EntitySizeInfo::__construct with PHP's default eye height
// (min(height / 2 + 0.1, height)).
func NewEntitySizeInfo(height, width float64) EntitySizeInfo {
	return EntitySizeInfo{height: height, width: width, eyeHeight: min(height/2+0.1, height)}
}

// NewEntitySizeInfoWithEyeHeight is EntitySizeInfo::__construct with an explicit eye height.
func NewEntitySizeInfoWithEyeHeight(height, width, eyeHeight float64) EntitySizeInfo {
	return EntitySizeInfo{height: height, width: width, eyeHeight: eyeHeight}
}

func (s EntitySizeInfo) GetHeight() float64 { return s.height }

func (s EntitySizeInfo) GetWidth() float64 { return s.width }

func (s EntitySizeInfo) GetEyeHeight() float64 { return s.eyeHeight }

// Scale is a port of EntitySizeInfo::scale.
func (s EntitySizeInfo) Scale(newScale float64) EntitySizeInfo {
	return EntitySizeInfo{height: s.height * newScale, width: s.width * newScale, eyeHeight: s.eyeHeight * newScale}
}
