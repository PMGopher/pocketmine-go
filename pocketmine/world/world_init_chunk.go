package world

import (
	"fmt"
	"sort"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	worldformatio "pocketmine-go/pocketmine/world/format/io"
)

// initChunk is a port of World::initChunk: recreates the entities (EntityFactory, through
// LoadEntityFunc) and tiles (TileFactory) saved with a freshly loaded chunk. Unknown or broken
// entities and tiles are logged and dropped, like PHP.
func (w *World) initChunk(chunkX, chunkZ int, chunkData *worldformatio.ChunkData, chunk *format.Chunk) {
	var logger log.Logger = log.Global()
	if w.logger != nil {
		logger = w.logger
	}
	logger = log.NewPrefixedLogger(logger, fmt.Sprintf("Loading chunk %d %d", chunkX, chunkZ))

	if entityNBT := chunkData.GetEntityNBT(); len(entityNBT) != 0 && LoadEntityFunc != nil {
		deletedEntities := map[string]int{}
		for k, tag := range entityNBT {
			entity, err := LoadEntityFunc(w, tag)
			if err != nil {
				logger.Error(fmt.Sprintf("Bad entity data at list position %d: %v", k, err))
				continue
			}
			if entity == nil {
				saveID := "<unknown>"
				idTag, ok := tag.GetTag("identifier")
				if !ok {
					idTag, _ = tag.GetTag("id")
				}
				switch v := idTag.(type) {
				case nbt.StringTag:
					saveID = string(v)
				case nbt.IntTag: // legacy MCPE format
					saveID = fmt.Sprintf("legacy(%d)", v)
				}
				deletedEntities[saveID]++
			}
			//TODO: we can't prevent entities getting added to unloaded chunks if they were saved in the wrong place
			//here, because entities currently add themselves to the world
		}
		for _, saveID := range sortedKeys(deletedEntities) {
			logger.Warning(fmt.Sprintf("Deleted unknown entity type %s x%d", saveID, deletedEntities[saveID]))
		}
	}

	if tileNBT := chunkData.GetTileNBT(); len(tileNBT) != 0 {
		tileFactory := tile.GetTileFactory()
		deletedTiles := map[string]int{}
		for k, tag := range tileNBT {
			t, err := tileFactory.CreateFromData(w, tag)
			if err != nil {
				logger.Error(fmt.Sprintf("Bad tile entity data at list position %d: %v", k, err))
				continue
			}
			if t == nil {
				deletedTiles[string(tag.GetStringOr(tile.TagID, "<unknown>"))]++
				continue
			}

			pos := t.GetPosition()
			x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()
			if _, loaded := w.chunks[chunkKey(x>>4, z>>4)]; !loaded {
				logger.Error("Found tile saved on wrong chunk - unable to fix due to correct chunk not loaded")
			} else if !w.IsInWorld(x, y, z) {
				logger.Error(fmt.Sprintf("Cannot add tile with position outside the world bounds: x=%d,y=%d,z=%d", x, y, z))
			} else if _, exists := w.GetTileAt(x, y, z); exists {
				logger.Error(fmt.Sprintf("Cannot add tile at x=%d,y=%d,z=%d: Another tile is already at that position", x, y, z))
			} else {
				w.AddTile(t)
			}
			if (x>>4) != chunkX || (z>>4) != chunkZ {
				continue
			}
			expectedStateID := chunk.GetBlockStateID(x&0xf, y, z&0xf)
			actualStateID := int32(w.GetBlockAt(x, y, z).GetStateId())
			if expectedStateID != actualStateID {
				// state ID was updated by readStateFromWorld - typically because the block pulled some data from the tile
				// make sure this is synced to the chunk
				//TODO: in the future we should pull tile reading logic out of readStateFromWorld() and do it only
				//when the tile is loaded - this would be cleaner and faster
				chunk.SetBlockStateID(x&0xf, y, z&0xf, actualStateID)
				logger.Debug(fmt.Sprintf("Tile %T at x=%d,y=%d,z=%d updated block state ID from %d to %d", t, x, y, z, expectedStateID, actualStateID))
			}
		}
		for _, saveID := range sortedKeys(deletedTiles) {
			logger.Warning(fmt.Sprintf("Deleted unknown tile entity type %s x%d", saveID, deletedTiles[saveID]))
		}
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
