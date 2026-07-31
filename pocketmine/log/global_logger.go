package log

import "sync"

// GlobalLogger is a port of \GlobalLogger: a global accessor for the process-wide logger.
var (
	globalMu     sync.Mutex
	globalLogger Logger
)

func Global() Logger {
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalLogger == nil {
		globalLogger = NewSimpleLogger()
	}
	return globalLogger
}

func SetGlobal(logger Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}
