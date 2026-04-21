package collector

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/disk"

	"github.com/yusupkhemraev/argus/internal/config"
)

type DiskCollector struct {
	cfg config.DiskConfig
}

func NewDiskCollector(cfg config.DiskConfig) *DiskCollector {
	return &DiskCollector{cfg: cfg}
}

func (d *DiskCollector) Name() string { return "disk" }
func (d *DiskCollector) Start()       {}

func (d *DiskCollector) Collect() (Metric, error) {
	usage, err := disk.Usage(d.cfg.Path)
	if err != nil {
		return Metric{}, fmt.Errorf("disk %s: %w", d.cfg.Path, err)
	}

	return Metric{
		Collector: d.Name(),
		Value:     usage.UsedPercent,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"path":     d.cfg.Path,
			"total_gb": fmt.Sprintf("%.1f", float64(usage.Total)/1024/1024/1024),
			"used_gb":  fmt.Sprintf("%.1f", float64(usage.Used)/1024/1024/1024),
			"free_gb":  fmt.Sprintf("%.1f", float64(usage.Free)/1024/1024/1024),
		},
	}, nil
}

func (d *DiskCollector) Check(metric Metric) *Alarm {
	if metric.Value < d.cfg.Threshold {
		return nil
	}

	return &Alarm{
		ID:        fmt.Sprintf("disk_high_%s", d.cfg.Path),
		Collector: d.Name(),
		Message: fmt.Sprintf("Disk %s %.1f%% / %.1f%%", d.cfg.Path, metric.Value, d.cfg.Threshold),
		Severity:  Severity(d.cfg.Severity),
		Value:     metric.Value,
		Threshold: d.cfg.Threshold,
		Timestamp: metric.Timestamp,
	}
}
