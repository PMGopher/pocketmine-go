package entity

import (
	"fmt"
	"sync"
)

// maxFloat32 is PHP's 340282346638528859811704183484516925440.00 (FLT_MAX), the upper bound of
// several attributes.
const maxFloat32 = 340282346638528859811704183484516925440.0

// AttributeFactory is a port of pocketmine\entity\AttributeFactory.
type AttributeFactory struct {
	attributes map[string]*Attribute
}

var (
	attributeFactoryOnce     sync.Once
	attributeFactoryInstance *AttributeFactory
)

// GetAttributeFactory is the port of AttributeFactory::getInstance().
func GetAttributeFactory() *AttributeFactory {
	attributeFactoryOnce.Do(func() {
		f := &AttributeFactory{attributes: map[string]*Attribute{}}
		f.Register(AttributeAbsorption, 0.00, maxFloat32, 0.00, true)
		f.Register(AttributeSaturation, 0.00, 20.00, 20.00, true)
		f.Register(AttributeExhaustion, 0.00, 5.00, 0.0, false)
		f.Register(AttributeKnockbackResistance, 0.00, 1.00, 0.00, true)
		f.Register(AttributeHealth, 0.00, 20.00, 20.00, true)
		f.Register(AttributeMovementSpeed, 0.00, maxFloat32, 0.10, true)
		f.Register(AttributeFollowRange, 0.00, 2048.00, 16.00, false)
		f.Register(AttributeHunger, 0.00, 20.00, 20.00, true)
		f.Register(AttributeAttackDamage, 0.00, maxFloat32, 1.00, false)
		f.Register(AttributeExperienceLevel, 0.00, 24791.00, 0.00, true)
		f.Register(AttributeExperience, 0.00, 1.00, 0.00, true)
		f.Register(AttributeUnderwaterMovement, 0.0, maxFloat32, 0.02, true)
		f.Register(AttributeLuck, -1024.0, 1024.0, 0.0, true)
		f.Register(AttributeFallDamage, 0.0, maxFloat32, 1.0, true)
		f.Register(AttributeHorseJumpStrength, 0.0, 2.0, 0.7, true)
		f.Register(AttributeZombieSpawnReinforcements, 0.0, 1.0, 0.0, true)
		f.Register(AttributeLavaMovement, 0.0, maxFloat32, 0.02, true)
		attributeFactoryInstance = f
	})
	return attributeFactoryInstance
}

// Get returns a copy of the attribute registered under id (nil if none).
func (f *AttributeFactory) Get(id string) *Attribute {
	if a, ok := f.attributes[id]; ok {
		return a.Clone()
	}
	return nil
}

// MustGet is Get, panicking if the attribute isn't registered.
func (f *AttributeFactory) MustGet(id string) *Attribute {
	result := f.Get(id)
	if result == nil {
		panic(fmt.Sprintf("Attribute %s is not registered", id))
	}
	return result
}

// Register registers a new attribute type (panicking on invalid ranges).
func (f *AttributeFactory) Register(id string, minValue, maxValue, defaultValue float64, shouldSend bool) *Attribute {
	a := NewAttribute(id, minValue, maxValue, defaultValue, shouldSend)
	f.attributes[id] = a
	return a
}
