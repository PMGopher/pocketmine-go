// Command pocketmine-go starts the server: a port of PocketMine.php, the entry point that prepares
// the data folder and hands over to pocketmine\Server.
//
// Options (PocketMine.php's BootstrapOptions): --data=<path> (defaults to the working directory),
// --version. Any server.properties key can be overridden with --key=value, e.g. --server-port=19133
// (ServerConfigGroup's getopt).
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/server"
)

// getoptString is PocketMine.php's getopt_string: the value of --opt=value, if given.
func getoptString(opt string) (string, bool) {
	for _, arg := range os.Args[1:] {
		if value, ok := strings.CutPrefix(arg, "--"+opt+"="); ok {
			return value, true
		}
	}
	return "", false
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--"+pocketmine.OptVersion {
			fmt.Printf("%s %s\n", pocketmine.Name, pocketmine.Version().GetFullVersion(true))
			return
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	dataPath := cwd
	if v, ok := getoptString(pocketmine.OptData); ok {
		dataPath = v
	}
	if err := os.MkdirAll(dataPath, 0o777); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create/access data directory at %s. Check that the target location is accessible by the current user.\n", dataPath)
		os.Exit(1)
	}
	if abs, err := filepath.Abs(dataPath); err == nil {
		dataPath = abs
	}

	logger := log.NewSimpleLogger()
	log.SetGlobal(logger)

	srv, err := server.New(dataPath, logger)
	if err != nil {
		logger.Emergency(err.Error())
		os.Exit(1)
	}

	// SignalHandler: Ctrl+C and SIGTERM shut the server down cleanly (saving worlds and players).
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		logger.Info("Stopping the server...")
		srv.Shutdown()
	}()

	if err := srv.Start(); err != nil {
		logger.Emergency(err.Error())
		os.Exit(1)
	}
}
