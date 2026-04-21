package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
)

type WebhookNotifier struct {
	cfg        config.WebhookConfig
	serverName string
	client     *http.Client
}

func NewWebhookNotifier(cfg config.WebhookConfig, serverName string) *WebhookNotifier {
	return &WebhookNotifier{
		cfg:        cfg,
		serverName: serverName,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookNotifier) Name() string { return "webhook" }

type webhookPayload struct {
	Server    string    `json:"server"`
	Collector string    `json:"collector"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

func (w *WebhookNotifier) Send(alarm collector.Alarm) error {
	payload := webhookPayload{
		Server:    w.serverName,
		Collector: alarm.Collector,
		Severity:  string(alarm.Severity),
		Message:   alarm.Message,
		Value:     alarm.Value,
		Threshold: alarm.Threshold,
		Timestamp: alarm.Timestamp,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	method := w.cfg.Method
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequest(method, w.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if w.cfg.APIKey != "" {
		header := w.cfg.APIKeyHeader
		if header == "" {
			header = "Authorization"
		}
		req.Header.Set(header, w.cfg.APIKey)
	}

	for k, v := range w.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: bad status: %d", resp.StatusCode)
	}

	return nil
}