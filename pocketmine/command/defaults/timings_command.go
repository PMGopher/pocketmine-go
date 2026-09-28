package defaults

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/timings"
)

// TimingsCommand is a port of pocketmine\command\defaults\TimingsCommand.
type TimingsCommand struct{ VanillaCommand }

func NewTimingsCommand() *TimingsCommand {
	return &TimingsCommand{newVanillaCommand("timings", lang.KnownTranslationFactory.PocketmineCommandTimingsDescription(), lang.KnownTranslationFactory.PocketmineCommandTimingsUsage(), nil, permission.CommandTimings)}
}

func (c *TimingsCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) != 1 {
		return syntaxError()
	}

	mode := strings.ToLower(args[0])

	if mode == "on" {
		if timings.IsEnabled() {
			sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandTimingsAlreadyEnabled())
			return true, nil
		}
		timings.SetEnabled(true)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsEnable(), true)
		return true, nil
	} else if mode == "off" {
		timings.SetEnabled(false)
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsDisable(), true)
		return true, nil
	}

	if !timings.IsEnabled() {
		sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandTimingsTimingsDisabled())
		return true, nil
	}

	paste := mode == "paste"

	switch {
	case mode == "reset":
		timings.Reload()
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsReset(), true)
	case paste:
		lines := timings.RequestPrintTimings()
		command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsCollect(), true)
		c.uploadReport(lines, sender)
	case mode == "merged" || mode == "report":
		timingsFile, err := timings.CreateReportFile(filepath.Join(srv(sender).GetDataPath(), "timings"), "")
		if err != nil {
			srv(sender).GetLogger().Error("Failed to create timings report file")
		} else {
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsTimingsWrite(timingsFile), true)
		}
	default:
		return syntaxError()
	}
	return true, nil
}

// uploadReport is a port of TimingsCommand::uploadReport.
func (c *TimingsCommand) uploadReport(lines []string, sender command.Sender) {
	s := srv(sender)
	agent := s.GetName() + " " + s.GetPocketMineVersion()
	data := url.Values{
		"browser": {agent},
		"data":    {strings.Join(lines, "\n")},
		"private": {"true"},
	}

	host := s.GetConfigGroup().GetPropertyString(server.YmlTimingsHost, "timings.pmmp.io")

	s.GetAsyncPool().SubmitTask(scheduler.NewBulkCurlTask(
		[]scheduler.BulkCurlTaskOperation{{
			Page:           "https://" + host + "?upload=true",
			TimeoutSeconds: 10,
			ExtraHeaders: map[string]string{
				"User-Agent":   agent,
				"Content-Type": "application/x-www-form-urlencoded",
			},
			PostFields: data.Encode(),
		}},
		func(results []scheduler.BulkCurlTaskResult) {
			if p, ok := sender.(*player.Player); ok && !p.IsOnline() { // TODO replace with a more generic API method for checking availability of CommandSender
				return
			}
			result := results[0]
			if result.Err != nil {
				s.GetLogger().LogException(result.Err, "")
				return
			}
			var response map[string]any
			if json.Unmarshal([]byte(result.Result.Body), &response) == nil {
				if id, ok := timingsResponseID(response["id"]); ok {
					reportURL := "https://" + host + "/?id=" + id
					if token, ok := response["access_token"].(string); ok {
						reportURL += "&access_token=" + token
					} else {
						s.GetLogger().Warning("Your chosen timings host does not support private reports. Anyone will be able to see your report if they guess the ID.")
					}
					command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsTimingsRead(reportURL), true)
					return
				}
			}
			s.GetLogger().Debug(fmt.Sprintf("Invalid response from timings server (%d): %s", result.Result.Code, result.Result.Body))
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.PocketmineCommandTimingsPasteError(), true)
		},
	))
}

// timingsResponseID is `is_int($response["id"]) || is_string($response["id"])`, as a string.
func timingsResponseID(v any) (string, bool) {
	switch id := v.(type) {
	case string:
		return id, true
	case float64:
		if id == float64(int64(id)) {
			return fmt.Sprintf("%d", int64(id)), true
		}
	}
	return "", false
}
