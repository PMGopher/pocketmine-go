package main

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/log"
	pmmath "pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
)

// serverMu serialises everything that touches the world and its entities: the world tick
// (runTickLoop) and every connection's packet handling. PocketMine-MP runs all of this on one
// thread; this port's World, entities and players aren't safe for concurrent use either, so each
// goroutine takes this lock for the duration of its work instead.
var serverMu sync.Mutex

// session is this port's connection-level wrapper around a real player.Player - a stand-in for
// the parts of PocketMine-MP's NetworkSession that cmd/pocketmine-go needs (the player's identity
// and skin for the player list).
type session struct {
	conn   *minecraft.Conn
	player *player.Player

	name string
	uuid uuid.UUID
	skin protocol.Skin
}

// newSession builds the session and its player.Player (which registers itself in w as an entity,
// like PHP's Player constructor). spawn is the player's feet position. An invalid skin is an error
// (PHP disconnects with disconnectionScreen.invalidSkin).
func newSession(conn *minecraft.Conn, w *world.World, spawn pmmath.Vector3) (*session, error) {
	id := conn.IdentityData()

	playerUUID, err := uuid.Parse(id.Identity)
	if err != nil {
		// Offline/unauthenticated connections (this port always runs with AuthenticationDisabled)
		// don't necessarily send a real XBL identity UUID - a locally generated one is fine, since
		// nothing here persists player identity across sessions yet.
		playerUUID = uuid.New()
	}

	networkSkin, err := buildSkin(conn.ClientData())
	if err != nil {
		return nil, fmt.Errorf("building skin for %s: %w", id.DisplayName, err)
	}
	skin, err := entity.SkinFromNetwork(networkSkin)
	if err != nil {
		return nil, fmt.Errorf("invalid skin for %s: %w", id.DisplayName, err)
	}

	plr := player.NewPlayer(id.DisplayName, playerUUID, id.XUID, w, spawn, player.GameModeSurvival, skin)
	plr.SetPacketSender(func(pk packet.Packet) { _ = conn.WritePacket(pk) })

	return &session{
		conn:   conn,
		player: plr,
		name:   id.DisplayName,
		uuid:   playerUUID,
		skin:   networkSkin,
	}, nil
}

// playerListEntry is the PlayerListEntry PocketMine-MP's Server sends every player for every other
// online player (Server::addOnlinePlayer / sendFullPlayerListData) - required before AddPlayer for
// a player to render (see AddPlayer's own doc comment in gophertunnel).
func (s *session) playerListEntry() protocol.PlayerListEntry {
	return protocol.PlayerListEntry{
		UUID:           s.uuid,
		EntityUniqueID: int64(s.player.GetID()),
		Username:       s.name,
		XUID:           s.player.GetXuid(),
		Skin:           s.skin,
	}
}

// buildSkin adapts login.ClientData's base64-encoded skin fields (as sent by the client in its
// login request) into the protocol.Skin shape AddPlayer/PlayerList need - a direct field-by-field
// format conversion (base64 decode, JSON-shape copy), not skin validation/processing logic, the
// same category of "protocol adaptation" as this port's other gophertunnel-adjacent conversions
// (see block_state_dictionary.go's NBT handling). Animated skins aren't converted (Animations is
// left empty) - a static skin is enough for a player to be visible at all, which is what this is
// for; nothing here plays animated skin frames yet regardless.
func buildSkin(cd login.ClientData) (protocol.Skin, error) {
	skinData, err := base64.StdEncoding.DecodeString(cd.SkinData)
	if err != nil {
		return protocol.Skin{}, fmt.Errorf("decoding SkinData: %w", err)
	}
	capeData, err := base64.StdEncoding.DecodeString(cd.CapeData)
	if err != nil {
		return protocol.Skin{}, fmt.Errorf("decoding CapeData: %w", err)
	}
	geometry, err := base64.StdEncoding.DecodeString(cd.SkinGeometry)
	if err != nil {
		return protocol.Skin{}, fmt.Errorf("decoding SkinGeometry: %w", err)
	}
	resourcePatch, err := base64.StdEncoding.DecodeString(cd.SkinResourcePatch)
	if err != nil {
		return protocol.Skin{}, fmt.Errorf("decoding SkinResourcePatch: %w", err)
	}

	pieces := make([]protocol.PersonaPiece, len(cd.PersonaPieces))
	for i, p := range cd.PersonaPieces {
		pieces[i] = protocol.PersonaPiece{
			PieceID:   p.PieceID,
			PieceType: p.PieceType,
			PackID:    p.PackID,
			Default:   p.Default,
			ProductID: p.ProductID,
		}
	}
	tints := make([]protocol.PersonaPieceTintColour, len(cd.PieceTintColours))
	for i, t := range cd.PieceTintColours {
		tints[i] = protocol.PersonaPieceTintColour{PieceType: t.PieceType, Colours: t.Colours[:]}
	}

	return protocol.Skin{
		SkinID:                   cd.SkinID,
		PlayFabID:                cd.PlayFabID,
		SkinResourcePatch:        resourcePatch,
		SkinImageWidth:           uint32(cd.SkinImageWidth),
		SkinImageHeight:          uint32(cd.SkinImageHeight),
		SkinData:                 skinData,
		CapeImageWidth:           uint32(cd.CapeImageWidth),
		CapeImageHeight:          uint32(cd.CapeImageHeight),
		CapeData:                 capeData,
		SkinGeometry:             geometry,
		PremiumSkin:              cd.PremiumSkin,
		PersonaSkin:              cd.PersonaSkin,
		PersonaCapeOnClassicSkin: cd.CapeOnClassicSkin,
		CapeID:                   cd.CapeID,
		SkinColour:               cd.SkinColour,
		ArmSize:                  cd.ArmSize,
		PersonaPieces:            pieces,
		PieceTintColours:         tints,
		Trusted:                  cd.TrustedSkin,
	}, nil
}

// registry tracks every currently-connected session - the online-player list half of
// PocketMine-MP's Server (addOnlinePlayer/removeOnlinePlayer and the PlayerList packets they send).
// Making players visible to each other is the entity system's job (Player.SetSpawned ->
// Entity.SpawnToAll / spawnEntitiesOnChunk), exactly as in PHP.
type registry struct {
	sessions map[int]*session
	logger   log.Logger
}

func newRegistry(logger log.Logger) *registry {
	return &registry{sessions: map[int]*session{}, logger: logger}
}

// Count returns the number of currently-connected sessions.
func (r *registry) Count() int { return len(r.sessions) }

// Join is a port of Server::addOnlinePlayer + sendFullPlayerListData: the new player is added to
// everyone's player list, and receives the whole list.
func (r *registry) Join(s *session) {
	entries := []protocol.PlayerListEntry{s.playerListEntry()}
	for _, other := range r.sessions {
		entries = append(entries, other.playerListEntry())
		_ = other.conn.WritePacket(&packet.PlayerList{
			ActionType: packet.PlayerListActionAdd,
			Entries:    []protocol.PlayerListEntry{s.playerListEntry()},
		})
	}
	_ = s.conn.WritePacket(&packet.PlayerList{ActionType: packet.PlayerListActionAdd, Entries: entries})
	r.sessions[s.player.GetID()] = s
}

// Leave is a port of Server::removeOnlinePlayer + the player's close on disconnect: the player
// entity is closed (despawning it from everyone and removing it from its world) and removed from
// everyone's player list.
func (r *registry) Leave(s *session) {
	delete(r.sessions, s.player.GetID())
	s.player.Close()

	for _, other := range r.sessions {
		_ = other.conn.WritePacket(&packet.PlayerList{
			ActionType: packet.PlayerListActionRemove,
			Entries:    []protocol.PlayerListEntry{{UUID: s.uuid}},
		})
	}
}
