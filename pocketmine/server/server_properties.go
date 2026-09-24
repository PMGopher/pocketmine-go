// Package server is a port of pocketmine\Server and the classes that live next to it
// (ServerProperties, ServerConfigGroup). In PHP they sit in the root pocketmine namespace; here they
// get their own package because the root package (VersionInfo, CoreConstants) is imported by
// low-level packages such as entity and world, which Server itself depends on.
package server

// Keys of server.properties, a port of pocketmine\ServerProperties.
const (
	PropertyAutoSave                      = "auto-save"
	PropertyDefaultWorldGenerator         = "level-type"
	PropertyDefaultWorldGeneratorSettings = "generator-settings"
	PropertyDefaultWorldName              = "level-name"
	PropertyDefaultWorldSeed              = "level-seed"
	PropertyDifficulty                    = "difficulty"
	PropertyEnableIPv6                    = "enable-ipv6"
	PropertyEnableQuery                   = "enable-query"
	PropertyForceGameMode                 = "force-gamemode"
	PropertyGameMode                      = "gamemode"
	PropertyHardcore                      = "hardcore"
	PropertyLanguage                      = "language"
	PropertyMaxPlayers                    = "max-players"
	PropertyMotd                          = "motd"
	PropertyPvp                           = "pvp"
	PropertyServerIPv4                    = "server-ip"
	PropertyServerIPv6                    = "server-ipv6"
	PropertyServerPortIPv4                = "server-port"
	PropertyServerPortIPv6                = "server-portv6"
	PropertyViewDistance                  = "view-distance"
	PropertyWhitelist                     = "white-list"
	PropertyXboxAuth                      = "xbox-auth"
)
