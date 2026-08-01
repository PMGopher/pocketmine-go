package blockutils

// DripleafState is a port of pocketmine\block\utils\DripleafState.
type DripleafState int

const (
	DripleafStateStable DripleafState = iota
	DripleafStateUnstable
	DripleafStatePartialTilt
	DripleafStateFullTilt
)

// GetScheduledUpdateDelayTicks is a port of DripleafState::getScheduledUpdateDelayTicks. The
// PHP original returns ?int; ok is false for Stable, matching the null case.
func (d DripleafState) GetScheduledUpdateDelayTicks() (delay int, ok bool) {
	switch d {
	case DripleafStateStable:
		return 0, false
	case DripleafStateUnstable, DripleafStatePartialTilt:
		return 10, true
	case DripleafStateFullTilt:
		return 100, true
	}
	return 0, false
}
