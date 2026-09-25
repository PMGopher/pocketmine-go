package world

import (
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/light"
	"pocketmine-go/pocketmine/world/utils"
)

// LightPopulationResult is what LightPopulationTask computes: per sub-chunk Y, the block and sky
// light arrays, and the height map.
type LightPopulationResult struct {
	BlockLight map[int]*format.LightArray
	SkyLight   map[int]*format.LightArray
	HeightMap  [256]int
}

// computeChunkLight is LightPopulationTask::onRun: the chunk (a copy) is lit on its own, at 0,0
// in a SimpleChunkManager.
func computeChunkLight(chunk *format.Chunk, registry *blockStateRegistry) LightPopulationResult {
	manager := NewSimpleChunkManager(YMin, YMax, registry)
	manager.SetChunk(0, 0, chunk)

	blockUpdate := light.NewBlockLightUpdate(utils.NewSubChunkExplorer(manager), registry.lightFilters, registry.lightEmitters)
	blockUpdate.RecalculateChunk(0, 0)
	blockUpdate.Execute()
	skyUpdate := light.NewSkyLightUpdate(utils.NewSubChunkExplorer(manager), registry.lightFilters, registry.directSkyLightBlockers)
	skyUpdate.RecalculateChunk(0, 0)
	skyUpdate.Execute()

	chunk.SetLightPopulated(true, true)

	result := LightPopulationResult{BlockLight: map[int]*format.LightArray{}, SkyLight: map[int]*format.LightArray{}, HeightMap: chunk.GetHeightMapArray()}
	for y, subChunk := range chunk.GetSubChunks() {
		result.SkyLight[y] = subChunk.GetBlockSkyLightArray()
		result.BlockLight[y] = subChunk.GetBlockLightArray()
	}
	return result
}

// LightPopulationTask is a port of pocketmine\world\light\LightPopulationTask. It lives in the
// world package because it needs SimpleChunkManager.
type LightPopulationTask struct {
	scheduler.AsyncTaskBase
	chunk        *format.Chunk
	registry     *blockStateRegistry
	onCompletion func(LightPopulationResult)
	result       LightPopulationResult
}

// NewLightPopulationTask is a port of LightPopulationTask::__construct: chunk must be a copy.
func NewLightPopulationTask(chunk *format.Chunk, registry *blockStateRegistry, onCompletion func(LightPopulationResult)) *LightPopulationTask {
	return &LightPopulationTask{chunk: chunk, registry: registry, onCompletion: onCompletion}
}

func (t *LightPopulationTask) OnRun() { t.result = computeChunkLight(t.chunk, t.registry) }

func (t *LightPopulationTask) OnCompletion() { t.onCompletion(t.result) }
