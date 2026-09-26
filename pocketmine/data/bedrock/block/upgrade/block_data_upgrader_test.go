package blockupgrade

import (
	"testing"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
)

func defaultUpgrader(t *testing.T) *BlockDataUpgrader {
	t.Helper()
	u, err := NewDefaultBlockDataUpgrader()
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestLoadsEverySchema(t *testing.T) {
	schemas, err := LoadSchemas(SchemaFS, "schema/nbt_upgrade_schema", int(^uint(0)>>1))
	if err != nil {
		t.Fatal(err)
	}
	if len(schemas) != 35 {
		t.Fatalf("got %d schemas, want 35", len(schemas))
	}
	for i := 1; i < len(schemas); i++ {
		if schemas[i-1].GetSchemaID() >= schemas[i].GetSchemaID() {
			t.Fatalf("schemas not ordered by schema ID")
		}
	}
	if u := NewBlockStateUpgrader(schemas); u.GetOutputVersion() != bedrock.CurrentBlockStateVersion {
		t.Fatalf("output version %d, want %d", u.GetOutputVersion(), bedrock.CurrentBlockStateVersion)
	}
}

// Current states (the 1.26.50 palette) must come out of the upgrader unchanged: they still go
// through every schema sharing the current version ID (see Upgrade's comment), which must not
// change them.
func TestCurrentStatesAreUnchanged(t *testing.T) {
	u := defaultUpgrader(t).GetBlockStateUpgrader()
	for _, state := range bedrock.BlockStates() {
		upgraded := u.Upgrade(state)
		// 0321 remaps red mushroom blocks with stem bits (10, 15) to mushroom stems, and PHP
		// re-applies it to current states too.
		if state.Name == "minecraft:red_mushroom_block" && upgraded.Name == "minecraft:mushroom_stem" {
			if bits := state.States["huge_mushroom_bits"]; bits == int32(10) || bits == int32(15) {
				continue
			}
		}
		if !upgraded.Equals(state) {
			t.Errorf("%s %v upgraded to %s %v", state.Name, state.States, upgraded.Name, upgraded.States)
		}
	}
}

func TestUpgradesOldStates(t *testing.T) {
	u := defaultUpgrader(t)
	for _, tc := range []struct {
		name    string
		states  map[string]any
		version int32
		want    string
		check   map[string]any
	}{
		// 1.19-era log with the old_log_type property, flattened in 1.20.
		{"minecraft:log", map[string]any{"old_log_type": "birch", "pillar_axis": "y"}, 17959425, "minecraft:birch_log", map[string]any{"pillar_axis": "y"}},
		{"minecraft:wool", map[string]any{"color": "red"}, 17959425, "minecraft:red_wool", map[string]any{}},
		{"minecraft:stone", map[string]any{"stone_type": "granite"}, 17959425, "minecraft:granite", map[string]any{}},
	} {
		tag := bedrock.BlockStateData{Name: tc.name, States: tc.states, Version: tc.version}.ToVanillaNbt()
		got, err := u.UpgradeBlockStateNbt(tag)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got.Name != tc.want {
			t.Errorf("%s %v upgraded to %s, want %s", tc.name, tc.states, got.Name, tc.want)
		}
		for k, v := range tc.check {
			if !bedrock.StateValuesEqual(got.States[k], v) {
				t.Errorf("%s: state %s = %v, want %v", tc.name, k, got.States[k], v)
			}
		}
	}
}

func TestUpgradesLegacyIdMeta(t *testing.T) {
	u := defaultUpgrader(t)
	for _, tc := range []struct {
		id, meta int
		want     string
	}{
		{1, 0, "minecraft:stone"},
		{1, 1, "minecraft:granite"},
		{2, 0, "minecraft:grass_block"},
		{17, 2, "minecraft:birch_log"},
		{35, 14, "minecraft:red_wool"},
		{54, 0, "minecraft:chest"},
	} {
		got, err := u.UpgradeIntIdMeta(tc.id, tc.meta)
		if err != nil {
			t.Fatalf("%d:%d: %v", tc.id, tc.meta, err)
		}
		if got.Name != tc.want {
			t.Errorf("%d:%d upgraded to %s, want %s", tc.id, tc.meta, got.Name, tc.want)
		}
		if got.Version != bedrock.CurrentBlockStateVersion {
			t.Errorf("%d:%d upgraded to version %s", tc.id, tc.meta, got.GetVersionAsString())
		}
	}
	if _, err := u.UpgradeIntIdMeta(4000, 0); err == nil {
		t.Errorf("unknown legacy ID should fail")
	}

	// Legacy name+val blockstate NBT (pre-1.13 saves).
	tag := nbt.NewCompoundTag()
	tag.SetString("name", "minecraft:planks")
	tag.SetShort("val", 2)
	got, err := u.UpgradeBlockStateNbt(tag)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "minecraft:birch_planks" {
		t.Errorf("planks:2 upgraded to %s", got.Name)
	}
}
