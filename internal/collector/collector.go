package collector

import "time"

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type Metric struct {
	Collector string            `json:"collector"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

type Alarm struct {
	ID        string   `json:"id"`
	Collector string   `json:"collector"`
	Message   string   `json:"message"`
	Severity  Severity `json:"severity"`
	Value     float64  `json:"value"`
	Threshold float64  `json:"threshold"`
	Timestamp time.Time `json:"triggered_at"`
	Resolved  bool     `json:"resolved"`
}

type Collector interface {
	Name() string
	Start()
	Collect() (Metric, error)
	Check(m Metric) *Alarm
}

func SeverityWeight(s Severity) int {
	switch s {
	case SeverityInfo:
		return 1
	case SeverityWarning:
		return 2
	case SeverityCritical:
		return 3
	default:
		return 0
	}
}
