package timings

import (
	"sync"
	"time"
)

// TimingsRecord is a port of pocketmine\timings\TimingsRecord.
//
// This whole subsystem (a single global "current record" pointer forming an implicit call
// stack) assumes timing start/stop calls happen sequentially on one execution context, same as
// PocketMine's own single-threaded-per-tick assumption. It is guarded by a mutex for basic
// safety, but concurrent goroutines interleaving StartTiming/StopTiming calls would produce
// nonsensical nesting — same caveat as event.CallOn's call-depth counter elsewhere in this port.
type TimingsRecord struct {
	id           int
	handler      *TimingsHandler
	parentRecord *TimingsRecord

	count        int
	curCount     int
	start        time.Time
	running      bool
	totalTime    time.Duration
	curTickTotal time.Duration
	violations   int
	ticksActive  int
	peakTime     time.Duration
}

var (
	recordsMu     sync.Mutex
	records       = map[int]*TimingsRecord{}
	nextRecordID  int
	currentRecord *TimingsRecord
)

func newTimingsRecord(handler *TimingsHandler, parent *TimingsRecord) *TimingsRecord {
	recordsMu.Lock()
	defer recordsMu.Unlock()
	nextRecordID++
	r := &TimingsRecord{id: nextRecordID, handler: handler, parentRecord: parent}
	records[r.id] = r
	return r
}

// ResetRecords is a port of TimingsRecord::reset().
func ResetRecords() {
	recordsMu.Lock()
	recs := make([]*TimingsRecord, 0, len(records))
	for _, r := range records {
		recs = append(recs, r)
	}
	records = map[int]*TimingsRecord{}
	currentRecord = nil
	recordsMu.Unlock()

	for _, r := range recs {
		r.handler.reset()
	}
}

func GetAllRecords() []*TimingsRecord {
	recordsMu.Lock()
	defer recordsMu.Unlock()
	result := make([]*TimingsRecord, 0, len(records))
	for _, r := range records {
		result = append(result, r)
	}
	return result
}

// TickRecords is a port of TimingsRecord::tick().
func TickRecords(measure bool) {
	recordsMu.Lock()
	defer recordsMu.Unlock()

	if measure {
		for _, r := range records {
			if r.curCount > 0 {
				if r.curTickTotal > TargetTimePerTick {
					r.violations += int(r.curTickTotal / TargetTimePerTick)
				}
				r.curTickTotal = 0
				r.curCount = 0
				r.ticksActive++
			}
		}
	} else {
		for _, r := range records {
			r.totalTime -= r.curTickTotal
			r.count -= r.curCount
			r.curTickTotal = 0
			r.curCount = 0
		}
	}
}

func (r *TimingsRecord) GetID() int { return r.id }
func (r *TimingsRecord) GetParentID() (int, bool) {
	if r.parentRecord == nil {
		return 0, false
	}
	return r.parentRecord.id, true
}
func (r *TimingsRecord) GetTimerID() int                { return r.handler.id }
func (r *TimingsRecord) GetName() string                { return r.handler.GetName() }
func (r *TimingsRecord) GetGroup() string               { return r.handler.GetGroup() }
func (r *TimingsRecord) GetCount() int                  { return r.count }
func (r *TimingsRecord) GetCurCount() int               { return r.curCount }
func (r *TimingsRecord) GetStart() time.Time            { return r.start }
func (r *TimingsRecord) GetTotalTime() time.Duration    { return r.totalTime }
func (r *TimingsRecord) GetCurTickTotal() time.Duration { return r.curTickTotal }
func (r *TimingsRecord) GetViolations() int             { return r.violations }
func (r *TimingsRecord) GetTicksActive() int            { return r.ticksActive }
func (r *TimingsRecord) GetPeakTime() time.Duration     { return r.peakTime }

func (r *TimingsRecord) startTiming(now time.Time) {
	recordsMu.Lock()
	r.start = now
	r.running = true
	currentRecord = r
	recordsMu.Unlock()
}

// stopTiming reports whether the timer was actually stopped (mirrors the PHP original silently
// returning early in the two cases described inline below).
func (r *TimingsRecord) stopTiming(now time.Time) {
	recordsMu.Lock()
	defer recordsMu.Unlock()

	if !r.running {
		return
	}
	if currentRecord != r {
		if currentRecord == nil {
			// timings may have been stopped while this timer was running
			return
		}
		panic("stopTiming() called on a non-current timer")
	}

	currentRecord = r.parentRecord
	diff := now.Sub(r.start)
	r.totalTime += diff
	r.curTickTotal += diff
	r.curCount++
	r.count++
	r.running = false
	if diff > r.peakTime {
		r.peakTime = diff
	}
}

func CurrentRecord() *TimingsRecord {
	recordsMu.Lock()
	defer recordsMu.Unlock()
	return currentRecord
}
