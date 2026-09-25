package defaults

import (
	"slices"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// HelpCommand is a port of pocketmine\command\defaults\HelpCommand.
type HelpCommand struct{ VanillaCommand }

func NewHelpCommand() *HelpCommand {
	return &HelpCommand{newVanillaCommand("help", lang.KnownTranslationFactory.PocketmineCommandHelpDescription(), lang.KnownTranslationFactory.CommandsHelpUsage(), []string{"?"}, permission.CommandHelp)}
}

// messageString is `$x instanceof Translatable ? $lang->translate($x) : $x`.
func messageString(language *lang.Language, message any) string {
	if t, ok := message.(*lang.Translatable); ok {
		return language.Translate(t)
	}
	if s, ok := message.(string); ok {
		return s
	}
	return ""
}

func (c *HelpCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	var commandName string
	pageNumber := 1
	if len(args) == 0 {
		commandName = ""
	} else if isNumeric(args[len(args)-1]) {
		pageNumber = phpInt(args[len(args)-1])
		args = args[:len(args)-1]
		if pageNumber <= 0 {
			pageNumber = 1
		}
		commandName = strings.Join(args, " ")
	} else {
		commandName = strings.Join(args, " ")
	}

	pageHeight := sender.GetScreenLineHeight()
	commandMap := srv(sender).GetSimpleCommandMap()

	if commandName == "" {
		byLabel := map[string]command.CommandLike{}
		for _, cmd := range commandMap.GetCommands() {
			if cmd.TestPermissionSilent(sender, nil) {
				byLabel[cmd.Label()] = cmd
			}
		}
		labels := make([]string, 0, len(byLabel))
		for label := range byLabel {
			labels = append(labels, label)
		}
		// ksort SORT_NATURAL | SORT_FLAG_CASE
		slices.SortFunc(labels, func(a, b string) int { return utils.NatCaseCompare(a, b) })
		var pages [][]string
		for i := 0; i < len(labels); i += pageHeight {
			pages = append(pages, labels[i:min(i+pageHeight, len(labels))])
		}
		pageNumber = min(len(pages), pageNumber)
		if pageNumber < 1 {
			pageNumber = 1
		}
		sender.SendMessage(lang.KnownTranslationFactory.CommandsHelpHeader(strconv.Itoa(pageNumber), strconv.Itoa(len(pages))))
		language := sender.GetLanguage()
		if pageNumber-1 < len(pages) {
			for _, label := range pages[pageNumber-1] {
				cmd := byLabel[label]
				sender.SendMessage(utils.DarkGreen + "/" + cmd.Label() + ": " + utils.Reset + messageString(language, cmd.Description()))
			}
		}
		return true, nil
	}

	if cmd := commandMap.GetCommand(strings.ToLower(commandName)); cmd != nil {
		if cmd.TestPermissionSilent(sender, nil) {
			language := sender.GetLanguage()
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandHelpSpecificCommandHeader(cmd.Label()).
				Format(utils.Yellow+"--------- "+utils.Reset, utils.Yellow+" ---------"))
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandHelpSpecificCommandDescription(utils.Reset + messageString(language, cmd.Description())).
				Prefix(utils.Gold))

			usageString := messageString(language, cmd.Usage())
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandHelpSpecificCommandUsage(utils.Reset + strings.Join(strings.Split(usageString, "\n"), "\n"+utils.Reset)).
				Prefix(utils.Gold))

			aliases := append([]string(nil), cmd.Aliases()...)
			slices.SortFunc(aliases, func(a, b string) int { return utils.NatCompare(a, b) })
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandHelpSpecificCommandAliases(utils.Reset + strings.Join(aliases, ", ")).
				Prefix(utils.Gold))
			return true, nil
		}
	}
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandNotFound(commandName, "/help").Prefix(utils.Red))
	return true, nil
}
