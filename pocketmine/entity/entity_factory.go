package entity

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// EntityFactory NBT keys, a port of EntityFactory::TAG_*.
const (
	TagIdentifier = "identifier" //TAG_String
	TagLegacyID   = "id"         //TAG_Int
)

// EntityCreationFunc creates an entity of a registered type from its saved data.
type EntityCreationFunc func(w *world.World, tag *nbt.CompoundTag) (world.Entity, error)

// EntityFactory is a port of pocketmine\entity\EntityFactory: a registry of entity types, keyed by
// their Go type (PHP's class name), with the save IDs they're stored under.
//
// PHP's constructor registers every vanilla entity class. Here each package registers its own
// types from init() (this package: Human, Squid, Villager, Zombie; entity/object and
// entity/projectile: theirs), since this package can't import those.
type EntityFactory struct {
	mu            sync.RWMutex
	creationFuncs map[string]EntityCreationFunc
	saveNames     map[reflect.Type]string
}

var (
	entityFactoryOnce     sync.Once
	entityFactoryInstance *EntityFactory
)

// GetEntityFactory is the port of EntityFactory::getInstance().
func GetEntityFactory() *EntityFactory {
	entityFactoryOnce.Do(func() {
		entityFactoryInstance = &EntityFactory{
			creationFuncs: map[string]EntityCreationFunc{},
			saveNames:     map[reflect.Type]string{},
		}
	})
	return entityFactoryInstance
}

// Register is a port of EntityFactory::register: registers an entity type (entityType is the
// pointer type, e.g. reflect.TypeFor[*Zombie]()) with the function that creates it from saved
// data and the save IDs it may be stored under. The first save name is the one it's saved with.
// Panics if no save names are given (PHP's InvalidArgumentException).
func (f *EntityFactory) Register(entityType reflect.Type, creationFunc EntityCreationFunc, saveNames []string) {
	if len(saveNames) == 0 {
		panic("At least one save name must be provided")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, name := range saveNames {
		f.creationFuncs[name] = creationFunc
	}
	f.saveNames[entityType] = saveNames[0]
}

// RegisterEntity is Register with the entity type as a type parameter.
func RegisterEntity[T world.Entity](f *EntityFactory, creationFunc func(w *world.World, tag *nbt.CompoundTag) (T, error), saveNames []string) {
	f.Register(reflect.TypeFor[T](), func(w *world.World, tag *nbt.CompoundTag) (world.Entity, error) {
		e, err := creationFunc(w, tag)
		if err != nil {
			return nil, err
		}
		return e, nil
	}, saveNames)
}

// IsRegistered is a port of EntityFactory::isRegistered.
func (f *EntityFactory) IsRegistered(entityType reflect.Type) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.saveNames[entityType]
	return ok
}

// CreateFromData is a port of EntityFactory::createFromData: creates an entity from saved data. A
// nil entity with a nil error means the data is for an unknown entity type. Invalid data is
// returned as a *data.SavedDataLoadingError.
func (f *EntityFactory) CreateFromData(w *world.World, tag *nbt.CompoundTag) (result world.Entity, err error) {
	defer func() {
		if r := recover(); r != nil {
			var loadErr *data.SavedDataLoadingError
			if e, ok := r.(error); ok && errors.As(e, &loadErr) {
				result, err = nil, loadErr
				return
			}
			if e, ok := r.(error); ok {
				result, err = nil, &data.SavedDataLoadingError{Message: e.Error(), Cause: e}
				return
			}
			result, err = nil, data.NewSavedDataLoadingError(fmt.Sprint(r))
		}
	}()

	saveID, ok := tag.GetTag(TagIdentifier)
	if !ok {
		saveID, ok = tag.GetTag(TagLegacyID)
	}
	var creationFunc EntityCreationFunc
	f.mu.RLock()
	switch id := saveID.(type) {
	case nbt.StringTag:
		creationFunc = f.creationFuncs[string(id)]
	case nbt.IntTag: //legacy MCPE format
		if stringID, ok := bedrock.LegacyEntityIdToStringIdMap().LegacyToString(int(id) & 0xff); ok {
			creationFunc = f.creationFuncs[stringID]
		}
	}
	f.mu.RUnlock()
	if creationFunc == nil {
		return nil, nil
	}

	return creationFunc(w, tag)
}

// InjectSaveID is a port of EntityFactory::injectSaveId. Panics if the type isn't registered (PHP's
// InvalidArgumentException).
func (f *EntityFactory) InjectSaveID(entityType reflect.Type, saveData *nbt.CompoundTag) {
	saveData.SetString(TagIdentifier, nbt.StringTag(f.GetSaveID(entityType)))
}

// GetSaveID is a port of EntityFactory::getSaveId. Panics if the type isn't registered.
func (f *EntityFactory) GetSaveID(entityType reflect.Type) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if name, ok := f.saveNames[entityType]; ok {
		return name
	}
	panic(fmt.Sprintf("Entity %s is not registered", entityType))
}

func init() {
	f := GetEntityFactory()
	RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Squid, error) {
		loc, err := ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewSquid(loc, tag), nil
	}, []string{"Squid", "minecraft:squid"})
	RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Villager, error) {
		loc, err := ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewVillager(loc, tag), nil
	}, []string{"Villager", "minecraft:villager"})
	RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Zombie, error) {
		loc, err := ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		return NewZombie(loc, tag), nil
	}, []string{"Zombie", "minecraft:zombie"})
	RegisterEntity(f, func(w *world.World, tag *nbt.CompoundTag) (*Human, error) {
		loc, err := ParseLocation(tag, w)
		if err != nil {
			return nil, err
		}
		skin, err := ParseSkinNBT(tag)
		if err != nil {
			return nil, err
		}
		return NewHuman(loc, skin, tag), nil
	}, []string{"Human"})

	world.LoadEntityFunc = func(w *world.World, tag *nbt.CompoundTag) (world.Entity, error) {
		return GetEntityFactory().CreateFromData(w, tag)
	}
}
