package block

import (
	"fmt"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Ageable is a port of pocketmine\block\utils\Ageable.
type Ageable interface {
	GetAge() int
	GetMaxAge() int
	SetAge(age int)
}

// AgeComponent is a port of pocketmine\block\utils\AgeableTrait's state. The trait requires a
// `MAX_AGE` class constant on the using class; Go has no per-type constant reachable from an
// embedded component, so MaxAge is a plain field set once by the concrete type's constructor
// instead.
type AgeComponent struct {
	Age    int
	MaxAge int
}

func NewAgeComponent(maxAge int) AgeComponent { return AgeComponent{MaxAge: maxAge} }

func (a *AgeComponent) DescribeAge(w runtime.DataDescriber) { w.BoundedIntAuto(0, a.MaxAge, &a.Age) }

func (a *AgeComponent) GetAge() int { return a.Age }

func (a *AgeComponent) GetMaxAge() int { return a.MaxAge }

// SetAge panics if age is out of range, mirroring the PHP original's \InvalidArgumentException (a
// programmer error at the call site).
func (a *AgeComponent) SetAge(age int) {
	if age < 0 || age > a.MaxAge {
		panic(fmt.Sprintf("Age must be in range 0 ... %d", a.MaxAge))
	}
	a.Age = age
}
