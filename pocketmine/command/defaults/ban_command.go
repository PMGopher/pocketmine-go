package defaults

import (
	"net"
	"slices"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

// BanCommand is a port of pocketmine\command\defaults\BanCommand.
type BanCommand struct{ VanillaCommand }

func NewBanCommand() *BanCommand {
	return &BanCommand{newVanillaCommand("ban", lang.KnownTranslationFactory.PocketmineCommandBanPlayerDescription(), lang.KnownTranslationFactory.CommandsBanUsage(), nil, permission.CommandBanPlayer)}
}

func (c *BanCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	name := args[0]
	reason := strings.Join(args[1:], " ")

	srv(sender).GetNameBans().AddBan(name, reason, nil, sender.GetName())

	if p := srv(sender).GetPlayerExact(name); p != nil {
		if reason != "" {
			p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectBan(reason), nil, nil)
		} else {
			p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectBanNoReason(), nil, nil)
		}
		name = p.GetName()
	}
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsBanSuccess(name), true)
	return true, nil
}

// BanIpCommand is a port of pocketmine\command\defaults\BanIpCommand.
type BanIpCommand struct{ VanillaCommand }

func NewBanIpCommand() *BanIpCommand {
	return &BanIpCommand{newVanillaCommand("ban-ip", lang.KnownTranslationFactory.PocketmineCommandBanIpDescription(), lang.KnownTranslationFactory.CommandsBanipUsage(), nil, permission.CommandBanIP)}
}

func (c *BanIpCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) == 0 {
		return syntaxError()
	}
	value := args[0]
	reason := strings.Join(args[1:], " ")

	if net.ParseIP(value) != nil {
		c.processIPBan(value, sender, reason)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsBanipSuccess(value), true)
	} else if p := srv(sender).GetPlayerByPrefix(value); p != nil {
		ip := p.GetNetworkSession().GetIp()
		c.processIPBan(ip, sender, reason)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsBanipSuccessPlayers(ip, p.GetName()), true)
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsBanipInvalid())
		return false, nil
	}
	return true, nil
}

func (c *BanIpCommand) processIPBan(ip string, sender command.Sender, reason string) {
	srv(sender).GetIPBans().AddBan(ip, reason, nil, sender.GetName())

	for _, p := range srv(sender).GetOnlinePlayers() {
		if p.GetNetworkSession().GetIp() == ip {
			var banReason any = reason
			if reason == "" {
				banReason = lang.KnownTranslationFactory.PocketmineDisconnectBanIp()
			}
			p.Kick(lang.KnownTranslationFactory.PocketmineDisconnectBan(banReason), nil, nil)
		}
	}
	srv(sender).GetNetwork().BlockAddress(ip, -1)
}

// BanListCommand is a port of pocketmine\command\defaults\BanListCommand.
type BanListCommand struct{ VanillaCommand }

func NewBanListCommand() *BanListCommand {
	return &BanListCommand{newVanillaCommand("banlist", lang.KnownTranslationFactory.PocketmineCommandBanlistDescription(), lang.KnownTranslationFactory.CommandsBanlistUsage(), nil, permission.CommandBanList)}
}

func (c *BanListCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	which := "players"
	list := srv(sender).GetNameBans()
	if len(args) > 0 {
		which = strings.ToLower(args[0])
		switch which {
		case "ips":
			list = srv(sender).GetIPBans()
		case "players":
		default:
			return syntaxError()
		}
	}

	var names []string
	for _, entry := range list.GetEntries() {
		names = append(names, entry.Name())
	}
	slices.Sort(names)

	if which == "ips" {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsBanlistIps(strconv.Itoa(len(names))))
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsBanlistPlayers(strconv.Itoa(len(names))))
	}
	sender.SendMessage(strings.Join(names, ", "))
	return true, nil
}

// PardonCommand is a port of pocketmine\command\defaults\PardonCommand.
type PardonCommand struct{ VanillaCommand }

func NewPardonCommand() *PardonCommand {
	return &PardonCommand{newVanillaCommand("pardon", lang.KnownTranslationFactory.PocketmineCommandUnbanPlayerDescription(), lang.KnownTranslationFactory.CommandsUnbanUsage(), []string{"unban"}, permission.CommandUnbanPlayer)}
}

func (c *PardonCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) != 1 {
		return syntaxError()
	}
	srv(sender).GetNameBans().Remove(args[0])
	command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsUnbanSuccess(args[0]), true)
	return true, nil
}

// PardonIpCommand is a port of pocketmine\command\defaults\PardonIpCommand.
type PardonIpCommand struct{ VanillaCommand }

func NewPardonIpCommand() *PardonIpCommand {
	return &PardonIpCommand{newVanillaCommand("pardon-ip", lang.KnownTranslationFactory.PocketmineCommandUnbanIpDescription(), lang.KnownTranslationFactory.CommandsUnbanipUsage(), []string{"unban-ip"}, permission.CommandUnbanIP)}
}

func (c *PardonIpCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) != 1 {
		return syntaxError()
	}
	if net.ParseIP(args[0]) != nil {
		srv(sender).GetIPBans().Remove(args[0])
		srv(sender).GetNetwork().UnblockAddress(args[0])
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsUnbanipSuccess(args[0]), true)
	} else {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsUnbanipInvalid())
	}
	return true, nil
}
