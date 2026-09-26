package world

import (
	"fmt"
	stdmath "math"
	"pocketmine-go/pocketmine/event"
	worldevent "pocketmine-go/pocketmine/event/world"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	worldformatio "pocketmine-go/pocketmine/world/format/io"
)

// Difficulty constants, a port of World::DIFFICULTY_*.
const (
	DifficultyPeaceful = 0
	DifficultyEasy     = 1
	DifficultyNormal   = 2
	DifficultyHard     = 3
)

// Entity is the World's view of a pocketmine\entity\Entity - and, since World sits below
// pocketmine/entity in the import graph, the polymorphic "any entity" type the rest of this port
// uses wherever PHP code is typed against the Entity base class (World::getEntity(),
// Entity::canCollideWith(Entity), Entity::getOwningEntity(), ...). *entity.Entity and everything
// embedding it satisfy it.
type Entity interface {
	block.Entity

	GetWorld() *World
	IsAlive() bool
	IsFlaggedForDespawn() bool
	FlagForDespawn()
	Close()

	// OnUpdate is a port of Entity::onUpdate: ticks the entity, returning whether it needs to keep
	// being ticked.
	OnUpdate(currentTick int64) bool
	OnNearbyBlockChange()
	OnRandomUpdate()

	CanBeCollidedWith() bool
	CanCollideWith(other Entity) bool
	SetMotion(motion math.Vector3) bool

	CanSaveWithChunk() bool
	SaveNBT() *nbt.CompoundTag

	SpawnTo(player EntityViewer)
	DespawnFrom(player EntityViewer, send bool)
	GetViewers() []EntityViewer
}

// EntityViewer is the World's view of a pocketmine\player\Player as a viewer of entities (the
// $hasSpawned/getViewersForPosition() side of PHP's Player). *player.Player satisfies it.
type EntityViewer interface {
	SendPacket(pk packet.Packet)
	GetWorld() *World
	HasReceivedChunk(chunkX, chunkZ int) bool
}

// DropItemFunc is World::dropItem's body - creating the ItemEntity lives in
// pocketmine/entity/object, which imports this package, so that package's init() installs it
// here (the same dependency-inversion hook as block.NewItemBlockFunc). See DropItem.
var DropItemFunc func(w *World, source math.Vector3, it item.Item, motion *math.Vector3, delay int) Entity

// DropExperienceFunc is World::dropExperience's body, installed by pocketmine/entity/object (see
// DropItemFunc).
var DropExperienceFunc func(w *World, pos math.Vector3, amount int) []Entity

// LoadEntityFunc is EntityFactory::createFromData, installed by pocketmine/entity. It's how
// entities saved with a chunk are recreated when the chunk is loaded (World::initChunk). A nil
// result with a nil error means an unknown entity type.
var LoadEntityFunc func(w *World, tag *nbt.CompoundTag) (Entity, error)

// orderedEntitySet is PHP's insertion-ordered $updateEntities array (keyed by entity ID).
type orderedEntitySet struct {
	order []int
	byID  map[int]Entity
}

func (s *orderedEntitySet) set(e Entity) {
	if s.byID == nil {
		s.byID = map[int]Entity{}
	}
	if _, ok := s.byID[e.GetID()]; !ok {
		s.order = append(s.order, e.GetID())
	}
	s.byID[e.GetID()] = e
}

func (s *orderedEntitySet) remove(id int) {
	if _, ok := s.byID[id]; !ok {
		return
	}
	delete(s.byID, id)
	for i, v := range s.order {
		if v == id {
			s.order = append(s.order[:i:i], s.order[i+1:]...)
			return
		}
	}
}

func chunkOf(pos math.Vector3) [2]int { return chunkKey(pos.FloorX()>>4, pos.FloorZ()>>4) }

// AddEntity is a port of World::addEntity. Panics on the same programmer errors as PHP's
// InvalidArgumentException/AssumptionFailedError (a closed entity, an entity from another world,
// or an entity ID already in use).
//
// PHP additionally refuses entity classes that aren't registered with EntityFactory (unless they
// implement NeverSavedWithChunkEntity); that check lives in entity.Entity's constructor here,
// since only pocketmine/entity knows the factory.
func (w *World) AddEntity(e Entity) {
	if e.IsClosed() {
		panic("world: attempted to add a garbage closed Entity to world")
	}
	if e.GetWorld() != w {
		panic("world: invalid Entity world")
	}
	if existing, ok := w.entities[e.GetID()]; ok {
		if existing == e {
			panic(fmt.Sprintf("world: entity %d has already been added to this world", e.GetID()))
		}
		panic(fmt.Sprintf("world: found two different entities sharing entity ID %d", e.GetID()))
	}

	pos := e.GetPosition()
	key := chunkOf(pos)
	if w.entitiesByChunk[key] == nil {
		w.entitiesByChunk[key] = map[int]Entity{}
	}
	w.entitiesByChunk[key][e.GetID()] = e
	w.entityLastKnownPositions[e.GetID()] = pos
	w.entities[e.GetID()] = e
}

// RemoveEntity is a port of World::removeEntity. Panics if the entity isn't tracked by this world.
func (w *World) RemoveEntity(e Entity) {
	if e.GetWorld() != w {
		panic("world: invalid Entity world")
	}
	if _, ok := w.entities[e.GetID()]; !ok {
		panic("world: entity is not tracked by this world (possibly already removed?)")
	}
	pos := w.entityLastKnownPositions[e.GetID()]
	key := chunkOf(pos)
	if chunkEntities, ok := w.entitiesByChunk[key]; ok {
		delete(chunkEntities, e.GetID())
		if len(chunkEntities) == 0 {
			delete(w.entitiesByChunk, key)
		}
	}
	delete(w.entityLastKnownPositions, e.GetID())
	delete(w.entities, e.GetID())
	w.updateEntities.remove(e.GetID())
}

// GetEntity is a port of World::getEntity.
func (w *World) GetEntity(id int) (Entity, bool) {
	e, ok := w.entities[id]
	return e, ok
}

// GetEntities is a port of World::getEntities.
func (w *World) GetEntities() map[int]Entity { return w.entities }

// GetChunkEntities is a port of World::getChunkEntities.
func (w *World) GetChunkEntities(chunkX, chunkZ int) []Entity {
	chunkEntities := w.entitiesByChunk[chunkKey(chunkX, chunkZ)]
	result := make([]Entity, 0, len(chunkEntities))
	for _, e := range chunkEntities {
		result = append(result, e)
	}
	return result
}

// ScheduleEntityUpdate is Entity::scheduleUpdate's effect on the world (PHP writes straight into
// the public World::$updateEntities array): the entity will be ticked on the next world tick.
func (w *World) ScheduleEntityUpdate(e Entity) {
	w.updateEntities.set(e)
}

// OnEntityMoved is a port of World::onEntityMoved: keeps the per-chunk entity index up to date,
// and spawns/despawns the entity for players as it crosses chunk boundaries.
func (w *World) OnEntityMoved(e Entity) {
	oldPosition, ok := w.entityLastKnownPositions[e.GetID()]
	if !ok {
		//this can happen if the entity was teleported before addEntity() was called
		return
	}
	newPosition := e.GetPosition()

	oldKey := chunkOf(oldPosition)
	newKey := chunkOf(newPosition)

	if oldKey != newKey {
		if chunkEntities, ok := w.entitiesByChunk[oldKey]; ok {
			delete(chunkEntities, e.GetID())
			if len(chunkEntities) == 0 {
				delete(w.entitiesByChunk, oldKey)
			}
		}

		newViewers := map[EntityViewer]bool{}
		for _, v := range w.GetViewersForPosition(newPosition) {
			newViewers[v] = true
		}
		for _, player := range e.GetViewers() {
			if !newViewers[player] {
				e.DespawnFrom(player, true)
			} else {
				delete(newViewers, player)
			}
		}
		for _, player := range w.GetViewersForPosition(newPosition) {
			if newViewers[player] {
				e.SpawnTo(player)
			}
		}

		if w.entitiesByChunk[newKey] == nil {
			w.entitiesByChunk[newKey] = map[int]Entity{}
		}
		w.entitiesByChunk[newKey][e.GetID()] = e
	}
	w.entityLastKnownPositions[e.GetID()] = newPosition
}

// GetNearbyEntities satisfies block.World: World::getNearbyEntities without an excluded entity,
// typed as block.Entity for the block package's local interface.
func (w *World) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity {
	nearby := w.GetNearbyEntitiesExcept(bb, nil)
	result := make([]block.Entity, len(nearby))
	for i, e := range nearby {
		result[i] = e
	}
	return result
}

// GetNearbyEntitiesExcept is a port of World::getNearbyEntities: every entity whose bounding box
// intersects bb, except exclude (which may be nil).
func (w *World) GetNearbyEntitiesExcept(bb math.AxisAlignedBB, exclude Entity) []Entity {
	var nearby []Entity

	minX := int(stdmath.Floor(bb.MinX-2)) >> 4
	maxX := int(stdmath.Floor(bb.MaxX+2)) >> 4
	minZ := int(stdmath.Floor(bb.MinZ-2)) >> 4
	maxZ := int(stdmath.Floor(bb.MaxZ+2)) >> 4

	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			for _, ent := range w.entitiesByChunk[chunkKey(x, z)] {
				if ent != exclude && ent.GetBoundingBox().IntersectsWith(bb, intersectEpsilon) {
					nearby = append(nearby, ent)
				}
			}
		}
	}
	return nearby
}

// intersectEpsilon matches AxisAlignedBB::intersectsWith's PHP default parameter.
const intersectEpsilon = 0.00001

// GetCollidingEntities is a port of World::getCollidingEntities: entities in bb that can be
// collided with (by entity, if non-nil).
func (w *World) GetCollidingEntities(bb math.AxisAlignedBB, entity Entity) []Entity {
	var nearby []Entity
	for _, ent := range w.GetNearbyEntitiesExcept(bb, entity) {
		if ent.CanBeCollidedWith() && (entity == nil || entity.CanCollideWith(ent)) {
			nearby = append(nearby, ent)
		}
	}
	return nearby
}

// GetNearestEntity is a port of World::getNearestEntity. filter replaces PHP's
// `string $entityType` class filter (nil matches any entity, like the Entity::class default).
func (w *World) GetNearestEntity(pos math.Vector3, maxDistance float64, includeDead bool, filter func(Entity) bool) Entity {
	minX := int(stdmath.Floor(pos.X-maxDistance)) >> 4
	maxX := int(stdmath.Floor(pos.X+maxDistance)) >> 4
	minZ := int(stdmath.Floor(pos.Z-maxDistance)) >> 4
	maxZ := int(stdmath.Floor(pos.Z+maxDistance)) >> 4

	currentTargetDistSq := maxDistance * maxDistance
	var currentTarget Entity

	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			for _, e := range w.entitiesByChunk[chunkKey(x, z)] {
				if (filter != nil && !filter(e)) || e.IsFlaggedForDespawn() || (!includeDead && !e.IsAlive()) {
					continue
				}
				if distSq := e.GetPosition().DistanceSquared(pos); distSq < currentTargetDistSq {
					currentTargetDistSq = distSq
					currentTarget = e
				}
			}
		}
	}
	return currentTarget
}

// collisionBoxProvider is the promoted-from-*block.Block surface GetBlockCollisionBoxes needs.
type collisionBoxProvider interface {
	GetCollisionBoxes() []math.AxisAlignedBB
}

// getLoadedBlockAt is GetBlockAt without the load-or-generate side effect: PHP's
// getBlockCollisionInfo treats unloaded chunks (and positions outside the world) as having no
// collision, rather than loading them.
func (w *World) getLoadedBlockAt(x, y, z int) (block.Behavior, bool) {
	if !w.IsInWorld(x, y, z) {
		return nil, false
	}
	if _, ok := w.GetChunk(x>>4, z>>4); !ok {
		return nil, false
	}
	return w.GetBlockAt(x, y, z), true
}

// GetBlockCollisionBoxes is a port of World::getBlockCollisionBoxes: every block collision box
// intersecting bb.
//
// PHP gathers each cell's boxes through a per-state collision-info cache (COLLISION_NONE/CUBE/
// CUSTOM/MAY_OVERFLOW) plus the overflow boxes of face-adjacent MAY_OVERFLOW blocks (fences,
// walls...). That cache is a performance device; the equivalent here is to scan one extra block
// of padding around bb and keep every box that actually intersects bb, which yields the same
// colliding boxes.
func (w *World) GetBlockCollisionBoxes(bb math.AxisAlignedBB) []math.AxisAlignedBB {
	minX := int(stdmath.Floor(bb.MinX)) - 1
	minY := int(stdmath.Floor(bb.MinY)) - 1
	minZ := int(stdmath.Floor(bb.MinZ)) - 1
	maxX := int(stdmath.Floor(bb.MaxX)) + 1
	maxY := int(stdmath.Floor(bb.MaxY)) + 1
	maxZ := int(stdmath.Floor(bb.MaxZ)) + 1

	var collides []math.AxisAlignedBB
	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				blk, ok := w.getLoadedBlockAt(x, y, z)
				if !ok {
					continue
				}
				provider, ok := blk.(collisionBoxProvider)
				if !ok {
					continue
				}
				for _, blockBB := range provider.GetCollisionBoxes() {
					if blockBB.IntersectsWith(bb, intersectEpsilon) {
						collides = append(collides, blockBB)
					}
				}
			}
		}
	}
	return collides
}

// GetCollisionBoxes is a port of World::getCollisionBoxes: block collision boxes, plus (if
// entities is true) the bounding boxes of entities entity can collide with.
func (w *World) GetCollisionBoxes(entity Entity, bb math.AxisAlignedBB, entities bool) []math.AxisAlignedBB {
	collides := w.GetBlockCollisionBoxes(bb)
	if entities {
		for _, ent := range w.GetCollidingEntities(bb.ExpandedCopy(0.25, 0.25, 0.25), entity) {
			collides = append(collides, ent.GetBoundingBox())
		}
	}
	return collides
}

// GetViewersForPosition is a port of World::getViewersForPosition: the players using the chunk
// containing pos.
func (w *World) GetViewersForPosition(pos math.Vector3) []EntityViewer {
	listeners := w.GetChunkListeners(pos.FloorX()>>4, pos.FloorZ()>>4)
	viewers := make([]EntityViewer, 0, len(listeners))
	for _, l := range listeners {
		if v, ok := l.(EntityViewer); ok {
			viewers = append(viewers, v)
		}
	}
	return viewers
}

// DropItem is a port of World::dropItem: spawns an item entity (nil for an empty item). motion nil
// means a small random motion; PHP's default delay is 10 ticks.
func (w *World) DropItem(source math.Vector3, it item.Item, motion *math.Vector3, delay int) Entity {
	if it.IsNull() || DropItemFunc == nil {
		return nil
	}
	return DropItemFunc(w, source, it, motion, delay)
}

// DropBlockItem is World::dropItem($pos, $item) (default motion and delay) for blocks, which only
// hold a block.Item.
func (w *World) DropBlockItem(source math.Vector3, it block.Item) {
	if real, ok := it.(item.Item); ok {
		w.DropItem(source, real, nil, 10)
	}
}

// DropExperience is a port of World::dropExperience: spawns experience orbs worth amount in total.
func (w *World) DropExperience(pos math.Vector3, amount int) []Entity {
	if DropExperienceFunc == nil {
		return nil
	}
	return DropExperienceFunc(w, pos, amount)
}

// GetDifficulty is a port of World::getDifficulty.
func (w *World) GetDifficulty() int { return w.difficulty }

// SetDifficulty is a port of World::setDifficulty (panicking on an invalid level, like PHP's
// InvalidArgumentException): fires WorldDifficultyChangeEvent and syncs the difficulty to every
// player in the world (NetworkSession::syncWorldDifficulty).
func (w *World) SetDifficulty(difficulty int) {
	if difficulty < 0 || difficulty > 3 {
		panic(fmt.Sprintf("Invalid difficulty level %d", difficulty))
	}
	event.Call(worldevent.NewWorldDifficultyChangeEvent(w, w.difficulty, difficulty))
	w.difficulty = difficulty

	pk := &packet.SetDifficulty{Difficulty: uint32(difficulty)}
	for _, p := range w.GetPlayers() {
		p.SendPacket(pk)
	}
}

// IsInLoadedTerrain is a port of World::isInLoadedTerrain.
func (w *World) IsInLoadedTerrain(pos math.Vector3) bool {
	return w.IsChunkLoaded(pos.FloorX()>>4, pos.FloorZ()>>4)
}

// GetBlock is a port of World::getBlock (Vector3 overload of GetBlockAt).
func (w *World) GetBlock(pos math.Vector3) block.Behavior {
	return w.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())
}

// SetBlockAt is a port of World::setBlockAt.
func (w *World) SetBlockAt(x, y, z int, blk block.Behavior) error {
	return w.SetBlock(block.NewPosition(float64(x), float64(y), float64(z), w), blk)
}

// tickEntities is the "Update entities that need update" pass of World::actuallyDoTick.
func (w *World) tickEntities(currentTick int64) {
	ids := append([]int(nil), w.updateEntities.order...)
	for _, id := range ids {
		entity, ok := w.updateEntities.byID[id]
		if !ok {
			continue
		}
		if entity.IsClosed() || entity.IsFlaggedForDespawn() || !entity.OnUpdate(currentTick) {
			w.updateEntities.remove(id)
		}
		if entity.IsFlaggedForDespawn() {
			entity.Close()
		}
	}
}

// saveChunkEntities is the entity half of World::saveChunks/unloadChunk: the NBT of every entity
// in the chunk that can be saved with it.
func (w *World) saveChunkEntities(chunkX, chunkZ int) []*nbt.CompoundTag {
	var tags []*nbt.CompoundTag
	for _, e := range w.GetChunkEntities(chunkX, chunkZ) {
		if e.CanSaveWithChunk() {
			tags = append(tags, e.SaveNBT())
		}
	}
	return tags
}

// closeChunkEntities is the entity half of World::unloadChunk: every non-player entity in the
// chunk is closed. Players are recognised as the entities that are also entity viewers (PHP's
// `$entity instanceof Player`).
func (w *World) closeChunkEntities(chunkX, chunkZ int) {
	for _, e := range w.GetChunkEntities(chunkX, chunkZ) {
		if _, isPlayer := e.(EntityViewer); isPlayer {
			continue
		}
		e.Close()
	}
}

// GetCurrentTick returns the tick number of the tick World is on (the currentTick of the last
// DoTick call) - what PHP entities read as Server::getTick().
func (w *World) GetCurrentTick() int64 { return w.currentTick }

// GetBlockAtIfLoaded is World::getBlockAt with PHP's semantics for positions outside loaded
// terrain: air is returned rather than the chunk being loaded or generated (this port's GetBlockAt
// loads/generates on demand, which populators rely on). Entity movement and collision use this so
// that an entity wandering near the edge of loaded terrain never triggers generation.
func (w *World) GetBlockAtIfLoaded(x, y, z int) block.Behavior {
	if blk, ok := w.getLoadedBlockAt(x, y, z); ok {
		return blk
	}
	air := block.VanillaAir()
	if p, ok := air.(positionable); ok {
		p.SetPosition(w, x, y, z)
	}
	return air
}

// SerializeBlockState is GlobalBlockStateHandlers::getSerializer()->serialize($stateId)->toNbt(),
// which entities use to save a block (FallingBlock).
func (w *World) SerializeBlockState(blk block.Behavior) (*nbt.CompoundTag, error) {
	w.registerTemplate(blk)
	data, err := w.lookupBlockState(int32(blk.GetStateId()))
	if err != nil {
		return nil, err
	}
	return data.ToNbt(), nil
}

// DeserializeBlockState is the reverse of SerializeBlockState: GlobalBlockStateHandlers'
// upgrader (upgradeBlockStateNbt, so blocks saved by older versions load too), deserializer and
// RuntimeBlockStateRegistry::fromStateId.
func (w *World) DeserializeBlockState(tag *nbt.CompoundTag) (block.Behavior, error) {
	data, err := worldformatio.GetBlockDataUpgrader().UpgradeBlockStateNbt(tag)
	if err != nil {
		return nil, err
	}
	return w.blockFromStateData(data)
}

// DeserializeLegacyBlock is the legacy (numeric ID + meta) form of DeserializeBlockState:
// GlobalBlockStateHandlers::getUpgrader()->upgradeIntIdMeta().
func (w *World) DeserializeLegacyBlock(id, meta int) (block.Behavior, error) {
	data, err := worldformatio.GetBlockDataUpgrader().UpgradeIntIdMeta(id, meta)
	if err != nil {
		return nil, err
	}
	return w.blockFromStateData(data)
}

func (w *World) blockFromStateData(data bedrock.BlockStateData) (block.Behavior, error) {
	stateID, err := worldformatio.GetBlockStateDeserializer().Deserialize(data)
	if err != nil {
		return nil, err
	}
	tpl, ok := w.stateTemplates[int32(stateID)]
	if !ok {
		return nil, fmt.Errorf("world: no block registered for state %d", stateID)
	}
	return tpl.Clone(), nil
}

// GetCollisionBlocks is a port of World::getCollisionBlocks: the blocks whose collision boxes
// intersect bb. With targetFirst, only the first such block is returned.
func (w *World) GetCollisionBlocks(bb math.AxisAlignedBB, targetFirst bool) []block.Behavior {
	minX := int(stdmath.Floor(bb.MinX - 1))
	minY := int(stdmath.Floor(bb.MinY - 1))
	minZ := int(stdmath.Floor(bb.MinZ - 1))
	maxX := int(stdmath.Floor(bb.MaxX + 1))
	maxY := int(stdmath.Floor(bb.MaxY + 1))
	maxZ := int(stdmath.Floor(bb.MaxZ + 1))

	var collides []block.Behavior

	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				blk := w.GetBlockAtIfLoaded(x, y, z)
				if collider, ok := blk.(interface {
					CollidesWithBB(bb math.AxisAlignedBB) bool
				}); ok && collider.CollidesWithBB(bb) {
					if targetFirst {
						return []block.Behavior{blk}
					}
					collides = append(collides, blk)
				}
			}
		}
	}

	return collides
}

func init() {
	block.GetEntityFunc = func(w block.World, id int) (block.Entity, bool) {
		if world, ok := w.(*World); ok {
			if e, ok := world.GetEntity(id); ok {
				return e, true
			}
		}
		return nil, false
	}
}

// GetDifficultyFromString is a port of World::getDifficultyFromString: -1 for an unknown name.
func GetDifficultyFromString(str string) int {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "0", "peaceful", "p":
		return DifficultyPeaceful
	case "1", "easy", "e":
		return DifficultyEasy
	case "2", "normal", "n":
		return DifficultyNormal
	case "3", "hard", "h":
		return DifficultyHard
	}
	return -1
}
