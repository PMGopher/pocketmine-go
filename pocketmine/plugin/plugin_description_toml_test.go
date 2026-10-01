package plugin

import (
	"testing"

	"pocketmine-go/pocketmine/permission"
)

const minimalTOML = `
name = "TestPlugin"
version = "1.0.0"
main = "testplugin.Main"
api = ["5.0.0"]
authors = ["Alice", "Bob"]
depend = ["Other"]
load = "POSTWORLD"
`

func TestNewDescriptionFromTOML(t *testing.T) {
	d, err := NewDescriptionFromTOML(minimalTOML + `
[commands.hello]
description = "Says hello"
usage = "/hello"
aliases = ["hi"]
permission = "testplugin.zeta"
`)
	if err != nil {
		t.Fatal(err)
	}
	if d.GetName() != "TestPlugin" || d.GetVersion() != "1.0.0" || d.GetMain() != "testplugin.Main" {
		t.Errorf("name/version/main = %q %q %q", d.GetName(), d.GetVersion(), d.GetMain())
	}
	if api := d.GetCompatibleApis(); len(api) != 1 || api[0] != "5.0.0" {
		t.Errorf("api = %v", api)
	}
	if a := d.GetAuthors(); len(a) != 2 || a[1] != "Bob" {
		t.Errorf("authors = %v", a)
	}
	if dep := d.GetDepend(); len(dep) != 1 || dep[0] != "Other" {
		t.Errorf("depend = %v", dep)
	}
	cmd, ok := d.GetCommands()["hello"]
	if !ok || cmd.Description == nil || *cmd.Description != "Says hello" || len(cmd.Aliases) != 1 || cmd.Permission != "testplugin.zeta" {
		t.Errorf("command hello = %+v", cmd)
	}
}

// A "default" carries forward to the permissions declared after it, so the TOML parser must keep
// the declaration order (here it differs from the alphabetical order).
func TestNewDescriptionFromTOMLPermissionsCarryDefaultForward(t *testing.T) {
	d, err := NewDescriptionFromTOML(minimalTOML + `
[permissions."testplugin.zeta"]
default = true
description = "first permission"

[permissions."testplugin.alpha"]
description = "inherits true from above"

[permissions."testplugin.mid"]
default = "op"
`)
	if err != nil {
		t.Fatal(err)
	}
	perms := d.GetPermissions()
	names := map[string]bool{}
	for _, p := range perms[permission.DefaultTrue] {
		names[p.Name()] = true
	}
	if len(names) != 2 || !names["testplugin.zeta"] || !names["testplugin.alpha"] {
		t.Errorf("perms[true] = %v, want zeta and alpha (alpha inherits zeta's default)", names)
	}
	if op := perms[permission.DefaultOp]; len(op) != 1 || op[0].Name() != "testplugin.mid" {
		t.Errorf("perms[op] = %v, want just testplugin.mid", op)
	}
}

func TestNewDescriptionFromTOMLRejectsBadTOML(t *testing.T) {
	if _, err := NewDescriptionFromTOML("name = \n"); err == nil {
		t.Error("invalid TOML accepted")
	}
}
