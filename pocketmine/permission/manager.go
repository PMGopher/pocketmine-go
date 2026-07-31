package permission

import "sync"

// Manager is a port of pocketmine\permission\PermissionManager.
//
// PHP keys its subscription maps by spl_object_id(Permissible) since PHP objects can't be map
// keys directly; a Go *Permissible is already a comparable pointer, so it's used as the map key
// with no identity-hash workaround needed.
type Manager struct {
	mu       sync.Mutex
	perms    map[string]*Permission
	permSubs map[string]map[*Permissible]struct{}
}

func newManager() *Manager {
	return &Manager{perms: map[string]*Permission{}, permSubs: map[string]map[*Permissible]struct{}{}}
}

var (
	globalManagerOnce sync.Once
	globalManager     *Manager
)

func GetManager() *Manager {
	globalManagerOnce.Do(func() { globalManager = newManager() })
	return globalManager
}

func (m *Manager) GetPermission(name string) *Permission {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.perms[name]
}

// AddPermission registers permission, returning false if a permission with that name is already registered.
func (m *Manager) AddPermission(perm *Permission) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.perms[perm.Name()]; exists {
		return false
	}
	m.perms[perm.Name()] = perm
	return true
}

func (m *Manager) RemovePermission(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.perms, name)
}

// Deprecated: superseded by server chat broadcast channels once the Server type is ported.
func (m *Manager) subscribeToPermission(permission string, p *Permissible) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs, ok := m.permSubs[permission]
	if !ok {
		subs = map[*Permissible]struct{}{}
		m.permSubs[permission] = subs
	}
	subs[p] = struct{}{}
}

// Deprecated: superseded by server chat broadcast channels once the Server type is ported.
func (m *Manager) unsubscribeFromPermission(permission string, p *Permissible) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs, ok := m.permSubs[permission]
	if !ok {
		return
	}
	delete(subs, p)
	if len(subs) == 0 {
		delete(m.permSubs, permission)
	}
}

func (m *Manager) unsubscribeFromAllPermissions(p *Permissible) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for permission, subs := range m.permSubs {
		delete(subs, p)
		if len(subs) == 0 {
			delete(m.permSubs, permission)
		}
	}
}

func (m *Manager) getPermissionSubscriptions(permission string) []*Permissible {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.permSubs[permission]
	result := make([]*Permissible, 0, len(subs))
	for p := range subs {
		result = append(result, p)
	}
	return result
}

func (m *Manager) GetPermissions() map[string]*Permission {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]*Permission, len(m.perms))
	for k, v := range m.perms {
		result[k] = v
	}
	return result
}

func (m *Manager) ClearPermissions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.perms = map[string]*Permission{}
}
