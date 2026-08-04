package plugin

import (
	"testing"

	"pocketmine-go/pocketmine/permission"
)

const minimalYAML = `
name: TestPlugin
version: "1.0.0"
main: testplugin\Main
`

func TestNewDescriptionFromYAMLParsesBasicFields(t *testing.T) {
	d, err := NewDescriptionFromYAML(minimalYAML)
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	if d.GetName() != "TestPlugin" {
		t.Errorf("GetName() = %q, want %q", d.GetName(), "TestPlugin")
	}
	if d.GetVersion() != "1.0.0" {
		t.Errorf("GetVersion() = %q, want %q", d.GetVersion(), "1.0.0")
	}
	if d.GetMain() != `testplugin\Main` {
		t.Errorf(`GetMain() = %q, want "testplugin\Main"`, d.GetMain())
	}
	if d.GetFullName() != "TestPlugin v1.0.0" {
		t.Errorf("GetFullName() = %q, want %q", d.GetFullName(), "TestPlugin v1.0.0")
	}
	if d.GetOrder() != EnableOrderPostworld {
		t.Errorf("GetOrder() = %v, want EnableOrderPostworld (the default)", d.GetOrder())
	}
}

func TestNewDescriptionFromYAMLReplacesSpacesInName(t *testing.T) {
	d, err := NewDescriptionFromYAML("name: My Test Plugin\nversion: \"1.0\"\nmain: a\\B\n")
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	if d.GetName() != "My_Test_Plugin" {
		t.Errorf("GetName() = %q, want %q", d.GetName(), "My_Test_Plugin")
	}
}

func TestNewDescriptionFromYAMLRejectsInvalidName(t *testing.T) {
	_, err := NewDescriptionFromYAML("name: \"bad/name\"\nversion: \"1.0\"\nmain: a\\B\n")
	if err == nil {
		t.Fatal("NewDescriptionFromYAML() error = nil, want an error for an invalid plugin name")
	}
}

func TestNewDescriptionFromYAMLRejectsMainInPocketmineNamespace(t *testing.T) {
	_, err := NewDescriptionFromYAML("name: Bad\nversion: \"1.0\"\nmain: pocketmine\\Evil\n")
	if err == nil {
		t.Fatal("NewDescriptionFromYAML() error = nil, want an error for main inside the pocketmine namespace")
	}
}

func TestNewDescriptionFromYAMLParsesCommands(t *testing.T) {
	yaml := minimalYAML + `
commands:
  test:
    description: "A test command"
    usage: "/test"
    aliases: [t, tst]
    permission: testplugin.test
`
	d, err := NewDescriptionFromYAML(yaml)
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	cmd, ok := d.GetCommands()["test"]
	if !ok {
		t.Fatal("GetCommands()[\"test\"] missing")
	}
	if cmd.Permission != "testplugin.test" {
		t.Errorf("Permission = %q, want %q", cmd.Permission, "testplugin.test")
	}
	if cmd.Description == nil || *cmd.Description != "A test command" {
		t.Errorf("Description = %v, want \"A test command\"", cmd.Description)
	}
	if len(cmd.Aliases) != 2 || cmd.Aliases[0] != "t" || cmd.Aliases[1] != "tst" {
		t.Errorf("Aliases = %v, want [t tst]", cmd.Aliases)
	}
}

func TestNewDescriptionFromYAMLCommandMissingPermissionIsAnError(t *testing.T) {
	yaml := minimalYAML + "\ncommands:\n  test:\n    description: x\n"
	if _, err := NewDescriptionFromYAML(yaml); err == nil {
		t.Fatal("NewDescriptionFromYAML() error = nil, want an error for a command missing its permission")
	}
}

func TestNewDescriptionFromYAMLPermissionsCarryDefaultForward(t *testing.T) {
	// Real PocketMine-MP behaviour: a "default" on one permission entry carries forward to every
	// later entry that doesn't specify its own, and entries are bucketed by that resolved
	// default - this is the exact behaviour LoadPermissions' own doc comment says is
	// order-dependent, and why PluginDescription must parse permissions in file order.
	yaml := minimalYAML + `
permissions:
  testplugin.first:
    default: true
    description: "first permission"
  testplugin.second:
    description: "inherits true from above"
  testplugin.third:
    default: op
    description: "explicit op"
`
	d, err := NewDescriptionFromYAML(yaml)
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	perms := d.GetPermissions()

	trueBucket := perms[permission.DefaultTrue]
	if len(trueBucket) != 2 {
		t.Fatalf("perms[%q] has %d entries, want 2 (first + second inheriting the default)", permission.DefaultTrue, len(trueBucket))
	}
	names := map[string]bool{}
	for _, p := range trueBucket {
		names[p.Name()] = true
	}
	if !names["testplugin.first"] || !names["testplugin.second"] {
		t.Errorf("perms[%q] = %v, want testplugin.first and testplugin.second", permission.DefaultTrue, names)
	}

	opBucket := perms[permission.DefaultOp]
	if len(opBucket) != 1 || opBucket[0].Name() != "testplugin.third" {
		t.Errorf("perms[%q] = %v, want just testplugin.third", permission.DefaultOp, opBucket)
	}
}

func TestNewDescriptionFromYAMLExtensionsListForm(t *testing.T) {
	yaml := minimalYAML + "\nextensions:\n  - mbstring\n  - curl\n"
	d, err := NewDescriptionFromYAML(yaml)
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	ext := d.GetRequiredExtensions()
	if got := ext["mbstring"]; len(got) != 1 || got[0] != "*" {
		t.Errorf("GetRequiredExtensions()[\"mbstring\"] = %v, want [*]", got)
	}
	if got := ext["curl"]; len(got) != 1 || got[0] != "*" {
		t.Errorf("GetRequiredExtensions()[\"curl\"] = %v, want [*]", got)
	}
}

func TestNewDescriptionFromYAMLAuthorAndAuthorsAreMerged(t *testing.T) {
	yaml := minimalYAML + "\nauthor: Alice\nauthors: [Bob, Carol]\n"
	d, err := NewDescriptionFromYAML(yaml)
	if err != nil {
		t.Fatalf("NewDescriptionFromYAML() error = %v", err)
	}
	want := []string{"Alice", "Bob", "Carol"}
	got := d.GetAuthors()
	if len(got) != len(want) {
		t.Fatalf("GetAuthors() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetAuthors()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
