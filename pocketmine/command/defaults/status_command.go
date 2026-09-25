package defaults

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// StatusCommand is a port of pocketmine\command\defaults\StatusCommand. "Threads" counts
// goroutines.
type StatusCommand struct{ VanillaCommand }

func NewStatusCommand() *StatusCommand {
	return &StatusCommand{newVanillaCommand("status", lang.KnownTranslationFactory.PocketmineCommandStatusDescription(), nil, nil, permission.CommandStatus)}
}

func statusSend(sender command.Sender, message *lang.Translatable) {
	sender.SendMessage(message.Prefix(utils.Gold))
}

func strval(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }

func formatTPS(tps, usage float64, tpsColor string) *lang.Translatable {
	return lang.KnownTranslationFactory.PocketmineCommandStatusTpsStat(strval(tps), strval(usage)).Prefix(tpsColor)
}

func formatBandwidth(bytes float64) *lang.Translatable {
	//TODO: this should probably be number formatted?
	return lang.KnownTranslationFactory.PocketmineCommandStatusNetworkStat(strval(round2(bytes / 1024))).Prefix(utils.Red)
}

func formatMemory(bytes uint64) *lang.Translatable {
	return lang.KnownTranslationFactory.PocketmineCommandStatusMemoryStat(fmt.Sprintf("%.2f", round2(float64(bytes)/1024/1024))).Prefix(utils.Red)
}

func (c *StatusCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	reserved, vmRSS, vmSize := utils.AdvancedMemoryUsage()

	server := srv(sender)
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandStatusHeader().Format(
		utils.Green+"---- "+utils.Reset,
		utils.Green+" ----"+utils.Reset,
	))

	t := int(time.Since(server.GetStartTime()).Seconds())
	seconds := strconv.Itoa(t % 60)
	var message *lang.Translatable
	if t >= 60 {
		minutes := strconv.Itoa((t % 3600) / 60)
		if t >= 3600 {
			hours := strconv.Itoa((t % (3600 * 24)) / 3600)
			if t >= 3600*24 {
				message = lang.KnownTranslationFactory.PocketmineCommandStatusUptimeDays(strconv.Itoa(t/(3600*24)), hours, minutes, seconds)
			} else {
				message = lang.KnownTranslationFactory.PocketmineCommandStatusUptimeHours(hours, minutes, seconds)
			}
		} else {
			message = lang.KnownTranslationFactory.PocketmineCommandStatusUptimeMinutes(minutes, seconds)
		}
	} else {
		message = lang.KnownTranslationFactory.PocketmineCommandStatusUptimeSeconds(seconds)
	}
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusUptime(message.Prefix(utils.Red)))

	tpsColor := utils.Green
	tps := server.GetTicksPerSecond()
	if tps < 12 {
		tpsColor = utils.Red
	} else if tps < 17 {
		tpsColor = utils.Gold
	}

	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusTpsCurrent(formatTPS(tps, server.GetTickUsage(), tpsColor)))
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusTpsAverage(formatTPS(server.GetTicksPerSecondAverage(), server.GetTickUsageAverage(), tpsColor)))

	bandwidth := server.GetNetwork().GetBandwidthTracker()
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusNetworkUpload(formatBandwidth(bandwidth.GetSend().GetAverageBytes())))
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusNetworkDownload(formatBandwidth(bandwidth.GetReceive().GetAverageBytes())))

	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusThreads(utils.Red+strconv.Itoa(runtime.NumGoroutine())))

	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusMemoryMainThread(formatMemory(reserved)))
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusMemoryTotal(formatMemory(vmRSS)))
	statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusMemoryVirtual(formatMemory(vmSize)))

	if globalLimit := server.GetMemoryManager().GetGlobalMemoryLimit(); globalLimit > 0 {
		statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusMemoryManager(formatMemory(globalLimit)))
	}

	for _, w := range server.GetWorldManager().GetWorlds() {
		worldName := ""
		if w.GetFolderName() != w.GetDisplayName() {
			worldName = " (" + w.GetDisplayName() + ")"
		}
		timeColor := utils.Yellow
		if w.GetTickRateTime() > 40 {
			timeColor = utils.Red
		}
		statusSend(sender, lang.KnownTranslationFactory.PocketmineCommandStatusWorld(
			"\""+w.GetFolderName()+"\""+worldName,
			utils.Red+numberFormat(len(w.GetLoadedChunks()))+utils.Green,
			utils.Red+numberFormat(len(w.GetTickingChunks()))+utils.Green,
			utils.Red+numberFormat(len(w.GetEntities()))+utils.Green,
			lang.KnownTranslationFactory.PocketmineCommandStatusWorldTimeStat(strval(round2(w.GetTickRateTime()))).Prefix(timeColor),
		))
	}
	return true, nil
}

// numberFormat is PHP's number_format($n): thousands separated by commas.
func numberFormat(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + numberFormat(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
