package utils

import (
	"path/filepath"
	"testing"
)

func TestTOMLConfigRoundTrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.toml")
	c, err := NewConfig(file, ConfigDetect, map[string]any{
		"name":  "Steve",
		"count": 3,
		"ratio": 0.5,
		"on":    true,
		"list":  []any{"a", "b"},
		"empty": nil,
		"nested": map[string]any{
			"level":  2,
			"deeper": map[string]any{"key.with.dots": "x"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Reload(); err != nil {
		t.Fatal(err)
	}
	if c.Get("name", nil) != "Steve" || c.Get("count", nil) != 3 || c.Get("ratio", nil) != 0.5 || c.Get("on", nil) != true {
		t.Errorf("scalars = %v", c.GetAll())
	}
	if list, _ := c.Get("list", nil).([]any); len(list) != 2 || list[1] != "b" {
		t.Errorf("list = %v", c.Get("list", nil))
	}
	if c.GetNested("nested.level", nil) != 2 {
		t.Errorf("nested.level = %v (%T), want int 2", c.GetNested("nested.level", nil), c.GetNested("nested.level", nil))
	}
	deeper, _ := c.GetNested("nested.deeper", nil).(map[string]any)
	if deeper["key.with.dots"] != "x" {
		t.Errorf("nested.deeper = %v", deeper)
	}
	if _, ok := c.GetAll()["empty"]; ok {
		t.Error("a nil value was written (TOML has no null)")
	}
}
