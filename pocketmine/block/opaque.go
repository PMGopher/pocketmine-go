package block

// Opaque is a port of pocketmine\block\Opaque.
//
// Opaque blocks do not allow light to pass through. They are usually collidable full-cube
// blocks. Most blocks in Minecraft fall into this category.
type Opaque struct {
	Block
}

func (o *Opaque) IsSolid() bool { return true }

// Clone lets Opaque be instantiated directly (PHP does this for many simple blocks with no extra
// state - e.g. VanillaBlocksInputs.php's `new Opaque($id, "Obsidian", ...)`). Concrete types that
// embed Opaque and add their own state (Wool, Bedrock, etc.) already define their own Clone that
// shadows this one, so this only matters for genuinely bare Opaque instances.
func (o *Opaque) Clone() Behavior {
	c := *o
	c.rebind(&c)
	return &c
}
