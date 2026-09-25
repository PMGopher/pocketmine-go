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
	if _, err := os.Stat(filepath.Join(dir, "resource_packs.yml")); err != nil {
		t.Fatalf("resource_packs.yml wasn't created: %v", err)
	}
	if m.ResourcePacksRequired() || len(m.GetResourceStack()) != 0 {
		t.Fatalf("default config must require nothing and load nothing")
	}

	if err := os.WriteFile(filepath.Join(dir, "resource_packs.yml"), []byte("force_resources: true\nresource_stack:\n  - missing.zip\n  - dir\n"), 0o666); err != nil {
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
