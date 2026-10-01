package resourcepacks

import (
	"os"
	"path/filepath"
	"testing"

	"pocketmine-go/pocketmine/log"
)

func TestManagerCreatesConfigAndReportsBadEntries(t *testing.T) {
	dir := t.TempDir()
	m, err := NewResourcePackManager(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "resource_packs.toml")); err != nil {
		t.Fatalf("resource_packs.toml wasn't created: %v", err)
	}
	if m.ResourcePacksRequired() || len(m.GetResourceStack()) != 0 {
		t.Fatalf("default config must require nothing and load nothing")
	}

	if err := os.WriteFile(filepath.Join(dir, "resource_packs.toml"), []byte("force_resources = true\nresource_stack = [\"missing.zip\", \"dir\"]\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	_ = os.Mkdir(filepath.Join(dir, "dir"), 0o777)
	m, err = NewResourcePackManager(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	if !m.ResourcePacksRequired() || len(m.GetResourceStack()) != 0 {
		t.Fatalf("bad packs must be skipped: %v", m.GetResourceStack())
	}
	if err := m.SetPackEncryptionKey("00000000-0000-0000-0000-000000000000", "x"); err == nil {
		t.Fatalf("setting a key for an unknown pack must fail")
	}
}

// A resource_packs.yml of an older version becomes resource_packs.toml with its settings.
func TestManagerConvertsOldConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "resource_packs.yml"), []byte("force_resources: true\nresource_stack: []\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	m, err := NewResourcePackManager(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	if !m.ResourcePacksRequired() {
		t.Error("force_resources from resource_packs.yml was lost")
	}
	for _, name := range []string{"resource_packs.toml", "resource_packs.yml.bak"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s missing: %v", name, err)
		}
	}
}
