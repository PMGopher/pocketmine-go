package blockutils

// CopperOxidation is a port of pocketmine\block\utils\CopperOxidation.
type CopperOxidation int

const (
	CopperOxidationNone CopperOxidation = iota
	CopperOxidationExposed
	CopperOxidationWeathered
	CopperOxidationOxidized
)

// GetPrevious returns the previous (less oxidized) stage, and ok=false if already CopperOxidationNone.
func (c CopperOxidation) GetPrevious() (CopperOxidation, bool) {
	if c <= CopperOxidationNone {
		return 0, false
	}
	return c - 1, true
}

// GetNext returns the next (more oxidized) stage, and ok=false if already CopperOxidationOxidized.
func (c CopperOxidation) GetNext() (CopperOxidation, bool) {
	if c >= CopperOxidationOxidized {
		return 0, false
	}
	return c + 1, true
}
