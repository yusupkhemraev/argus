package collector

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/yusupkhemraev/argus/internal/config"
)

type CPUCollector struct {
	cfg config.CPUConfig
}

func NewCPUCollector(cfg config.CPUConfig) *CPUCollector {
	return &CPUCollector{cfg: cfg}
}

func (c *CPUCollector) Name() string { return "cpu" }
func (c *CPUCollector) Start()       {}

func (c *CPUCollector) Collect() (Metric, error) {
	percents, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return Metric{}, fmt.Errorf("cpu: %w", err)
	}

	if len(percents) == 0 {
		return Metric{}, fmt.Errorf("cpu: no data returned")
	}

	cores, _ := cpu.Counts(true)

	return Metric{
		Collector: c.Name(),
		Value:     percents[0],
		Timestamp: time.Now(),
		Labels: map[string]string{
			"cores": fmt.Sprintf("%d", cores),
		},
	}, nil
}

func (c *CPUCollector) Check(metric Metric) *Alarm {
	if metric.Value < c.cfg.Threshold {
		return nil
	}

	return &Alarm{
		ID:        "cpu_high",
		Collector: c.Name(),
		Message:   fmt.Sprintf("CPU %.1f%% / %.1f%%", metric.Value, c.cfg.Threshold),
		Severity:  Severity(c.cfg.Severity),
		Value:     metric.Value,
		Threshold: c.cfg.Threshold,
		Timestamp: metric.Timestamp,
	}
}
