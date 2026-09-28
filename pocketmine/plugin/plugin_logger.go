package plugin

import (
	"sync"

	"pocketmine-go/pocketmine/log"
)

// PluginLogger is a port of pocketmine\plugin\PluginLogger: a PrefixedLogger with attachments.
type PluginLogger struct {
	*log.PrefixedLogger

	mu          sync.Mutex
	attachments map[log.AttachmentHandle]log.Attachment
	order       []log.AttachmentHandle
	nextHandle  log.AttachmentHandle
}

func NewPluginLogger(delegate log.Logger, prefix string) *PluginLogger {
	return &PluginLogger{PrefixedLogger: log.NewPrefixedLogger(delegate, prefix), attachments: map[log.AttachmentHandle]log.Attachment{}}
}

func (l *PluginLogger) AddAttachment(attachment log.Attachment) log.AttachmentHandle {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextHandle++
	l.attachments[l.nextHandle] = attachment
	l.order = append(l.order, l.nextHandle)
	return l.nextHandle
}

func (l *PluginLogger) RemoveAttachment(handle log.AttachmentHandle) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attachments, handle)
	for i, h := range l.order {
		if h == handle {
			l.order = append(l.order[:i], l.order[i+1:]...)
			break
		}
	}
}

func (l *PluginLogger) RemoveAttachments() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attachments = map[log.AttachmentHandle]log.Attachment{}
	l.order = nil
}

func (l *PluginLogger) GetAttachments() []log.Attachment {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]log.Attachment, 0, len(l.order))
	for _, h := range l.order {
		result = append(result, l.attachments[h])
	}
	return result
}

func (l *PluginLogger) Emergency(m string) { l.Log(log.Emergency, m) }
func (l *PluginLogger) Alert(m string)     { l.Log(log.Alert, m) }
func (l *PluginLogger) Critical(m string)  { l.Log(log.Critical, m) }
func (l *PluginLogger) Error(m string)     { l.Log(log.Error, m) }
func (l *PluginLogger) Warning(m string)   { l.Log(log.Warning, m) }
func (l *PluginLogger) Notice(m string)    { l.Log(log.Notice, m) }
func (l *PluginLogger) Info(m string)      { l.Log(log.Info, m) }
func (l *PluginLogger) Debug(m string)     { l.Log(log.Debug, m) }

func (l *PluginLogger) Log(level log.Level, message string) {
	l.PrefixedLogger.Log(level, message)
	for _, attachment := range l.GetAttachments() {
		attachment(level, message)
	}
}

var _ log.AttachableLogger = (*PluginLogger)(nil)
