package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	respawnAnchorMinCharges = 0
	respawnAnchorMaxCharges = 4
)

// itemTypeIDsGlowstone mirrors ItemTypeIds::fromBlockTypeId(BlockTypeIds::GLOWSTONE), i.e.
// -GLOWSTONE (negative item type IDs are treated as block IDs in the PHP original) - not imported
// from a real item package since it doesn't exist yet, same reasoning as itemTypeIDsHoneycomb.
const itemTypeIDsGlowstone = -GLOWSTONE

// RespawnAnchor is a port of pocketmine\block\RespawnAnchor.
type RespawnAnchor struct {
	Opaque

	Charges int
}

func NewRespawnAnchor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RespawnAnchor {
	r := &RespawnAnchor{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, Charges: respawnAnchorMinCharges}
	r.Init(r)
	return r
}

func (r *RespawnAnchor) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RespawnAnchor) DescribeBlockOnlyState(w runtime.DataDescriber) {
	charges := r.Charges
	w.BoundedIntAuto(respawnAnchorMinCharges, respawnAnchorMaxCharges, &charges)
	r.Charges = charges
}

func (r *RespawnAnchor) GetCharges() int { return r.Charges }

func (r *RespawnAnchor) SetCharges(charges int) {
	if charges < respawnAnchorMinCharges || charges > respawnAnchorMaxCharges {
		panic("Charges must be between 0 and 4")
	}
	r.Charges = charges
}

func (r *RespawnAnchor) GetLightLevel() int {
	if r.Charges > 0 {
		return r.Charges*4 - 1
	}
	return 0
}

// ExplodeFunc is `new Explosion($source, $radius, $what)` with setFireChance, explodeA (only when
// blockBreaking) and explodeB. This package can't import world, which sets it in its init().
var ExplodeFunc func(source Position, radius float64, what Behavior, fireChance float64, blockBreaking bool)

// Player spawn hooks, set by the player package (the block Player interface can't name
// player.Position): Player::getSpawn and Player::setSpawn.
var (
	PlayerSpawnFunc    func(p Player) (Position, bool)
	SetPlayerSpawnFunc func(p Player, pos Position)
)

// OnInteract is a port of RespawnAnchor::onInteract.
func (r *RespawnAnchor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := r.position.GetWorld()
	if err != nil {
		return false
	}
	if item.GetTypeId() == itemTypeIDsGlowstone && r.Charges < respawnAnchorMaxCharges {
		r.Charges++
		if err := world.SetBlock(r.position, r.self); err != nil {
			panic(err)
		}
		world.AddSound(r.position.AsVector3(), sound.RespawnAnchorChargeSound{})
		item.Pop()
		return true
	}

	if r.Charges > respawnAnchorMinCharges {
		eventPlayer, ok := player.(playerevent.Player)
		if player == nil || !ok {
			return false
		}
		ev := playerevent.NewPlayerRespawnAnchorUseEvent(eventPlayer, r.self, playerevent.RespawnAnchorActionExplode)
		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
		switch ev.GetAction() {
		case playerevent.RespawnAnchorActionExplode:
			r.explode(player)
			return true
		case playerevent.RespawnAnchorActionSetSpawn:
			if PlayerSpawnFunc != nil {
				if spawn, ok := PlayerSpawnFunc(player); ok && spawn.world == r.position.world && spawn.Vector3.Equals(r.position.Vector3) {
					return true
				}
			}
			if SetPlayerSpawnFunc != nil {
				SetPlayerSpawnFunc(player, r.position)
			}
			world.AddSound(r.position.AsVector3(), sound.RespawnAnchorSetSpawnSound{})
			sendPlayerMessage(player, lang.KnownTranslationFactory.TileRespawnAnchorRespawnSet().Prefix(utils.Gray))
			return true
		}
	}
	return false
}

// explode is a port of RespawnAnchor::explode.
func (r *RespawnAnchor) explode(player Player) {
	var causingPlayer blockevent.Player
	if player != nil {
		causingPlayer = player
	}
	ev := blockevent.NewBlockPreExplodeEvent(r.self, 5, causingPlayer, 0)
	ev.SetIncendiary(true)
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}

	world, err := r.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(r.position, VanillaAir())

	if ExplodeFunc != nil {
		ExplodeFunc(NewPosition(r.position.X+0.5, r.position.Y+0.5, r.position.Z+0.5, world), ev.GetRadius(), r.self, ev.GetFireChance(), ev.IsBlockBreaking())
	}
}
