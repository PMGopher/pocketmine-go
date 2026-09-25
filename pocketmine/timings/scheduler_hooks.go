package timings

import "pocketmine-go/pocketmine/scheduler"

func init() {
	scheduler.TimingsSetEnabledFunc = SetEnabled
	scheduler.TimingsReloadFunc = Reload
	scheduler.TimingsPrintCurrentThreadRecordsFunc = PrintCurrentThreadRecords
}
