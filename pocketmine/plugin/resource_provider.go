package plugin

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ResourceProvider is a port of pocketmine\plugin\ResourceProvider.
type ResourceProvider interface {
	// GetResource gets an embedded resource on the plugin file (nil if there's none). The caller
	// must close it.
	GetResource(filename string) io.ReadCloser
	// GetResources returns all the resources packaged with the plugin, keyed by their path in
	// the resources folder.
	GetResources() map[string]fs.FileInfo
}

// DiskResourceProvider is a port of pocketmine\plugin\DiskResourceProvider: resources from the
// given plugin directory on disk.
type DiskResourceProvider struct {
	file string
}

func NewDiskResourceProvider(path string) *DiskResourceProvider {
	return &DiskResourceProvider{file: strings.TrimRight(strings.ReplaceAll(path, string(filepath.Separator), "/"), "/") + "/"}
}

func (p *DiskResourceProvider) GetResource(filename string) io.ReadCloser {
	filename = strings.TrimRight(strings.ReplaceAll(filename, string(filepath.Separator), "/"), "/")
	f, err := os.Open(p.file + filename)
	if err != nil {
		return nil
	}
	if info, err := f.Stat(); err != nil || info.IsDir() {
		_ = f.Close()
		return nil
	}
	return f
}

func (p *DiskResourceProvider) GetResources() map[string]fs.FileInfo {
	resources := map[string]fs.FileInfo{}
	_ = filepath.WalkDir(p.file, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		resources[strings.ReplaceAll(strings.TrimPrefix(filepath.ToSlash(path), p.file), string(filepath.Separator), "/")] = info
		return nil
	})
	return resources
}

// FSResourceProvider is the ResourceProvider of a compiled-in Go plugin: the "resources" folder
// of the fs.FS it registered (usually an embed.FS). It plays the part DiskResourceProvider plays
// for a phar's resources/ folder.
type FSResourceProvider struct {
	fsys fs.FS
}

func NewFSResourceProvider(fsys fs.FS) *FSResourceProvider { return &FSResourceProvider{fsys: fsys} }

func (p *FSResourceProvider) GetResource(filename string) io.ReadCloser {
	if p.fsys == nil {
		return nil
	}
	f, err := p.fsys.Open(strings.Trim(strings.ReplaceAll(filename, "\\", "/"), "/"))
	if err != nil {
		return nil
	}
	if info, err := f.Stat(); err != nil || info.IsDir() {
		_ = f.Close()
		return nil
	}
	return f
}

func (p *FSResourceProvider) GetResources() map[string]fs.FileInfo {
	resources := map[string]fs.FileInfo{}
	if p.fsys == nil {
		return resources
	}
	_ = fs.WalkDir(p.fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			resources[path] = info
		}
		return nil
	})
	return resources
}
