package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// DetectorRail is a port of pocketmine\block\DetectorRail.
type DetectorRail struct {
	StraightOnlyRail

	Activated bool
}

func NewDetectorRail(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DetectorRail {
	d := &DetectorRail{StraightOnlyRail: StraightOnlyRail{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	d.Init(d)
	return d
}

func (d *DetectorRail) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DetectorRail) DescribeBlockOnlyState(w runtime.DataDescriber) {
	d.StraightOnlyRail.DescribeBlockOnlyState(w)
	w.Bool(&d.Activated)
}

func (d *DetectorRail) IsActivated() bool { return d.Activated }

func (d *DetectorRail) SetActivated(activated bool) { d.Activated = activated }
