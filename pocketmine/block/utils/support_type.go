package blockutils

// SupportType is a port of pocketmine\block\utils\SupportType.
type SupportType int

const (
	SupportTypeFull SupportType = iota
	SupportTypeCenter
	SupportTypeEdge
	SupportTypeNone
)

func (s SupportType) HasEdgeSupport() bool   { return s == SupportTypeEdge || s == SupportTypeFull }
func (s SupportType) HasCenterSupport() bool { return s == SupportTypeCenter || s == SupportTypeFull }
