package defaults

import (
	"slices"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/server"
)

// WhitelistCommand is a port of pocketmine\command\defaults\WhitelistCommand.
type WhitelistCommand struct{ VanillaCommand }

func NewWhitelistCommand() *WhitelistCommand {
	c := &WhitelistCommand{VanillaCommand{Command: command.InitCommand("whitelist", lang.KnownTranslationFactory.PocketmineCommandWhitelistDescription(), lang.KnownTranslationFactory.CommandsWhitelistUsage(), nil)}}
	_ = c.SetPermissions([]string{
		permission.CommandWhitelistReload,
		permission.CommandWhitelistEnable,
		permission.CommandWhitelistDisable,
		permission.CommandWhitelistList,
		permission.CommandWhitelistAdd,
		permission.CommandWhitelistRemove,
	})
	return c
}

func (c *WhitelistCommand) test(sender command.Sender, perm string) bool {
	return c.TestPermission(sender, &perm)
}

func (c *WhitelistCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	s := srv(sender)
	if len(args) == 1 {
		switch strings.ToLower(args[0]) {
		case "reload":
			if c.test(sender, permission.CommandWhitelistReload) {
				if err := s.GetWhitelisted().Reload(); err != nil {
					s.GetLogger().Error(err.Error())
				}
				if s.HasWhitelist() {
					kickNonWhitelistedPlayers(s)
				}
				command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsWhitelistReloaded(), true)
			}
			return true, nil
		case "on":
			if c.test(sender, permission.CommandWhitelistEnable) {
				s.GetConfigGroup().SetConfigBool(server.PropertyWhitelist, true)
				kickNonWhitelistedPlayers(s)
				command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsWhitelistEnabled(), true)
			}
			return true, nil
		case "off":
			if c.test(sender, permission.CommandWhitelistDisable) {
				s.GetConfigGroup().SetConfigBool(server.PropertyWhitelist, false)
				command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsWhitelistDisabled(), true)
			}
			return true, nil
		case "list":
			if c.test(sender, permission.CommandWhitelistList) {
				entries := s.GetWhitelisted().GetKeys()
				slices.Sort(entries)
				count := strconv.Itoa(len(entries))
				sender.SendMessage(lang.KnownTranslationFactory.CommandsWhitelistList(count, count))
				sender.SendMessage(strings.Join(entries, ", "))
			}
			return true, nil
		case "add":
			sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericUsage(lang.KnownTranslationFactory.CommandsWhitelistAddUsage()))
			return true, nil
		case "remove":
			sender.SendMessage(lang.KnownTranslationFactory.CommandsGenericUsage(lang.KnownTranslationFactory.CommandsWhitelistRemoveUsage()))
			return true, nil
		}
	} else if len(args) == 2 {
		if !player.IsValidUserName(args[1]) {
			return syntaxError()
		}
		switch strings.ToLower(args[0]) {
		case "add":
			if c.test(sender, permission.CommandWhitelistAdd) {
				s.AddWhitelist(args[1])
				command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsWhitelistAddSuccess(args[1]), true)
			}
			return true, nil
		case "remove":
			if c.test(sender, permission.CommandWhitelistRemove) {
				s.RemoveWhitelist(args[1])
				if !s.IsWhitelisted(args[1]) {
					if p := s.GetPlayerExact(args[1]); p != nil {
						p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectKick(lang.KnownTranslationFactory.PocketmineDisconnectWhitelisted()), nil, nil)
					}
				}
				command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsWhitelistRemoveSuccess(args[1]), true)
			}
			return true, nil
		}
	}
	return syntaxError()
}

func kickNonWhitelistedPlayers(s *server.Server) {
	message := lang.KnownTranslationFactory.PocketmineDisconnectKick(lang.KnownTranslationFactory.PocketmineDisconnectWhitelisted())
	for _, p := range s.GetOnlinePlayers() {
		if !s.IsWhitelisted(p.GetName()) {
			p.Kick(message, nil, nil)
		}
	}
}
