package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestMonsterSpawnerSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	m := NewMonsterSpawner(w, math.Vector3{})
	m.EntityTypeID = "minecraft:zombie"
	m.SpawnRange = 8
	m.DisplayEntityScale = 1.5

	saved := m.SaveNBT()
	decoded := NewMonsterSpawner(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}

	if decoded.EntityTypeID != "minecraft:zombie" {
		t.Errorf("EntityTypeID = %q, want %q", decoded.EntityTypeID, "minecraft:zombie")
	}
	if decoded.SpawnRange != 8 {
		t.Errorf("SpawnRange = %d, want 8", decoded.SpawnRange)
	}
	if decoded.DisplayEntityScale != 1.5 {
		t.Errorf("DisplayEntityScale = %v, want 1.5", decoded.DisplayEntityScale)
	}
}

func TestMonsterSpawnerReadSaveDataDefaultsWithEmptyTag(t *testing.T) {
	w := &fakeWorld{}
	m := NewMonsterSpawner(w, math.Vector3{})

	if err := m.ReadSaveData(nbt.NewCompoundTag()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if m.EntityTypeID != ":" {
		t.Errorf("EntityTypeID = %q, want %q (default)", m.EntityTypeID, ":")
	}
	if m.MinSpawnDelay != MonsterSpawnerDefaultMinSpawnDelay {
		t.Errorf("MinSpawnDelay = %d, want %d", m.MinSpawnDelay, MonsterSpawnerDefaultMinSpawnDelay)
	}
}

func TestMonsterSpawnerLegacyEntityIDFallsBackToDefault(t *testing.T) {
	w := &fakeWorld{}
	m := NewMonsterSpawner(w, math.Vector3{})
	m.EntityTypeID = "should be overwritten"

	tag := nbt.NewCompoundTag()
	tag.SetInt(monsterSpawnerTagLegacyEntityTypeID, 32)

	if err := m.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if m.EntityTypeID != ":" {
		t.Errorf("EntityTypeID = %q, want %q (LegacyEntityIdToStringIdMap not ported, falls back)", m.EntityTypeID, ":")
	}
}
