package bus

import (
	"encoding/json"
	"sync"
)

type Event struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Collector string          `json:"-"`
}

type Snapshot struct {
	Metrics map[string]json.RawMessage
}

type Bus struct {
	mu       sync.RWMutex
	subs     map[string]chan Event
	snapshot Snapshot
}

func New() *Bus {
	return &Bus{
		subs: make(map[string]chan Event),
		snapshot: Snapshot{
			Metrics: make(map[string]json.RawMessage),
		},
	}
}

func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	if e.Type == "metric" && e.Collector != "" {
		b.snapshot.Metrics[e.Collector] = e.Data
	}
	subs := make([]chan Event, 0, len(b.subs))
	for _, ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (b *Bus) Subscribe(id string) <-chan Event {
	ch := make(chan Event, 100)
	b.mu.Lock()
	b.subs[id] = ch
	b.mu.Unlock()
	return ch
}

func (b *Bus) Unsubscribe(id string) {
	b.mu.Lock()
	delete(b.subs, id)
	b.mu.Unlock()
}

func (b *Bus) Reset() {
	b.mu.Lock()
	b.snapshot.Metrics = make(map[string]json.RawMessage)
	b.mu.Unlock()
}

func (b *Bus) Snapshot() Snapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cp := Snapshot{Metrics: make(map[string]json.RawMessage, len(b.snapshot.Metrics))}
	for k, v := range b.snapshot.Metrics {
		cp.Metrics[k] = v
	}
	return cp
}
