package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	monsterSpawnerTagLegacyEntityTypeID  = "EntityId"
	monsterSpawnerTagEntityTypeID        = "EntityIdentifier"
	monsterSpawnerTagSpawnDelay          = "Delay"
	monsterSpawnerTagSpawnPotentials     = "SpawnPotentials"
	monsterSpawnerTagSpawnData           = "SpawnData"
	monsterSpawnerTagMinSpawnDelay       = "MinSpawnDelay"
	monsterSpawnerTagMaxSpawnDelay       = "MaxSpawnDelay"
	monsterSpawnerTagSpawnPerAttempt     = "SpawnCount"
	monsterSpawnerTagMaxNearbyEntities   = "MaxNearbyEntities"
	monsterSpawnerTagRequiredPlayerRange = "RequiredPlayerRange"
	monsterSpawnerTagSpawnRange          = "SpawnRange"
	monsterSpawnerTagEntityWidth         = "DisplayEntityWidth"
	monsterSpawnerTagEntityHeight        = "DisplayEntityHeight"
	monsterSpawnerTagEntityScale         = "DisplayEntityScale"

	MonsterSpawnerDefaultMinSpawnDelay = 200
	MonsterSpawnerDefaultMaxSpawnDelay = 800

	MonsterSpawnerDefaultMaxNearbyEntities   = 6
	MonsterSpawnerDefaultSpawnRange          = 4
	MonsterSpawnerDefaultRequiredPlayerRange = 16
)

// MonsterSpawner is a port of pocketmine\block\tile\MonsterSpawner.
//
// Deprecated in the PHP original too.
//
// spawnPotentials/spawnData are kept as raw NBT (not deserialized into a structured form) -
// matching the PHP original's own TODOs, which note PC/PE formats differ and full deserialization
// isn't implemented upstream either. The legacy numeric EntityId tag needs
// LegacyEntityIdToStringIdMap (an unported data table), so that branch falls back to the default
// entity type ID, documented below.
type MonsterSpawner struct {
	SpawnableBase

	EntityTypeID        string
	SpawnPotentials     *nbt.ListTag
	SpawnData           *nbt.CompoundTag
	DisplayEntityWidth  float64
	DisplayEntityHeight float64
	DisplayEntityScale  float64
	SpawnDelay          int
	MinSpawnDelay       int
	MaxSpawnDelay       int
	SpawnPerAttempt     int
	MaxNearbyEntities   int
	SpawnRange          int
	RequiredPlayerRange int
}

func NewMonsterSpawner(world World, pos math.Vector3) *MonsterSpawner {
	m := &MonsterSpawner{
		SpawnableBase:       SpawnableBase{TileBase: NewTileBase(world, pos)},
		EntityTypeID:        ":",
		DisplayEntityWidth:  1.0,
		DisplayEntityHeight: 1.0,
		DisplayEntityScale:  1.0,
		SpawnDelay:          MonsterSpawnerDefaultMinSpawnDelay,
		MinSpawnDelay:       MonsterSpawnerDefaultMinSpawnDelay,
		MaxSpawnDelay:       MonsterSpawnerDefaultMaxSpawnDelay,
		SpawnPerAttempt:     1,
		MaxNearbyEntities:   MonsterSpawnerDefaultMaxNearbyEntities,
		SpawnRange:          MonsterSpawnerDefaultSpawnRange,
		RequiredPlayerRange: MonsterSpawnerDefaultRequiredPlayerRange,
	}
	m.Init(m)
	return m
}

func (m *MonsterSpawner) SaveID() string { return "MobSpawner" }

func (m *MonsterSpawner) ReadSaveData(tag *nbt.CompoundTag) error {
	if _, ok := tag.GetTag(monsterSpawnerTagLegacyEntityTypeID); ok {
		// TODO: LegacyEntityIdToStringIdMap isn't ported, so this always falls back to the
		// default entity type ID instead of resolving the legacy numeric ID (same category of
		// gap as everywhere else an unported data table is needed).
		m.EntityTypeID = ":"
	} else if idTag, ok := tag.GetTag(monsterSpawnerTagEntityTypeID); ok {
		if strTag, ok := idTag.(nbt.StringTag); ok {
			m.EntityTypeID = string(strTag)
		} else {
			m.EntityTypeID = ":"
		}
	} else {
		m.EntityTypeID = ":"
	}

	if spawnData, ok, err := tag.GetCompoundTag(monsterSpawnerTagSpawnData); err == nil {
		if ok {
			m.SpawnData = spawnData
		} else {
			m.SpawnData = nil
		}
	}
	if spawnPotentials, ok, err := tag.GetListTag(monsterSpawnerTagSpawnPotentials); err == nil {
		if ok {
			m.SpawnPotentials = spawnPotentials
		} else {
			m.SpawnPotentials = nil
		}
	}

	m.SpawnDelay = int(tag.GetShortOr(monsterSpawnerTagSpawnDelay, MonsterSpawnerDefaultMinSpawnDelay))
	m.MinSpawnDelay = int(tag.GetShortOr(monsterSpawnerTagMinSpawnDelay, MonsterSpawnerDefaultMinSpawnDelay))
	m.MaxSpawnDelay = int(tag.GetShortOr(monsterSpawnerTagMaxSpawnDelay, MonsterSpawnerDefaultMaxSpawnDelay))
	m.SpawnPerAttempt = int(tag.GetShortOr(monsterSpawnerTagSpawnPerAttempt, 1))
	m.MaxNearbyEntities = int(tag.GetShortOr(monsterSpawnerTagMaxNearbyEntities, MonsterSpawnerDefaultMaxNearbyEntities))
	m.RequiredPlayerRange = int(tag.GetShortOr(monsterSpawnerTagRequiredPlayerRange, MonsterSpawnerDefaultRequiredPlayerRange))
	m.SpawnRange = int(tag.GetShortOr(monsterSpawnerTagSpawnRange, MonsterSpawnerDefaultSpawnRange))

	m.DisplayEntityWidth = float64(tag.GetFloatOr(monsterSpawnerTagEntityWidth, 1.0))
	m.DisplayEntityHeight = float64(tag.GetFloatOr(monsterSpawnerTagEntityHeight, 1.0))
	m.DisplayEntityScale = float64(tag.GetFloatOr(monsterSpawnerTagEntityScale, 1.0))
	return nil
}

func (m *MonsterSpawner) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetString(monsterSpawnerTagEntityTypeID, nbt.StringTag(m.EntityTypeID))
	if m.SpawnData != nil {
		tag.SetTag(monsterSpawnerTagSpawnData, m.SpawnData.Clone())
	}
	if m.SpawnPotentials != nil {
		tag.SetTag(monsterSpawnerTagSpawnPotentials, m.SpawnPotentials.Clone())
	}

	tag.SetShort(monsterSpawnerTagSpawnDelay, nbt.ShortTag(m.SpawnDelay))
	tag.SetShort(monsterSpawnerTagMinSpawnDelay, nbt.ShortTag(m.MinSpawnDelay))
	tag.SetShort(monsterSpawnerTagMaxSpawnDelay, nbt.ShortTag(m.MaxSpawnDelay))
	tag.SetShort(monsterSpawnerTagSpawnPerAttempt, nbt.ShortTag(m.SpawnPerAttempt))
	tag.SetShort(monsterSpawnerTagMaxNearbyEntities, nbt.ShortTag(m.MaxNearbyEntities))
	tag.SetShort(monsterSpawnerTagRequiredPlayerRange, nbt.ShortTag(m.RequiredPlayerRange))
	tag.SetShort(monsterSpawnerTagSpawnRange, nbt.ShortTag(m.SpawnRange))

	tag.SetFloat(monsterSpawnerTagEntityWidth, nbt.FloatTag(m.DisplayEntityWidth))
	tag.SetFloat(monsterSpawnerTagEntityHeight, nbt.FloatTag(m.DisplayEntityHeight))
	tag.SetFloat(monsterSpawnerTagEntityScale, nbt.FloatTag(m.DisplayEntityScale))
}

// AddAdditionalSpawnData is a port of MonsterSpawner::addAdditionalSpawnData. SpawnData is
// deliberately not included here, matching the PHP original's own TODO: it might crash the
// client if it's from a PC world, since full deserialization isn't implemented.
func (m *MonsterSpawner) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetString(monsterSpawnerTagEntityTypeID, nbt.StringTag(m.EntityTypeID))
	tag.SetFloat(monsterSpawnerTagEntityScale, nbt.FloatTag(m.DisplayEntityScale))
}
