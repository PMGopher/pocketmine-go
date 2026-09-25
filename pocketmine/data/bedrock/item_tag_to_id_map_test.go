package bedrock

import "testing"

func TestItemTagToIdMap(t *testing.T) {
	m := GetItemTagToIdMap()
	if !m.TagContainsId("minecraft:boat", "minecraft:acacia_boat") {
		t.Error("minecraft:boat should contain minecraft:acacia_boat")
	}
	if m.TagContainsId("minecraft:boat", "minecraft:stick") {
		t.Error("minecraft:boat shouldn't contain minecraft:stick")
	}
	m.AddIdToTag("test:tag", "test:id")
	if ids := m.GetIdsForTag("test:tag"); len(ids) != 1 || ids[0] != "test:id" {
		t.Errorf("GetIdsForTag = %v", ids)
	}
}
