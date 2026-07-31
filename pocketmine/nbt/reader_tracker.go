package nbt

import "fmt"

// ReaderTracker is a port of pocketmine\nbt\ReaderTracker: guards against excessively deep NBT
// nesting during decode.
type ReaderTracker struct {
	maxDepth     int
	currentDepth int
}

func NewReaderTracker(maxDepth int) *ReaderTracker {
	return &ReaderTracker{maxDepth: maxDepth}
}

// ProtectDepth runs execute(), returning an error instead of recursing further if maxDepth
// (when > 0) has been exceeded.
func (t *ReaderTracker) ProtectDepth(execute func() error) error {
	if t.maxDepth > 0 {
		t.currentDepth++
		if t.currentDepth > t.maxDepth {
			t.currentDepth--
			return NewNbtDataException(fmt.Sprintf("Nesting level too deep: reached max depth of %d tags", t.maxDepth))
		}
		defer func() { t.currentDepth-- }()
	}
	return execute()
}
