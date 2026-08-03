package main

import (
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/log"
)

// nextEntityRuntimeID hands out a unique EntityRuntimeID/EntityUniqueID to every connecting
// player - PocketMine-MP's own World tracks a similar per-world incrementing entity ID counter.
// Starts at 1 (0 is reserved/invalid in the protocol).
var nextEntityRuntimeID atomic.Uint64

func init() { nextEntityRuntimeID.Store(1) }

// session is this port's minimal stand-in for pocketmine\player\Player: just enough player state
// (identity, skin, position/rotation) to make one connected player visible to every other one.
// A real Player (health, inventory, gamemode transitions, ...) is a separate, much larger
// undertaking this doesn't attempt yet.
type session struct {
	conn            *minecraft.Conn
	name            string
	uuid            uuid.UUID
	entityRuntimeID uint64
	entityUniqueID  int64
	skin            protocol.Skin

	mu         sync.Mutex
	position   mgl32.Vec3
	pitch, yaw float32
	headYaw    float32
}

func newSession(conn *minecraft.Conn, spawn mgl32.Vec3) (*session, error) {
	id := conn.IdentityData()

	playerUUID, err := uuid.Parse(id.Identity)
	if err != nil {
		// Offline/unauthenticated connections (this port always runs with AuthenticationDisabled)
		// don't necessarily send a real XBL identity UUID - a locally generated one is fine, since
		// nothing here persists player identity across sessions yet.
		playerUUID = uuid.New()
	}

	skin, err := buildSkin(conn.ClientData())
	if err != nil {
		return nil, fmt.Errorf("building skin for %s: %w", id.DisplayName, err)
	}

	runtimeID := nextEntityRuntimeID.Add(1) - 1
	return &session{
		conn:            conn,
		name:            id.DisplayName,
		uuid:            playerUUID,
		entityRuntimeID: runtimeID,
		entityUniqueID:  int64(runtimeID),
		skin:            skin,
		position:        spawn,
	}, nil
}

// Position/Rotation/SetPositionAndRotation are the only pieces of session state anything outside
// this file touches - guarded by a mutex since PlayerAuthInput handling (writer) and any future
// broadcast/tick code (reader) run from different connections' goroutines.
func (s *session) Position() mgl32.Vec3 { s.mu.Lock(); defer s.mu.Unlock(); return s.position }

func (s *session) Rotation() (pitch, yaw, headYaw float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pitch, s.yaw, s.headYaw
}

func (s *session) SetPositionAndRotation(pos mgl32.Vec3, pitch, yaw, headYaw float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.position, s.pitch, s.yaw, s.headYaw = pos, pitch, yaw, headYaw
}

// addPlayerPacket is a port of the AddPlayer packet real PocketMine-MP sends to make one player's
// entity visible to another (Player::spawnTo, in spirit).
func (s *session) addPlayerPacket() *packet.AddPlayer {
	pos := s.Position()
	pitch, yaw, headYaw := s.Rotation()
	return &packet.AddPlayer{
		UUID:            s.uuid,
		Username:        s.name,
		EntityRuntimeID: s.entityRuntimeID,
		Position:        pos,
		Pitch:           pitch,
		Yaw:             yaw,
		HeadYaw:         headYaw,
		GameType:        packet.GameTypeSurvival,
		AbilityData: protocol.AbilityData{
			EntityUniqueID:     s.entityUniqueID,
			PlayerPermissions:  packet.PermissionLevelMember,
			CommandPermissions: protocol.CommandPermissionLevelAny,
		},
	}
}

// playerListEntry is a port of the PlayerListEntry real PocketMine-MP builds from a player's
// login/skin data for the PlayerList packet - required before AddPlayer for the player to actually
// render for anyone else (see AddPlayer's own doc comment in gophertunnel).
func (s *session) playerListEntry() protocol.PlayerListEntry {
	return protocol.PlayerListEntry{
		UUID:           s.uuid,
		EntityUniqueID: s.entityUniqueID,
		Username:       s.name,
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

// registry tracks every currently-connected session so a newly joined player can be shown every
// existing one (and vice versa), and so movement/disconnects can be relayed to everyone else - the
// minimal multiplayer-visibility slice of PocketMine-MP's World player tracking
// (World::addPlayer/removePlayer plus the network broadcast side of Player::spawnTo/despawnFrom).
type registry struct {
	mu       sync.Mutex
	sessions map[uint64]*session
	logger   log.Logger
}

func newRegistry(logger log.Logger) *registry {
	return &registry{sessions: map[uint64]*session{}, logger: logger}
}

// Count returns the number of currently-connected sessions.
func (r *registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sessions)
}

// Join is a port of the network-visible half of World::addPlayer: shows every already-connected
// player to the new session, the new session to every already-connected player, and finally
// registers it so future joins/moves/leaves reach it too.
func (r *registry) Join(s *session) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, other := range r.sessions {
		r.sendPlayerTo(other, s)
		r.sendPlayerTo(s, other)
	}
	r.sessions[s.entityRuntimeID] = s
}

// sendPlayerTo sends the PlayerList entry + AddPlayer needed for target to see subject.
func (r *registry) sendPlayerTo(target, subject *session) {
	_ = target.conn.WritePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionAdd,
		Entries:    []protocol.PlayerListEntry{subject.playerListEntry()},
	})
	_ = target.conn.WritePacket(subject.addPlayerPacket())
}

// Leave is a port of the network-visible half of World::removePlayer: tells every other connected
// player this session is gone.
func (r *registry) Leave(s *session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, s.entityRuntimeID)

	for _, other := range r.sessions {
		_ = other.conn.WritePacket(&packet.RemoveActor{EntityUniqueID: s.entityUniqueID})
		_ = other.conn.WritePacket(&packet.PlayerList{
			ActionType: packet.PlayerListActionRemove,
			Entries:    []protocol.PlayerListEntry{{UUID: s.uuid}},
		})
	}
}

// BroadcastMove relays s's latest position/rotation (see PlayerAuthInput handling in main.go) to
// every other connected player, so they see s move - PocketMine-MP's equivalent of
// Player::broadcastMovement / the MovePlayer packets Human::sendMovement fans out.
func (r *registry) BroadcastMove(s *session) {
	pos := s.Position()
	pitch, yaw, headYaw := s.Rotation()
	pk := &packet.MovePlayer{
		EntityRuntimeID: s.entityRuntimeID,
		Position:        pos,
		Pitch:           pitch,
		Yaw:             yaw,
		HeadYaw:         headYaw,
		Mode:            packet.MoveModeNormal,
		OnGround:        true,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for id, other := range r.sessions {
		if id == s.entityRuntimeID {
			continue
		}
		_ = other.conn.WritePacket(pk)
	}
}
