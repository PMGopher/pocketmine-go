package permission

// RootConsole, RootOperator and RootUser are a port of DefaultPermissions::ROOT_*.
const (
	RootConsole  = GroupConsole
	RootOperator = GroupOperator
	RootUser     = GroupUser
)

// RegisterPermission is a port of DefaultPermissions::registerPermission().
func RegisterPermission(candidate *Permission, grantedBy []*Permission, deniedBy []*Permission) *Permission {
	for _, perm := range grantedBy {
		perm.AddChild(candidate.Name(), true)
	}
	for _, perm := range deniedBy {
		perm.AddChild(candidate.Name(), false)
	}
	GetManager().AddPermission(candidate)
	return GetManager().GetPermission(candidate.Name())
}

// registerNoArgsDesc is a port of DefaultPermissions::registerNoArgsDesc().
//
// PHP looks up a translated description string (a Translatable) keyed by
// "pocketmine.permission.<name>" via KnownTranslationParameterInfo — the lang package isn't
// ported yet, so this uses the permission name itself as a placeholder description. Swap this
// for a real *lang.Translatable lookup once that package exists; every call site here is
// unaffected either way since Permission.Description() is typed `any`.
func registerNoArgsDesc(permission string, grantedBy []*Permission) *Permission {
	return RegisterPermission(NewPermission(permission, permission, nil), grantedBy, nil)
}

// RegisterCorePermissions is a port of DefaultPermissions::registerCorePermissions().
func RegisterCorePermissions() {
	consoleRoot := registerNoArgsDesc(RootConsole, nil)
	operatorRoot := registerNoArgsDesc(RootOperator, []*Permission{consoleRoot})
	everyoneRoot := registerNoArgsDesc(RootUser, []*Permission{operatorRoot})

	registerNoArgsDesc(CommandDumpMemory, []*Permission{consoleRoot})

	forOperator := []string{
		BroadcastAdmin, CommandBanIP, CommandBanList, CommandBanPlayer, CommandClearOther,
		CommandDefaultGamemode, CommandDifficulty, CommandEffectOther, CommandEffectSelf,
		CommandEnchantOther, CommandEnchantSelf, CommandGamemodeOther, CommandGamemodeSelf,
		CommandGC, CommandGiveOther, CommandGiveSelf, CommandKick, CommandKillOther, CommandList,
		CommandOpGive, CommandOpTake, CommandParticle, CommandPlugins, CommandSaveDisable,
		CommandSaveEnable, CommandSavePerform, CommandSay, CommandSeed, CommandSetWorldSpawn,
		CommandSpawnpointOther, CommandSpawnpointSelf, CommandStatus, CommandStop,
		CommandTeleportOther, CommandTeleportSelf, CommandTimeAdd, CommandTimeQuery,
		CommandTimeSet, CommandTimeStart, CommandTimeStop, CommandTimings, CommandTitleOther,
		CommandTitleSelf, CommandTransferServer, CommandUnbanIP, CommandUnbanPlayer,
		CommandWhitelistAdd, CommandWhitelistDisable, CommandWhitelistEnable, CommandWhitelistList,
		CommandWhitelistReload, CommandWhitelistRemove, CommandXPOther, CommandXPSelf,
	}
	for _, name := range forOperator {
		registerNoArgsDesc(name, []*Permission{operatorRoot})
	}

	forEveryone := []string{
		CommandKillSelf, CommandMe, CommandHelp, BroadcastUser, CommandClearSelf, CommandTell, CommandVersion,
	}
	for _, name := range forEveryone {
		registerNoArgsDesc(name, []*Permission{everyoneRoot})
	}
}
