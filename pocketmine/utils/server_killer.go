package utils

import (
	"fmt"
	"time"
)

// ServerKiller is a port of pocketmine\utils\ServerKiller: force-kills the process if a
// graceful stop takes longer than the given timeout.
//
// PHP's original is a pthreads Thread using synchronized wait()/notify(); Go's time.Timer plus
// a stop channel does the same job far more simply, without needing a real OS thread.
type ServerKiller struct {
	timeout time.Duration
	stop    chan struct{}
}

func NewServerKiller(timeoutSeconds int) *ServerKiller {
	return &ServerKiller{
		timeout: time.Duration(timeoutSeconds) * time.Second,
		stop:    make(chan struct{}),
	}
}

// Start begins the countdown in a new goroutine, mirroring Thread::start()+onRun().
func (k *ServerKiller) Start() {
	go func() {
		timer := time.NewTimer(k.timeout)
		defer timer.Stop()
		select {
		case <-timer.C:
			fmt.Println("\nTook too long to stop, server was killed forcefully!")
			_ = KillProcess(Pid())
		case <-k.stop:
		}
	}()
}

// Quit cancels the countdown, mirroring ServerKiller::quit().
func (k *ServerKiller) Quit() {
	close(k.stop)
}
