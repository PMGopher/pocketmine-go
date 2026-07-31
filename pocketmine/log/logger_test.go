package log

import (
	"errors"
	"testing"
)

// mockLogger is a fully independent Logger implementation (no embedding), used so tests aren't
// tripped up by the same "embedding doesn't give virtual dispatch" gotcha documented on SimpleLogger.
type mockLogger struct {
	lastLevel    Level
	lastMessage  string
	lastErr      error
	lastErrTrace string
}

func (m *mockLogger) Emergency(s string) { m.Log(Emergency, s) }
func (m *mockLogger) Alert(s string)     { m.Log(Alert, s) }
func (m *mockLogger) Critical(s string)  { m.Log(Critical, s) }
func (m *mockLogger) Error(s string)     { m.Log(Error, s) }
func (m *mockLogger) Warning(s string)   { m.Log(Warning, s) }
func (m *mockLogger) Notice(s string)    { m.Log(Notice, s) }
func (m *mockLogger) Info(s string)      { m.Log(Info, s) }
func (m *mockLogger) Debug(s string)     { m.Log(Debug, s) }
func (m *mockLogger) Log(level Level, message string) {
	m.lastLevel = level
	m.lastMessage = message
}
func (m *mockLogger) LogException(err error, trace string) {
	m.lastErr = err
	m.lastErrTrace = trace
}

func TestPrefixedLoggerPrependsPrefix(t *testing.T) {
	delegate := &mockLogger{}
	logger := NewPrefixedLogger(delegate, "Server thread")
	logger.Info("hello")

	if delegate.lastLevel != Info {
		t.Fatalf("lastLevel = %v, want Info", delegate.lastLevel)
	}
	if delegate.lastMessage != "[Server thread] hello" {
		t.Fatalf("lastMessage = %q, want %q", delegate.lastMessage, "[Server thread] hello")
	}
}

func TestPrefixedLoggerEachLevelRoutesThroughItsOwnLog(t *testing.T) {
	// Regression test for the embedding pitfall documented on SimpleLogger: PrefixedLogger must
	// NOT embed SimpleLogger, or Emergency() etc. would bypass PrefixedLogger.Log()'s prefixing.
	delegate := &mockLogger{}
	logger := NewPrefixedLogger(delegate, "X")
	logger.Emergency("boom")
	if delegate.lastLevel != Emergency || delegate.lastMessage != "[X] boom" {
		t.Fatalf("got level=%v message=%q, want Emergency/[X] boom", delegate.lastLevel, delegate.lastMessage)
	}
}

func TestPrefixedLoggerSetPrefix(t *testing.T) {
	delegate := &mockLogger{}
	logger := NewPrefixedLogger(delegate, "old")
	logger.SetPrefix("new")
	logger.Notice("hi")
	if delegate.lastMessage != "[new] hi" {
		t.Fatalf("lastMessage = %q, want %q", delegate.lastMessage, "[new] hi")
	}
}

func TestPrefixedLoggerLogExceptionDelegates(t *testing.T) {
	delegate := &mockLogger{}
	logger := NewPrefixedLogger(delegate, "X")
	err := errors.New("boom")
	logger.LogException(err, "trace here")
	if delegate.lastErr != err || delegate.lastErrTrace != "trace here" {
		t.Fatalf("LogException did not delegate correctly: err=%v trace=%q", delegate.lastErr, delegate.lastErrTrace)
	}
}

func TestGlobalLoggerDefaultsToSimpleLogger(t *testing.T) {
	SetGlobal(nil)
	if _, ok := Global().(*SimpleLogger); !ok {
		t.Fatalf("Global() = %T, want *SimpleLogger by default", Global())
	}
}

func TestGlobalLoggerSetGlobal(t *testing.T) {
	custom := &mockLogger{}
	SetGlobal(custom)
	defer SetGlobal(nil)
	if Global() != Logger(custom) {
		t.Fatalf("Global() did not return the logger set via SetGlobal")
	}
}
