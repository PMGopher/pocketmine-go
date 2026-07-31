package log

// Level is a port of the global \LogLevel interface from pocketmine/log (PSR-3-style level names).
type Level string

const (
	Emergency Level = "emergency"
	Alert     Level = "alert"
	Critical  Level = "critical"
	Error     Level = "error"
	Warning   Level = "warning"
	Notice    Level = "notice"
	Info      Level = "info"
	Debug     Level = "debug"
)
