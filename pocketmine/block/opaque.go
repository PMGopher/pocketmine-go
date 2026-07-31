package block

// Opaque is a port of pocketmine\block\Opaque.
//
// Opaque blocks do not allow light to pass through. They are usually collidable full-cube
// blocks. Most blocks in Minecraft fall into this category.
type Opaque struct {
	Block
}

func (o *Opaque) IsSolid() bool { return true }
