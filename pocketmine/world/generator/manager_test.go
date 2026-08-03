package generator

import "testing"

func TestGetFactoryFindsNormalAndItsDefaultAlias(t *testing.T) {
	for _, name := range []string{"normal", "default"} {
		factory, ok := GetFactory(name)
		if !ok {
			t.Fatalf("GetFactory(%q) not found", name)
		}
		gen, err := factory(42, "")
		if err != nil {
			t.Fatalf("factory(%q) returned an error: %v", name, err)
		}
		if _, ok := gen.(*Normal); !ok {
			t.Errorf("factory(%q) returned a %T, want *Normal", name, gen)
		}
	}
}

func TestGetFactoryReturnsFalseForAnUnregisteredName(t *testing.T) {
	if _, ok := GetFactory("some-unregistered-generator"); ok {
		t.Error("GetFactory found a factory for a name nothing registered")
	}
}

func TestRegisterGeneratorMakesANewNameFindable(t *testing.T) {
	const name = "test-only-generator"
	RegisterGenerator(name, func(seed int64, options string) (Generator, error) {
		return NewNormal(int(seed)), nil
	})

	factory, ok := GetFactory(name)
	if !ok {
		t.Fatal("GetFactory did not find the just-registered generator")
	}
	if _, err := factory(1, ""); err != nil {
		t.Errorf("registered factory returned an error: %v", err)
	}
}

func TestUnknownGeneratorErrorMessageIncludesTheName(t *testing.T) {
	err := &UnknownGeneratorError{Name: "some-name"}
	if got := err.Error(); got == "" {
		t.Fatal("Error() returned an empty string")
	}
}
