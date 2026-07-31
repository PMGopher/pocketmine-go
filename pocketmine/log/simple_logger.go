package log

import (
	"fmt"
	"strings"
)

// SimpleLogger is a port of \SimpleLogger: a bare-bones stdout logger.
//
// PHP's SimpleLogger has Emergency/Alert/etc. each call $this->log(...), and PrefixedLogger
// overrides log() so that inherited call still ends up going through the override. Go has no
// such virtual dispatch through embedding — if PrefixedLogger embedded SimpleLogger, calling
// PrefixedLogger.Emergency() would run the embedded SimpleLogger's own Log(), not the override.
// So each level method here calls s.Log(...) directly, and PrefixedLogger duplicates the same
// eight one-liners against its own Log() instead of embedding this type.
type SimpleLogger struct{}

func NewSimpleLogger() *SimpleLogger { return &SimpleLogger{} }

func (s *SimpleLogger) Emergency(m string) { s.Log(Emergency, m) }
func (s *SimpleLogger) Alert(m string)     { s.Log(Alert, m) }
func (s *SimpleLogger) Critical(m string)  { s.Log(Critical, m) }
func (s *SimpleLogger) Error(m string)     { s.Log(Error, m) }
func (s *SimpleLogger) Warning(m string)   { s.Log(Warning, m) }
func (s *SimpleLogger) Notice(m string)    { s.Log(Notice, m) }
func (s *SimpleLogger) Info(m string)      { s.Log(Info, m) }
func (s *SimpleLogger) Debug(m string)     { s.Log(Debug, m) }

func (s *SimpleLogger) Log(level Level, message string) {
	fmt.Println("[" + strings.ToUpper(string(level)) + "] " + message)
}

func (s *SimpleLogger) LogException(err error, trace string) {
	s.Critical(err.Error())
	if trace != "" {
		fmt.Println(trace)
	}
}
