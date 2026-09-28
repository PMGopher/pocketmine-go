package defaults

import (
	"os"
	"strings"
	"testing"

	"pocketmine-go/pocketmine/timings"
)

func TestTimingsCommand(t *testing.T) {
	s, sender := newTestServer(t)
	defer timings.SetEnabled(false)

	s.DispatchCommand(sender, "timings report", false)
	if msg := lastMessage(sender); !strings.Contains(strings.ToLower(msg), "timings") {
		t.Errorf("report while disabled = %q", msg)
	}

	s.DispatchCommand(sender, "timings on", false)
	if !timings.IsEnabled() {
		t.Fatal("/timings on didn't enable timings")
	}
	h := timings.NewTimingsHandler("Test Handler", nil, "Test Group")
	h.StartTiming()
	h.StopTiming()

	s.DispatchCommand(sender, "timings report", false)
	msg := lastMessage(sender)
	i := strings.Index(msg, s.GetDataPath())
	if i < 0 {
		t.Fatalf("report message = %q, want the report path", msg)
	}
	path := strings.TrimSpace(msg[i:])
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	report := string(data)
	for _, want := range []string{"Test Group", "    Test Handler Time: ", "Count: 1", "# FormatVersion 3", "Sample time "} {
		if !strings.Contains(report, want) {
			t.Errorf("report is missing %q:\n%s", want, report)
		}
	}

	s.DispatchCommand(sender, "timings off", false)
	if timings.IsEnabled() {
		t.Error("/timings off didn't disable timings")
	}
}
