package timings

import (
	"sync"
	"time"
)

// GroupMinecraft mirrors Timings::GROUP_MINECRAFT.
const GroupMinecraft = "Minecraft"

// TargetTimePerTick mirrors Server::TARGET_NANOSECONDS_PER_TICK (20 TPS -> 50ms/tick), used to
// count "violations" (ticks where a timer ran long).
const TargetTimePerTick = 50 * time.Millisecond

var (
	handlerIDMu   sync.Mutex
	nextHandlerID int

	enabledMu   sync.Mutex
	enabled     bool
	timingStart time.Time

	toggleCallbacks []func(enable bool)
	reloadCallbacks []func()
)

func IsEnabled() bool {
	enabledMu.Lock()
	defer enabledMu.Unlock()
	return enabled
}

// SetEnabled is a port of TimingsHandler::setEnabled().
func SetEnabled(enable bool) {
	enabledMu.Lock()
	if enable == enabled {
		enabledMu.Unlock()
		return
	}
	enabled = enable
	internalReloadLocked()
	callbacks := append([]func(bool){}, toggleCallbacks...)
	enabledMu.Unlock()

	for _, cb := range callbacks {
		cb(enable)
	}
}

func AddToggleCallback(cb func(enable bool)) {
	enabledMu.Lock()
	defer enabledMu.Unlock()
	toggleCallbacks = append(toggleCallbacks, cb)
}

func AddReloadCallback(cb func()) {
	enabledMu.Lock()
	defer enabledMu.Unlock()
	reloadCallbacks = append(reloadCallbacks, cb)
}

func GetStartTime() time.Time {
	enabledMu.Lock()
	defer enabledMu.Unlock()
	return timingStart
}

func internalReloadLocked() {
	ResetRecords()
	if enabled {
		timingStart = time.Now()
	}
}

// Reload is a port of TimingsHandler::reload().
func Reload() {
	enabledMu.Lock()
	internalReloadLocked()
	callbacks := append([]func(){}, reloadCallbacks...)
	enabledMu.Unlock()

	for _, cb := range callbacks {
		cb()
	}
}

// Tick is a port of TimingsHandler::tick().
func Tick(measure bool) {
	if IsEnabled() {
		TickRecords(measure)
	}
}

// TimingsHandler is a port of pocketmine\timings\TimingsHandler.
type TimingsHandler struct {
	id     int
	name   string
	parent *TimingsHandler
	group  string

	mu              sync.Mutex
	rootRecord      *TimingsRecord
	timingDepth     int
	recordsByParent map[int]*TimingsRecord // keyed by parent record's id
}

// NewTimingsHandler mirrors the TimingsHandler constructor. group defaults to GroupMinecraft.
func NewTimingsHandler(name string, parent *TimingsHandler, group string) *TimingsHandler {
	if group == "" {
		group = GroupMinecraft
	}
	handlerIDMu.Lock()
	nextHandlerID++
	id := nextHandlerID
	handlerIDMu.Unlock()

	return &TimingsHandler{id: id, name: name, parent: parent, group: group, recordsByParent: map[int]*TimingsRecord{}}
}

func (h *TimingsHandler) GetName() string  { return h.name }
func (h *TimingsHandler) GetGroup() string { return h.group }

func (h *TimingsHandler) StartTiming() {
	if IsEnabled() {
		h.internalStartTiming(time.Now())
	}
}

func (h *TimingsHandler) internalStartTiming(now time.Time) {
	h.mu.Lock()
	h.timingDepth++
	isOutermost := h.timingDepth == 1
	h.mu.Unlock()

	if !isOutermost {
		return
	}

	if h.parent != nil {
		h.parent.internalStartTiming(now)
	}

	current := CurrentRecord()

	h.mu.Lock()
	var record *TimingsRecord
	if current != nil {
		record = h.recordsByParent[current.id]
		if record == nil {
			record = newTimingsRecord(h, current)
			h.recordsByParent[current.id] = record
		}
	} else {
		if h.rootRecord == nil {
			h.rootRecord = newTimingsRecord(h, nil)
		}
		record = h.rootRecord
	}
	h.mu.Unlock()

	record.startTiming(now)
}

func (h *TimingsHandler) StopTiming() {
	if IsEnabled() {
		h.internalStopTiming(time.Now())
	}
}

func (h *TimingsHandler) internalStopTiming(now time.Time) {
	h.mu.Lock()
	if h.timingDepth == 0 {
		// TODO: it would be nice to bail here, but tracking timing depth across resets and
		// enable/disable would cost more than the limited usefulness of bailing is worth.
		h.mu.Unlock()
		return
	}
	h.timingDepth--
	isOutermost := h.timingDepth == 0
	h.mu.Unlock()

	if !isOutermost {
		return
	}

	record := CurrentRecord()
	for record != nil && record.GetTimerID() != h.id {
		// A timer higher up the stack should have been stopped before this one; this shouldn't
		// normally happen, but matches the PHP original's defensive recovery rather than panicking.
		record.stopTiming(now)
		record = CurrentRecord()
	}
	if record != nil {
		record.stopTiming(now)
	}

	if h.parent != nil {
		h.parent.internalStopTiming(now)
	}
}

// Time runs fn, timed by h. A free function (not a method) since Go methods can't have their own
// type parameters.
func Time[T any](h *TimingsHandler, fn func() T) T {
	h.StartTiming()
	defer h.StopTiming()
	return fn()
}

// reset is called by ResetRecords for every handler that had live records.
func (h *TimingsHandler) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rootRecord = nil
	h.recordsByParent = map[int]*TimingsRecord{}
	h.timingDepth = 0
}
