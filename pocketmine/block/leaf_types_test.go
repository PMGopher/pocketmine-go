package block

// Compile-time proof that every concrete leaf block type in this file set fully satisfies
// Behavior (i.e. nothing was missed when overriding methods).
var (
	_ Behavior = (*StoneButton)(nil)
	_ Behavior = (*Air)(nil)
	_ Behavior = (*Cobweb)(nil)
	_ Behavior = (*EndRod)(nil)
	_ Behavior = (*Flower)(nil)
	_ Behavior = (*DeadBush)(nil)
	_ Behavior = (*WitherRose)(nil)
	_ Behavior = (*RedMushroom)(nil)
	_ Behavior = (*Torch)(nil)
	_ Behavior = (*Lever)(nil)
	_ Behavior = (*NetherSprouts)(nil)
	_ Behavior = (*CactusFlower)(nil)
	_ Behavior = (*DoublePlant)(nil)
	_ Behavior = (*NetherWartPlant)(nil)
	_ Behavior = (*Carpet)(nil)
	_ Behavior = (*WaterLily)(nil)
	_ Behavior = (*Vine)(nil)
	_ Behavior = (*Wool)(nil)
	_ Behavior = (*PinkPetals)(nil)
)
