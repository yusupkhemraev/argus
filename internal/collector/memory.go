package collector

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/yusupkhemraev/argus/internal/config"
)

type MemoryCollector struct {
	cfg config.MemoryConfig
}

func NewMemoryCollector(cfg config.MemoryConfig) *MemoryCollector {
	return &MemoryCollector{cfg: cfg}
}

func (m *MemoryCollector) Name() string { return "memory" }
func (m *MemoryCollector) Start()       {}

func (m *MemoryCollector) Collect() (Metric, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return Metric{}, fmt.Errorf("memory: %w", err)
	}

	return Metric{
		Collector: m.Name(),
		Value:     stat.UsedPercent,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"total_gb": fmt.Sprintf("%.1f", float64(stat.Total)/1024/1024/1024),
			"used_gb":  fmt.Sprintf("%.1f", float64(stat.Used)/1024/1024/1024),
			"free_gb":  fmt.Sprintf("%.1f", float64(stat.Available)/1024/1024/1024),
		},
	}, nil
}

func (m *MemoryCollector) Check(metric Metric) *Alarm {
	if metric.Value < m.cfg.Threshold {
		return nil
	}

	return &Alarm{
		ID:        "memory_high",
		Collector: m.Name(),
		Message:   fmt.Sprintf("Memory %.1f%% / %.1f%%", metric.Value, m.cfg.Threshold),
		Severity:  Severity(m.cfg.Severity),
		Value:     metric.Value,
		Threshold: m.cfg.Threshold,
		Timestamp: metric.Timestamp,
	}
}
