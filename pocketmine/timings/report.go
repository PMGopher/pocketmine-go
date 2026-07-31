package timings

import (
	"fmt"
	"sort"
	"time"
)

// PrintCurrentThreadRecords is a port of TimingsHandler::printCurrentThreadRecords().
//
// PHP suffixes each group name with " ThreadId: N" when running on a pthreads worker thread.
// This port has no equivalent worker-thread model (see the scheduler package's doc comment —
// Go's goroutines replace it), so that suffix is simply omitted; there's only "the" thread.
func PrintCurrentThreadRecords() []string {
	groups := map[string][]string{}
	var groupOrder []string

	for _, r := range GetAllRecords() {
		count := r.GetCount()
		if count == 0 {
			// this should never happen - a timings record shouldn't exist if it hasn't been used
			continue
		}
		total := r.GetTotalTime()
		avg := total / time.Duration(count)

		group := r.GetGroup()
		if _, exists := groups[group]; !exists {
			groupOrder = append(groupOrder, group)
		}

		parentID, hasParent := r.GetParentID()
		parentStr := "none"
		if hasParent {
			parentStr = fmt.Sprintf("%d", parentID)
		}

		groups[group] = append(groups[group], fmt.Sprintf(
			"%s Time: %d Count: %d Avg: %v Violations: %d RecordId: %d ParentRecordId: %s TimerId: %d Ticks: %d Peak: %d",
			r.GetName(), total.Nanoseconds(), count, avg, r.GetViolations(), r.GetID(), parentStr, r.GetTimerID(), r.GetTicksActive(), r.GetPeakTime().Nanoseconds(),
		))
	}

	sort.Strings(groupOrder)

	var result []string
	for _, group := range groupOrder {
		result = append(result, group)
		for _, line := range groups[group] {
			result = append(result, "    "+line)
		}
	}
	return result
}
