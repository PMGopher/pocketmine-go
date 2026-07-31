package permission

import (
	"fmt"
	"strings"
	"time"
)

// BanEntryDateFormat mirrors BanEntry::$format — Go's reference-time layout equivalent of PHP's
// "Y-m-d H:i:s O".
const BanEntryDateFormat = "2006-01-02 15:04:05 -0700"

// BanEntry is a port of pocketmine\permission\BanEntry.
type BanEntry struct {
	name    string
	created time.Time
	source  string
	expires *time.Time
	reason  string
}

func NewBanEntry(name string) *BanEntry {
	return &BanEntry{
		name:    strings.ToLower(name),
		created: time.Now(),
		source:  "(Unknown)",
		reason:  "Banned by an operator.",
	}
}

func (b *BanEntry) Name() string            { return b.name }
func (b *BanEntry) Created() time.Time      { return b.created }
func (b *BanEntry) SetCreated(t time.Time)  { b.created = t }
func (b *BanEntry) Source() string          { return b.source }
func (b *BanEntry) SetSource(source string) { b.source = source }
func (b *BanEntry) Expires() *time.Time     { return b.expires }
func (b *BanEntry) SetExpires(t *time.Time) { b.expires = t }
func (b *BanEntry) Reason() string          { return b.reason }
func (b *BanEntry) SetReason(reason string) { b.reason = reason }

func (b *BanEntry) HasExpired() bool {
	return b.expires != nil && b.expires.Before(time.Now())
}

func (b *BanEntry) String() string {
	expires := "Forever"
	if b.expires != nil {
		expires = b.expires.Format(BanEntryDateFormat)
	}
	return strings.Join([]string{
		b.name,
		b.created.Format(BanEntryDateFormat),
		b.source,
		expires,
		b.reason,
	}, "|")
}

// BanEntryFromString is a port of BanEntry::fromString().
func BanEntryFromString(str string) (*BanEntry, error) {
	if len(str) < 2 {
		return nil, nil
	}

	// At most 5 parts expected, but accept 6 in case of an extra unexpected delimiter — we don't
	// want to include unexpected data in the ban reason.
	parts := strings.SplitN(strings.TrimSpace(str), "|", 6)
	entry := NewBanEntry(strings.TrimSpace(parts[0]))
	parts = parts[1:]

	if len(parts) > 0 {
		t, err := time.Parse(BanEntryDateFormat, strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("corrupted date/time: %w", err)
		}
		entry.SetCreated(t)
		parts = parts[1:]
	}
	if len(parts) > 0 {
		entry.SetSource(strings.TrimSpace(parts[0]))
		parts = parts[1:]
	}
	if len(parts) > 0 {
		expire := strings.TrimSpace(parts[0])
		if expire != "" && strings.ToLower(expire) != "forever" {
			t, err := time.Parse(BanEntryDateFormat, expire)
			if err != nil {
				return nil, fmt.Errorf("corrupted date/time: %w", err)
			}
			entry.SetExpires(&t)
		}
		parts = parts[1:]
	}
	if len(parts) > 0 {
		entry.SetReason(strings.TrimSpace(parts[0]))
	}

	return entry, nil
}
