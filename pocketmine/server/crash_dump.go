package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"pocketmine-go/pocketmine/crash"
	"pocketmine-go/pocketmine/lang"
)

// crashExit ends the process after a crash dump (PHP: `@Process::kill(Process::pid()); exit(1)`).
// Tests replace it.
var crashExit = func() { os.Exit(1) }

// crashThrottleSleep is the sleep of Server::crashDump's minimum uptime. Tests replace it.
var crashThrottleSleep = time.Sleep

// tickOrCrash runs one tick; a panic in it is handled like an uncaught exception on PHP's main
// thread. The server lock is held.
func (s *Server) tickOrCrash() {
	defer func() {
		if p := recover(); p != nil {
			pcs := make([]uintptr, 64)
			pcs = pcs[:runtime.Callers(2, pcs)]
			s.mu.Unlock()
			s.ExceptionHandler(p, pcs)
			s.mu.Lock() // only reached when the server was already stopping (crashDump returns)
		}
	}()
	s.tick()
}

var whitespaceRun = regexp.MustCompile(`\s+`)

// ExceptionHandler is a port of Server::exceptionHandler: it logs a panic recovered from the tick
// or from packet handling (p, with the stack pcs of the goroutine that panicked) and writes a
// crash dump. The server lock must not be held. It only returns if the server isn't running.
func (s *Server) ExceptionHandler(p any, pcs []uintptr) {
	c := crash.FromPanic(p, pcs, "Main")
	var trace strings.Builder
	for i, frame := range c.Trace {
		trace.WriteString(frame.PrintableFrame(i) + "\n")
	}
	s.logger.LogException(fmt.Errorf("%s: %s", c.Type, c.Message), trace.String())

	c.Message = whitespaceRun.ReplaceAllString(strings.TrimSpace(c.Message), " ")
	s.lastExceptionError = &c
	s.crashDump()
}

// writeCrashDumpFile is a port of Server::writeCrashDumpFile.
func (s *Server) writeCrashDumpFile(dump *crash.CrashDump) (string, error) {
	crashFolder := filepath.Join(s.dataPath, "crashdumps")
	if err := os.MkdirAll(crashFolder, 0o777); err != nil {
		return "", err
	}
	crashTime := time.Unix(0, int64(dump.GetData().Time*1e9))
	crashDumpPath := filepath.Join(crashFolder, crashTime.Format("2006-01-02_15.04.05_MST")+".log")

	fp, err := os.Create(crashDumpPath)
	if err != nil {
		return "", fmt.Errorf("Unable to open new file to generate crashdump")
	}
	defer fp.Close()
	writer := crash.NewCrashDumpRenderer(fp, dump.GetData())
	writer.RenderHumanReadable()
	dump.EncodeData(writer)

	return crashDumpPath, nil
}

// crashDump is a port of Server::crashDump.
//
// PHP then submits the dump to the crash archive (auto-report.host, crash.pmmp.io by default).
// That archive only accepts PocketMine-MP (PHP) crash dumps, so nothing is submitted; the dump is
// only written to crashdumps/.
func (s *Server) crashDump() {
	if !s.isRunning.Load() {
		return
	}
	s.hasStopped = false

	func() {
		defer func() {
			if p := recover(); p != nil {
				err := fmt.Errorf("%v", p)
				s.logger.LogException(err, "")
				s.logger.Critical(s.language.Translate(lang.KnownTranslationFactory.PocketmineCrashError(err.Error())))
			}
		}()
		s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineCrashCreate()))
		c := crash.Crash{Type: "Unknown", Message: "Crash error information missing", Thread: "Main"}
		if s.lastExceptionError != nil {
			c = *s.lastExceptionError
		}
		dump, err := crash.NewCrashDump(s, s.pluginManager, c)
		if err != nil {
			panic(err)
		}

		crashDumpPath, err := s.writeCrashDumpFile(dump)
		if err != nil {
			panic(err)
		}

		s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineCrashSubmit(crashDumpPath)))
	}()

	s.ForceShutdown()
	s.isRunning.Store(false)

	//Force minimum uptime to be >= 120 seconds, to reduce the impact of spammy crash loops
	uptime := int(time.Since(s.startTime).Seconds())
	minUptime := 120
	if spacing := minUptime - uptime; spacing > 0 {
		fmt.Printf("--- Uptime %ds - waiting %ds to throttle automatic restart (you can kill the process safely now) ---\n", uptime, spacing)
		crashThrottleSleep(time.Duration(spacing) * time.Second)
	}
	crashExit()
}
