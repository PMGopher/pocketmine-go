package defaults

import (
	"fmt"
	"runtime"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// GarbageCollectorCommand is a port of pocketmine\command\defaults\GarbageCollectorCommand. Go's
// garbage collector doesn't report collected cycles, so that line shows 0.
type GarbageCollectorCommand struct{ VanillaCommand }

func NewGarbageCollectorCommand() *GarbageCollectorCommand {
	return &GarbageCollectorCommand{newVanillaCommand("gc", lang.KnownTranslationFactory.PocketmineCommandGcDescription(), nil, nil, permission.CommandGC)}
}

func (c *GarbageCollectorCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	chunksCollected, entitiesCollected := 0, 0

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for _, w := range srv(sender).GetWorldManager().GetWorlds() {
		chunks, entities := len(w.GetLoadedChunks()), len(w.GetEntities())
		w.DoChunkGarbageCollection()
		w.UnloadChunks(true)
		chunksCollected += chunks - len(w.GetLoadedChunks())
		entitiesCollected += entities - len(w.GetEntities())
		// World::clearCache(true): this port's worlds have no block cache.
	}

	srv(sender).GetMemoryManager().TriggerGarbageCollector()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGcHeader().Format(utils.Green+"---- "+utils.Reset, utils.Green+" ----"+utils.Reset))
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGcChunks(utils.Red + numberFormat(chunksCollected)).Prefix(utils.Gold))
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGcEntities(utils.Red + numberFormat(entitiesCollected)).Prefix(utils.Gold))
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGcCycles(utils.Red + "0").Prefix(utils.Gold))
	freed := (float64(before.HeapAlloc) - float64(after.HeapAlloc)) / 1024 / 1024
	sender.SendMessage(lang.KnownTranslationFactory.PocketmineCommandGcMemoryFreed(utils.Red + fmt.Sprintf("%.2f", round2(freed))).Prefix(utils.Gold))
	return true, nil
}
