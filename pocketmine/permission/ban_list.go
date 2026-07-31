package permission

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"pocketmine-go/pocketmine/log"
)

// BanList is a port of pocketmine\permission\BanList.
type BanList struct {
	mu      sync.Mutex
	file    string
	list    map[string]*BanEntry
	enabled bool
}

func NewBanList(file string) *BanList {
	return &BanList{file: file, list: map[string]*BanEntry{}, enabled: true}
}

func (b *BanList) IsEnabled() bool   { return b.enabled }
func (b *BanList) SetEnabled(v bool) { b.enabled = v }

func (b *BanList) GetEntry(name string) *BanEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpiredLocked()
	return b.list[strings.ToLower(name)]
}

func (b *BanList) GetEntries() map[string]*BanEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpiredLocked()
	result := make(map[string]*BanEntry, len(b.list))
	for k, v := range b.list {
		result[k] = v
	}
	return result
}

func (b *BanList) IsBanned(name string) bool {
	name = strings.ToLower(name)
	if !b.IsEnabled() {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpiredLocked()
	_, ok := b.list[name]
	return ok
}

func (b *BanList) Add(entry *BanEntry) {
	b.mu.Lock()
	b.list[entry.Name()] = entry
	b.mu.Unlock()
	b.Save(true)
}

// AddBan mirrors BanList::addBan(): reason/expires/source default to BanEntry's own defaults
// when nil/empty.
func (b *BanList) AddBan(target string, reason string, expires *time.Time, source string) *BanEntry {
	entry := NewBanEntry(target)
	if source != "" {
		entry.SetSource(source)
	}
	entry.SetExpires(expires)
	if reason != "" {
		entry.SetReason(reason)
	}

	b.mu.Lock()
	b.list[entry.Name()] = entry
	b.mu.Unlock()
	b.Save(true)

	return entry
}

func (b *BanList) Remove(name string) {
	name = strings.ToLower(name)
	b.mu.Lock()
	_, exists := b.list[name]
	if exists {
		delete(b.list, name)
	}
	b.mu.Unlock()
	if exists {
		b.Save(true)
	}
}

func (b *BanList) RemoveExpired() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpiredLocked()
}

func (b *BanList) removeExpiredLocked() {
	for name, entry := range b.list {
		if entry.HasExpired() {
			delete(b.list, name)
		}
	}
}

func (b *BanList) Load() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.list = map[string]*BanEntry{}

	f, err := os.Open(b.file)
	if err != nil {
		log.Global().Error("Could not load ban list")
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entry, err := BanEntryFromString(line)
		if err != nil {
			log.Global().Critical(fmt.Sprintf("Failed to parse ban entry from string %q: %s", strings.TrimSpace(line), err.Error()))
			continue
		}
		if entry != nil {
			b.list[entry.Name()] = entry
		}
	}
}

func (b *BanList) Save(writeHeader bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpiredLocked()

	f, err := os.Create(b.file)
	if err != nil {
		log.Global().Error("Could not save ban list")
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if writeHeader {
		w.WriteString("# victim name | ban date | banned by | banned until | reason\n\n")
	}
	for _, entry := range b.list {
		w.WriteString(entry.String() + "\n")
	}
	w.Flush()
}
