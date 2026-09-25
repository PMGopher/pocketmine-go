package server

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/event"
	serverevent "pocketmine-go/pocketmine/event/server"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/timings"
	"pocketmine-go/pocketmine/utils"
)

// MemoryManager constants, MemoryManager::DEFAULT_*.
const (
	memoryDefaultCheckRate             = TargetTicksPerSecond
	memoryDefaultContinuousTriggerRate = TargetTicksPerSecond * 2
	memoryDefaultTicksPerGC            = 30 * 60 * TargetTicksPerSecond
)

// PruneChunkCachesFunc is ChunkCache::pruneCaches, installed by the network/mcpe package.
var PruneChunkCachesFunc func()

// MemoryManager is a port of pocketmine\MemoryManager: watches memory usage against the limits in
// pocketmine.yml, fires LowMemoryEvent and lowers the allowed view distance while memory is low.
//
// PHP's hard limit is memory_limit, which kills the process when exceeded; Go has no such hard
// limit, so memory.main-hard-limit is applied as the runtime's soft limit (debug.SetMemoryLimit),
// which makes the garbage collector work harder near it instead. Memory dumps (MemoryDump) and the
// cycle collector (GarbageCollectorManager) are PHP-runtime specific and not ported.
type MemoryManager struct {
	server *Server
	logger log.Logger

	memoryLimit       uint64
	globalMemoryLimit uint64
	checkRate         int
	checkTicker       int
	lowMemory         bool

	continuousTrigger       bool
	continuousTriggerRate   int
	continuousTriggerCount  int
	continuousTriggerTicker int

	garbageCollectionPeriod int
	garbageCollectionTicker int

	lowMemChunkRadiusOverride int
}

var memoryLimitPattern = regexp.MustCompile(`([0-9]+)([KMGkmg])`)

func NewMemoryManager(server *Server) *MemoryManager {
	m := &MemoryManager{server: server, logger: log.NewPrefixedLogger(server.GetLogger(), "Memory Manager")}
	m.init(server.GetConfigGroup())
	return m
}

func (m *MemoryManager) init(config *ServerConfigGroup) {
	m.memoryLimit = uint64(max(0, config.GetPropertyInt(YmlMemoryMainLimit, 0))) * 1024 * 1024

	defaultMemory := 1024
	if matches := memoryLimitPattern.FindStringSubmatch(config.GetConfigString("memory-limit", "")); matches != nil {
		v, _ := strconv.Atoi(matches[1])
		if v <= 0 {
			defaultMemory = 0
		} else {
			switch strings.ToUpper(matches[2]) {
			case "K":
				defaultMemory = v / 1024
			case "M":
				defaultMemory = v
			case "G":
				defaultMemory = v * 1024
			default:
				defaultMemory = v
			}
		}
	}

	if hardLimit := config.GetPropertyInt(YmlMemoryMainHardLimit, defaultMemory); hardLimit > 0 {
		debug.SetMemoryLimit(int64(hardLimit) * 1024 * 1024)
	}

	m.globalMemoryLimit = uint64(max(0, config.GetPropertyInt(YmlMemoryGlobalLimit, 0))) * 1024 * 1024
	m.checkRate = config.GetPropertyInt(YmlMemoryCheckRate, memoryDefaultCheckRate)
	m.continuousTrigger = config.GetPropertyBool(YmlMemoryContinuousTrigger, true)
	m.continuousTriggerRate = config.GetPropertyInt(YmlMemoryContinuousTriggerRate, memoryDefaultContinuousTriggerRate)
	m.garbageCollectionPeriod = config.GetPropertyInt(YmlMemoryGarbageCollectionPeriod, memoryDefaultTicksPerGC)
	m.lowMemChunkRadiusOverride = config.GetPropertyInt(YmlMemoryMaxChunksChunkRadius, 4)
}

func (m *MemoryManager) IsLowMemory() bool { return m.lowMemory }

func (m *MemoryManager) GetGlobalMemoryLimit() uint64 { return m.globalMemoryLimit }

func (m *MemoryManager) CanUseChunkCache() bool { return !m.lowMemory }

// GetViewDistance returns the allowed chunk radius based on the current memory usage.
func (m *MemoryManager) GetViewDistance(distance int) int {
	if m.lowMemory && m.lowMemChunkRadiusOverride > 0 {
		return min(m.lowMemChunkRadiusOverride, distance)
	}
	return distance
}

// Trigger is a port of MemoryManager::trigger: triggers garbage collection and cache cleanup to
// try and free memory.
func (m *MemoryManager) Trigger(memory, limit uint64, global bool, triggerCount int) {
	prefix := ""
	if global {
		prefix = "Global "
	}
	m.logger.Debug(fmt.Sprintf("%sLow memory triggered, limit %gMB, using %gMB", prefix, round2(float64(limit)/1024/1024), round2(float64(memory)/1024/1024)))
	if PruneChunkCachesFunc != nil {
		PruneChunkCachesFunc()
	}

	ev := serverevent.NewLowMemoryEvent(memory, limit, global, triggerCount)
	event.Call(ev)

	m.TriggerGarbageCollector()

	m.logger.Debug(fmt.Sprintf("Freed %gMB", round2(float64(ev.GetMemoryFreed())/1024/1024)))
}

// Check is a port of MemoryManager::check, called every tick.
func (m *MemoryManager) Check() {
	timings.MemoryManager.StartTiming()
	defer timings.MemoryManager.StopTiming()

	if (m.memoryLimit > 0 || m.globalMemoryLimit > 0) && func() bool { m.checkTicker++; return m.checkTicker >= m.checkRate }() {
		m.checkTicker = 0
		_, mainUsage, globalUsage := utils.AdvancedMemoryUsage()
		trigger := -1
		usage := uint64(0)
		if m.memoryLimit > 0 && mainUsage > m.memoryLimit {
			trigger, usage = 0, mainUsage
		} else if m.globalMemoryLimit > 0 && globalUsage > m.globalMemoryLimit {
			trigger, usage = 1, globalUsage
		}

		if trigger != -1 {
			if m.lowMemory && m.continuousTrigger {
				m.continuousTriggerTicker++
				if m.continuousTriggerTicker >= m.continuousTriggerRate {
					m.continuousTriggerTicker = 0
					m.continuousTriggerCount++
					m.Trigger(usage, m.memoryLimit, trigger > 0, m.continuousTriggerCount)
				}
			} else {
				m.lowMemory = true
				m.continuousTriggerCount = 0
				m.Trigger(usage, m.memoryLimit, trigger > 0, 0)
			}
		} else {
			m.lowMemory = false
		}
	}

	if m.garbageCollectionPeriod > 0 {
		m.garbageCollectionTicker++
		if m.garbageCollectionTicker >= m.garbageCollectionPeriod {
			m.garbageCollectionTicker = 0
			m.TriggerGarbageCollector()
		}
	}
}

// TriggerGarbageCollector is a port of MemoryManager::triggerGarbageCollector: shuts down idle
// async workers and runs the garbage collector.
func (m *MemoryManager) TriggerGarbageCollector() {
	timings.GarbageCollector.StartTiming()
	defer timings.GarbageCollector.StopTiming()

	if pool := m.server.GetAsyncPool(); pool != nil {
		if w := pool.ShutdownUnusedWorkers(); w > 0 {
			m.logger.Debug(fmt.Sprintf("Shut down %d idle async pool workers", w))
		}
	}
	runtime.GC()
	debug.FreeOSMemory()
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
