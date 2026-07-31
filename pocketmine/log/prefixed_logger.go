package log

import (
	"fmt"
	"sync"
)

// PrefixedLogger is a port of \PrefixedLogger: wraps another Logger, prepending a prefix to
// every message. See SimpleLogger's doc comment for why this doesn't embed SimpleLogger.
type PrefixedLogger struct {
	mu       sync.Mutex
	delegate Logger
	prefix   string
}

func NewPrefixedLogger(delegate Logger, prefix string) *PrefixedLogger {
	return &PrefixedLogger{delegate: delegate, prefix: prefix}
}

func (p *PrefixedLogger) Emergency(m string) { p.Log(Emergency, m) }
func (p *PrefixedLogger) Alert(m string)     { p.Log(Alert, m) }
func (p *PrefixedLogger) Critical(m string)  { p.Log(Critical, m) }
func (p *PrefixedLogger) Error(m string)     { p.Log(Error, m) }
func (p *PrefixedLogger) Warning(m string)   { p.Log(Warning, m) }
func (p *PrefixedLogger) Notice(m string)    { p.Log(Notice, m) }
func (p *PrefixedLogger) Info(m string)      { p.Log(Info, m) }
func (p *PrefixedLogger) Debug(m string)     { p.Log(Debug, m) }

func (p *PrefixedLogger) Log(level Level, message string) {
	p.mu.Lock()
	prefix := p.prefix
	p.mu.Unlock()
	p.delegate.Log(level, fmt.Sprintf("[%s] %s", prefix, message))
}

func (p *PrefixedLogger) LogException(err error, trace string) {
	p.delegate.LogException(err, trace)
}

func (p *PrefixedLogger) GetPrefix() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.prefix
}

func (p *PrefixedLogger) SetPrefix(prefix string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prefix = prefix
}
