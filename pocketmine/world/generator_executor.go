package world

import (
	"fmt"
	"strconv"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator"
)

// This file ports pocketmine\world\generator\executor, PopulationTask and PopulationUtils. They
// live in the world package (not world/generator/executor) because they need SimpleChunkManager
// and the World's block-state registry snapshot, and world already imports the generator package.

// PopulationResult is what a population run hands back to the main thread: the populated centre
// chunk, the adjacent chunks it generated or modified (relative [dx, dz] => chunk; unchanged ones
// are left out), and block states written that the registry snapshot didn't know yet.
type PopulationResult struct {
	Center       *format.Chunk
	Adjacent     map[[2]int]*format.Chunk
	NewTemplates map[int32]block.Behavior
}

// GeneratorExecutor is a port of pocketmine\world\generator\executor\GeneratorExecutor.
type GeneratorExecutor interface {
	// Populate generates (if nil) and populates the chunk, generating any of its missing (nil)
	// adjacent chunks, then calls onCompletion on the main thread.
	Populate(chunkX, chunkZ int, centerChunk *format.Chunk, adjacentChunks map[[2]int]*format.Chunk, registry *blockStateRegistry, onCompletion func(PopulationResult))
	Shutdown()
}

// GeneratorExecutorSetupParameters is a port of
// pocketmine\world\generator\executor\GeneratorExecutorSetupParameters: what a worker needs to
// build its own generator instance (generators aren't safe to share between goroutines).
type GeneratorExecutorSetupParameters struct {
	WorldMinY int
	WorldMaxY int
	// NewGenerator creates a generator for the world (PHP: generator class, seed and options).
	NewGenerator func() (generator.Generator, error)
}

// setOrGenerateChunk is a port of PopulationUtils::setOrGenerateChunk.
func setOrGenerateChunk(manager *SimpleChunkManager, gen generator.Generator, chunkX, chunkZ int, chunk *format.Chunk) *format.Chunk {
	if chunk == nil {
		chunk = gen.GenerateChunk(chunkX, chunkZ)
	}
	manager.SetChunk(chunkX, chunkZ, chunk)
	return chunk
}

// PopulateChunkWithAdjacents is a port of PopulationUtils::populateChunkWithAdjacents.
func PopulateChunkWithAdjacents(minY, maxY int, gen generator.Generator, registry *blockStateRegistry, chunkX, chunkZ int, centerChunk *format.Chunk, adjacentChunks map[[2]int]*format.Chunk) (*format.Chunk, map[[2]int]*format.Chunk, map[int32]block.Behavior) {
	manager := NewSimpleChunkManager(minY, maxY, registry)
	setOrGenerateChunk(manager, gen, chunkX, chunkZ, centerChunk)

	resultChunks := make(map[[2]int]*format.Chunk, len(adjacentChunks))
	for relative, c := range adjacentChunks {
		resultChunks[relative] = setOrGenerateChunk(manager, gen, chunkX+relative[0], chunkZ+relative[1], c)
	}

	gen.PopulateChunk(manager, chunkX, chunkZ)
	center, ok := manager.GetChunk(chunkX, chunkZ)
	if !ok {
		panic("We just generated this chunk, so it must exist")
	}
	center.SetPopulated(true)
	return center, resultChunks, manager.newTemplates
}

// runPopulation is PopulationTask::onRun: adjacent chunks that existed are marked clean first, so
// only the ones population changed (or that were generated) go back to the main thread.
func runPopulation(params GeneratorExecutorSetupParameters, gen generator.Generator, registry *blockStateRegistry, chunkX, chunkZ int, center *format.Chunk, adjacent map[[2]int]*format.Chunk) PopulationResult {
	for _, c := range adjacent {
		if c != nil {
			//this allows us to avoid sending existing chunks back to the main thread if they haven't changed during generation
			c.ClearTerrainDirtyFlags()
		}
	}
	generated := map[[2]int]bool{}
	for relative, c := range adjacent {
		if c == nil {
			generated[relative] = true
		}
	}
	populated, chunks, newTemplates := PopulateChunkWithAdjacents(params.WorldMinY, params.WorldMaxY, gen, registry, chunkX, chunkZ, center, adjacent)
	result := PopulationResult{Center: populated, Adjacent: map[[2]int]*format.Chunk{}, NewTemplates: newTemplates}
	for relative, c := range chunks {
		if generated[relative] || c.IsTerrainDirty() {
			result.Adjacent[relative] = c
		}
	}
	return result
}

// SyncGeneratorExecutor is a port of pocketmine\world\generator\executor\SyncGeneratorExecutor:
// population runs immediately on the calling goroutine.
type SyncGeneratorExecutor struct {
	params    GeneratorExecutorSetupParameters
	generator generator.Generator
}

// NewSyncGeneratorExecutor is a port of SyncGeneratorExecutor::__construct. gen is the generator
// to use (the World's own; it's only ever used from the main thread).
func NewSyncGeneratorExecutor(params GeneratorExecutorSetupParameters, gen generator.Generator) *SyncGeneratorExecutor {
	return &SyncGeneratorExecutor{params: params, generator: gen}
}

func (e *SyncGeneratorExecutor) Populate(chunkX, chunkZ int, centerChunk *format.Chunk, adjacentChunks map[[2]int]*format.Chunk, registry *blockStateRegistry, onCompletion func(PopulationResult)) {
	onCompletion(runPopulation(e.params, e.generator, registry, chunkX, chunkZ, centerChunk, adjacentChunks))
}

func (e *SyncGeneratorExecutor) Shutdown() {}

// ThreadLocalGeneratorContext is a port of
// pocketmine\world\generator\executor\ThreadLocalGeneratorContext: a worker's own generator for
// one world, kept in the worker's thread store.
type ThreadLocalGeneratorContext struct {
	generator generator.Generator
	worldMinY int
	worldMaxY int
}

func threadLocalGeneratorContextKey(contextID int) string {
	return "pocketmine.world.generatorContext." + strconv.Itoa(contextID)
}

// fetchThreadLocalGeneratorContext is ThreadLocalGeneratorContext::fetch.
func fetchThreadLocalGeneratorContext(worker *scheduler.AsyncWorker, contextID int) *ThreadLocalGeneratorContext {
	ctx, _ := worker.GetFromThreadStore(threadLocalGeneratorContextKey(contextID)).(*ThreadLocalGeneratorContext)
	return ctx
}

// AsyncGeneratorRegisterTask is a port of
// pocketmine\world\generator\executor\AsyncGeneratorRegisterTask: creates the worker's generator.
type AsyncGeneratorRegisterTask struct {
	scheduler.AsyncTaskBase
	params    GeneratorExecutorSetupParameters
	contextID int
}

func (t *AsyncGeneratorRegisterTask) OnRun() {
	gen, err := t.params.NewGenerator()
	if err != nil {
		panic(fmt.Sprintf("failed to create generator: %v", err))
	}
	t.GetWorker().SaveToThreadStore(threadLocalGeneratorContextKey(t.contextID), &ThreadLocalGeneratorContext{generator: gen, worldMinY: t.params.WorldMinY, worldMaxY: t.params.WorldMaxY})
}

// AsyncGeneratorUnregisterTask is a port of
// pocketmine\world\generator\executor\AsyncGeneratorUnregisterTask.
type AsyncGeneratorUnregisterTask struct {
	scheduler.AsyncTaskBase
	contextID int
}

func (t *AsyncGeneratorUnregisterTask) OnRun() {
	t.GetWorker().RemoveFromThreadStore(threadLocalGeneratorContextKey(t.contextID))
}

// PopulationTask is a port of pocketmine\world\generator\PopulationTask. The chunks it's given are
// copies (PHP serializes them); the results are handed back to the main thread in OnCompletion.
type PopulationTask struct {
	scheduler.AsyncTaskBase
	contextID    int
	chunkX       int
	chunkZ       int
	center       *format.Chunk
	adjacent     map[[2]int]*format.Chunk
	registry     *blockStateRegistry
	onCompletion func(PopulationResult)
	result       PopulationResult
}

func (t *PopulationTask) OnRun() {
	ctx := fetchThreadLocalGeneratorContext(t.GetWorker(), t.contextID)
	if ctx == nil {
		panic("Generator context should have been initialized before any PopulationTask execution")
	}
	t.result = runPopulation(GeneratorExecutorSetupParameters{WorldMinY: ctx.worldMinY, WorldMaxY: ctx.worldMaxY}, ctx.generator, t.registry, t.chunkX, t.chunkZ, t.center, t.adjacent)
}

func (t *PopulationTask) OnCompletion() {
	t.onCompletion(t.result)
}

var nextAsyncContextID = 1

// AsyncGeneratorExecutor is a port of pocketmine\world\generator\executor\AsyncGeneratorExecutor:
// population runs on the server's AsyncPool workers, each with its own generator.
type AsyncGeneratorExecutor struct {
	logger                    log.Logger
	workerPool                *scheduler.AsyncPool
	setupParameters           GeneratorExecutorSetupParameters
	asyncContextID            int
	generatorRegisteredWorker map[int]bool
	workerStartHook           *func(worker int)
}

// NewAsyncGeneratorExecutor is a port of AsyncGeneratorExecutor::__construct.
func NewAsyncGeneratorExecutor(logger log.Logger, workerPool *scheduler.AsyncPool, setupParameters GeneratorExecutorSetupParameters) *AsyncGeneratorExecutor {
	e := &AsyncGeneratorExecutor{
		logger:                    log.NewPrefixedLogger(logger, "AsyncGeneratorExecutor"),
		workerPool:                workerPool,
		setupParameters:           setupParameters,
		asyncContextID:            nextAsyncContextID,
		generatorRegisteredWorker: map[int]bool{},
	}
	nextAsyncContextID++
	hook := func(workerID int) {
		if e.generatorRegisteredWorker[workerID] {
			e.logger.Debug(fmt.Sprintf("Worker %d with previously registered generator restarted, flagging as unregistered", workerID))
			delete(e.generatorRegisteredWorker, workerID)
		}
	}
	e.workerStartHook = &hook
	workerPool.AddWorkerStartHook(e.workerStartHook)
	return e
}

func (e *AsyncGeneratorExecutor) registerGeneratorToWorker(worker int) {
	e.logger.Debug(fmt.Sprintf("Registering generator on worker %d", worker))
	e.workerPool.SubmitTaskToWorker(&AsyncGeneratorRegisterTask{params: e.setupParameters, contextID: e.asyncContextID}, worker)
	e.generatorRegisteredWorker[worker] = true
}

func (e *AsyncGeneratorExecutor) unregisterGenerator() {
	running := map[int]bool{}
	for _, id := range e.workerPool.GetRunningWorkers() {
		running[id] = true
	}
	for id := range e.generatorRegisteredWorker {
		if running[id] {
			e.workerPool.SubmitTaskToWorker(&AsyncGeneratorUnregisterTask{contextID: e.asyncContextID}, id)
		}
	}
	e.generatorRegisteredWorker = map[int]bool{}
}

// Populate is a port of AsyncGeneratorExecutor::populate.
func (e *AsyncGeneratorExecutor) Populate(chunkX, chunkZ int, centerChunk *format.Chunk, adjacentChunks map[[2]int]*format.Chunk, registry *blockStateRegistry, onCompletion func(PopulationResult)) {
	task := &PopulationTask{
		contextID:    e.asyncContextID,
		chunkX:       chunkX,
		chunkZ:       chunkZ,
		center:       centerChunk,
		adjacent:     adjacentChunks,
		registry:     registry,
		onCompletion: onCompletion,
	}
	workerID := e.workerPool.SelectWorker()
	running := false
	for _, id := range e.workerPool.GetRunningWorkers() {
		if id == workerID {
			running = true
		}
	}
	if !running && e.generatorRegisteredWorker[workerID] {
		e.logger.Debug(fmt.Sprintf("Selected worker %d previously had generator registered, but is now offline", workerID))
		delete(e.generatorRegisteredWorker, workerID)
	}
	if !e.generatorRegisteredWorker[workerID] {
		e.registerGeneratorToWorker(workerID)
	}
	e.workerPool.SubmitTaskToWorker(task, workerID)
}

// Shutdown is a port of AsyncGeneratorExecutor::shutdown.
func (e *AsyncGeneratorExecutor) Shutdown() {
	e.unregisterGenerator()
	e.workerPool.RemoveWorkerStartHook(e.workerStartHook)
}
