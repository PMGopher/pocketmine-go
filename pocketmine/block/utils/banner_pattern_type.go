package blockutils

// BannerPatternType is a port of pocketmine\block\utils\BannerPatternType.
type BannerPatternType int

const (
	BannerPatternTypeBorder BannerPatternType = iota
	BannerPatternTypeBricks
	BannerPatternTypeCircle
	BannerPatternTypeCreeper
	BannerPatternTypeCross
	BannerPatternTypeCurlyBorder
	BannerPatternTypeDiagonalLeft
	BannerPatternTypeDiagonalRight
	BannerPatternTypeDiagonalUpLeft
	BannerPatternTypeDiagonalUpRight
	BannerPatternTypeFlow
	BannerPatternTypeFlower
	BannerPatternTypeGlobe
	BannerPatternTypeGradient
	BannerPatternTypeGradientUp
	BannerPatternTypeGuster
	BannerPatternTypeHalfHorizontal
	BannerPatternTypeHalfHorizontalBottom
	BannerPatternTypeHalfVertical
	BannerPatternTypeHalfVerticalRight
	BannerPatternTypeMojang
	BannerPatternTypePiglin
	BannerPatternTypeRhombus
	BannerPatternTypeSkull
	BannerPatternTypeSmallStripes
	BannerPatternTypeSquareBottomLeft
	BannerPatternTypeSquareBottomRight
	BannerPatternTypeSquareTopLeft
	BannerPatternTypeSquareTopRight
	BannerPatternTypeStraightCross
	BannerPatternTypeStripeBottom
	BannerPatternTypeStripeCenter
	BannerPatternTypeStripeDownleft
	BannerPatternTypeStripeDownright
	BannerPatternTypeStripeLeft
	BannerPatternTypeStripeMiddle
	BannerPatternTypeStripeRight
	BannerPatternTypeStripeTop
	BannerPatternTypeTriangleBottom
	BannerPatternTypeTriangleTop
	BannerPatternTypeTrianglesBottom
	BannerPatternTypeTrianglesTop
)

// bannerPatternTypeIDs is a port of pocketmine\data\bedrock\BannerPatternTypeIdMap's registration
// table - kept as a plain map in this package rather than a separate data/bedrock package/
// singleton, since it's the only thing that map is used for right now.
var bannerPatternTypeIDs = map[BannerPatternType]string{
	BannerPatternTypeBorder:               "bo",
	BannerPatternTypeBricks:               "bri",
	BannerPatternTypeCircle:               "mc",
	BannerPatternTypeCreeper:              "cre",
	BannerPatternTypeCross:                "cr",
	BannerPatternTypeCurlyBorder:          "cbo",
	BannerPatternTypeDiagonalLeft:         "lud",
	BannerPatternTypeDiagonalRight:        "rd",
	BannerPatternTypeDiagonalUpLeft:       "ld",
	BannerPatternTypeDiagonalUpRight:      "rud",
	BannerPatternTypeFlower:               "flo",
	BannerPatternTypeFlow:                 "flw",
	BannerPatternTypeGlobe:                "glb",
	BannerPatternTypeGradient:             "gra",
	BannerPatternTypeGradientUp:           "gru",
	BannerPatternTypeGuster:               "gus",
	BannerPatternTypeHalfHorizontal:       "hh",
	BannerPatternTypeHalfHorizontalBottom: "hhb",
	BannerPatternTypeHalfVertical:         "vh",
	BannerPatternTypeHalfVerticalRight:    "vhr",
	BannerPatternTypeMojang:               "moj",
	BannerPatternTypePiglin:               "pig",
	BannerPatternTypeRhombus:              "mr",
	BannerPatternTypeSkull:                "sku",
	BannerPatternTypeSmallStripes:         "ss",
	BannerPatternTypeSquareBottomLeft:     "bl",
	BannerPatternTypeSquareBottomRight:    "br",
	BannerPatternTypeSquareTopLeft:        "tl",
	BannerPatternTypeSquareTopRight:       "tr",
	BannerPatternTypeStraightCross:        "sc",
	BannerPatternTypeStripeBottom:         "bs",
	BannerPatternTypeStripeCenter:         "cs",
	BannerPatternTypeStripeDownleft:       "dls",
	BannerPatternTypeStripeDownright:      "drs",
	BannerPatternTypeStripeLeft:           "ls",
	BannerPatternTypeStripeMiddle:         "ms",
	BannerPatternTypeStripeRight:          "rs",
	BannerPatternTypeStripeTop:            "ts",
	BannerPatternTypeTriangleBottom:       "bt",
	BannerPatternTypeTriangleTop:          "tt",
	BannerPatternTypeTrianglesBottom:      "bts",
	BannerPatternTypeTrianglesTop:         "tts",
}

var bannerPatternTypeByID map[string]BannerPatternType

func init() {
	bannerPatternTypeByID = make(map[string]BannerPatternType, len(bannerPatternTypeIDs))
	for t, id := range bannerPatternTypeIDs {
		bannerPatternTypeByID[id] = t
	}
}

// BannerPatternTypeToID is a port of BannerPatternTypeIdMap::toId.
func BannerPatternTypeToID(t BannerPatternType) (string, bool) {
	id, ok := bannerPatternTypeIDs[t]
	return id, ok
}

// BannerPatternTypeFromID is a port of BannerPatternTypeIdMap::fromId.
func BannerPatternTypeFromID(id string) (BannerPatternType, bool) {
	t, ok := bannerPatternTypeByID[id]
	return t, ok
}

// BannerPatternLayer is a port of pocketmine\block\utils\BannerPatternLayer.
type BannerPatternLayer struct {
	PatternType BannerPatternType
	Color       DyeColor
}

func NewBannerPatternLayer(patternType BannerPatternType, color DyeColor) BannerPatternLayer {
	return BannerPatternLayer{PatternType: patternType, Color: color}
}

func (b BannerPatternLayer) GetType() BannerPatternType { return b.PatternType }

func (b BannerPatternLayer) GetColor() DyeColor { return b.Color }
