package handler

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe"
)

// TranslateItemStackContainerID is a port of ItemStackContainerIdTranslator::translate: maps a
// container UI ID and slot to the window ID and slot InventoryManager knows it by.
func TranslateItemStackContainerID(containerInterfaceID byte, currentWindowID, slotID int) (int, int, error) {
	switch containerInterfaceID {
	case protocol.ContainerArmor:
		return mcpe.ContainerIDArmor, slotID, nil

	case protocol.ContainerHotBar, protocol.ContainerInventory, protocol.ContainerCombinedHotBarAndInventory:
		return mcpe.ContainerIDInventory, slotID, nil

	//TODO: HACK! The client sends an incorrect slot ID for the offhand as of 1.19.70 (though this doesn't really matter since the offhand has only 1 slot anyway)
	case protocol.ContainerOffhand:
		return mcpe.ContainerIDOffhand, 0, nil

	case protocol.ContainerAnvilInput, protocol.ContainerAnvilMaterial, protocol.ContainerBeaconPayment,
		protocol.ContainerCartographyAdditional, protocol.ContainerCartographyInput, protocol.ContainerCompoundCreatorInput,
		protocol.ContainerCraftingInput, protocol.ContainerCreatedOutput, protocol.ContainerCursor,
		protocol.ContainerEnchantingInput, protocol.ContainerEnchantingMaterial, protocol.ContainerGrindstoneAdditional,
		protocol.ContainerGrindstoneInput, protocol.ContainerLabTableInput, protocol.ContainerLoomDye,
		protocol.ContainerLoomInput, protocol.ContainerLoomMaterial, protocol.ContainerMaterialReducerInput,
		protocol.ContainerMaterialReducerOutput, protocol.ContainerSmithingTableInput, protocol.ContainerSmithingTableMaterial,
		protocol.ContainerSmithingTableTemplate, protocol.ContainerStonecutterInput, protocol.ContainerTradeTwoIngredientOne,
		protocol.ContainerTradeTwoIngredientTwo, protocol.ContainerTradeIngredientOne, protocol.ContainerTradeIngredientTwo:
		return mcpe.ContainerIDUI, slotID, nil

	case protocol.ContainerBarrel, protocol.ContainerBlastFurnaceIngredient, protocol.ContainerBrewingStandFuel,
		protocol.ContainerBrewingStandInput, protocol.ContainerBrewingStandResult, protocol.ContainerFurnaceFuel,
		protocol.ContainerFurnaceIngredient, protocol.ContainerFurnaceResult, protocol.ContainerHorseEquip,
		protocol.ContainerLevelEntity, //chest
		protocol.ContainerShulkerBox, protocol.ContainerSmokerIngredient:
		return currentWindowID, slotID, nil
	}
	//all preview slots are ignored, since the client shouldn't be modifying those directly
	return 0, 0, &network.PacketHandlingError{Message: fmt.Sprintf("Unexpected container UI ID %d", containerInterfaceID)}
}
