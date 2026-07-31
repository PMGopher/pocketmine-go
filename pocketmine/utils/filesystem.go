package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	CleanPathSrcPrefix     = "pmsrc"
	CleanPathPluginsPrefix = "plugins"
)

var (
	cleanedPathsMu sync.Mutex
	cleanedPaths   = map[string]string{}
)

// RecursiveUnlink deletes a file or directory tree.
//
// PHP's original manually recurses scandir() to delete children before rmdir(); Go's
// os.RemoveAll already does exactly that (and is a no-op if the path doesn't exist).
func RecursiveUnlink(path string) error {
	return os.RemoveAll(path)
}

// RecursiveCopy recursively copies a directory to a new location. The parent directory of
// destination must already exist.
func RecursiveCopy(origin, destination string) error {
	info, err := os.Stat(origin)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%s does not exist, or is not a directory", origin)
	}
	if destInfo, err := os.Stat(destination); err == nil {
		if !destInfo.IsDir() {
			return fmt.Errorf("%s already exists, and is not a directory", destination)
		}
	} else {
		parentInfo, parentErr := os.Stat(filepath.Dir(destination))
		if parentErr != nil || !parentInfo.IsDir() {
			return fmt.Errorf("the parent directory of %s does not exist, or is not a directory", destination)
		}
		if err := os.Mkdir(destination, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", destination, err)
		}
	}
	return recursiveCopyInternal(origin, destination)
}

func recursiveCopyInternal(origin, destination string) error {
	entries, err := os.ReadDir(origin)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(origin, entry.Name())
		dst := filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			if err := recursiveCopyInternal(src, dst); err != nil {
				return err
			}
		} else if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// AddCleanedPath registers a path to be redacted (replaced with replacement) when CleanPath is
// called, e.g. to strip the local filesystem layout out of crash report traces.
func AddCleanedPath(path, replacement string) {
	cleanedPathsMu.Lock()
	defer cleanedPathsMu.Unlock()
	cleanedPaths[path] = replacement
}

// GetCleanedPaths returns a snapshot of the registered path redactions.
func GetCleanedPaths() map[string]string {
	cleanedPathsMu.Lock()
	defer cleanedPathsMu.Unlock()
	result := make(map[string]string, len(cleanedPaths))
	for k, v := range cleanedPaths {
		result[k] = v
	}
	return result
}

// CleanPath sanitizes a filesystem path for logs/crash reports, replacing registered path
// prefixes (the longest match first) and stripping OS-specific separators and phar/.php noise.
func CleanPath(path string) string {
	result := strings.NewReplacer(string(os.PathSeparator), "/", ".php", "", "phar://", "").Replace(path)

	cleanedPathsMu.Lock()
	prefixes := make([]string, 0, len(cleanedPaths))
	for k := range cleanedPaths {
		prefixes = append(prefixes, k)
	}
	cleanedPathsMu.Unlock()
	sort.Slice(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) }) // longest first

	for _, cleanPathPrefix := range prefixes {
		replacement := cleanedPaths[cleanPathPrefix]
		normalized := strings.TrimRight(strings.NewReplacer(string(os.PathSeparator), "/", "phar://", "").Replace(cleanPathPrefix), "/")
		if strings.HasPrefix(result, normalized) {
			result = strings.TrimPrefix(strings.Replace(result, normalized, replacement, 1), "/")
		}
	}
	return result
}

// SafeFilePutContents writes to a temporary file before renaming over the original, so a
// full disk fails the write before the original file is touched, rather than truncating it.
func SafeFilePutContents(fileName string, contents []byte) error {
	directory := filepath.Dir(fileName)
	if info, err := os.Stat(directory); err != nil || !info.IsDir() {
		return fmt.Errorf("target directory path does not exist or is not a directory")
	}
	if info, err := os.Stat(fileName); err == nil && info.IsDir() {
		return fmt.Errorf("target file path already exists and is not a file")
	}

	counter := 0
	var temporaryFileName string
	for {
		temporaryFileName = fmt.Sprintf("%s.%d.tmp", fileName, counter)
		if info, err := os.Stat(temporaryFileName); err != nil || !info.IsDir() {
			break
		}
		counter++
	}

	if err := os.WriteFile(temporaryFileName, contents, 0o644); err != nil {
		os.Remove(temporaryFileName)
		return fmt.Errorf("failed to write to temporary file %s: %w", temporaryFileName, err)
	}

	if err := os.Rename(temporaryFileName, fileName); err != nil {
		if err := copyFile(temporaryFileName, fileName); err != nil {
			return fmt.Errorf("failed to move temporary file contents into target file: %w", err)
		}
		os.Remove(temporaryFileName)
	}
	return nil
}
