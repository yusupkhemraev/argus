package alarm

import (
	"encoding/json"
	"time"

	"github.com/yusupkhemraev/argus/internal/bus"
	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
	"github.com/yusupkhemraev/argus/internal/history"
	"github.com/yusupkhemraev/argus/internal/notifier"
)

type Manager struct {
	collectors []collector.Collector
	notifiers  []notifier.Notifier
	history    *history.Writer
	bus        *bus.Bus
	alarmCh    chan collector.Alarm
	stopCh     chan struct{}
	cfg        config.Config
}

func NewManager(cfg config.Config, collectors []collector.Collector, notifiers []notifier.Notifier, hist *history.Writer, b *bus.Bus) *Manager {
	return &Manager{
		collectors: collectors,
		notifiers:  notifiers,
		history:    hist,
		bus:        b,
		alarmCh:    make(chan collector.Alarm, 100),
		stopCh:     make(chan struct{}),
		cfg:        cfg,
	}
}

func (m *Manager) Start() {
	for _, c := range m.collectors {
		c.Start()
		go m.runCollector(c)
	}
	go m.processAlarms()
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) runCollector(c collector.Collector) {
	interval := m.intervalFor(c.Name())
	refresh := m.refreshFor(c.Name())

	refreshTicker := time.NewTicker(refresh)
	defer refreshTicker.Stop()

	intervalTicker := time.NewTicker(interval)
	defer intervalTicker.Stop()

	m.collectAndCheck(c)

	for {
		select {
		case <-refreshTicker.C:
			m.collectMetric(c)
		case <-intervalTicker.C:
			m.collectAndCheck(c)
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) collectMetric(c collector.Collector) {
	metric, err := c.Collect()
	if err != nil {
		return
	}

	if m.bus == nil {
		return
	}

	data, _ := json.Marshal(metric)
	m.bus.Publish(bus.Event{
		Type:      "metric",
		Data:      data,
		Collector: c.Name(),
	})
}

func (m *Manager) collectAndCheck(c collector.Collector) {
	metric, err := c.Collect()
	if err != nil {
		return
	}

	if m.bus != nil {
		data, _ := json.Marshal(metric)
		m.bus.Publish(bus.Event{
			Type:      "metric",
			Data:      data,
			Collector: c.Name(),
		})
	}

	alarm := c.Check(metric)
	if alarm == nil {
		return
	}

	select {
	case m.alarmCh <- *alarm:
	default:
	}
}

func (m *Manager) processAlarms() {
	for {
		select {
		case alarm := <-m.alarmCh:
			if m.bus != nil {
				data, _ := json.Marshal(alarm)
				m.bus.Publish(bus.Event{Type: "alarm", Data: data})
			}

			if m.history != nil {
				m.history.Write(alarm)
			}

			for _, n := range m.notifiers {
				n.Send(alarm)
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) intervalFor(name string) time.Duration {
	switch name {
	case "memory":
		return m.cfg.Collectors.Memory.Interval
	case "cpu":
		return m.cfg.Collectors.CPU.Interval
	case "disk":
		return m.cfg.Collectors.Disk.Interval
	case "nginx":
		return m.cfg.Collectors.Nginx.Interval
	case "rabbitmq":
		return m.cfg.Collectors.RabbitMQ.Interval
	default:
		return 30 * time.Second
	}
}

func (m *Manager) refreshFor(name string) time.Duration {
	var d time.Duration
	switch name {
	case "memory":
		d = m.cfg.Collectors.Memory.Refresh
	case "cpu":
		d = m.cfg.Collectors.CPU.Refresh
	case "disk":
		d = m.cfg.Collectors.Disk.Refresh
	case "nginx":
		d = m.cfg.Collectors.Nginx.Refresh
	}
	if d <= 0 {
		d = 5 * time.Second
	}
	return d
}
