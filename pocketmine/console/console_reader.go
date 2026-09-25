package console

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"

	"pocketmine-go/pocketmine/log"
)

// ConsoleReader is a port of pocketmine\console\ConsoleReader together with
// ConsoleReaderChildProcessDaemon: stdin is read line by line in the background, and ReadLine
// returns the next complete command without blocking (PHP polls stdin through a child process
// because stream_select() on stdin doesn't work on Windows; a goroutine does the same job here,
// so the child process and its IPC token scheme aren't needed).
type ConsoleReader struct {
	logger log.Logger
	lines  chan string
	done   chan struct{}
}

// NewConsoleReader starts reading from os.Stdin.
func NewConsoleReader(logger log.Logger) *ConsoleReader {
	return NewConsoleReaderFrom(logger, os.Stdin)
}

// NewConsoleReaderFrom starts reading lines from r.
func NewConsoleReaderFrom(logger log.Logger, r io.Reader) *ConsoleReader {
	c := &ConsoleReader{logger: log.NewPrefixedLogger(logger, "Console Reader Daemon"), lines: make(chan string, 64), done: make(chan struct{})}
	go c.run(r)
	return c
}

func (c *ConsoleReader) run(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case c.lines <- scanner.Text():
		case <-c.done:
			return
		}
	}
	if err := scanner.Err(); err != nil {
		c.logger.Debug("Stopped reading stdin: " + err.Error())
	}
}

var (
	// escapeSequences are the terminal escape sequences (arrow keys etc.) stripped from commands.
	escapeSequences = regexp.MustCompile(`\x1b\x5b([^\x1b]*\x7e|[\x40-\x50])`)
	controlChars    = regexp.MustCompile(`[[:cntrl:]]`)
)

// ReadLine is a port of ConsoleReaderChildProcessDaemon::readLine: the next command typed in the
// console, or "" (PHP's null) if there is none right now.
func (c *ConsoleReader) ReadLine() (string, bool) {
	for {
		select {
		case raw := <-c.lines:
			command := escapeSequences.ReplaceAllString(strings.TrimSpace(raw), "")
			command = controlChars.ReplaceAllString(command, "")
			if command != "" {
				return command, true
			}
		default:
			return "", false
		}
	}
}

// Quit is a port of ConsoleReaderChildProcessDaemon::quit: stops handing out lines.
func (c *ConsoleReader) Quit() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}
