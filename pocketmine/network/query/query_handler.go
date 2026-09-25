package query

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network"
)

// Query packet types, QueryHandler::HANDSHAKE/STATISTICS.
const (
	Handshake  = 9
	Statistics = 0
)

// QueryHandler is a port of pocketmine\network\query\QueryHandler: the raw packet handler
// answering query packets (0xFE 0xFD ...) arriving on the game port.
//
// The Query protocol is built on top of the existing Minecraft PE UDP network stack. Because the
// 0xFE packet does not exist in the MCPE protocol, we can identify Query packets and remove them
// from the packet queue.
type QueryHandler struct {
	server    Server
	logger    log.Logger
	lastToken []byte
	token     []byte
}

func NewQueryHandler(server Server, logger log.Logger) *QueryHandler {
	h := &QueryHandler{server: server, logger: log.NewPrefixedLogger(logger, "Query Handler")}
	h.token = generateToken()
	h.lastToken = h.token
	return h
}

// Matches is QueryHandler::getPattern ('/^\xfe\xfd.+$/s').
func (h *QueryHandler) Matches(packet []byte) bool {
	return len(packet) > 2 && packet[0] == 0xfe && packet[1] == 0xfd
}

func generateToken() []byte {
	token := make([]byte, 16)
	_, _ = rand.Read(token)
	return token
}

// RegenerateToken is a port of QueryHandler::regenerateToken.
func (h *QueryHandler) RegenerateToken() {
	h.lastToken = h.token
	h.token = generateToken()
}

// GetTokenString is a port of QueryHandler::getTokenString.
func GetTokenString(token []byte, salt string) int32 {
	sum := sha512.Sum512(append([]byte(salt+":"), token...))
	return int32(binary.BigEndian.Uint32(sum[7:11]))
}

// Handle is a port of QueryHandler::handle.
func (h *QueryHandler) Handle(iface network.AdvancedNetworkInterface, address string, port int, packet []byte) (bool, error) {
	r := bytes.NewReader(packet)
	header := make([]byte, 2)
	if _, err := r.Read(header); err != nil || header[0] != 0xfe || header[1] != 0xfd {
		return false, nil
	}
	var packetType uint8
	var sessionID uint32
	if binary.Read(r, binary.BigEndian, &packetType) != nil || binary.Read(r, binary.BigEndian, &sessionID) != nil {
		h.logger.Debug(fmt.Sprintf("Bad packet from %s %d: not enough bytes", address, port))
		return false, nil
	}
	switch packetType {
	case Handshake:
		var w bytes.Buffer
		w.WriteByte(Handshake)
		_ = binary.Write(&w, binary.BigEndian, sessionID)
		w.WriteString(fmt.Sprint(GetTokenString(h.token, address)) + "\x00")
		iface.SendRawPacket(address, port, w.Bytes())
		return true, nil
	case Statistics:
		var token int32
		if binary.Read(r, binary.BigEndian, &token) != nil {
			h.logger.Debug(fmt.Sprintf("Bad packet from %s %d: not enough bytes", address, port))
			return false, nil
		}
		t1, t2 := GetTokenString(h.token, address), GetTokenString(h.lastToken, address)
		if token != t1 && token != t2 {
			h.logger.Debug(fmt.Sprintf("Bad token %d from %s %d, expected %d or %d", token, address, port, t1, t2))
			return true, nil
		}
		var w bytes.Buffer
		w.WriteByte(Statistics)
		_ = binary.Write(&w, binary.BigEndian, sessionID)
		info := h.server.GetQueryInformation()
		if r.Len() == 4 { //TODO: check this! according to the spec, this should always be here and always be FF FF FF 01
			w.Write(info.GetLongQuery())
		} else {
			w.Write(info.GetShortQuery())
		}
		iface.SendRawPacket(address, port, w.Bytes())
		return true, nil
	}
	return false, nil
}
