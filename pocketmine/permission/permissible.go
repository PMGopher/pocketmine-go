package permission

import (
	"sync"
)

// Permissible is a port of pocketmine\permission\PermissibleInternal, combined with what
// PermissibleBase/PermissibleDelegateTrait added on top.
//
// PHP splits this into PermissibleInternal (the real logic) plus a PermissibleBase wrapper whose
// entire purpose is calling destroyCycles() from __destruct(), because PermissibleInternal holds
// references (permission subscriptions, attachments) that form reference cycles PHP's refcounting
// GC can't collect on its own. Go's tracing garbage collector doesn't have that problem — cycles
// are collected like anything else — so there's no need for a second wrapper type here. Close()
// is kept anyway, not for GC correctness, but as a hygiene measure: it proactively drops this
// Permissible's subscriptions from the global Manager and its attachments, so a long-lived
// Manager doesn't accumulate stale subscription entries for objects that are logically "done"
// (e.g. a disconnected player) but haven't been garbage collected yet.
type Permissible struct {
	mu                     sync.Mutex
	rootPermissions        map[string]bool
	attachments            map[*PermissionAttachment]struct{}
	permissions            map[string]*AttachmentInfo
	recalculationCallbacks []func(changedPermissionsOldValues map[string]bool)
}

func NewPermissible(basePermissions map[string]bool) *Permissible {
	root := make(map[string]bool, len(basePermissions))
	for k, v := range basePermissions {
		root[k] = v
	}
	p := &Permissible{
		rootPermissions: root,
		attachments:     map[*PermissionAttachment]struct{}{},
		permissions:     map[string]*AttachmentInfo{},
	}
	p.RecalculatePermissions()
	return p
}

func (p *Permissible) SetBasePermission(name string, grant bool) {
	p.mu.Lock()
	p.rootPermissions[name] = grant
	p.mu.Unlock()
	p.RecalculatePermissions()
}

func (p *Permissible) UnsetBasePermission(name string) {
	p.mu.Lock()
	delete(p.rootPermissions, name)
	p.mu.Unlock()
	p.RecalculatePermissions()
}

func (p *Permissible) IsPermissionSet(name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.permissions[name]
	return ok
}

func (p *Permissible) HasPermission(name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if info, ok := p.permissions[name]; ok {
		return info.Value()
	}
	return false
}

// AddAttachment mirrors Permissible::addAttachment(). If value is nil, name is ignored (matching
// PHP's `$name !== null && $value !== null` gate on whether to seed an initial permission).
func (p *Permissible) AddAttachment(plugin Plugin, name string, value *bool) (*PermissionAttachment, error) {
	attachment, err := NewPermissionAttachment(plugin)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.attachments[attachment] = struct{}{}
	p.mu.Unlock()

	if name != "" && value != nil {
		attachment.SetPermission(name, *value)
	}
	attachment.subscribePermissible(p)
	p.RecalculatePermissions()

	return attachment, nil
}

func (p *Permissible) RemoveAttachment(attachment *PermissionAttachment) {
	p.mu.Lock()
	_, exists := p.attachments[attachment]
	if exists {
		delete(p.attachments, attachment)
	}
	p.mu.Unlock()

	if exists {
		attachment.unsubscribePermissible(p)
		p.RecalculatePermissions()
	}
}

// RecalculatePermissions rebuilds the effective permission set from root permissions and
// attachments, and returns a map of changed permission names to their PREVIOUS value (matching
// PHP's $changedPermissionsOldValues — a brand-new permission's "previous value" is reported as
// false, since it had no prior state).
func (p *Permissible) RecalculatePermissions() map[string]bool {
	mgr := GetManager()
	mgr.unsubscribeFromAllPermissions(p)

	p.mu.Lock()
	oldPermissions := p.permissions
	p.permissions = map[string]*AttachmentInfo{}

	for name, isGranted := range p.rootPermissions {
		perm := mgr.GetPermission(name)
		if perm == nil {
			panic("Unregistered root permission " + name)
		}
		info := NewAttachmentInfo(name, nil, isGranted, nil)
		p.permissions[name] = info
		mgr.subscribeToPermission(name, p)
		p.calculateChildPermissionsLocked(mapToNamed(perm.Children()), !isGranted, nil, info)
	}

	for attachment := range p.attachments {
		p.calculateChildPermissionsLocked(attachment.OrderedPermissions(), false, attachment, nil)
	}

	diff := map[string]bool{}
	for name, info := range p.permissions {
		oldInfo, existed := oldPermissions[name]
		if !existed {
			diff[name] = false
			continue
		}
		if oldInfo.Value() != info.Value() {
			continue
		}
		delete(oldPermissions, name)
	}
	for name, oldInfo := range oldPermissions {
		diff[name] = oldInfo.Value()
	}

	callbacks := append([]func(map[string]bool){}, p.recalculationCallbacks...)
	p.mu.Unlock()

	if len(diff) > 0 {
		for _, cb := range callbacks {
			cb(diff)
		}
	}
	return diff
}

// mapToNamed converts an unordered map into a slice for calculateChildPermissionsLocked. Unlike
// PermissionAttachment (see its type doc comment), Permission's own child-permission order isn't
// called out anywhere in the original as behaviorally significant, so a plain map — with no
// defined iteration order — is an acceptable simplification for Permission.children specifically.
func mapToNamed(m map[string]bool) []NamedPermission {
	result := make([]NamedPermission, 0, len(m))
	for k, v := range m {
		result = append(result, NamedPermission{Name: k, Value: v})
	}
	return result
}

// calculateChildPermissionsLocked must be called with p.mu held.
func (p *Permissible) calculateChildPermissionsLocked(children []NamedPermission, invert bool, attachment *PermissionAttachment, parent *AttachmentInfo) {
	mgr := GetManager()
	for _, np := range children {
		perm := mgr.GetPermission(np.Name)
		value := np.Value != invert // XOR
		info := NewAttachmentInfo(np.Name, attachment, value, parent)
		p.permissions[np.Name] = info
		mgr.subscribeToPermission(np.Name, p)
		if perm != nil {
			p.calculateChildPermissionsLocked(mapToNamed(perm.Children()), !value, attachment, info)
		}
	}
}

func (p *Permissible) AddPermissionRecalculationCallback(cb func(changedPermissionsOldValues map[string]bool)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.recalculationCallbacks = append(p.recalculationCallbacks, cb)
}

func (p *Permissible) GetEffectivePermissions() map[string]*AttachmentInfo {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make(map[string]*AttachmentInfo, len(p.permissions))
	for k, v := range p.permissions {
		result[k] = v
	}
	return result
}

// Close unsubscribes this Permissible from the global Manager and all of its attachments. See the
// type doc comment: this is a hygiene measure, not a GC necessity.
func (p *Permissible) Close() {
	GetManager().unsubscribeFromAllPermissions(p)

	p.mu.Lock()
	p.permissions = map[string]*AttachmentInfo{}
	attachments := make([]*PermissionAttachment, 0, len(p.attachments))
	for a := range p.attachments {
		attachments = append(attachments, a)
	}
	p.attachments = map[*PermissionAttachment]struct{}{}
	p.recalculationCallbacks = nil
	p.mu.Unlock()

	for _, a := range attachments {
		a.unsubscribePermissible(p)
	}
}
