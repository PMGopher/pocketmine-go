package block

// DriedKelp is a port of pocketmine\block\DriedKelp.
type DriedKelp struct {
	Opaque
}

func NewDriedKelp(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DriedKelp {
	d := &DriedKelp{Opaque{NewBlock(idInfo, name, typeInfo)}}
	d.Init(d)
	return d
}

func (d *DriedKelp) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DriedKelp) GetFlameEncouragement() int { return 30 }

func (d *DriedKelp) GetFlammability() int { return 60 }

func (d *DriedKelp) GetFuelTime() int { return 4000 }
