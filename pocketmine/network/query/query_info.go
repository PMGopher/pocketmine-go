// Package query is a port of pocketmine\network\query: the GameSpy4-style query protocol used by
// server list sites to read the server name, player list and plugins.
package query

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// GameID is QueryInfo::GAME_ID.
const GameID = "MINECRAFTPE"

// Plugin is what QueryInfo needs from pocketmine\plugin\Plugin.
type Plugin interface {
	GetName() string
	GetVersion() string
}

// Server is what QueryInfo and QueryHandler need from pocketmine\Server.
type Server interface {
	GetMotd() string
	// QueryListPlugins is pocketmine.yml's settings.query-plugins.
	QueryListPlugins() bool
	GetQueryPlugins() []Plugin
	GetOnlinePlayerNames() []string
	// IsSurvivalLike reports whether the default game mode is survival or adventure.
	IsSurvivalLike() bool
	GetVersion() string
	GetName() string
	GetPocketMineVersion() string
	GetDefaultWorldName() (string, bool)
	GetMaxPlayers() int
	HasWhitelist() bool
	GetPort() int
	GetIp() string
	GetQueryInformation() *QueryInfo
}

// QueryInfo is a port of pocketmine\network\query\QueryInfo: the information the query protocol
// hands out, rebuilt every second (Server::tick) and editable through QueryRegenerateEvent.
type QueryInfo struct {
	serverName      string
	listPlugins     bool
	plugins         []Plugin
	players         []string
	gametype        string
	version         string
	serverEngine    string
	worldName       string
	numPlayers      int
	maxPlayers      int
	whitelist       string
	port            int
	ip              string
	extraKeys       []string
	extraData       map[string]string
	longQueryCache  *string
	shortQueryCache *string
}

func NewQueryInfo(server Server) *QueryInfo {
	q := &QueryInfo{
		serverName:   server.GetMotd(),
		listPlugins:  server.QueryListPlugins(),
		plugins:      server.GetQueryPlugins(),
		players:      server.GetOnlinePlayerNames(),
		gametype:     "CMP",
		version:      server.GetVersion(),
		serverEngine: server.GetName() + " " + server.GetPocketMineVersion(),
		maxPlayers:   server.GetMaxPlayers(),
		whitelist:    "off",
		port:         server.GetPort(),
		ip:           server.GetIp(),
		extraData:    map[string]string{},
	}
	if server.IsSurvivalLike() {
		q.gametype = "SMP"
	}
	q.worldName = "unknown"
	if name, ok := server.GetDefaultWorldName(); ok {
		q.worldName = name
	}
	q.numPlayers = len(q.players)
	if server.HasWhitelist() {
		q.whitelist = "on"
	}
	return q
}

func (q *QueryInfo) destroyCache() {
	q.longQueryCache = nil
	q.shortQueryCache = nil
}

func (q *QueryInfo) GetServerName() string { return q.serverName }

func (q *QueryInfo) SetServerName(serverName string) {
	q.serverName = serverName
	q.destroyCache()
}

func (q *QueryInfo) CanListPlugins() bool { return q.listPlugins }

func (q *QueryInfo) SetListPlugins(value bool) {
	q.listPlugins = value
	q.destroyCache()
}

func (q *QueryInfo) GetPlugins() []Plugin { return q.plugins }

func (q *QueryInfo) SetPlugins(plugins []Plugin) {
	q.plugins = plugins
	q.destroyCache()
}

func (q *QueryInfo) GetPlayerList() []string { return q.players }

func (q *QueryInfo) SetPlayerList(players []string) {
	q.players = players
	q.destroyCache()
}

func (q *QueryInfo) GetPlayerCount() int { return q.numPlayers }

func (q *QueryInfo) SetPlayerCount(count int) {
	q.numPlayers = count
	q.destroyCache()
}

func (q *QueryInfo) GetMaxPlayerCount() int { return q.maxPlayers }

func (q *QueryInfo) SetMaxPlayerCount(count int) {
	q.maxPlayers = count
	q.destroyCache()
}

func (q *QueryInfo) GetWorld() string { return q.worldName }

func (q *QueryInfo) SetWorld(world string) {
	q.worldName = world
	q.destroyCache()
}

// GetExtraData returns the extra Query data in key => value form.
func (q *QueryInfo) GetExtraData() map[string]string { return q.extraData }

// SetExtraData sets the extra Query data; keys keep the order given in keys.
func (q *QueryInfo) SetExtraData(keys []string, extraData map[string]string) {
	q.extraKeys = keys
	q.extraData = extraData
	q.destroyCache()
}

func sanitize(s string) string {
	return strings.NewReplacer(";", "", ":", "", " ", "_").Replace(s)
}

// GetLongQuery is a port of QueryInfo::getLongQuery.
func (q *QueryInfo) GetLongQuery() []byte {
	if q.longQueryCache != nil {
		return []byte(*q.longQueryCache)
	}
	var query strings.Builder

	plist := q.serverEngine
	if len(q.plugins) > 0 && q.listPlugins {
		plist += ":"
		for _, p := range q.plugins {
			plist += " " + sanitize(p.GetName()) + " " + sanitize(p.GetVersion()) + ";"
		}
		plist = plist[:len(plist)-1]
	}

	kv := [][2]string{
		{"splitnum", string([]byte{128})},
		{"hostname", q.serverName},
		{"gametype", q.gametype},
		{"game_id", GameID},
		{"version", q.version},
		{"server_engine", q.serverEngine},
		{"plugins", plist},
		{"map", q.worldName},
		{"numplayers", fmt.Sprint(q.numPlayers)},
		{"maxplayers", fmt.Sprint(q.maxPlayers)},
		{"whitelist", q.whitelist},
		{"hostip", q.ip},
		{"hostport", fmt.Sprint(q.port)},
	}
	for _, e := range kv {
		query.WriteString(e[0] + "\x00" + e[1] + "\x00")
	}
	for _, key := range q.extraKeys {
		query.WriteString(key + "\x00" + q.extraData[key] + "\x00")
	}
	query.WriteString("\x00\x01player_\x00\x00")
	for _, player := range q.players {
		query.WriteString(player + "\x00")
	}
	query.WriteString("\x00")

	s := query.String()
	q.longQueryCache = &s
	return []byte(s)
}

// GetShortQuery is a port of QueryInfo::getShortQuery.
func (q *QueryInfo) GetShortQuery() []byte {
	if q.shortQueryCache == nil {
		port := make([]byte, 2)
		binary.LittleEndian.PutUint16(port, uint16(q.port))
		s := q.serverName + "\x00" + q.gametype + "\x00" + q.worldName + "\x00" + fmt.Sprint(q.numPlayers) + "\x00" + fmt.Sprint(q.maxPlayers) + "\x00" + string(port) + q.ip + "\x00"
		q.shortQueryCache = &s
	}
	return []byte(*q.shortQueryCache)
}
