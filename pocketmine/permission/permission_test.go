package permission

import (
	"testing"
	"time"
)

type fakePlugin struct {
	name    string
	enabled bool
}

func (p *fakePlugin) IsEnabled() bool { return p.enabled }
func (p *fakePlugin) Name() string    { return p.name }

func freshManagerForTest(t *testing.T) {
	t.Helper()
	m := GetManager()
	m.ClearPermissions()
}

func TestPermissibleBasePermissionAndChild(t *testing.T) {
	freshManagerForTest(t)
	defer freshManagerForTest(t)

	root := NewPermission("test.root", "root perm", nil)
	GetManager().AddPermission(root)
	child := NewPermission("test.root.child", "child perm", nil)
	GetManager().AddPermission(child)
	root.AddChild(child.Name(), true)

	p := NewPermissible(map[string]bool{"test.root": true})
	defer p.Close()

	if !p.HasPermission("test.root") {
		t.Fatalf("expected test.root to be granted")
	}
	if !p.HasPermission("test.root.child") {
		t.Fatalf("expected test.root.child to be granted via parent")
	}
}

func TestPermissibleDeniedRootInvertsChildren(t *testing.T) {
	freshManagerForTest(t)
	defer freshManagerForTest(t)

	root := NewPermission("test.root", "", nil)
	GetManager().AddPermission(root)
	child := NewPermission("test.child", "", nil)
	GetManager().AddPermission(child)
	root.AddChild(child.Name(), true)

	p := NewPermissible(map[string]bool{"test.root": false})
	defer p.Close()

	if p.HasPermission("test.root") {
		t.Fatalf("root should be denied")
	}
	if p.HasPermission("test.child") {
		t.Fatalf("child granted=true under a denied root should be inverted to denied")
	}
}

func TestPermissionAttachmentOverridesBase(t *testing.T) {
	freshManagerForTest(t)
	defer freshManagerForTest(t)

	root := NewPermission("test.root", "", nil)
	GetManager().AddPermission(root)

	p := NewPermissible(map[string]bool{"test.root": true})
	defer p.Close()

	plugin := &fakePlugin{name: "TestPlugin", enabled: true}
	value := false
	attachment, err := p.AddAttachment(plugin, "test.root", &value)
	if err != nil {
		t.Fatalf("AddAttachment() error = %v", err)
	}

	if p.HasPermission("test.root") {
		t.Fatalf("attachment should override base permission to false")
	}

	p.RemoveAttachment(attachment)
	if !p.HasPermission("test.root") {
		t.Fatalf("removing the attachment should restore the base permission")
	}
}

func TestAddAttachmentRejectsDisabledPlugin(t *testing.T) {
	freshManagerForTest(t)
	defer freshManagerForTest(t)

	p := NewPermissible(nil)
	defer p.Close()

	plugin := &fakePlugin{name: "Disabled", enabled: false}
	if _, err := p.AddAttachment(plugin, "", nil); err == nil {
		t.Fatalf("expected an error adding an attachment for a disabled plugin")
	}
}

func TestRecalculationCallbackReceivesOldValues(t *testing.T) {
	freshManagerForTest(t)
	defer freshManagerForTest(t)

	root := NewPermission("test.root", "", nil)
	GetManager().AddPermission(root)

	p := NewPermissible(map[string]bool{"test.root": true})
	defer p.Close()

	var lastDiff map[string]bool
	p.AddPermissionRecalculationCallback(func(diff map[string]bool) { lastDiff = diff })

	p.SetBasePermission("test.root", false)

	if lastDiff == nil {
		t.Fatalf("expected the recalculation callback to fire")
	}
	if old, ok := lastDiff["test.root"]; !ok || old != true {
		t.Fatalf("diff[test.root] = %v, %v, want true, true (the previous value)", old, ok)
	}
}

func TestBanEntryRoundTrip(t *testing.T) {
	entry := NewBanEntry("Griefer42")
	entry.SetReason("spamming")
	str := entry.String()

	parsed, err := BanEntryFromString(str)
	if err != nil {
		t.Fatalf("BanEntryFromString() error = %v", err)
	}
	if parsed.Name() != "griefer42" {
		t.Fatalf("Name() = %q, want griefer42", parsed.Name())
	}
	if parsed.Reason() != "spamming" {
		t.Fatalf("Reason() = %q, want spamming", parsed.Reason())
	}
	if parsed.Expires() != nil {
		t.Fatalf("Expires() should be nil for a non-expiring ban")
	}
}

func TestBanListSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/banned-players.txt"

	list := NewBanList(file)
	future := time.Now().Add(24 * time.Hour)
	list.AddBan("Notch", "testing", &future, "console")

	reloaded := NewBanList(file)
	reloaded.Load()

	if !reloaded.IsBanned("Notch") {
		t.Fatalf("expected Notch to be banned after reload")
	}
	entry := reloaded.GetEntry("notch")
	if entry == nil || entry.Reason() != "testing" {
		t.Fatalf("GetEntry() = %v, want reason=testing", entry)
	}
}

func TestPermissionParserDefaultFromString(t *testing.T) {
	v, err := DefaultFromString("isAdmin")
	if err != nil || v != DefaultOp {
		t.Fatalf("DefaultFromString(isAdmin) = %v, %v, want DefaultOp", v, err)
	}
	if _, err := DefaultFromString("bogus"); err == nil {
		t.Fatalf("expected an error for an unknown default name")
	}
}

func TestLoadPermissionsCarriesDefaultForward(t *testing.T) {
	entries := []PermissionEntry{
		{Name: "test.a", Default: "op"},
		{Name: "test.b"}, // no explicit default -> should inherit "op" from the previous entry
		{Name: "test.c", Default: "true"},
	}
	result, err := LoadPermissions(entries, DefaultFalse)
	if err != nil {
		t.Fatalf("LoadPermissions() error = %v", err)
	}
	if len(result[DefaultOp]) != 2 {
		t.Fatalf("expected 2 permissions in the %q bucket, got %d", DefaultOp, len(result[DefaultOp]))
	}
	if len(result[DefaultTrue]) != 1 {
		t.Fatalf("expected 1 permission in the %q bucket, got %d", DefaultTrue, len(result[DefaultTrue]))
	}
}
