// Command pocketmine-go starts the server: a port of PocketMine.php, the entry point that prepares
// the data folder and hands over to pocketmine\Server.
//
// Options (PocketMine.php's BootstrapOptions): --data=<path> (defaults to the working directory),
// --plugins=<path>, --version, --enable-ansi, --disable-ansi, --no-log-file. Any server.properties
// or pocketmine.yml key can be overridden with --key=value, e.g. --server-port=19133
// (ServerConfigGroup's getopt).
//
// Not ported: the platform/PHP dependency checks, the setup wizard (--no-wizard is accepted and
// ignored; server.properties is created with defaults) and ThreadManager (goroutines end with the
// process).
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/utils"

	// Linked for its init(): the default commands (SimpleCommandMap::setDefaultCommands).
	_ "pocketmine-go/pocketmine/command/defaults"
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

// hasOpt reports whether --opt was given.
func hasOpt(opt string) bool {
	for _, arg := range os.Args[1:] {
		if arg == "--"+opt || strings.HasPrefix(arg, "--"+opt+"=") {
			return true
		}
	}
	return false
}

// criticalError is PocketMine.php's critical_error.
func criticalError(message string) {
	fmt.Println("[ERROR] " + message)
}

func main() {
	os.Exit(run())
}

// run is PocketMine.php's server(): it returns the process exit code.
func run() int {
	if hasOpt(pocketmine.OptVersion) {
		fmt.Printf("%s %s (git hash %s) for Minecraft: Bedrock Edition %s\n", pocketmine.Name, pocketmine.Version().GetFullVersion(true), pocketmine.GitHash(), protocol.CurrentVersion)
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	dataPath := cwd
	if v, ok := getoptString(pocketmine.OptData); ok {
		dataPath = v
	}
	pluginPath := filepath.Join(cwd, "plugins")
	if v, ok := getoptString(pocketmine.OptPlugins); ok {
		pluginPath = v
	}

	if err := os.MkdirAll(dataPath, 0o777); err != nil {
		criticalError(fmt.Sprintf("Unable to create/access data directory at %s. Check that the target location is accessible by the current user.", dataPath))
		return 1
	}
	if abs, err := filepath.Abs(dataPath); err == nil {
		dataPath = abs
	}

	lockFilePath := filepath.Join(dataPath, "server.lock")
	pid, err := utils.CreateLockFile(lockFilePath)
	if err != nil {
		criticalError(err.Error())
		criticalError("Please ensure that there is enough space on the disk and that the current user has read/write permissions to the selected data directory " + dataPath + ".")
		return 1
	}
	if pid != nil {
		criticalError(fmt.Sprintf("Another %s instance (PID %d) is already using this folder (%s).", pocketmine.Name, *pid, dataPath))
		criticalError("Please stop the other server first before running a new one.")
		return 1
	}
	defer func() { _ = utils.ReleaseLockFile(lockFilePath) }()

	if err := os.MkdirAll(pluginPath, 0o777); err != nil {
		criticalError(fmt.Sprintf("Unable to create plugin directory at %s. Check that the target location is accessible by the current user.", pluginPath))
		return 1
	}
	if abs, err := filepath.Abs(pluginPath); err == nil {
		pluginPath = abs
	}

	utils.InitTimezone()

	switch {
	case hasOpt(pocketmine.OptEnableANSI):
		enable := true
		utils.InitTerminal(&enable)
	case hasOpt(pocketmine.OptDisableANSI):
		enable := false
		utils.InitTerminal(&enable)
	default:
		utils.InitTerminal(nil)
	}
	logFile := filepath.Join(dataPath, "server.log")
	if hasOpt(pocketmine.OptNoLogFile) {
		logFile = ""
	}

	location, err := time.LoadLocation(utils.GetTimezone())
	if err != nil {
		location = time.Local
	}
	logger, err := utils.NewMainLogger(logFile, utils.HasFormattingCodes(), "Server", location, false, filepath.Join(dataPath, "log_archive"))
	if err != nil {
		criticalError(err.Error())
		return 1
	}
	defer logger.Shutdown()
	if logFile == "" {
		logger.Notice("Logging to file disabled. Ensure logs are collected by other means (e.g. Docker logs).")
	}
	log.SetGlobal(logger)

	srv, err := server.NewWithPluginPath(dataPath, pluginPath, logger)
	if err != nil {
		logger.Emergency(err.Error())
		return 1
	}

	// SignalHandler: Ctrl+C and SIGTERM shut the server down cleanly (saving worlds and players).
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		logger.Info("Received signal interrupt, stopping the server")
		srv.Shutdown()
	}()

	if err := srv.Start(); err != nil {
		logger.Emergency(err.Error())
		return 1
	}

	logger.Info("Stopping other threads")
	fmt.Println(utils.FormatReset)
	return 0
}
