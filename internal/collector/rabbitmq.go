package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yusupkhemraev/argus/internal/config"
)

type RabbitMQCollector struct {
	cfg    config.RabbitMQConfig
	client *http.Client
}

func NewRabbitMQCollector(cfg config.RabbitMQConfig) *RabbitMQCollector {
	return &RabbitMQCollector{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *RabbitMQCollector) Name() string { return "rabbitmq" }
func (r *RabbitMQCollector) Start()       {}

type rabbitQueue struct {
	Name                   string `json:"name"`
	Messages               int64  `json:"messages"`
	MessagesUnacknowledged int64  `json:"messages_unacknowledged"`
}

func (r *RabbitMQCollector) fetchQueues() ([]rabbitQueue, error) {
	url := r.cfg.ManagementURL + "/api/queues"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: %w", err)
	}
	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rabbitmq: management API returned %d", resp.StatusCode)
	}

	var queues []rabbitQueue
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return nil, fmt.Errorf("rabbitmq: decode: %w", err)
	}
	return queues, nil
}

func (r *RabbitMQCollector) Collect() (Metric, error) {
	queues, err := r.fetchQueues()
	if err != nil {
		return Metric{}, err
	}

	queueMap := make(map[string]rabbitQueue, len(queues))
	for _, q := range queues {
		queueMap[q.Name] = q
	}

	var maxMessages int64
	labels := make(map[string]string)

	for _, qcfg := range r.cfg.Queues {
		q, ok := queueMap[qcfg.Name]
		if !ok {
			labels[qcfg.Name] = "not found"
			continue
		}
		labels[qcfg.Name] = fmt.Sprintf("%d msg / %d unacked", q.Messages, q.MessagesUnacknowledged)
		if q.Messages > maxMessages {
			maxMessages = q.Messages
		}
	}

	return Metric{
		Collector: r.Name(),
		Value:     float64(maxMessages),
		Timestamp: time.Now(),
		Labels:    labels,
	}, nil
}

func (r *RabbitMQCollector) Check(metric Metric) *Alarm {
	queues, err := r.fetchQueues()
	if err != nil {
		return nil
	}

	queueMap := make(map[string]rabbitQueue, len(queues))
	for _, q := range queues {
		queueMap[q.Name] = q
	}

	var lines []string
	worstSev := Severity("")
	var worstValue, worstThreshold float64

	for _, qcfg := range r.cfg.Queues {
		q, ok := queueMap[qcfg.Name]
		if !ok {
			continue
		}

		msgTriggered := qcfg.Threshold > 0 && q.Messages >= qcfg.Threshold
		unackedTriggered := qcfg.UnackedThreshold > 0 && q.MessagesUnacknowledged >= qcfg.UnackedThreshold

		if !msgTriggered && !unackedTriggered {
			continue
		}

		sev := Severity(qcfg.Severity)
		if sev == "" {
			sev = SeverityWarning
		}

		if worstSev == "" || SeverityWeight(sev) > SeverityWeight(worstSev) {
			worstSev = sev
			worstValue = float64(q.Messages)
			worstThreshold = float64(qcfg.Threshold)
		}

		if unackedTriggered {
			lines = append(lines, fmt.Sprintf("%s: %d unacked / %d", qcfg.Name, q.MessagesUnacknowledged, qcfg.UnackedThreshold))
		}
		if msgTriggered {
			lines = append(lines, fmt.Sprintf("%s: %d messages / %d", qcfg.Name, q.Messages, qcfg.Threshold))
		}
	}

	if len(lines) == 0 {
		return nil
	}

	return &Alarm{
		ID:        "rabbitmq_queues",
		Collector: r.Name(),
		Message:   strings.Join(lines, "\n"),
		Severity:  worstSev,
		Value:     worstValue,
		Threshold: worstThreshold,
		Timestamp: metric.Timestamp,
	}
}
