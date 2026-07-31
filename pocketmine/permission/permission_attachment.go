package permission

import (
	"fmt"
	"sync"
)

// Plugin is the minimal surface PermissionAttachment needs from a plugin. It's declared locally
// (rather than importing the not-yet-ported plugin package) so any future plugin.Plugin type
// satisfies it automatically just by having these two methods — Go interfaces are structural, so
// no import (and no risk of an import cycle with the future plugin package) is needed here.
type Plugin interface {
	IsEnabled() bool
	Name() string
}

// NamedPermission is a (name, value) pair, used where PermissionAttachment's permissions need to
// be read back in insertion order (see OrderedPermissions).
type NamedPermission struct {
	Name  string
	Value bool
}

// PermissionAttachment is a port of pocketmine\permission\PermissionAttachment.
//
// SetPermission's overwrite-moves-to-end-of-order behavior is deliberately preserved (see its
// doc comment) because PocketMine's own source flags it as load-bearing for how conflicting
// child permissions resolve — the same insertion-order concern as CompoundTag/ObjectSet
// elsewhere in this port, here made explicit by re-appending after a delete on overwrite.
type PermissionAttachment struct {
	mu          sync.Mutex
	plugin      Plugin
	permissions map[string]bool
	order       []string
	subscribers map[*Permissible]struct{}
}

// NewPermissionAttachment mirrors the PermissionAttachment constructor's disabled-plugin check.
func NewPermissionAttachment(plugin Plugin) (*PermissionAttachment, error) {
	if !plugin.IsEnabled() {
		return nil, fmt.Errorf("plugin %s is disabled", plugin.Name())
	}
	return &PermissionAttachment{
		plugin:      plugin,
		permissions: map[string]bool{},
		subscribers: map[*Permissible]struct{}{},
	}, nil
}

func (a *PermissionAttachment) Plugin() Plugin { return a.plugin }

func (a *PermissionAttachment) Subscribers() []*Permissible {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]*Permissible, 0, len(a.subscribers))
	for s := range a.subscribers {
		result = append(result, s)
	}
	return result
}

// GetPermissions returns a snapshot map (unordered). Use OrderedPermissions where insertion
// order matters (child-permission conflict resolution).
func (a *PermissionAttachment) GetPermissions() map[string]bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make(map[string]bool, len(a.permissions))
	for k, v := range a.permissions {
		result[k] = v
	}
	return result
}

func (a *PermissionAttachment) OrderedPermissions() []NamedPermission {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]NamedPermission, len(a.order))
	for i, name := range a.order {
		result[i] = NamedPermission{Name: name, Value: a.permissions[name]}
	}
	return result
}

func (a *PermissionAttachment) recalculatePermissibles() {
	for _, s := range a.Subscribers() {
		s.RecalculatePermissions()
	}
}

func (a *PermissionAttachment) ClearPermissions() {
	a.mu.Lock()
	a.permissions = map[string]bool{}
	a.order = nil
	a.mu.Unlock()
	a.recalculatePermissibles()
}

func (a *PermissionAttachment) SetPermissions(permissions map[string]bool) {
	a.mu.Lock()
	for k, v := range permissions {
		if _, exists := a.permissions[k]; !exists {
			a.order = append(a.order, k)
		}
		a.permissions[k] = v
	}
	a.mu.Unlock()
	a.recalculatePermissibles()
}

func (a *PermissionAttachment) UnsetPermissions(names []string) {
	a.mu.Lock()
	for _, name := range names {
		a.removeFromOrderLocked(name)
		delete(a.permissions, name)
	}
	a.mu.Unlock()
	a.recalculatePermissibles()
}

// SetPermission sets a single permission. If it already had this exact value, this is a no-op
// (no recalculation triggered). If it existed with a different value, its position in the
// insertion order is moved to the end — see the type doc comment for why that matters.
func (a *PermissionAttachment) SetPermission(name string, value bool) {
	a.mu.Lock()
	if existing, ok := a.permissions[name]; ok {
		if existing == value {
			a.mu.Unlock()
			return
		}
		a.removeFromOrderLocked(name)
	}
	a.permissions[name] = value
	a.order = append(a.order, name)
	a.mu.Unlock()
	a.recalculatePermissibles()
}

func (a *PermissionAttachment) UnsetPermission(name string) {
	a.mu.Lock()
	_, exists := a.permissions[name]
	if exists {
		delete(a.permissions, name)
		a.removeFromOrderLocked(name)
	}
	a.mu.Unlock()
	if exists {
		a.recalculatePermissibles()
	}
}

func (a *PermissionAttachment) removeFromOrderLocked(name string) {
	for i, n := range a.order {
		if n == name {
			a.order = append(a.order[:i:i], a.order[i+1:]...)
			return
		}
	}
}

func (a *PermissionAttachment) subscribePermissible(p *Permissible) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.subscribers[p] = struct{}{}
}

func (a *PermissionAttachment) unsubscribePermissible(p *Permissible) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.subscribers, p)
}
