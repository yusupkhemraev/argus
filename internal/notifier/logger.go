package notifier

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yusupkhemraev/argus/internal/collector"
)

type LogEntry struct {
	Timestamp string `json:"ts"`
	Notifier  string `json:"notifier"`
	AlarmID   string `json:"alarm_id"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type LogPublisher func(entry LogEntry)

type NotificationLogger struct {
	inner   Notifier
	logFile *os.File
	mu      sync.Mutex
	publish LogPublisher
}

func NewNotificationLogger(inner Notifier, logPath string, publishers ...LogPublisher) *NotificationLogger {
	dir := filepath.Dir(logPath)
	os.MkdirAll(dir, 0755)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return &NotificationLogger{inner: inner}
	}

	nl := &NotificationLogger{inner: inner, logFile: f}
	if len(publishers) > 0 {
		nl.publish = publishers[0]
	}
	return nl
}

func (l *NotificationLogger) Send(alarm collector.Alarm) error {
	err := l.inner.Send(alarm)

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Notifier:  l.inner.Name(),
		AlarmID:   alarm.ID,
		Status:    "sent",
	}

	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
	}

	l.writeEntry(entry)
	if l.publish != nil {
		l.publish(entry)
	}

	return err
}

func (l *NotificationLogger) Name() string {
	return l.inner.Name()
}

func (l *NotificationLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *NotificationLogger) writeEntry(entry LogEntry) {
	if l.logFile == nil {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.logFile, "%s\n", data)
}

func ReadLogEntries(logPath string, n int, statusFilter string) ([]LogEntry, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var entries []LogEntry

	for _, line := range lines {
		if line == "" {
			continue
		}
		var e LogEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if statusFilter != "" && e.Status != statusFilter {
			continue
		}
		entries = append(entries, e)
	}

	start := len(entries) - n
	if start < 0 {
		start = 0
	}
	return entries[start:], nil
}
