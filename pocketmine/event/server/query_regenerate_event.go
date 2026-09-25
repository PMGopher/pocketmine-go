package server

// QueryInfo is the surface of pocketmine\network\query\QueryInfo (*query.QueryInfo) that
// QueryRegenerateEvent handlers can use without importing the query package (which imports
// network, which fires this package's events). Assert to *query.QueryInfo for the plugin list.
type QueryInfo interface {
	GetServerName() string
	SetServerName(serverName string)
	CanListPlugins() bool
	SetListPlugins(value bool)
	GetPlayerList() []string
	SetPlayerList(players []string)
	GetPlayerCount() int
	SetPlayerCount(count int)
	GetMaxPlayerCount() int
	SetMaxPlayerCount(count int)
	GetWorld() string
	SetWorld(world string)
	GetExtraData() map[string]string
	SetExtraData(keys []string, extraData map[string]string)
	GetLongQuery() []byte
	GetShortQuery() []byte
}

// QueryRegenerateEvent is a port of pocketmine\event\server\QueryRegenerateEvent: called every
// second when the query information is regenerated; handlers may change it.
type QueryRegenerateEvent struct {
	queryInfo QueryInfo
}

func NewQueryRegenerateEvent(queryInfo QueryInfo) *QueryRegenerateEvent {
	return &QueryRegenerateEvent{queryInfo: queryInfo}
}

func (e *QueryRegenerateEvent) GetQueryInfo() QueryInfo { return e.queryInfo }
