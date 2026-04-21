package notifier

import (
	"github.com/yusupkhemraev/argus/internal/collector"
)

type Notifier interface {
	Send(alarm collector.Alarm) error
	Name() string
}
