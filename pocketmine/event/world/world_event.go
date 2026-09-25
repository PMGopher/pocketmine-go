// Package world is a port of pocketmine\event\world: events about worlds and chunks.
//
// It sits below pocketmine/world (which fires these events), so the world is a small local
// interface that *world.World satisfies. Chunks are *format.Chunk.
//
// Importers conventionally alias this package as worldevent.
package world

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// World is the surface these events need from pocketmine\world\World.
type World interface {
	GetDisplayName() string
	GetFolderName() string
}

// WorldEvent is a port of pocketmine\event\world\WorldEvent.
type WorldEvent struct {
	world World
}

func (e *WorldEvent) GetWorld() World { return e.world }

// ChunkEvent is a port of pocketmine\event\world\ChunkEvent: chunk-related events.
type ChunkEvent struct {
	WorldEvent

	chunkX, chunkZ int
	chunk          *format.Chunk
}

func newChunkEvent(world World, chunkX, chunkZ int, chunk *format.Chunk) ChunkEvent {
	return ChunkEvent{WorldEvent: WorldEvent{world: world}, chunkX: chunkX, chunkZ: chunkZ, chunk: chunk}
}

func (e *ChunkEvent) GetChunk() *format.Chunk { return e.chunk }
func (e *ChunkEvent) GetChunkX() int          { return e.chunkX }
func (e *ChunkEvent) GetChunkZ() int          { return e.chunkZ }

// ChunkLoadEvent is a port of pocketmine\event\world\ChunkLoadEvent: called when a chunk is
// loaded or created by the world generator.
type ChunkLoadEvent struct {
	ChunkEvent

	newChunk bool
}

func NewChunkLoadEvent(world World, chunkX, chunkZ int, chunk *format.Chunk, newChunk bool) *ChunkLoadEvent {
	return &ChunkLoadEvent{ChunkEvent: newChunkEvent(world, chunkX, chunkZ, chunk), newChunk: newChunk}
}

// IsNewChunk returns whether the chunk is newly generated. If false, the chunk was loaded from
// storage.
func (e *ChunkLoadEvent) IsNewChunk() bool { return e.newChunk }

// ChunkPopulateEvent is a port of pocketmine\event\world\ChunkPopulateEvent: called when a chunk
// is populated (after receiving it on the main thread).
type ChunkPopulateEvent struct {
	ChunkEvent
}

func NewChunkPopulateEvent(world World, chunkX, chunkZ int, chunk *format.Chunk) *ChunkPopulateEvent {
	return &ChunkPopulateEvent{ChunkEvent: newChunkEvent(world, chunkX, chunkZ, chunk)}
}

// ChunkUnloadEvent is a port of pocketmine\event\world\ChunkUnloadEvent: called when a chunk is
// unloaded.
type ChunkUnloadEvent struct {
	ChunkEvent
	event.CancellableTrait
}

func NewChunkUnloadEvent(world World, chunkX, chunkZ int, chunk *format.Chunk) *ChunkUnloadEvent {
	return &ChunkUnloadEvent{ChunkEvent: newChunkEvent(world, chunkX, chunkZ, chunk)}
}

// SpawnChangeEvent is a port of pocketmine\event\world\SpawnChangeEvent: an event that is called
// when a world spawn changes. The previous spawn is included.
type SpawnChangeEvent struct {
	WorldEvent

	previousSpawn entityevent.Position
}

func NewSpawnChangeEvent(world World, previousSpawn entityevent.Position) *SpawnChangeEvent {
	return &SpawnChangeEvent{WorldEvent: WorldEvent{world: world}, previousSpawn: previousSpawn}
}

func (e *SpawnChangeEvent) GetPreviousSpawn() entityevent.Position { return e.previousSpawn }

// WorldDifficultyChangeEvent is a port of pocketmine\event\world\WorldDifficultyChangeEvent.
type WorldDifficultyChangeEvent struct {
	WorldEvent

	oldDifficulty, newDifficulty int
}

func NewWorldDifficultyChangeEvent(world World, oldDifficulty, newDifficulty int) *WorldDifficultyChangeEvent {
	return &WorldDifficultyChangeEvent{WorldEvent: WorldEvent{world: world}, oldDifficulty: oldDifficulty, newDifficulty: newDifficulty}
}

func (e *WorldDifficultyChangeEvent) GetOldDifficulty() int { return e.oldDifficulty }
func (e *WorldDifficultyChangeEvent) GetNewDifficulty() int { return e.newDifficulty }

// WorldDisplayNameChangeEvent is a port of pocketmine\event\world\WorldDisplayNameChangeEvent.
type WorldDisplayNameChangeEvent struct {
	WorldEvent

	oldName, newName string
}

func NewWorldDisplayNameChangeEvent(world World, oldName, newName string) *WorldDisplayNameChangeEvent {
	return &WorldDisplayNameChangeEvent{WorldEvent: WorldEvent{world: world}, oldName: oldName, newName: newName}
}

func (e *WorldDisplayNameChangeEvent) GetOldName() string { return e.oldName }
func (e *WorldDisplayNameChangeEvent) GetNewName() string { return e.newName }

// WorldInitEvent is a port of pocketmine\event\world\WorldInitEvent: called when a world is
// initialized, before it is loaded.
type WorldInitEvent struct {
	WorldEvent
}

func NewWorldInitEvent(world World) *WorldInitEvent {
	return &WorldInitEvent{WorldEvent: WorldEvent{world: world}}
}

// WorldLoadEvent is a port of pocketmine\event\world\WorldLoadEvent: called when a world is
// loaded.
type WorldLoadEvent struct {
	WorldEvent
}

func NewWorldLoadEvent(world World) *WorldLoadEvent {
	return &WorldLoadEvent{WorldEvent: WorldEvent{world: world}}
}

// WorldSaveEvent is a port of pocketmine\event\world\WorldSaveEvent: called when a world is
// saved.
type WorldSaveEvent struct {
	WorldEvent
}

func NewWorldSaveEvent(world World) *WorldSaveEvent {
	return &WorldSaveEvent{WorldEvent: WorldEvent{world: world}}
}

// WorldUnloadEvent is a port of pocketmine\event\world\WorldUnloadEvent: called when a world is
// unloaded.
type WorldUnloadEvent struct {
	WorldEvent
	event.CancellableTrait
}

func NewWorldUnloadEvent(world World) *WorldUnloadEvent {
	return &WorldUnloadEvent{WorldEvent: WorldEvent{world: world}}
}

// WorldParticleEvent is a port of pocketmine\event\world\WorldParticleEvent. Recipients are the
// players (world.EntityViewer values) the particle will be sent to.
type WorldParticleEvent struct {
	WorldEvent
	event.CancellableTrait

	particle   particle.Particle
	position   math.Vector3
	recipients []any
}

func NewWorldParticleEvent(world World, p particle.Particle, position math.Vector3, recipients []any) *WorldParticleEvent {
	return &WorldParticleEvent{WorldEvent: WorldEvent{world: world}, particle: p, position: position, recipients: recipients}
}

func (e *WorldParticleEvent) GetParticle() particle.Particle  { return e.particle }
func (e *WorldParticleEvent) SetParticle(p particle.Particle) { e.particle = p }
func (e *WorldParticleEvent) GetPosition() math.Vector3       { return e.position }
func (e *WorldParticleEvent) GetRecipients() []any            { return e.recipients }
func (e *WorldParticleEvent) SetRecipients(recipients []any)  { e.recipients = recipients }

// WorldSoundEvent is a port of pocketmine\event\world\WorldSoundEvent. Recipients are the players
// (world.EntityViewer values) the sound will be sent to.
type WorldSoundEvent struct {
	WorldEvent
	event.CancellableTrait

	sound      sound.Sound
	position   math.Vector3
	recipients []any
}

func NewWorldSoundEvent(world World, s sound.Sound, position math.Vector3, recipients []any) *WorldSoundEvent {
	return &WorldSoundEvent{WorldEvent: WorldEvent{world: world}, sound: s, position: position, recipients: recipients}
}

func (e *WorldSoundEvent) GetSound() sound.Sound          { return e.sound }
func (e *WorldSoundEvent) SetSound(s sound.Sound)         { e.sound = s }
func (e *WorldSoundEvent) GetPosition() math.Vector3      { return e.position }
func (e *WorldSoundEvent) GetRecipients() []any           { return e.recipients }
func (e *WorldSoundEvent) SetRecipients(recipients []any) { e.recipients = recipients }
