// Package resourcepacks is a port of pocketmine\resourcepacks: the resource packs offered to
// clients when they join (resource_packs/resource_packs.yml). gophertunnel sends the packs to the
// client; this package decides which ones.
package resourcepacks

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// ResourcePack is a port of pocketmine\resourcepacks\ResourcePack.
type ResourcePack interface {
	// GetPackName returns the human-readable name of the resource pack.
	GetPackName() string
	// GetPackId returns the pack's UUID as a human-readable string.
	GetPackId() string
	// GetPackSize returns the size of the pack on disk in bytes.
	GetPackSize() int
	// GetPackVersion returns a version number for the pack in the format major.minor.patch.
	GetPackVersion() string
	// GetSha256 returns the raw SHA256 sum of the compressed resource pack zip.
	GetSha256() []byte
	// GetPackChunk returns a chunk of the resource pack zip as a byte-array for sending to
	// clients.
	GetPackChunk(start, length int) ([]byte, error)
}

// ResourcePackError is a port of pocketmine\resourcepacks\ResourcePackException.
type ResourcePackError struct {
	Message string
	Cause   error
}

func (e *ResourcePackError) Error() string { return e.Message }
func (e *ResourcePackError) Unwrap() error { return e.Cause }

// ZippedResourcePack is a port of pocketmine\resourcepacks\ZippedResourcePack: a .zip or .mcpack
// resource pack. The manifest is read by gophertunnel's resource package, which also delivers the
// pack to clients (see Pack).
type ZippedResourcePack struct {
	path   string
	pack   *resource.Pack
	sha256 []byte
}

// NewZippedResourcePack is a port of ZippedResourcePack::__construct.
func NewZippedResourcePack(zipPath string) (*ZippedResourcePack, error) {
	info, err := os.Stat(zipPath)
	if err != nil {
		return nil, &ResourcePackError{Message: "File not found", Cause: err}
	}
	if info.Size() == 0 {
		return nil, &ResourcePackError{Message: "Empty file, probably corrupted"}
	}
	pack, err := resource.ReadPath(zipPath)
	if err != nil {
		return nil, &ResourcePackError{Message: "Invalid manifest.json contents: " + err.Error(), Cause: err}
	}
	if pack.UUID() == uuid.Nil {
		return nil, &ResourcePackError{Message: "Resource pack has an invalid UUID"}
	}
	return &ZippedResourcePack{path: zipPath, pack: pack}, nil
}

func (p *ZippedResourcePack) GetPath() string { return p.path }

func (p *ZippedResourcePack) GetPackName() string { return p.pack.Name() }

func (p *ZippedResourcePack) GetPackVersion() string { return p.pack.Version() }

func (p *ZippedResourcePack) GetPackId() string { return p.pack.UUID().String() }

func (p *ZippedResourcePack) GetPackSize() int { return p.pack.Len() }

func (p *ZippedResourcePack) GetSha256() []byte {
	if p.sha256 == nil {
		sum := sha256.Sum256(nil)
		if data, err := os.ReadFile(p.path); err == nil {
			sum = sha256.Sum256(data)
		}
		p.sha256 = sum[:]
	}
	return p.sha256
}

// GetPackChunk is a port of ZippedResourcePack::getPackChunk.
func (p *ZippedResourcePack) GetPackChunk(start, length int) ([]byte, error) {
	if length < 1 {
		return nil, errors.New("Pack length must be positive")
	}
	if start >= p.pack.Len() {
		return nil, errors.New("Requested a resource pack chunk with invalid start offset")
	}
	buf := make([]byte, min(length, p.pack.Len()-start))
	n, err := p.pack.ReadAt(buf, int64(start))
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

// Pack returns the gophertunnel pack that delivers this pack to clients.
func (p *ZippedResourcePack) Pack() *resource.Pack { return p.pack }

// loadPackFromPath is a port of ResourcePackManager::loadPackFromPath.
func loadPackFromPath(packPath string) (ResourcePack, error) {
	info, err := os.Stat(packPath)
	if err != nil {
		return nil, &ResourcePackError{Message: "File or directory not found"}
	}
	if info.IsDir() {
		return nil, &ResourcePackError{Message: "Directory resource packs are unsupported"}
	}
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(packPath), ".")) {
	case "zip", "mcpack":
		return NewZippedResourcePack(packPath)
	}
	return nil, &ResourcePackError{Message: "Format not recognized"}
}
