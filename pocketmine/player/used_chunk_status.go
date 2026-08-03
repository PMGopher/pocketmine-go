package player

// UsedChunkStatus is a port of pocketmine\player\UsedChunkStatus.
type UsedChunkStatus int

const (
	UsedChunkStatusNeeded UsedChunkStatus = iota
	UsedChunkStatusRequestedGeneration
	UsedChunkStatusRequestedSending
	UsedChunkStatusSent
)
