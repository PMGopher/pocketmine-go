package player

import (
	"os"
	"path/filepath"
	"testing"

	"pocketmine-go/pocketmine/nbt"
)

func TestDatFilePlayerDataProviderHasDataFalseBeforeAnyDataIsSaved(t *testing.T) {
	p := NewDatFilePlayerDataProvider(t.TempDir())
	if p.HasData("Steve") {
		t.Error("HasData(\"Steve\") = true before anything was ever saved")
	}
}

func TestDatFilePlayerDataProviderSaveThenLoadRoundTrips(t *testing.T) {
	p := NewDatFilePlayerDataProvider(t.TempDir())
	data := nbt.NewCompoundTag().SetString("Foo", "bar").SetInt("Baz", 42)

	if err := p.SaveData("Steve", data); err != nil {
		t.Fatalf("SaveData: %v", err)
	}
	if !p.HasData("Steve") {
		t.Error("HasData(\"Steve\") = false after SaveData")
	}

	loaded, err := p.LoadData("Steve")
	if err != nil {
		t.Fatalf("LoadData: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadData returned (nil, nil) after a successful save")
	}
	if got, err := loaded.GetString("Foo"); err != nil || string(got) != "bar" {
		t.Errorf("loaded Foo = %q (err=%v), want %q", got, err, "bar")
	}
	if got, err := loaded.GetInt("Baz"); err != nil || int(got) != 42 {
		t.Errorf("loaded Baz = %d (err=%v), want 42", got, err)
	}
}

func TestDatFilePlayerDataProviderIsCaseInsensitive(t *testing.T) {
	p := NewDatFilePlayerDataProvider(t.TempDir())
	data := nbt.NewCompoundTag().SetString("Foo", "bar")

	if err := p.SaveData("Steve", data); err != nil {
		t.Fatal(err)
	}
	if !p.HasData("STEVE") {
		t.Error("HasData(\"STEVE\") = false - player data lookups should be case-insensitive")
	}
	loaded, err := p.LoadData("sTeVe")
	if err != nil || loaded == nil {
		t.Fatalf("LoadData(\"sTeVe\") = (%v, %v), want the same data back regardless of case", loaded, err)
	}
}

func TestDatFilePlayerDataProviderLoadDataReturnsNilNilWhenNothingSaved(t *testing.T) {
	p := NewDatFilePlayerDataProvider(t.TempDir())
	loaded, err := p.LoadData("NeverPlayed")
	if err != nil {
		t.Fatalf("LoadData for a never-played name returned an error: %v", err)
	}
	if loaded != nil {
		t.Errorf("LoadData for a never-played name = %v, want nil", loaded)
	}
}

func TestDatFilePlayerDataProviderLoadDataOnCorruptFileRenamesItAndReturnsAnError(t *testing.T) {
	dir := t.TempDir()
	p := NewDatFilePlayerDataProvider(dir)

	path := filepath.Join(dir, "steve.dat")
	if err := os.WriteFile(path, []byte("not a valid gzip stream"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := p.LoadData("Steve"); err == nil {
		t.Error("LoadData on corrupt data = nil error, want an error")
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Errorf("corrupt file was not renamed to .bak: %v", err)
	}
}
