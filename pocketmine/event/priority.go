package event

import (
	"fmt"
	"strings"
)

// Priority is a port of pocketmine\event\EventPriority.
//
// Events are called in this order: Lowest -> Low -> Normal -> High -> Highest -> Monitor.
// Monitor handlers should not change the event outcome or contents.
type Priority int

const (
	Monitor Priority = 0
	Highest Priority = 1
	High    Priority = 2
	Normal  Priority = 3
	Low     Priority = 4
	Lowest  Priority = 5
)

// AllPriorities lists every valid Priority, in call order (Lowest first, Monitor last).
var AllPriorities = []Priority{Lowest, Low, Normal, High, Highest, Monitor}

func PriorityFromString(name string) (Priority, error) {
	switch strings.ToUpper(name) {
	case "LOWEST":
		return Lowest, nil
	case "LOW":
		return Low, nil
	case "NORMAL":
		return Normal, nil
	case "HIGH":
		return High, nil
	case "HIGHEST":
		return Highest, nil
	case "MONITOR":
		return Monitor, nil
	default:
		return 0, fmt.Errorf("unable to resolve priority %q", name)
	}
}
