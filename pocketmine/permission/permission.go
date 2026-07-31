package permission

import "sync"

// Permission is a port of pocketmine\permission\Permission.
//
// Description holds either a plain string or (once the lang package is ported) a
// *lang.Translatable — modeled as `any` for now rather than taking a hard dependency on an
// unported package. Most callers just pass a string.
type Permission struct {
	mu          sync.Mutex
	name        string
	description any
	children    map[string]bool
}

func NewPermission(name string, description any, children map[string]bool) *Permission {
	if description == nil {
		description = ""
	}
	c := make(map[string]bool, len(children))
	for k, v := range children {
		c[k] = v
	}
	p := &Permission{name: name, description: description, children: c}
	p.RecalculatePermissibles()
	return p
}

func (p *Permission) Name() string { return p.name }

// Children returns a snapshot copy of this permission's child permissions.
func (p *Permission) Children() map[string]bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make(map[string]bool, len(p.children))
	for k, v := range p.children {
		result[k] = v
	}
	return result
}

func (p *Permission) Description() any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.description
}

func (p *Permission) SetDescription(value any) {
	p.mu.Lock()
	p.description = value
	p.mu.Unlock()
}

// Permissibles returns every Permissible currently subscribed to this permission name.
func (p *Permission) Permissibles() []*Permissible {
	return GetManager().getPermissionSubscriptions(p.name)
}

func (p *Permission) RecalculatePermissibles() {
	for _, perm := range p.Permissibles() {
		perm.RecalculatePermissions()
	}
}

func (p *Permission) AddChild(name string, value bool) {
	p.mu.Lock()
	p.children[name] = value
	p.mu.Unlock()
	p.RecalculatePermissibles()
}

func (p *Permission) RemoveChild(name string) {
	p.mu.Lock()
	delete(p.children, name)
	p.mu.Unlock()
	p.RecalculatePermissibles()
}
