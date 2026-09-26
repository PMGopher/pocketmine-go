package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	stdmath "math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory/transaction"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
)

// InGamePacketHandler limits, InGamePacketHandler::*.
const (
	maxFormResponseSize       = 10 * 1024 //10 KB
	maxFormResponseDepth      = 2         //modal/simple will be 1, custom forms 2 - they will never contain anything other than string|int|float|bool|null
	pageLengthSoftLimitChars  = 256
	pageLengthHardLimitBytes  = 1024 * 4 // WritableBookPage::PAGE_LENGTH_HARD_LIMIT_BYTES
	int16Max                  = 0x7fff   // Limits::INT16_MAX
	actionMagicSlotDropItem   = 0        // NetworkInventoryAction::ACTION_MAGIC_SLOT_DROP_ITEM
	noBlockRuntimeID          = 0        // ItemTranslator::NO_BLOCK_RUNTIME_ID
	predictedResultSuccess    = protocol.ClientPredictionSuccess
	interactActionMouseover   = packet.InteractActionMouseOverEntity
	interactActionOpenInvetry = packet.InteractActionOpenInventory
)

// errFilterNoisyPacket is FilterNoisyPacketException.
var errFilterNoisyPacket = network.ErrFilterNoisyPacket

// InGamePacketHandler is a port of pocketmine\network\mcpe\handler\InGamePacketHandler: handles
// the packets a spawned player sends.
//
// FilterNoisyPacketException (Animate, the right-click spam bug) only makes the packet be ignored:
// gophertunnel has already decoded it, so there's no raw buffer to filter repeats of.
type InGamePacketHandler struct {
	player           *player.Player
	session          *mcpe.NetworkSession
	inventoryManager *mcpe.InventoryManager

	lastRightClickData *protocol.UseItemTransactionData
	lastRightClickTime time.Time

	lastPlayerAuthInputPosition *mgl32.Vec3
	lastPlayerAuthInputYaw      float32
	lastPlayerAuthInputPitch    float32
	lastPlayerAuthInputFlags    *protocol.InputFlags

	lastBlockAttacked *protocol.BlockPos

	// forceMoveSync is InGamePacketHandler::$forceMoveSync: set after a teleport until the
	// client's movement catches up.
	forceMoveSync bool

	lastRequestedFullSkinID *string
}

func NewInGamePacketHandler(p *player.Player, session *mcpe.NetworkSession, inventoryManager *mcpe.InventoryManager) *InGamePacketHandler {
	return &InGamePacketHandler{
		player:                   p,
		session:                  session,
		inventoryManager:         inventoryManager,
		lastPlayerAuthInputYaw:   float32(stdmath.NaN()),
		lastPlayerAuthInputPitch: float32(stdmath.NaN()),
	}
}

// SetUp is PacketHandler::setUp (empty for InGamePacketHandler).
func (h *InGamePacketHandler) SetUp() {}

// SetForceMoveSync sets $forceMoveSync (NetworkSession::syncMovement).
func (h *InGamePacketHandler) SetForceMoveSync(v bool) { h.forceMoveSync = v }

// CanHandle reports the packets InGamePacketHandler handles (PacketHandlerInspector).
func (h *InGamePacketHandler) CanHandle(pk packet.Packet) bool {
	switch pk.(type) {
	case *packet.Text, *packet.PlayerAuthInput, *packet.InventoryTransaction, *packet.ItemStackRequest,
		*packet.MobEquipment, *packet.Interact, *packet.BlockPickRequest, *packet.ActorPickRequest,
		*packet.PlayerAction, *packet.Animate, *packet.ContainerClose, *packet.BlockActorData,
		*packet.SetPlayerGameType, *packet.RequestChunkRadius, *packet.CommandRequest, *packet.PlayerSkin,
		*packet.BookEdit, *packet.ModalFormResponse, *packet.LecternUpdate, *packet.Emote,
		*packet.SubChunkRequest, *packet.ClientCacheBlobStatus, *packet.RequestAbility:
		return true
	}
	return false
}

// HandleDataPacket dispatches pk to the handleX method for its type.
func (h *InGamePacketHandler) HandleDataPacket(pk packet.Packet) (bool, error) {
	switch pk := pk.(type) {
	case *packet.Text:
		return h.handleText(pk), nil
	case *packet.PlayerAuthInput:
		return h.handlePlayerAuthInput(pk)
	case *packet.InventoryTransaction:
		return h.handleInventoryTransaction(pk)
	case *packet.ItemStackRequest:
		return h.handleItemStackRequest(pk)
	case *packet.MobEquipment:
		return h.handleMobEquipment(pk), nil
	case *packet.Interact:
		return h.handleInteract(pk), nil
	case *packet.BlockPickRequest:
		return h.player.PickBlock(math.NewVector3(float64(pk.Position[0]), float64(pk.Position[1]), float64(pk.Position[2])), pk.AddBlockNBT), nil
	case *packet.ActorPickRequest:
		return h.player.PickEntity(int(pk.EntityUniqueID)), nil
	case *packet.PlayerAction:
		return h.handlePlayerActionFromData(pk.ActionType, pk.BlockPosition, pk.BlockFace)
	case *packet.Animate:
		//this spams harder than a firehose on left click if "Improved Input Response" is enabled, and we don't even
		//use it anyway :<
		return true, nil
	case *packet.ContainerClose:
		h.inventoryManager.OnClientRemoveWindow(int(pk.WindowID))
		return true, nil
	case *packet.BlockActorData:
		return h.handleBlockActorData(pk)
	case *packet.SetPlayerGameType:
		return h.handleSetPlayerGameType(pk), nil
	case *packet.RequestChunkRadius:
		h.player.SetViewDistance(int(pk.ChunkRadius))
		return true, nil
	case *packet.CommandRequest:
		if strings.HasPrefix(pk.CommandLine, "/") {
			h.player.Chat(pk.CommandLine)
			return true, nil
		}
		return false, nil
	case *packet.PlayerSkin:
		return h.handlePlayerSkin(pk)
	case *packet.BookEdit:
		return h.handleBookEdit(pk)
	case *packet.ModalFormResponse:
		return h.handleModalFormResponse(pk)
	case *packet.LecternUpdate:
		return h.handleLecternUpdate(pk), nil
	case *packet.Emote:
		h.player.Emote(pk.EmoteID)
		return true, nil

	// No InGamePacketHandler counterparts: PocketMine-MP 5.44 sends full chunks, this port sends
	// them in sub-chunk request mode and through the client blob cache (see mcpe.ClientBlobCache).
	case *packet.SubChunkRequest:
		h.session.SendDataPacket(mcpe.HandleSubChunkRequest(h.player.GetWorld(), pk, h.session.GetBlobCache()))
		return true, nil
	case *packet.ClientCacheBlobStatus:
		if cache := h.session.GetBlobCache(); cache != nil {
			if resp := cache.HandleBlobStatus(pk); resp != nil {
				h.session.SendDataPacket(resp)
			}
			return true, nil
		}
		return false, nil
	case *packet.RequestAbility:
		return h.handleRequestAbility(pk), nil
	}
	return false, nil
}

// handleText is a port of InGamePacketHandler::handleText.
func (h *InGamePacketHandler) handleText(pk *packet.Text) bool {
	if pk.TextType == packet.TextTypeChat {
		return h.player.Chat(pk.Message)
	}
	return false
}

// inputFlagsEqual is PHP's BitSet::equals for gophertunnel's InputFlags.
func inputFlagsEqual(a, b protocol.InputFlags) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if a.Load(i) != b.Load(i) {
			return false
		}
	}
	return true
}

// resolveOnOffInputFlags is a port of InGamePacketHandler::resolveOnOffInputFlags: nil when
// neither flag was set, or both were set.
func resolveOnOffInputFlags(inputFlags protocol.InputFlags, startFlag, stopFlag int) *bool {
	enabled := inputFlags.Load(startFlag)
	disabled := inputFlags.Load(stopFlag)
	if enabled != disabled {
		return &enabled
	}
	return nil
}

func vec(v mgl32.Vec3) math.Vector3 {
	return math.NewVector3(float64(v[0]), float64(v[1]), float64(v[2]))
}

func blockVec(p protocol.BlockPos) math.Vector3 {
	return math.NewVector3(float64(p[0]), float64(p[1]), float64(p[2]))
}

// handlePlayerAuthInput is a port of InGamePacketHandler::handlePlayerAuthInput.
func (h *InGamePacketHandler) handlePlayerAuthInput(pk *packet.PlayerAuthInput) (bool, error) {
	rawPos := pk.Position
	rawYaw, rawPitch := pk.Yaw, pk.Pitch
	for _, f := range []float32{rawPos[0], rawPos[1], rawPos[2], rawYaw, pk.HeadYaw, rawPitch} {
		if stdmath.IsInf(float64(f), 0) || stdmath.IsNaN(float64(f)) {
			h.session.GetLogger().Debug("Invalid movement received, contains NAN/INF components")
			return false, nil
		}
	}

	if rawYaw != h.lastPlayerAuthInputYaw || rawPitch != h.lastPlayerAuthInputPitch {
		h.lastPlayerAuthInputYaw, h.lastPlayerAuthInputPitch = rawYaw, rawPitch

		yaw := stdmath.Mod(float64(rawYaw), 360)
		pitch := stdmath.Mod(float64(rawPitch), 360)
		if yaw < 0 {
			yaw += 360
		}
		h.player.SetRotation(yaw, pitch)
	}

	hasMoved := h.lastPlayerAuthInputPosition == nil || *h.lastPlayerAuthInputPosition != rawPos
	newPos := vec(rawPos).Subtract(0, 1.62, 0).Round(4)

	if h.forceMoveSync && hasMoved {
		curPos := h.player.GetLocation().Vector3
		if newPos.DistanceSquared(curPos) > 1 { //Tolerate up to 1 block to avoid problems with client-sided physics when spawning in blocks
			h.session.GetLogger().Debug(fmt.Sprintf("Got outdated pre-teleport movement, received %v, expected %v", newPos, curPos))
			//Still getting movements from before teleport, ignore them
			return true, nil
		}
		// Once we get a movement within a reasonable distance, treat it as a teleport ACK and remove position lock
		h.forceMoveSync = false
	}

	inputFlags := pk.InputData
	if h.lastPlayerAuthInputFlags == nil || !inputFlagsEqual(*h.lastPlayerAuthInputFlags, inputFlags) {
		flags := inputFlags
		h.lastPlayerAuthInputFlags = &flags

		sneakPressed := inputFlags.Load(packet.InputFlagSneaking)

		sneaking := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSneaking, packet.InputFlagStopSneaking)
		sprinting := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSprinting, packet.InputFlagStopSprinting)
		swimming := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSwimming, packet.InputFlagStopSwimming)
		gliding := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartGliding, packet.InputFlagStopGliding)
		flying := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartFlying, packet.InputFlagStopFlying)

		sneak := h.player.IsSneaking()
		if sneaking != nil {
			sneak = *sneaking
		}
		// PHP evaluates every toggle (bitwise |), so every one of them runs.
		mismatch := !h.player.ToggleSneak(sneak, sneakPressed)
		if sprinting != nil && !h.player.ToggleSprint(*sprinting) {
			mismatch = true
		}
		if swimming != nil && !h.player.ToggleSwim(*swimming) {
			mismatch = true
		}
		if gliding != nil && !h.player.ToggleGlide(*gliding) {
			mismatch = true
		}
		if flying != nil && !h.player.ToggleFlight(*flying) {
			mismatch = true
		}
		if mismatch {
			h.player.SendData([]world.EntityViewer{h.player}, nil)
			// The client already applied a flight toggle it predicted; PHP resyncs the abilities
			// from Player::setFlying, which isn't reached when flight is refused.
			if flying != nil {
				h.session.SyncAbilities(h.player)
			}
		}

		if inputFlags.Load(packet.InputFlagStartJumping) {
			h.player.Jump()
		}
		if inputFlags.Load(packet.InputFlagMissedSwing) {
			h.player.MissSwing()
		}
	}

	if !h.forceMoveSync && hasMoved {
		pos := rawPos
		h.lastPlayerAuthInputPosition = &pos
		//TODO: this packet has WAYYYYY more useful information that we're not using
		h.player.HandleMovement(newPos)
	}

	packetHandled := true

	if useItemTransaction, ok := pk.ItemInteractionData.Value(); ok {
		if len(useItemTransaction.Actions) > 100 {
			return false, &network.PacketHandlingError{Message: "Too many actions in item use transaction"}
		}

		requestID := useItemTransaction.LegacyRequestID
		h.inventoryManager.SetCurrentItemStackRequestID(&requestID)
		if err := h.inventoryManager.AddRawPredictedSlotChanges(useItemTransaction.Actions); err != nil {
			return false, err
		}
		handled, err := h.handleUseItemTransaction(&useItemTransaction)
		if err != nil && !errors.Is(err, errFilterNoisyPacket) {
			return false, err
		}
		if !handled {
			packetHandled = false
			h.session.GetLogger().Debug(fmt.Sprintf("Unhandled transaction in PlayerAuthInputPacket (type %d)", useItemTransaction.ActionType))
		} else {
			h.inventoryManager.SyncMismatchedPredictedSlotChanges()
		}
		h.inventoryManager.SetCurrentItemStackRequestID(nil)
	}

	itemStackRequest, hasItemStackRequest := pk.ItemStackRequest.Value()
	var itemStackResponseBuilder *ItemStackResponseBuilder
	if hasItemStackRequest {
		var err error
		if itemStackResponseBuilder, err = h.handleSingleItemStackRequest(itemStackRequest); err != nil {
			return false, err
		}
	}

	//itemstack request or transaction may set predictions for the outcome of these actions, so these need to be
	//processed last
	if blockActions, ok := pk.BlockActions.Value(); ok {
		if len(blockActions) > 100 {
			return false, &network.PacketHandlingError{Message: "Too many block actions in PlayerAuthInputPacket"}
		}
		for k, blockAction := range blockActions {
			var actionHandled bool
			var err error
			if blockAction.Action == protocol.PlayerActionStopBreak {
				actionHandled, err = h.handlePlayerActionFromData(blockAction.Action, protocol.BlockPos{}, int32(math.Down))
			} else {
				actionHandled, err = h.handlePlayerActionFromData(blockAction.Action, blockAction.BlockPos, blockAction.Face)
			}
			if err != nil {
				return false, err
			}
			if !actionHandled {
				packetHandled = false
				h.session.GetLogger().Debug(fmt.Sprintf("Unhandled player block action at offset %d in PlayerAuthInputPacket", k))
			}
		}
	}

	if hasItemStackRequest {
		response := protocol.ItemStackResponse{Status: protocol.ItemStackResponseStatusError, RequestID: itemStackRequest.RequestID}
		if itemStackResponseBuilder != nil {
			response = itemStackResponseBuilder.Build()
		}
		h.session.SendDataPacket(&packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{response}})
	}

	return packetHandled, nil
}

// handleInventoryTransaction is a port of InGamePacketHandler::handleInventoryTransaction.
func (h *InGamePacketHandler) handleInventoryTransaction(pk *packet.InventoryTransaction) (bool, error) {
	result := true

	if len(pk.Actions) > 50 {
		return false, &network.PacketHandlingError{Message: "Too many actions in inventory transaction"}
	}
	if len(pk.LegacySetItemSlots) > 10 {
		return false, &network.PacketHandlingError{Message: "Too many slot sync requests in inventory transaction"}
	}

	requestID := pk.LegacyRequestID
	h.inventoryManager.SetCurrentItemStackRequestID(&requestID)
	if err := h.inventoryManager.AddRawPredictedSlotChanges(pk.Actions); err != nil {
		return false, err
	}

	var err error
	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		result, err = h.handleNormalTransaction(pk.Actions, pk.LegacyRequestID)
	case *protocol.MismatchTransactionData:
		h.session.GetLogger().Debug("Mismatch transaction received")
		h.inventoryManager.RequestSyncAll()
		result = true
	case *protocol.UseItemTransactionData:
		data.Actions = pk.Actions
		result, err = h.handleUseItemTransaction(data)
		if errors.Is(err, errFilterNoisyPacket) {
			err = nil
		}
	case *protocol.UseItemOnEntityTransactionData:
		result = h.handleUseItemOnEntityTransaction(data)
	case *protocol.ReleaseItemTransactionData:
		result = h.handleReleaseItemTransaction(data)
	}
	if err != nil {
		return false, err
	}

	h.inventoryManager.SyncMismatchedPredictedSlotChanges()

	//requestChangedSlots asks the server to always send out the contents of the specified slots, even if they
	//haven't changed. Handling these is necessary to ensure the client inventory stays in sync if the server
	//rejects the transaction. The most common example of this is equipping armor by right-click, which doesn't send
	//a legacy prediction action for the destination armor slot.
	for _, containerInfo := range pk.LegacySetItemSlots {
		for _, netSlot := range containerInfo.Slots {
			windowID, slot, err := TranslateItemStackContainerID(containerInfo.ContainerID, h.inventoryManager.GetCurrentWindowID(), int(netSlot))
			if err != nil {
				return false, err
			}
			if inv, slot, ok := h.inventoryManager.LocateWindowAndSlot(windowID, slot); ok { //trigger the normal slot sync logic
				h.inventoryManager.OnSlotChange(inv, slot)
			}
		}
	}

	h.inventoryManager.SetCurrentItemStackRequestID(nil)
	return result, nil
}

// executableTransaction is an InventoryTransaction or one of its subclasses (CraftingTransaction,
// EnchantingTransaction).
type executableTransaction interface {
	GetActions() []transaction.InventoryAction
	Execute() error
}

// executeInventoryTransaction is a port of InGamePacketHandler::executeInventoryTransaction.
func (h *InGamePacketHandler) executeInventoryTransaction(tx executableTransaction, requestID int32) bool {
	h.player.SetUsingItem(false)

	h.inventoryManager.SetCurrentItemStackRequestID(&requestID)
	for _, action := range tx.GetActions() {
		if slotChange, ok := action.(*transaction.SlotChangeAction); ok {
			//TODO: ItemStackRequestExecutor can probably build these predictions with much lower overhead
			h.inventoryManager.AddPredictedSlotChange(slotChange.GetInventory(), slotChange.GetSlot(), slotChange.GetTargetItem())
		}
	}
	defer func() {
		h.inventoryManager.SyncMismatchedPredictedSlotChanges()
		h.inventoryManager.SetCurrentItemStackRequestID(nil)
	}()
	if err := tx.Execute(); err != nil {
		var cancelled *transaction.TransactionCancelledError
		if errors.As(err, &cancelled) {
			h.session.GetLogger().Debug(fmt.Sprintf("Inventory transaction %d cancelled by a plugin", requestID))
			return false
		}
		h.inventoryManager.RequestSyncAll()
		h.session.GetLogger().Debug(fmt.Sprintf("Invalid inventory transaction %d: %s", requestID, err.Error()))
		return false
	}
	return true
}

// handleNormalTransaction is a port of InGamePacketHandler::handleNormalTransaction.
func (h *InGamePacketHandler) handleNormalTransaction(actions []protocol.InventoryAction, itemStackRequestID int32) (bool, error) {
	//When the ItemStackRequest system is used, this transaction type is used for dropping items by pressing Q.
	//I don't know why they don't just use ItemStackRequest for that too, which already supports dropping items by
	//clicking them outside an open inventory menu, but for now it is what it is.
	//Fortunately, this means we can be much stricter about the validation criteria.

	actionCount := len(actions)
	if actionCount > 2 {
		if actionCount > 5 {
			return false, &network.PacketHandlingError{Message: fmt.Sprintf("Too many actions (%d) in normal inventory transaction", actionCount)}
		}

		//Due to a bug in the game, this transaction type is still sent when a player edits a book. We don't need
		//these transactions for editing books, since we have BookEditPacket, so we can just ignore them.
		h.session.GetLogger().Debug(fmt.Sprintf("Ignoring normal inventory transaction with %d actions (drop-item should have exactly 2 actions)", actionCount))
		return false, nil
	}

	sourceSlot := -1
	var clientItemStack *protocol.ItemStack
	droppedCount := -1

	for _, networkInventoryAction := range actions {
		windowID, hasWindow := networkInventoryAction.WindowID.Value()
		if networkInventoryAction.SourceType == protocol.InventoryActionSourceWorld && networkInventoryAction.InventorySlot == actionMagicSlotDropItem {
			droppedCount = int(networkInventoryAction.NewItem.Stack.Count)
			if droppedCount <= 0 {
				return false, &network.PacketHandlingError{Message: "Expected positive count for dropped item"}
			}
		} else if networkInventoryAction.SourceType == protocol.InventoryActionSourceContainer && hasWindow && windowID == mcpe.ContainerIDInventory {
			//mobile players can drop an item from a non-selected hotbar slot
			sourceSlot = int(networkInventoryAction.InventorySlot)
			stack := networkInventoryAction.OldItem.Stack
			clientItemStack = &stack
		} else {
			h.session.GetLogger().Debug(fmt.Sprintf("Unexpected inventory action type %d in drop item transaction", networkInventoryAction.SourceType))
			return false, nil
		}
	}
	if sourceSlot == -1 || clientItemStack == nil || droppedCount == -1 {
		h.session.GetLogger().Debug("Missing information in drop item transaction, need source slot, client item stack and dropped count")
		return false, nil
	}

	inv := h.player.GetInventory()
	if !inv.SlotExists(sourceSlot) {
		return false, nil //TODO: size desync??
	}

	sourceSlotItem := inv.GetItem(sourceSlot)
	if sourceSlotItem.GetCount() < droppedCount {
		return false, nil
	}
	serverItemStack := convert.CoreItemStackToNet(sourceSlotItem)
	//Sadly we don't have itemstack IDs here, so we have to compare the basic item properties to ensure that we're
	//dropping the item the client expects (inventory might be out of sync with the client).
	if serverItemStack.NetworkID != clientItemStack.NetworkID ||
		serverItemStack.MetadataValue != clientItemStack.MetadataValue ||
		serverItemStack.Count != clientItemStack.Count ||
		serverItemStack.BlockRuntimeID != clientItemStack.BlockRuntimeID {
		//Raw extraData may not match because of TAG_Compound key ordering differences, and decoding it to compare
		//is costly. Assume that we're in sync if id+meta+count+runtimeId match.
		//NB: Make sure $clientItemStack isn't used to create the dropped item, as that would allow the client
		//to change the item NBT since we're not validating it.
		return false, nil
	}

	//this modifies $sourceSlotItem
	droppedItem := sourceSlotItem.PopCount(droppedCount)

	builder := transaction.NewTransactionBuilder()
	builder.GetInventory(inv).SetItem(sourceSlot, sourceSlotItem)
	builder.AddAction(transaction.NewDropItemAction(droppedItem))

	tx, err := transaction.NewInventoryTransaction(h.player, builder.GenerateActions())
	if err != nil {
		h.inventoryManager.RequestSyncAll()
		h.session.GetLogger().Debug(fmt.Sprintf("Invalid inventory transaction %d: %s", itemStackRequestID, err.Error()))
		return false, nil
	}
	return h.executeInventoryTransaction(tx, itemStackRequestID), nil
}

// handleUseItemTransaction is a port of InGamePacketHandler::handleUseItemTransaction.
func (h *InGamePacketHandler) handleUseItemTransaction(data *protocol.UseItemTransactionData) (bool, error) {
	h.player.SelectHotbarSlot(int(data.HotBarSlot))

	switch data.ActionType {
	case protocol.UseItemActionClickBlock:
		//TODO: start hack for client spam bug
		clickPos := vec(data.ClickedPosition)
		last := h.lastRightClickData
		spamBug := last != nil &&
			time.Since(h.lastRightClickTime) < 100*time.Millisecond && //100ms
			last.BlockFace == data.BlockFace &&
			vec(last.Position).DistanceSquared(vec(data.Position)) < 0.00001 &&
			last.BlockPosition == data.BlockPosition &&
			vec(last.ClickedPosition).DistanceSquared(clickPos) < 0.00001 //signature spam bug has 0 distance, but allow some error
		//get rid of continued spam if the player clicks and holds right-click
		copied := *data
		h.lastRightClickData = &copied
		h.lastRightClickTime = time.Now()
		if spamBug {
			return true, errFilterNoisyPacket
		}
		//TODO: end hack for client spam bug

		if err := validateFacing(data.BlockFace); err != nil {
			return false, err
		}

		vBlockPos := blockVec(data.BlockPosition)
		h.player.InteractBlock(vBlockPos, math.Facing(data.BlockFace), clickPos)
		if data.ClientPrediction == predictedResultSuccess {
			//If the item has an associated blockstate ID, this means it will only place one block.
			//We can avoid syncing the adjacent blocks of the place position in this case, since that's only
			//necessary if there might be multiple blocks around the placement location affected.
			//Adjacents of the clicked block are still always synced, since it's too complicated to figure out
			//if the client might've predicted something in this case. However, since the clicked block is always
			//"behind" the placed block, this shouldn't affect bridging or fast placement.
			//This would be much easier if the client would just tell us which blocks it thinks changed...
			var syncAdjacentFace *math.Facing
			if data.HeldItem.Stack.BlockRuntimeID == noBlockRuntimeID {
				h.session.GetLogger().Debug("Placing held item might place multiple blocks client-side; doing full adjacent sync")
				face := math.Facing(data.BlockFace)
				syncAdjacentFace = &face
			}
			h.syncBlocksNearby(vBlockPos, syncAdjacentFace)
		}
		return true, nil
	case protocol.UseItemActionClickAir:
		if h.player.IsUsingItem() {
			if !h.player.ConsumeHeldItem() {
				hungerAttr := h.player.GetAttributeMap().Get(entity.AttributeHunger)
				if hungerAttr == nil {
					panic("hunger attribute should exist")
				}
				hungerAttr.MarkSynchronized(false)
			}
			//TODO: workaround goat horns getting stuck in the "using item" state
			//this timed-trigger behaviour is also used for other items apart from food
			//in the future we'll generalise this logic and add proper hooks for it
			h.player.SetUsingItem(false)
			return true, nil
		}
		h.player.UseHeldItem()
		return true, nil
	}
	return false, nil
}

// validateFacing is a port of InGamePacketHandler::validateFacing.
func validateFacing(facing int32) error {
	if facing < 0 || facing > 5 {
		return &network.PacketHandlingError{Message: fmt.Sprintf("Invalid facing value %d", facing)}
	}
	return nil
}

// syncBlocksNearby is a port of InGamePacketHandler::syncBlocksNearby: syncs blocks nearby to
// ensure that the client and server agree on the world's blocks after a block interaction.
func (h *InGamePacketHandler) syncBlocksNearby(blockPos math.Vector3, face *math.Facing) {
	if blockPos.DistanceSquared(h.player.GetLocation().Vector3) >= 10000 {
		return
	}
	var blocks []math.Vector3
	for _, side := range math.AllFacing {
		blocks = append(blocks, blockPos.GetSide(side, 1))
	}
	if face != nil {
		sidePos := blockPos.GetSide(*face, 1)
		//getAllSides() on each of these will include $blockPos and $sidePos because they are next to each other
		for _, side := range math.AllFacing {
			blocks = append(blocks, sidePos.GetSide(side, 1))
		}
	} else {
		blocks = append(blocks, blockPos)
	}
	for _, pk := range h.player.GetWorld().CreateBlockUpdatePackets(blocks) {
		h.session.SendDataPacket(pk)
	}
}

// handleUseItemOnEntityTransaction is a port of InGamePacketHandler::handleUseItemOnEntityTransaction.
func (h *InGamePacketHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData) bool {
	target, ok := h.player.GetWorld().GetEntity(int(data.TargetEntityRuntimeID))
	//TODO: HACK! We really shouldn't be keeping disconnected players (and generally flagged-for-despawn entities)
	//in the world's entity table, but changing that is too risky for a hotfix. This workaround will do for now.
	if !ok || target.IsFlaggedForDespawn() {
		return false
	}

	h.player.SelectHotbarSlot(int(data.HotBarSlot))

	switch data.ActionType {
	case protocol.UseItemOnEntityActionInteract:
		h.player.InteractEntity(target, vec(data.ClickedPosition))
		return true
	case protocol.UseItemOnEntityActionAttack:
		h.player.AttackEntity(target)
		return true
	}
	return false
}

// handleReleaseItemTransaction is a port of InGamePacketHandler::handleReleaseItemTransaction.
func (h *InGamePacketHandler) handleReleaseItemTransaction(data *protocol.ReleaseItemTransactionData) bool {
	h.player.SelectHotbarSlot(int(data.HotBarSlot))

	if data.ActionType == protocol.ReleaseItemActionRelease {
		h.player.ReleaseHeldItem()
		return true
	}
	return false
}

// handleSingleItemStackRequest is a port of InGamePacketHandler::handleSingleItemStackRequest.
func (h *InGamePacketHandler) handleSingleItemStackRequest(request protocol.ItemStackRequest) (*ItemStackResponseBuilder, error) {
	if len(request.Actions) > 60 {
		//recipe book auto crafting can affect all slots of the inventory when consuming inputs or producing outputs
		//this means there could be as many as 50 CraftingConsumeInput actions or Place (taking the result) actions
		//in a single request (there are certain ways items can be arranged which will result in the same stack
		//being taken from multiple times, but this is behaviour with a calculable limit)
		//this means there SHOULD be AT MOST 53 actions in a single request, but 60 is a nice round number.
		return nil, &network.PacketHandlingError{Message: "Too many actions in ItemStackRequest"}
	}
	executor := NewItemStackRequestExecutor(h.player, h.inventoryManager, request)
	result := false
	tx, err := executor.GenerateInventoryTransaction()
	if err != nil {
		var handlingErr *network.PacketHandlingError
		if errors.As(err, &handlingErr) {
			return nil, err
		}
		h.session.GetLogger().Debug(fmt.Sprintf("ItemStackRequest #%d failed: %s", request.RequestID, err.Error()))
		h.inventoryManager.RequestSyncAll()
	} else if tx != nil {
		result = h.executeInventoryTransaction(tx, request.RequestID)
	} else {
		result = true //predictions only, just send responses
	}

	if !result {
		return nil, nil
	}
	return executor.GetItemStackResponseBuilder(), nil
}

// handleItemStackRequest is a port of InGamePacketHandler::handleItemStackRequest.
func (h *InGamePacketHandler) handleItemStackRequest(pk *packet.ItemStackRequest) (bool, error) {
	if len(pk.Requests) > 80 {
		//TODO: we can probably lower this limit, but this will do for now
		return false, &network.PacketHandlingError{Message: "Too many requests in ItemStackRequestPacket"}
	}
	responses := make([]protocol.ItemStackResponse, 0, len(pk.Requests))
	for _, request := range pk.Requests {
		builder, err := h.handleSingleItemStackRequest(request)
		if err != nil {
			return false, err
		}
		if builder != nil {
			responses = append(responses, builder.Build())
		} else {
			responses = append(responses, protocol.ItemStackResponse{Status: protocol.ItemStackResponseStatusError, RequestID: request.RequestID})
		}
	}
	h.session.SendDataPacket(&packet.ItemStackResponse{Responses: responses})
	return true, nil
}

// handleMobEquipment is a port of InGamePacketHandler::handleMobEquipment.
func (h *InGamePacketHandler) handleMobEquipment(pk *packet.MobEquipment) bool {
	if pk.WindowID == mcpe.ContainerIDOffhand {
		return true //this happens when we put an item into the offhand
	}
	if pk.WindowID == mcpe.ContainerIDInventory {
		h.inventoryManager.OnClientSelectHotbarSlot(int(pk.HotBarSlot))
		if !h.player.SelectHotbarSlot(int(pk.HotBarSlot)) {
			h.inventoryManager.SyncSelectedHotbarSlot()
		}
		return true
	}
	return false
}

// handleInteract is a port of InGamePacketHandler::handleInteract.
func (h *InGamePacketHandler) handleInteract(pk *packet.Interact) bool {
	if pk.ActionType == interactActionMouseover {
		//TODO HACK: silence useless spam (MCPE 1.8)
		//due to some messy Mojang hacks, it sends this when changing the held item now, which causes us to think
		//the inventory was closed when it wasn't.
		//this is also sent whenever entity metadata updates, which can get really spammy.
		//TODO: implement handling for this where it matters
		return true
	}
	target, ok := h.player.GetWorld().GetEntity(int(pk.TargetEntityRuntimeID))
	if !ok {
		return false
	}
	if pk.ActionType == interactActionOpenInvetry && target.GetID() == h.player.GetID() {
		h.inventoryManager.OnClientOpenMainInventory()
		return true
	}
	return false //TODO
}

// handlePlayerActionFromData is a port of InGamePacketHandler::handlePlayerActionFromData.
func (h *InGamePacketHandler) handlePlayerActionFromData(action int32, blockPosition protocol.BlockPos, face int32) (bool, error) {
	pos := blockVec(blockPosition)

	switch action {
	case protocol.PlayerActionStartBreak, protocol.PlayerActionContinueDestroyBlock: //destroy the next block while holding down left click
		if err := validateFacing(face); err != nil {
			return false, err
		}
		if h.lastBlockAttacked != nil && *h.lastBlockAttacked == blockPosition {
			//the client will send CONTINUE_DESTROY_BLOCK for the currently targeted block directly before it
			//sends PREDICT_DESTROY_BLOCK, but also when it starts to break the block
			//this seems like a bug in the client and would cause spurious left-click events if we allowed it to
			//be delivered to the player
			h.session.GetLogger().Debug(fmt.Sprintf("Ignoring PlayerAction %d on %v because we were already destroying this block", action, pos))
			break
		}
		if !h.player.AttackBlock(pos, math.Facing(face)) {
			f := math.Facing(face)
			h.syncBlocksNearby(pos, &f)
		}
		attacked := blockPosition
		h.lastBlockAttacked = &attacked

	case protocol.PlayerActionAbortBreak, protocol.PlayerActionStopBreak:
		h.player.StopBreakBlock(pos)
		h.lastBlockAttacked = nil
	case protocol.PlayerActionStartSleeping:
		//unused
	case protocol.PlayerActionStopSleeping:
		h.player.StopSleep()
	case protocol.PlayerActionCrackBreak:
		if err := validateFacing(face); err != nil {
			return false, err
		}
		h.player.ContinueBreakBlock(pos, math.Facing(face))
		attacked := blockPosition
		h.lastBlockAttacked = &attacked
	case protocol.PlayerActionInteractWithBlock: //TODO: ignored (for now)
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		//in server auth block breaking, we get PREDICT_DESTROY_BLOCK anyway, so this action is redundant
	case protocol.PlayerActionPredictDestroyBlock:
		if err := validateFacing(face); err != nil {
			return false, err
		}
		if !h.player.BreakBlock(pos) {
			f := math.Facing(face)
			h.syncBlocksNearby(pos, &f)
		}
		h.lastBlockAttacked = nil
	case protocol.PlayerActionStartItemUseOn, protocol.PlayerActionStopItemUseOn:
		//TODO: this has no obvious use and seems only used for analytics in vanilla - ignore it
	default:
		h.session.GetLogger().Debug(fmt.Sprintf("Unhandled/unknown player action type %d", action))
		return false, nil
	}

	h.player.SetUsingItem(false)

	return true, nil
}

// updateSignText is a port of InGamePacketHandler::updateSignText.
func (h *InGamePacketHandler) updateSignText(tag map[string]any, tagName string, frontFace bool, sign signBlock, pos math.Vector3) (bool, error) {
	textTag, ok := tag[tagName].(map[string]any)
	if !ok {
		return false, &network.PacketHandlingError{Message: fmt.Sprintf("Invalid tag type %T for tag %q in sign update data", tag[tagName], tagName)}
	}
	textBlob, ok := textTag[tile.SignTagTextBlob].(string)
	if !ok {
		return false, &network.PacketHandlingError{Message: fmt.Sprintf("Invalid tag type %T for tag %q in sign update data", textTag[tile.SignTagTextBlob], tile.SignTagTextBlob)}
	}

	text := blockutils.SignTextFromBlob(textBlob, nil, false)

	oldText := sign.GetFaceText(frontFace)
	if text.GetLines() == oldText.GetLines() {
		return false, nil
	}

	updated, err := sign.UpdateFaceText(h.player, h.player.GetName(), frontFace, text)
	if err != nil {
		return false, network.WrapPacketHandlingError(err, "")
	}
	if !updated {
		for _, updatePacket := range h.player.GetWorld().CreateBlockUpdatePackets([]math.Vector3{pos}) {
			h.session.SendDataPacket(updatePacket)
		}
		return false, nil
	}
	return true, nil
}

// signBlock is block.BaseSign, which every sign block embeds (PHP's instanceof BaseSign).
type signBlock interface {
	GetFaceText(frontFace bool) blockutils.SignText
	UpdateFaceText(author block.Player, authorName string, frontFace bool, text blockutils.SignText) (bool, error)
}

// handleBlockActorData is a port of InGamePacketHandler::handleBlockActorData.
func (h *InGamePacketHandler) handleBlockActorData(pk *packet.BlockActorData) (bool, error) {
	pos := blockVec(pk.Position)
	if pos.DistanceSquared(h.player.GetLocation().Vector3) > 10000 {
		return false, nil
	}

	blk := h.player.GetWorld().GetBlock(pos)
	if sign, ok := blk.(signBlock); ok {
		updated, err := h.updateSignText(pk.NBTData, tile.SignTagFrontText, true, sign, pos)
		if err != nil {
			return false, err
		}
		if !updated {
			//only one side can be updated at a time
			if _, err := h.updateSignText(pk.NBTData, tile.SignTagBackText, false, sign, pos); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

// handleSetPlayerGameType is a port of InGamePacketHandler::handleSetPlayerGameType.
func (h *InGamePacketHandler) handleSetPlayerGameType(pk *packet.SetPlayerGameType) bool {
	gameMode, ok := convert.ProtocolGameModeToCore(pk.GameType)
	if !ok || player.GameMode(gameMode) != h.player.GetGamemode() {
		//Set this back to default. TODO: handle this properly
		h.session.SyncGameMode(h.player.GetGamemode(), true)
	}
	return true
}

// handlePlayerSkin is a port of InGamePacketHandler::handlePlayerSkin.
func (h *InGamePacketHandler) handlePlayerSkin(pk *packet.PlayerSkin) (bool, error) {
	fullSkinID := pk.Skin.FullID
	if h.lastRequestedFullSkinID != nil && fullSkinID == *h.lastRequestedFullSkinID {
		//TODO: HACK! In 1.19.60, the client sends its skin back to us if we sent it a skin different from the one
		//it's using. We need to prevent this from causing a feedback loop.
		h.session.GetLogger().Debug("Refused duplicate skin change request")
		return true, nil
	}
	h.lastRequestedFullSkinID = &fullSkinID

	h.session.GetLogger().Debug("Processing skin change request")
	skin, err := entity.SkinFromNetwork(pk.Skin)
	if err != nil {
		return false, network.WrapPacketHandlingError(err, "Invalid skin in PlayerSkinPacket")
	}
	return h.player.ChangeSkin(skin, pk.NewSkinName, pk.OldSkinName), nil
}

// checkBookText is a port of InGamePacketHandler::checkBookText.
func (h *InGamePacketHandler) checkBookText(s, fieldName string, softLimit, hardLimit int, cancel *bool) (string, error) {
	if len(s) > hardLimit {
		return "", &network.PacketHandlingError{Message: fmt.Sprintf("Book %s must be at most %d bytes, but have %d bytes", fieldName, hardLimit, len(s))}
	}

	result := utils.Clean(s, false)
	//strlen() is O(1), mb_strlen() is O(n)
	if len(result) > softLimit*4 || utf8.RuneCountInString(result) > softLimit {
		*cancel = true
		h.session.GetLogger().Debug(fmt.Sprintf("Cancelled book edit due to %s exceeded soft limit of %d chars", fieldName, softLimit))
	}
	return result, nil
}

// handleBookEdit is a port of InGamePacketHandler::handleBookEdit.
func (h *InGamePacketHandler) handleBookEdit(pk *packet.BookEdit) (bool, error) {
	inv := h.player.GetInventory()
	slot := int(pk.InventorySlot)
	if !inv.SlotExists(slot) {
		return false, nil
	}
	//TODO: break this up into book API things
	oldBook, ok := inv.GetItem(slot).(*item.WritableBook)
	if !ok {
		return false, nil
	}

	newBook := oldBook.Clone().(*item.WritableBook)
	var newItem item.Item = newBook
	var modifiedPages []int
	cancel := false
	pageNumber := int(pk.PageNumber)
	var action int
	switch pk.ActionType {
	case packet.BookActionReplacePage:
		text, err := h.checkBookText(pk.Text, "page text", pageLengthSoftLimitChars, pageLengthHardLimitBytes, &cancel)
		if err != nil {
			return false, err
		}
		if pageNumber < 0 {
			return false, &network.PacketHandlingError{Message: "Page number cannot be negative"}
		}
		newBook.SetPageText(pageNumber, text)
		modifiedPages = append(modifiedPages, pageNumber)
		action = playerevent.EditBookActionReplacePage
	case packet.BookActionAddPage:
		if !newBook.PageExists(pageNumber) {
			//this may only come before a page which already exists
			//TODO: the client can send insert-before actions on trailing client-side pages which cause odd behaviour on the server
			return false, nil
		}
		text, err := h.checkBookText(pk.Text, "page text", pageLengthSoftLimitChars, pageLengthHardLimitBytes, &cancel)
		if err != nil {
			return false, err
		}
		newBook.InsertPage(pageNumber, text)
		modifiedPages = append(modifiedPages, pageNumber)
		action = playerevent.EditBookActionAddPage
	case packet.BookActionDeletePage:
		if !newBook.PageExists(pageNumber) {
			return false, nil
		}
		newBook.DeletePage(pageNumber)
		modifiedPages = append(modifiedPages, pageNumber)
		action = playerevent.EditBookActionDeletePage
	case packet.BookActionSwapPages:
		secondary := int(pk.SecondaryPageNumber)
		if pageNumber < 0 || secondary < 0 {
			return false, &network.PacketHandlingError{Message: "Page numbers cannot be negative"}
		}
		if !newBook.PageExists(pageNumber) || !newBook.PageExists(secondary) {
			//the client will create pages on its own without telling us until it tries to switch them
			newBook.AddPage(max(pageNumber, secondary))
		}
		newBook.SwapPages(pageNumber, secondary)
		modifiedPages = []int{pageNumber, secondary}
		action = playerevent.EditBookActionSwapPages
	case packet.BookActionSign:
		title, err := h.checkBookText(pk.Title, "title", 16, int16Max, &cancel)
		if err != nil {
			return false, err
		}
		//this one doesn't have a limit in vanilla, so we have to improvise
		author, err := h.checkBookText(pk.Author, "author", 256, int16Max, &cancel)
		if err != nil {
			return false, err
		}

		written := item.VanillaWrittenBook().(*item.WrittenBook)
		written.SetPages(oldBook.GetPages())
		written.SetAuthor(author)
		written.SetTitle(title)
		written.SetGeneration(item.WrittenBookGenerationOriginal)
		newItem = written
		action = playerevent.EditBookActionSignBook
	default:
		return false, nil
	}

	/*
	 * Plugins may have created books with more than 50 pages; we allow plugins to do this, but not players.
	 * Don't allow the page count to grow past 50, but allow deleting, swapping or altering text of existing pages.
	 */
	oldPageCount := len(oldBook.GetPages())
	newPageCount := oldPageCount
	if pages, ok := newItem.(interface {
		GetPages() []item.WritableBookPage
	}); ok {
		newPageCount = len(pages.GetPages())
	}
	if newPageCount > oldPageCount && newPageCount > 50 {
		h.session.GetLogger().Debug(fmt.Sprintf("Cancelled book edit due to adding too many pages (new page count would be %d)", newPageCount))
		cancel = true
	}

	ev := playerevent.NewPlayerEditBookEvent(h.player, oldBook, newItem, action, modifiedPages)
	if cancel {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return true, nil
	}

	if newBook, ok := ev.GetNewBook().(item.Item); ok {
		h.player.GetInventory().SetItem(slot, newBook)
	}
	return true, nil
}

// handleModalFormResponse is a port of InGamePacketHandler::handleModalFormResponse.
func (h *InGamePacketHandler) handleModalFormResponse(pk *packet.ModalFormResponse) (bool, error) {
	formID := int(pk.FormID)
	if _, cancelled := pk.CancelReason.Value(); cancelled {
		//TODO: make APIs for this to allow plugins to use this information
		return h.player.OnFormSubmit(formID, nil), nil
	}
	formData, ok := pk.ResponseData.Value()
	if !ok {
		return false, &network.PacketHandlingError{Message: "Expected either formData or cancelReason to be set in ModalFormResponsePacket"}
	}
	if len(formData) > maxFormResponseSize {
		return false, &network.PacketHandlingError{Message: fmt.Sprintf("Form response data too large, refusing to decode (received%d bytes, max %d bytes)", len(formData), maxFormResponseSize)}
	}
	if !h.player.HasPendingForm(formID) {
		h.session.GetLogger().Debug(fmt.Sprintf("Got unexpected response for form %d", formID))
		return false, nil
	}
	var responseData any
	if err := json.Unmarshal(formData, &responseData); err != nil {
		return false, network.WrapPacketHandlingError(err, "Failed to decode form response data")
	}
	if jsonDepth(responseData) > maxFormResponseDepth {
		return false, &network.PacketHandlingError{Message: "Failed to decode form response data: Maximum stack depth exceeded"}
	}
	return h.player.OnFormSubmit(formID, responseData), nil
}

// jsonDepth is the nesting depth of a decoded JSON value, as json_decode's $depth counts it.
func jsonDepth(v any) int {
	switch v := v.(type) {
	case []any:
		d := 0
		for _, e := range v {
			d = max(d, jsonDepth(e))
		}
		return d + 1
	case map[string]any:
		d := 0
		for _, e := range v {
			d = max(d, jsonDepth(e))
		}
		return d + 1
	}
	return 0
}

// handleLecternUpdate is a port of InGamePacketHandler::handleLecternUpdate.
func (h *InGamePacketHandler) handleLecternUpdate(pk *packet.LecternUpdate) bool {
	pos := pk.Position
	chunkX, chunkZ := int(pos[0])>>4, int(pos[2])>>4
	w := h.player.GetWorld()
	// World::isChunkLocked: chunks are only locked by async population, which this port doesn't
	// have (generation is synchronous), so no chunk is ever locked.
	if !w.IsChunkLoaded(chunkX, chunkZ) {
		return false
	}

	lectern, ok := w.GetBlockAt(int(pos[0]), int(pos[1]), int(pos[2])).(*block.Lectern)
	if ok && h.player.CanInteract(blockVec(pos).Add(0.5, 0.5, 0.5), 15) {
		if !lectern.OnPageTurn(int(pk.Page)) {
			h.syncBlocksNearby(blockVec(pos), nil)
		}
		return true
	}
	return false
}

// handleRequestAbility has no InGamePacketHandler counterpart: PocketMine-MP 5.44 doesn't handle
// RequestAbility. The client predicts the requested ability (flying, when the swim-up gesture is
// read as a double jump) until told otherwise, so the abilities are resynced to refuse anything
// the player isn't allowed (see commit 697dfa7).
func (h *InGamePacketHandler) handleRequestAbility(pk *packet.RequestAbility) bool {
	if pk.Ability == protocol.AbilityFlying {
		if value, ok := pk.Value.(bool); ok && h.player.ToggleFlight(value) {
			return true
		}
	}
	h.session.SyncAbilities(h.player)
	return true
}
