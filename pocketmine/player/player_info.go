package player

import (
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/utils"
)

// PlayerInfo is a port of pocketmine\player\PlayerInfo: encapsulates the data needed to create a
// player.
type PlayerInfo struct {
	username  string
	uuid      string
	skin      *entity.Skin
	locale    string
	extraData map[string]any
}

// NewPlayerInfo is a port of PlayerInfo::__construct - `TextFormat::clean($username)` is the
// already-ported utils.Clean(username, true).
func NewPlayerInfo(username string, uuid string, skin *entity.Skin, locale string, extraData map[string]any) *PlayerInfo {
	return &PlayerInfo{
		username:  utils.Clean(username, true),
		uuid:      uuid,
		skin:      skin,
		locale:    locale,
		extraData: extraData,
	}
}

func (i *PlayerInfo) GetUsername() string          { return i.username }
func (i *PlayerInfo) GetUUID() string              { return i.uuid }
func (i *PlayerInfo) GetSkin() *entity.Skin        { return i.skin }
func (i *PlayerInfo) GetLocale() string            { return i.locale }
func (i *PlayerInfo) GetExtraData() map[string]any { return i.extraData }

// XboxLivePlayerInfo is a port of pocketmine\player\XboxLivePlayerInfo: PlayerInfo plus the XUID
// that identifies an Xbox Live-authenticated player.
type XboxLivePlayerInfo struct {
	PlayerInfo
	xuid string
}

// NewXboxLivePlayerInfo is a port of XboxLivePlayerInfo::__construct.
func NewXboxLivePlayerInfo(xuid, username, uuid string, skin *entity.Skin, locale string, extraData map[string]any) *XboxLivePlayerInfo {
	return &XboxLivePlayerInfo{
		PlayerInfo: *NewPlayerInfo(username, uuid, skin, locale, extraData),
		xuid:       xuid,
	}
}

func (i *XboxLivePlayerInfo) GetXuid() string { return i.xuid }

// WithoutXboxData is a port of XboxLivePlayerInfo::withoutXboxData: returns a new PlayerInfo with
// XBL data stripped, used to ensure non-XBL players can't spoof XBL data.
func (i *XboxLivePlayerInfo) WithoutXboxData() *PlayerInfo {
	return NewPlayerInfo(i.GetUsername(), i.GetUUID(), i.GetSkin(), i.GetLocale(), i.GetExtraData())
}
