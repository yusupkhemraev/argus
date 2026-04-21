package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
)

type TelegramNotifier struct {
	cfg        config.TelegramConfig
	serverName string
	mentions   []string
	client     *http.Client
}

func NewTelegramNotifier(cfg config.TelegramConfig, serverName string) *TelegramNotifier {
	return &TelegramNotifier{
		cfg:        cfg,
		serverName: serverName,
		mentions:   cfg.Mentions,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TelegramNotifier) Name() string {
	return "telegram"
}

func (t *TelegramNotifier) Send(alarm collector.Alarm) error {
	text := formatMessage(alarm, t.serverName, t.mentions)

	payload := struct {
		ChatID    int64  `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		ChatID:    t.cfg.ChatID,
		Text:      text,
		ParseMode: "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.cfg.Token)

	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: bad status: %d", resp.StatusCode)
	}

	return nil
}

func severityIcon(s collector.Severity) string {
	switch s {
	case collector.SeverityCritical:
		return "🔴"
	case collector.SeverityWarning:
		return "⚠️"
	default:
		return "ℹ️"
	}
}

func collectorTitle(name string) string {
	switch name {
	case "memory":
		return "Argus Memory Alert"
	case "cpu":
		return "Argus CPU Alert"
	case "disk":
		return "Argus Disk Alert"
	case "nginx":
		return "Argus Nginx Alert"
	case "rabbitmq":
		return "Argus RabbitMQ Alert"
	default:
		return "Argus " + strings.ToUpper(name[:1]) + name[1:] + " Alert"
	}
}

func formatMessage(alarm collector.Alarm, serverName string, mentions []string) string {
	var b strings.Builder

	switch alarm.Collector {
	case "nginx":
		b.WriteString(fmt.Sprintf("%s <b>%s</b> • %s\n\n", severityIcon(alarm.Severity), collectorTitle(alarm.Collector), serverName))
		if alarm.Message != "" {
			b.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>", alarm.Message))
		}
	default:
		b.WriteString(fmt.Sprintf("%s <b>%s</b>\n\n", severityIcon(alarm.Severity), collectorTitle(alarm.Collector)))
		b.WriteString("<blockquote>")
		b.WriteString(fmt.Sprintf("🖥 Server: <b>%s</b>\n", serverName))
		b.WriteString(fmt.Sprintf("📊 Usage: <b>%.1f%%</b>\n", alarm.Value))
		b.WriteString(fmt.Sprintf("⚠️ Threshold: <b>%.1f%%</b>", alarm.Threshold))
		b.WriteString("</blockquote>")
	}

	if len(mentions) > 0 {
		tags := make([]string, len(mentions))
		for i, m := range mentions {
			if m != "" && m[0] != '@' {
				tags[i] = "@" + m
			} else {
				tags[i] = m
			}
		}
		b.WriteString("\n<blockquote>")
		b.WriteString(strings.Join(tags, " "))
		b.WriteString("</blockquote>")
	}

	return b.String()
}
