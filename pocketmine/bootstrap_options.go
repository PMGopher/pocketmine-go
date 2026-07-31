package pocketmine

// BootstrapOptions is a port of pocketmine\BootstrapOptions: names of the command-line options
// PocketMine-MP supports. Other options not listed here can be used to override
// server.properties/pocketmine.yml values temporarily.
const (
	OptNoWizard    = "no-wizard"    // Disables the setup wizard on first startup
	OptDisableANSI = "disable-ansi" // Force-disables console text colour and formatting
	OptEnableANSI  = "enable-ansi"  // Force-enables console text colour and formatting
	OptPlugins     = "plugins"      // Path to look in for plugins
	OptData        = "data"         // Path to store and load server data
	OptVersion     = "version"      // Shows basic server version information and exits
	OptNoLogFile   = "no-log-file"  // Disables writing logs to server.log
)
