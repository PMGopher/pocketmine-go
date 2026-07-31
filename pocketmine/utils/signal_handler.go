package utils

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler is a port of pocketmine\utils\SignalHandler.
//
// PHP's original branches between sapi_windows_set_ctrl_handler (Windows) and pcntl_signal
// (Unix: SIGTERM/SIGINT/SIGHUP) because those are two unrelated PHP extension APIs. Go's
// os/signal package abstracts this already: signal.Notify listens for os.Interrupt (SIGINT,
// and Windows Ctrl+C/Break) and syscall.SIGTERM on every platform PocketMine targets, so no
// OS branching is needed here.
type SignalHandler struct {
	ch   chan os.Signal
	stop chan struct{}
}

// NewSignalHandler registers interruptCallback to run once when the process receives an
// interrupt/terminate signal.
func NewSignalHandler(interruptCallback func()) *SignalHandler {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	h := &SignalHandler{ch: ch, stop: make(chan struct{})}
	go func() {
		select {
		case <-ch:
			interruptCallback()
		case <-h.stop:
		}
	}()
	return h
}

// Unregister stops listening for signals.
func (h *SignalHandler) Unregister() {
	close(h.stop)
	signal.Stop(h.ch)
}
