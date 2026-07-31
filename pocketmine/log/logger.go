package log

// Logger is a port of the global \Logger interface from pocketmine/log.
type Logger interface {
	Emergency(message string)
	Alert(message string)
	Critical(message string)
	Error(message string)
	Warning(message string)
	Notice(message string)
	Info(message string)
	Debug(message string)
	Log(level Level, message string)
	LogException(err error, trace string)
}

// Attachment is a port of the LoggerAttachment closure type: `\Closure(mixed $level, string $message) : void`.
type Attachment func(level Level, message string)

// AttachmentHandle identifies a registered Attachment for later removal.
//
// PHP removes attachments by Closure object identity (spl_object_id); Go func values aren't
// comparable, so AddAttachment returns this opaque handle instead, and RemoveAttachment takes
// it back — the caller holds onto the handle the way a PHP caller would hold onto the Closure.
type AttachmentHandle int

// AttachableLogger is a port of \AttachableLogger.
type AttachableLogger interface {
	Logger
	AddAttachment(attachment Attachment) AttachmentHandle
	RemoveAttachment(handle AttachmentHandle)
	RemoveAttachments()
	GetAttachments() []Attachment
}

// BufferedLogger is a port of \BufferedLogger.
type BufferedLogger interface {
	Logger
	// Buffer runs buffered(), intended to batch a block of related log lines.
	//
	// PHP's synchronized() (used to implement this) is reentrant for the same thread, so
	// buffered() can safely call back into other logging methods on the same logger. Go's
	// sync.Mutex is not reentrant, so implementations here cannot hold a lock across the call
	// to buffered() without risking deadlock if it logs anything — see MainLogger.Buffer for
	// the resulting tradeoff.
	Buffer(buffered func())
}
