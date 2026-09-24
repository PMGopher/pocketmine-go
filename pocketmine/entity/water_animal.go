package entity

import entityevent "pocketmine-go/pocketmine/event/entity"

// WaterAnimal is a port of the abstract pocketmine\entity\WaterAnimal (also Ageable).
type WaterAnimal struct {
	Living

	baby bool
}

func (w *WaterAnimal) IsBaby() bool { return w.baby }

// CanBreathe is a port of WaterAnimal::canBreathe: water animals breathe underwater.
func (w *WaterAnimal) CanBreathe() bool { return w.lself.IsUnderwater() }

// OnAirExpired is a port of WaterAnimal::onAirExpired.
func (w *WaterAnimal) OnAirExpired() {
	ev := entityevent.NewEntityDamageEvent(w.lself, entityevent.CauseSuffocation, 2, nil)
	w.lself.Attack(ev)
}

// SyncNetworkData is a port of WaterAnimal::syncNetworkData.
func (w *WaterAnimal) SyncNetworkData(properties *MetadataCollection) {
	w.Living.SyncNetworkData(properties)
	properties.SetGenericFlag(FlagBaby, w.baby)
}
