package wizard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestWizard(dir, input string) (*SetupWizard, *bytes.Buffer) {
	out := &bytes.Buffer{}
	w := NewSetupWizardWithIO(dir, strings.NewReader(input), out)
	w.sleep = func(time.Duration) {}
	w.getIP = func() (string, bool) { return "1.2.3.4", true }
	w.getInternalIP = func() (string, error) { return "10.0.0.2", nil }
	return w, out
}

func TestSetupWizardFullRun(t *testing.T) {
	dir := t.TempDir()
	// language, license, don't skip, name, port v4 (invalid then valid), port v6, game mode
	// (invalid then creative), max players, view distance, op, whitelist, disable query, IP
	// confirmation.
	input := strings.Join([]string{"eng", "y", "n", "My Server", "70000", "19140", "", "5", "1", "10", "8", "Alex", "y", "y", ""}, "\n") + "\n"
	w, out := newTestWizard(dir, input)
	if !w.Run() {
		t.Fatalf("Run() = false; output:\n%s", out)
	}
	props, err := os.ReadFile(filepath.Join(dir, "server.properties"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"language=eng", "motd=My Server", "server-port=19140", "server-portv6=19133", "gamemode=CREATIVE", "max-players=10", "view-distance=8", "white-list=on", "enable-query=off"} {
		if !strings.Contains(string(props), want) {
			t.Errorf("server.properties is missing %q:\n%s", want, props)
		}
	}
	ops, _ := os.ReadFile(filepath.Join(dir, "ops.txt"))
	if strings.TrimSpace(string(ops)) != "alex" {
		t.Errorf("ops.txt = %q, want alex", ops)
	}
	if !strings.Contains(out.String(), "1.2.3.4") || !strings.Contains(out.String(), "[!]") {
		t.Errorf("output is missing the IP details:\n%s", out)
	}
}

func TestSetupWizardLicenseRefused(t *testing.T) {
	dir := t.TempDir()
	w, _ := newTestWizard(dir, "eng\n\n")
	if w.Run() {
		t.Error("Run() = true without accepting the license")
	}
	if _, err := os.Stat(filepath.Join(dir, "server.properties")); err == nil {
		t.Error("server.properties was written before the license was accepted")
	}
}

func TestSetupWizardSkip(t *testing.T) {
	dir := t.TempDir()
	w, _ := newTestWizard(dir, "nope\neng\ny\ny\n\n")
	if !w.Run() {
		t.Fatal("Run() = false")
	}
	props, _ := os.ReadFile(filepath.Join(dir, "server.properties"))
	if !strings.Contains(string(props), "language=eng") || strings.Contains(string(props), "motd") {
		t.Errorf("skipping should only write the language:\n%s", props)
	}
}
