package resourcepacks

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/resource"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/utils"
)

// defaultResourcePacksYml is resources/resource_packs.yml, copied into the resource pack folder on
// first start.
//
//go:embed resource_packs.yml
var defaultResourcePacksYml []byte

// ResourcePackManager is a port of pocketmine\resourcepacks\ResourcePackManager.
type ResourcePackManager struct {
	path                 string
	serverForceResources bool
	resourcePacks        []ResourcePack
	uuidList             map[string]ResourcePack
	encryptionKeys       map[string]string
}

// NewResourcePackManager is a port of ResourcePackManager::__construct: path is the folder the
// packs and resource_packs.yml are in.
func NewResourcePackManager(path string, logger log.Logger) (*ResourcePackManager, error) {
	m := &ResourcePackManager{path: path, uuidList: map[string]ResourcePack{}, encryptionKeys: map[string]string{}}

	if info, err := os.Stat(path); os.IsNotExist(err) {
		logger.Debug(fmt.Sprintf("Resource packs path %s does not exist, creating directory", path))
		if err := os.MkdirAll(path, 0o777); err != nil {
			return nil, err
		}
	} else if err == nil && !info.IsDir() {
		return nil, fmt.Errorf("Resource packs path %s exists and is not a directory", path)
	}

	resourcePacksYml := filepath.Join(path, "resource_packs.yml")
	if _, err := os.Stat(resourcePacksYml); os.IsNotExist(err) {
		if err := os.WriteFile(resourcePacksYml, defaultResourcePacksYml, 0o666); err != nil {
			return nil, err
		}
	}
	config, err := utils.NewConfig(resourcePacksYml, utils.ConfigYAML, map[string]any{})
	if err != nil {
		return nil, err
	}

	switch v := config.Get("force_resources", false).(type) {
	case bool:
		m.serverForceResources = v
	case string:
		m.serverForceResources = v != "" && v != "0"
	case int:
		m.serverForceResources = v != 0
	}

	logger.Info("Loading resource packs...")

	stack := config.Get("resource_stack", []any{})
	resourceStack, ok := stack.([]any)
	if !ok && stack != nil {
		return nil, fmt.Errorf("\"resource_stack\" key should contain a list of pack names")
	}
	for pos, entry := range resourceStack {
		switch entry.(type) {
		case string, int, int64, float64:
		default:
			logger.Critical(fmt.Sprintf("Found invalid entry in resource pack list at offset %d of type %T", pos, entry))
			continue
		}
		pack := fmt.Sprint(entry)
		if err := m.loadStackEntry(pack); err != nil {
			logger.Critical(fmt.Sprintf("Could not load resource pack %q: %v", pack, err))
		}
	}

	logger.Debug(fmt.Sprintf("Successfully loaded %d resource packs", len(m.resourcePacks)))
	return m, nil
}

func (m *ResourcePackManager) loadStackEntry(pack string) error {
	newPack, err := loadPackFromPath(filepath.Join(m.path, pack))
	if err != nil {
		return err
	}
	index := strings.ToLower(newPack.GetPackId())
	if _, err := uuid.Parse(index); err != nil {
		return &ResourcePackError{Message: fmt.Sprintf("Invalid UUID (%s)", index)}
	}
	m.uuidList[index] = newPack
	m.resourcePacks = append(m.resourcePacks, newPack)

	keyPath := filepath.Join(m.path, pack+".key")
	if _, err := os.Stat(keyPath); err == nil {
		raw, err := os.ReadFile(keyPath)
		if err != nil {
			return &ResourcePackError{Message: "Could not read encryption key file: " + err.Error(), Cause: err}
		}
		key := strings.TrimRight(string(raw), "\r\n")
		if len(key) != 32 {
			return &ResourcePackError{Message: "Invalid encryption key length, must be exactly 32 bytes"}
		}
		m.encryptionKeys[index] = key
	}
	return nil
}

// GetPath returns the directory which resource packs are loaded from.
func (m *ResourcePackManager) GetPath() string { return m.path + string(filepath.Separator) }

// ResourcePacksRequired returns whether players must accept resource packs in order to join.
func (m *ResourcePackManager) ResourcePacksRequired() bool { return m.serverForceResources }

// SetResourcePacksRequired sets whether players must accept resource packs in order to join.
func (m *ResourcePackManager) SetResourcePacksRequired(value bool) { m.serverForceResources = value }

// GetResourceStack returns an array of resource packs in use, sorted in order of priority.
func (m *ResourcePackManager) GetResourceStack() []ResourcePack {
	return append([]ResourcePack(nil), m.resourcePacks...)
}

// SetResourceStack sets the resource packs to use. Packs earlier in the list will override the
// effects of later packs.
func (m *ResourcePackManager) SetResourceStack(resourceStack []ResourcePack) error {
	uuidList := map[string]ResourcePack{}
	var resourcePacks []ResourcePack
	for _, pack := range resourceStack {
		id := strings.ToLower(pack.GetPackId())
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("Invalid resource pack UUID (%s)", id)
		}
		if _, ok := uuidList[id]; ok {
			return fmt.Errorf("Cannot load two resource pack with the same UUID (%s)", id)
		}
		uuidList[id] = pack
		resourcePacks = append(resourcePacks, pack)
	}
	m.resourcePacks = resourcePacks
	m.uuidList = uuidList
	return nil
}

// GetPackById returns the resource pack matching the specified UUID string, or nil if the ID was
// not recognized.
func (m *ResourcePackManager) GetPackById(id string) ResourcePack {
	return m.uuidList[strings.ToLower(id)]
}

// GetPackIdList returns an array of pack IDs for packs currently in use.
func (m *ResourcePackManager) GetPackIdList() []string {
	ids := make([]string, 0, len(m.uuidList))
	for _, p := range m.resourcePacks {
		ids = append(ids, strings.ToLower(p.GetPackId()))
	}
	return ids
}

// GetPackEncryptionKey returns the key with which the pack was encrypted, or "" if the pack has
// no key.
func (m *ResourcePackManager) GetPackEncryptionKey(id string) (string, bool) {
	key, ok := m.encryptionKeys[strings.ToLower(id)]
	return key, ok
}

// SetPackEncryptionKey sets the encryption key to use for decrypting the specified resource
// pack. The pack must be loaded (except when removing a key with an empty key).
func (m *ResourcePackManager) SetPackEncryptionKey(id string, key string) error {
	id = strings.ToLower(id)
	if key == "" {
		//allow deprovisioning keys for resource packs that have been removed
		delete(m.encryptionKeys, id)
		return nil
	}
	if _, ok := m.uuidList[id]; !ok {
		return fmt.Errorf("Unknown pack ID %s", id)
	}
	if len(key) != 32 {
		return fmt.Errorf("Encryption key must be exactly 32 bytes long")
	}
	m.encryptionKeys[id] = key
	return nil
}

// NetworkPacks converts packs to the gophertunnel packs sent to clients, applying their
// encryption keys. Packs that aren't ZippedResourcePacks can't be delivered by gophertunnel and
// are skipped.
func NetworkPacks(packs []ResourcePack, keys map[string]string) []*resource.Pack {
	result := make([]*resource.Pack, 0, len(packs))
	for _, p := range packs {
		z, ok := p.(*ZippedResourcePack)
		if !ok {
			continue
		}
		pack := z.Pack()
		if key, ok := keys[strings.ToLower(p.GetPackId())]; ok {
			pack = pack.WithContentKey(key)
		}
		result = append(result, pack)
	}
	return result
}
