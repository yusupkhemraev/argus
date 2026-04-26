package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
)

type Record struct {
	Timestamp string  `json:"ts"`
	Severity  string  `json:"severity"`
	Collector string  `json:"collector"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Resolved  bool    `json:"resolved,omitempty"`
}

type Writer struct {
	cfg  config.HistoryConfig
	file *os.File
	mu   sync.Mutex
}

func New(cfg config.HistoryConfig) (*Writer, error) {
	if !cfg.Enabled {
		return &Writer{cfg: cfg}, nil
	}

	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("history: cannot create dir %s: %w", dir, err)
	}

	f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("history: cannot open %s: %w", cfg.FilePath, err)
	}

	return &Writer{cfg: cfg, file: f}, nil
}

func (w *Writer) Write(alarm collector.Alarm) error {
	if !w.cfg.Enabled || w.file == nil {
		return nil
	}

	record := Record{
		Timestamp: alarm.Timestamp.UTC().Format(time.RFC3339),
		Severity:  string(alarm.Severity),
		Collector: alarm.Collector,
		Message:   alarm.Message,
		Value:     alarm.Value,
		Threshold: alarm.Threshold,
		Resolved:  alarm.Resolved,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	_, err = fmt.Fprintf(w.file, "%s\n", data)
	return err
}

func (w *Writer) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// ReadRecords reads up to n records from filePath, optionally filtered by severity.
func ReadRecords(filePath string, n int, severity string) ([]Record, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r Record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		if severity != "" && r.Severity != severity {
			continue
		}
		records = append(records, r)
	}

	start := len(records) - n
	if start < 0 {
		start = 0
	}
	return records[start:], nil
}
