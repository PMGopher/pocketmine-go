package network

import "sync"

// BandwidthStatsTracker is a port of pocketmine\network\BandwidthStatsTracker. It's safe for
// concurrent use (bytes are counted on the connection goroutines).
type BandwidthStatsTracker struct {
	mu                     sync.Mutex
	history                []int64
	nextHistoryIndex       int
	bytesSinceLastRotation int64
	totalBytes             int64
}

func NewBandwidthStatsTracker(historySize int) *BandwidthStatsTracker {
	return &BandwidthStatsTracker{history: make([]int64, historySize)}
}

func (t *BandwidthStatsTracker) Add(bytes int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.totalBytes += int64(bytes)
	t.bytesSinceLastRotation += int64(bytes)
}

func (t *BandwidthStatsTracker) GetTotalBytes() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.totalBytes
}

func (t *BandwidthStatsTracker) RotateHistory() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history[t.nextHistoryIndex] = t.bytesSinceLastRotation
	t.bytesSinceLastRotation = 0
	t.nextHistoryIndex = (t.nextHistoryIndex + 1) % len(t.history)
}

// GetAverageBytes returns the average bytes per second recorded over the history window.
func (t *BandwidthStatsTracker) GetAverageBytes() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	var sum int64
	for _, h := range t.history {
		sum += h
	}
	return float64(sum) / float64(len(t.history))
}

func (t *BandwidthStatsTracker) ResetHistory() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.history {
		t.history[i] = 0
	}
}

// BidirectionalBandwidthStatsTracker is a port of
// pocketmine\network\BidirectionalBandwidthStatsTracker.
type BidirectionalBandwidthStatsTracker struct {
	send, receive *BandwidthStatsTracker
}

func NewBidirectionalBandwidthStatsTracker(historySize int) *BidirectionalBandwidthStatsTracker {
	return &BidirectionalBandwidthStatsTracker{send: NewBandwidthStatsTracker(historySize), receive: NewBandwidthStatsTracker(historySize)}
}

func (t *BidirectionalBandwidthStatsTracker) GetSend() *BandwidthStatsTracker    { return t.send }
func (t *BidirectionalBandwidthStatsTracker) GetReceive() *BandwidthStatsTracker { return t.receive }

func (t *BidirectionalBandwidthStatsTracker) Add(sendBytes, recvBytes int) {
	t.send.Add(sendBytes)
	t.receive.Add(recvBytes)
}

func (t *BidirectionalBandwidthStatsTracker) RotateAverageHistory() {
	t.send.RotateHistory()
	t.receive.RotateHistory()
}

func (t *BidirectionalBandwidthStatsTracker) ResetHistory() {
	t.send.ResetHistory()
	t.receive.ResetHistory()
}
