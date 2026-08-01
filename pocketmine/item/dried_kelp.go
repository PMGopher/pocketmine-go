package item

// DriedKelp is a port of pocketmine\item\DriedKelp.
type DriedKelp struct {
	Food
}

func NewDriedKelp(identifier ItemIdentifier, name string) *DriedKelp {
	d := &DriedKelp{}
	d.Init(d, identifier, name)
	return d
}

func (d *DriedKelp) Clone() Item {
	cl := *d
	cl.rebind(&cl)
	return &cl
}

func (d *DriedKelp) GetFoodRestore() int { return 1 }

func (d *DriedKelp) GetSaturationRestore() float64 { return 0.6 }
