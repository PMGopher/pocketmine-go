package player

import "pocketmine-go/pocketmine/nbt"

// TagFirstPlayed/TagLastPlayed mirror Player::TAG_FIRST_PLAYED/TAG_LAST_PLAYED.
const (
	TagFirstPlayed = "firstPlayed"
	TagLastPlayed  = "lastPlayed"
)

// IPlayer is a port of pocketmine\player\IPlayer.
type IPlayer interface {
	GetName() string
	// GetFirstPlayed/GetLastPlayed return (0, false) in place of real PHP's ?int null, matching
	// this port's established "ok bool" convention for optional values elsewhere.
	GetFirstPlayed() (int64, bool)
	GetLastPlayed() (int64, bool)
	HasPlayedBefore() bool
}

// OfflinePlayer is a port of pocketmine\player\OfflinePlayer: an IPlayer backed by that player's
// saved NBT data (if any), for looking up a player's info without them being currently connected.
type OfflinePlayer struct {
	name     string
	namedTag *nbt.CompoundTag
}

// NewOfflinePlayer is a port of OfflinePlayer::__construct. namedTag may be nil, matching real
// PHP's `?CompoundTag $namedtag` (no saved data found for this name).
func NewOfflinePlayer(name string, namedTag *nbt.CompoundTag) *OfflinePlayer {
	return &OfflinePlayer{name: name, namedTag: namedTag}
}

func (p *OfflinePlayer) GetName() string { return p.name }

// GetFirstPlayed is a port of OfflinePlayer::getFirstPlayed.
func (p *OfflinePlayer) GetFirstPlayed() (int64, bool) {
	if p.namedTag == nil {
		return 0, false
	}
	v, err := p.namedTag.GetLong(TagFirstPlayed)
	if err != nil {
		return 0, false
	}
	return int64(v), true
}

// GetLastPlayed is a port of OfflinePlayer::getLastPlayed.
func (p *OfflinePlayer) GetLastPlayed() (int64, bool) {
	if p.namedTag == nil {
		return 0, false
	}
	v, err := p.namedTag.GetLong(TagLastPlayed)
	if err != nil {
		return 0, false
	}
	return int64(v), true
}

// HasPlayedBefore is a port of OfflinePlayer::hasPlayedBefore.
func (p *OfflinePlayer) HasPlayedBefore() bool { return p.namedTag != nil }
