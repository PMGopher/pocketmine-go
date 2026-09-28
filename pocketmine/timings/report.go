package timings

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
		total := r.GetTotalTime().Nanoseconds()

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
			"%s Time: %d Count: %d Avg: %s Violations: %d RecordId: %d ParentRecordId: %s TimerId: %d Ticks: %d Peak: %d",
			r.GetName(), total, count, phpDivide(total, int64(count)), r.GetViolations(), r.GetID(), parentStr, r.GetTimerID(), r.GetTicksActive(), r.GetPeakTime().Nanoseconds(),
		))
	}

	var result []string
	for _, group := range groupOrder {
		result = append(result, group)
		for _, line := range groups[group] {
			result = append(result, "    "+line)
		}
	}
	return result
}

// phpDivide is `(string) ($a / $b)` for the average: PHP's `/` gives an int when the division is
// exact and a float otherwise.
func phpDivide(a, b int64) string {
	if a%b == 0 {
		return strconv.FormatInt(a/b, 10)
	}
	return strconv.FormatFloat(float64(a)/float64(b), 'G', 14, 64)
}

// formatVersion is TimingsHandler::FORMAT_VERSION.
const formatVersion = 3 //thread timings collection

// ServerInfoFunc gives printFooter what it reads from Server::getInstance(): getVersion(),
// getName() and getPocketMineVersion(). It's set by the server package (which imports this one).
var ServerInfoFunc func() (version, name, pocketMineVersion string)

// printFooter is a port of TimingsHandler::printFooter.
func printFooter() []string {
	var result []string

	version, name, pocketMineVersion := "", "", ""
	if ServerInfoFunc != nil {
		version, name, pocketMineVersion = ServerInfoFunc()
	}
	result = append(result, "# Version "+version)
	result = append(result, "# "+name+" "+pocketMineVersion)

	result = append(result, "# FormatVersion "+strconv.Itoa(formatVersion))

	sampleTime := time.Since(GetStartTime()).Nanoseconds()
	result = append(result, fmt.Sprintf("Sample time %d (%ss)", sampleTime, strconv.FormatFloat(float64(sampleTime)/1000000000, 'G', 14, 64)))

	return result
}

// PrintTimings is a port of TimingsHandler::printTimings: this thread's records and the footer.
func PrintTimings() []string {
	return append(PrintCurrentThreadRecords(), printFooter()...)
}

// RequestPrintTimings is a port of TimingsHandler::requestPrintTimings. PHP also collects the
// records of the async worker threads (getCollectCallbacks) and returns a promise; the async
// pool's goroutines record into the same records as the main thread here, so there's nothing
// else to wait for and the lines are returned directly.
func RequestPrintTimings() []string {
	return PrintTimings()
}

// CreateReportFile is a port of TimingsHandler::createReportFile: writes the timings report to
// directory/fileName.txt (fileName defaults to timings_<date>) and returns the file's path.
func CreateReportFile(directory string, fileName string) (string, error) {
	lines := RequestPrintTimings()
	if fileName == "" {
		fileName = "timings_" + time.Now().Format("2006-01-02_15.04.05_MST")
	}
	if err := os.MkdirAll(directory, 0o777); err != nil {
		return "", err
	}
	timingsFile := filepath.Join(directory, fileName+".txt")
	handle, err := os.OpenFile(timingsFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		return "", err
	}
	defer handle.Close()
	for _, line := range lines {
		if _, err := handle.WriteString(line + phpEOL); err != nil {
			return "", err
		}
	}
	return timingsFile, nil
}

// phpEOL is PHP_EOL.
var phpEOL = func() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}()
