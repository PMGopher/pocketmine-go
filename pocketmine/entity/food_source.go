package entity

// FoodSource is a port of pocketmine\entity\FoodSource: consumables which restore hunger.
type FoodSource interface {
	Consumable

	GetFoodRestore() int
	GetSaturationRestore() float64

	// RequiresHunger returns whether a Human eating this FoodSource must have a non-full hunger bar.
	RequiresHunger() bool
}
