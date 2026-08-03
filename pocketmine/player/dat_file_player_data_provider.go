package player

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pocketmine-go/pocketmine/nbt"
)

// datFileNbtMaxDepth mirrors io/leveldb's own constant of the same name/value - a generic NBT-
// reader safety limit, not specific to any one package (see io/leveldb's own doc comment).
const datFileNbtMaxDepth = 512

// DatFilePlayerDataProvider is a port of pocketmine\player\DatFilePlayerDataProvider: stores
// player data in a single .dat file per player, gzipped big-endian NBT - real PHP's own
// `zlib_encode(..., ZLIB_ENCODING_GZIP)`/`zlib_decode` is exactly the gzip container format Go's
// compress/gzip already speaks, not raw zlib/deflate despite the confusingly-named PHP constant.
type DatFilePlayerDataProvider struct {
	path string
}

func NewDatFilePlayerDataProvider(path string) *DatFilePlayerDataProvider {
	return &DatFilePlayerDataProvider{path: path}
}

func (p *DatFilePlayerDataProvider) getPlayerDataPath(username string) string {
	return filepath.Join(p.path, strings.ToLower(username)+".dat")
}

// handleCorruptedPlayerData is a port of DatFilePlayerDataProvider::handleCorruptedPlayerData -
// renames the offending file out of the way (appending .bak) so a corrupt file doesn't keep
// failing to load on every subsequent attempt.
func (p *DatFilePlayerDataProvider) handleCorruptedPlayerData(name string) {
	path := p.getPlayerDataPath(name)
	_ = os.Rename(path, path+".bak")
}

// HasData is a port of DatFilePlayerDataProvider::hasData.
func (p *DatFilePlayerDataProvider) HasData(name string) bool {
	_, err := os.Stat(p.getPlayerDataPath(name))
	return err == nil
}

// LoadData is a port of DatFilePlayerDataProvider::loadData.
func (p *DatFilePlayerDataProvider) LoadData(name string) (*nbt.CompoundTag, error) {
	name = strings.ToLower(name)
	path := p.getPlayerDataPath(name)

	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("player data: reading player data file %q: %w", path, err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(contents))
	if err != nil {
		p.handleCorruptedPlayerData(name)
		return nil, fmt.Errorf("player data: failed to decompress raw player data for %q: %w", name, err)
	}
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		p.handleCorruptedPlayerData(name)
		return nil, fmt.Errorf("player data: failed to decompress raw player data for %q: %w", name, err)
	}

	root, _, err := nbt.NewBigEndianSerializer().Read(decompressed, 0, datFileNbtMaxDepth)
	if err != nil {
		p.handleCorruptedPlayerData(name)
		return nil, fmt.Errorf("player data: failed to decode NBT data for %q: %w", name, err)
	}
	tag, err := root.MustGetCompoundTag()
	if err != nil {
		p.handleCorruptedPlayerData(name)
		return nil, fmt.Errorf("player data: failed to decode NBT data for %q: %w", name, err)
	}
	return tag, nil
}

// SaveData is a port of DatFilePlayerDataProvider::saveData.
func (p *DatFilePlayerDataProvider) SaveData(name string, data *nbt.CompoundTag) error {
	root, err := nbt.NewTreeRoot(data, "")
	if err != nil {
		return err
	}
	encoded, err := nbt.NewBigEndianSerializer().Write(root)
	if err != nil {
		return fmt.Errorf("player data: encoding player data for %q: %w", name, err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(encoded); err != nil {
		return fmt.Errorf("player data: compressing player data for %q: %w", name, err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("player data: compressing player data for %q: %w", name, err)
	}

	if err := os.MkdirAll(p.path, 0o755); err != nil {
		return fmt.Errorf("player data: creating player data directory %q: %w", p.path, err)
	}
	if err := os.WriteFile(p.getPlayerDataPath(name), buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("player data: writing player data file: %w", err)
	}
	return nil
}
