package console

import (
	"strings"
	"testing"
	"time"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/permission"
)

func TestConsoleReaderStripsControlCharactersAndEmptyLines(t *testing.T) {
	r := NewConsoleReaderFrom(log.NewSimpleLogger(), strings.NewReader("\n  stop\x1b[A \n\tlist\n"))
	var got []string
	deadline := time.Now().Add(time.Second)
	for len(got) < 2 && time.Now().Before(deadline) {
		if line, ok := r.ReadLine(); ok {
			got = append(got, line)
		}
	}
	if len(got) != 2 || got[0] != "stop" || got[1] != "list" {
		t.Fatalf("got %q", got)
	}
	r.Quit()
}

func TestConsoleSenderHasConsolePermission(t *testing.T) {
	permission.RegisterCorePermissions()
	s := NewConsoleCommandSender(nil, nil)
	if !s.HasPermission(permission.RootOperator) {
		t.Fatalf("the console must have operator permissions")
	}
	if s.GetName() != "CONSOLE" {
		t.Fatalf("name = %q", s.GetName())
	}
}
