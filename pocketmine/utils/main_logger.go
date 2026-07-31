package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pocketmine-go/pocketmine/log"
)

// attachmentEntry pairs a registered log.Attachment with the handle used to remove it later —
// see log.AttachmentHandle's doc comment for why a handle is needed instead of identity.
type attachmentEntry struct {
	handle log.AttachmentHandle
	fn     log.Attachment
}

// MainLogger is a combined port of pocketmine\utils\MainLogger and MainLoggerThread, implementing
// the pocketmine/log AttachableLogger and BufferedLogger interfaces (PHP: AttachableThreadSafeLogger
// implements \BufferedLogger).
//
// PHP splits this into two classes because pthreads Threads need their own class with a run()
// loop and a ThreadSafeArray for cross-thread buffering. A goroutine reading from a buffered
// channel does that job natively in Go, so this single type owns both the synchronous logging
// API (Info/Warning/etc.) and the background writer that persists lines to disk and rotates the
// log file — there's no separate "thread" type needed.
type MainLogger struct {
	mu                   sync.Mutex
	logDebug             bool
	useFormattingCodes   bool
	mainThreadName       string
	location             *time.Location
	format               string
	attachments          []attachmentEntry
	nextAttachmentHandle log.AttachmentHandle

	logFile     string
	archiveDir  string
	maxFileSize int64

	lines    chan string
	flush    chan chan struct{}
	shutdown chan struct{}
	done     chan struct{}
}

var _ log.AttachableLogger = (*MainLogger)(nil)
var _ log.BufferedLogger = (*MainLogger)(nil)

// NewMainLogger constructs a MainLogger. If logFile is "", nothing is persisted to disk (only
// the terminal/attachments receive messages) — this covers PHP's `?string $logFile` being null.
func NewMainLogger(logFile string, useFormattingCodes bool, mainThreadName string, location *time.Location, logDebug bool, logArchiveDir string) (*MainLogger, error) {
	l := &MainLogger{
		logDebug:           logDebug,
		useFormattingCodes: useFormattingCodes,
		mainThreadName:     mainThreadName,
		location:           location,
		format:             Aqua + "[%s] " + Reset + "%s[%s/%s]: %s" + Reset,
		maxFileSize:        32 * 1024 * 1024,
		lines:              make(chan string, 4096),
		flush:              make(chan chan struct{}),
		shutdown:           make(chan struct{}),
		done:               make(chan struct{}),
	}
	if logFile != "" {
		l.logFile = logFile
		l.archiveDir = logArchiveDir
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			f, createErr := os.Create(logFile)
			if createErr != nil {
				return nil, fmt.Errorf("couldn't create log file: %w", createErr)
			}
			f.Close()
		}
		if logArchiveDir != "" {
			if err := os.MkdirAll(logArchiveDir, 0o755); err != nil {
				return nil, fmt.Errorf("unable to create archive directory: %w", err)
			}
		}
		go l.writerLoop()
	} else {
		close(l.done)
	}
	return l, nil
}

func (l *MainLogger) GetFormat() string { return l.format }
func (l *MainLogger) SetFormat(format string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = format
}

func (l *MainLogger) AddAttachment(a log.Attachment) log.AttachmentHandle {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextAttachmentHandle++
	handle := l.nextAttachmentHandle
	l.attachments = append(l.attachments, attachmentEntry{handle: handle, fn: a})
	return handle
}

func (l *MainLogger) RemoveAttachment(handle log.AttachmentHandle) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, e := range l.attachments {
		if e.handle == handle {
			l.attachments = append(l.attachments[:i], l.attachments[i+1:]...)
			break
		}
	}
}

func (l *MainLogger) RemoveAttachments() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attachments = nil
}

func (l *MainLogger) GetAttachments() []log.Attachment {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]log.Attachment, len(l.attachments))
	for i, e := range l.attachments {
		result[i] = e.fn
	}
	return result
}

// Buffer runs buffered() with no additional synchronization. PHP's synchronized() is reentrant
// for the same thread, so buffered() can safely call Info()/Warning()/etc. on the same logger;
// Go's sync.Mutex is not reentrant, so holding a lock here across the call would deadlock the
// moment buffered() logs anything. Each individual line remains atomic (send() still locks
// around its own output), so this only gives up the "no other goroutine's lines interleave
// between mine" guarantee for a whole batch — not per-line corruption.
func (l *MainLogger) Buffer(buffered func()) {
	buffered()
}

func (l *MainLogger) Emergency(message string) { l.send(message, log.Emergency, "EMERGENCY", Red) }
func (l *MainLogger) Alert(message string)     { l.send(message, log.Alert, "ALERT", Red) }
func (l *MainLogger) Critical(message string)  { l.send(message, log.Critical, "CRITICAL", Red) }
func (l *MainLogger) Error(message string)     { l.send(message, log.Error, "ERROR", DarkRed) }
func (l *MainLogger) Warning(message string)   { l.send(message, log.Warning, "WARNING", Yellow) }
func (l *MainLogger) Notice(message string)    { l.send(message, log.Notice, "NOTICE", Aqua) }
func (l *MainLogger) Info(message string)      { l.send(message, log.Info, "INFO", White) }

func (l *MainLogger) Debug(message string) {
	if !l.logDebug {
		return
	}
	l.send(message, log.Debug, "DEBUG", Gray)
}

// DebugForced logs a debug message even when debug logging is disabled, matching
// MainLogger::debug($message, force: true) in PHP.
func (l *MainLogger) DebugForced(message string) {
	l.send(message, log.Debug, "DEBUG", Gray)
}

func (l *MainLogger) SetLogDebug(logDebug bool) { l.logDebug = logDebug }

func (l *MainLogger) Log(level log.Level, message string) {
	switch level {
	case log.Emergency:
		l.Emergency(message)
	case log.Alert:
		l.Alert(message)
	case log.Critical:
		l.Critical(message)
	case log.Error:
		l.Error(message)
	case log.Warning:
		l.Warning(message)
	case log.Notice:
		l.Notice(message)
	case log.Info:
		l.Info(message)
	case log.Debug:
		l.Debug(message)
	}
}

// LogException mirrors MainLogger::logException(): logs a Go error (with its stack trace, if
// one was captured) and forces a synchronous flush.
func (l *MainLogger) LogException(err error, stackTrace string) {
	msg := err.Error()
	if stackTrace != "" {
		msg += "\n--- Stack trace ---\n" + stackTrace
	}
	l.Critical(msg)
	l.SyncFlushBuffer()
}

func (l *MainLogger) send(message string, level log.Level, prefix string, color string) {
	now := time.Now()
	if l.location != nil {
		now = now.In(l.location)
	}

	threadName := l.mainThreadName + " thread"

	formatted := fmt.Sprintf(l.format, now.Format("15:04:05.000"), color, threadName, prefix, AddBase(color, Clean(message, false)))

	if !IsTerminalInit() {
		InitTerminal(&l.useFormattingCodes) //lazy-init colour codes, matching the PHP original
	}

	l.mu.Lock()
	WriteTerminalLine(formatted)
	if l.logFile != "" {
		select {
		case l.lines <- now.Format("2006-01-02") + " " + Clean(formatted, true) + "\n":
		case <-l.done:
		}
	}
	attachments := l.attachments
	l.mu.Unlock()

	for _, a := range attachments {
		a.fn(level, formatted)
	}
}

// SyncFlushBuffer blocks until every line queued so far has been written to disk.
func (l *MainLogger) SyncFlushBuffer() {
	if l.logFile == "" {
		return
	}
	ack := make(chan struct{})
	select {
	case l.flush <- ack:
		<-ack
	case <-l.done:
	}
}

// Shutdown stops the background writer goroutine, flushing any remaining lines first. Callers
// must invoke this explicitly (see DestructorCallbacks) — Go has no equivalent to PHP's
// __destruct() to do it automatically.
func (l *MainLogger) Shutdown() {
	if l.logFile == "" {
		return
	}
	select {
	case <-l.shutdown:
		// already closed
	default:
		close(l.shutdown)
	}
	<-l.done
}

func (l *MainLogger) writerLoop() {
	defer close(l.done)

	logResource, size, err := openLogFile(l.logFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "MainLogger: "+err.Error())
		return
	}
	defer logResource.Close()

	if l.archiveDir != "" && size >= l.maxFileSize {
		logResource, size = l.archiveLogFile(logResource, size)
	}

	for {
		select {
		case line := <-l.lines:
			logResource, size = l.writeLine(logResource, size, line)
		case ack := <-l.flush:
			logResource, size = l.drainPending(logResource, size)
			close(ack)
		case <-l.shutdown:
			logResource, size = l.drainPending(logResource, size)
			return
		}
	}
}

// drainPending writes every currently-queued line and returns the (possibly rotated) file
// handle and size, since archiving closes the old handle and opens a new one.
func (l *MainLogger) drainPending(f *os.File, size int64) (*os.File, int64) {
	for {
		select {
		case line := <-l.lines:
			f, size = l.writeLine(f, size, line)
		default:
			return f, size
		}
	}
}

// writeLine appends line to f and returns the (possibly rotated) file handle and new size.
func (l *MainLogger) writeLine(f *os.File, size int64, line string) (*os.File, int64) {
	f.WriteString(line)
	size += int64(len(line))
	if l.archiveDir != "" && size >= l.maxFileSize {
		f, size = l.archiveLogFile(f, size)
	}
	return f, size
}

func openLogFile(path string) (*os.File, int64, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, 0, fmt.Errorf("couldn't open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, fmt.Errorf("couldn't stat log file: %w", err)
	}
	return f, info.Size(), nil
}

func (l *MainLogger) archiveLogFile(f *os.File, size int64) (*os.File, int64) {
	f.Close()

	baseName := strings.TrimSuffix(filepath.Base(l.logFile), filepath.Ext(l.logFile))
	extension := strings.TrimPrefix(filepath.Ext(l.logFile), ".")
	dateStr := time.Now().Format("2006-01-02T15.04.05")

	var fileName, out string
	for i := 0; ; i++ {
		fileName = fmt.Sprintf("%s.%s.%d.%s", baseName, dateStr, i, extension)
		out = filepath.Join(l.archiveDir, fileName)
		if _, err := os.Stat(out); os.IsNotExist(err) {
			break
		}
	}

	os.MkdirAll(l.archiveDir, 0o755)
	os.Rename(l.logFile, out)

	newF, newSize, err := openLogFile(l.logFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "MainLogger: "+err.Error())
		return f, size
	}
	newF.WriteString(fmt.Sprintf("--- Starting new log file - old log file archived as %s ---\n", fileName))
	return newF, newSize
}
