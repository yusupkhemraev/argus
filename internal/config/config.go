package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Name       string           `mapstructure:"name" yaml:"name" json:"name"`
	Collectors CollectorsConfig `mapstructure:"collectors" yaml:"collectors" json:"collectors"`
	Notifiers  NotifiersConfig  `mapstructure:"notifiers" yaml:"notifiers" json:"notifiers"`
	History    HistoryConfig    `mapstructure:"history" yaml:"history" json:"history"`
	Server     ServerConfig     `mapstructure:"server" yaml:"server" json:"server"`
}

type CollectorsConfig struct {
	Memory   MemoryConfig   `mapstructure:"memory" yaml:"memory" json:"memory"`
	CPU      CPUConfig      `mapstructure:"cpu" yaml:"cpu" json:"cpu"`
	Disk     DiskConfig     `mapstructure:"disk" yaml:"disk" json:"disk"`
	Nginx    NginxConfig    `mapstructure:"nginx" yaml:"nginx" json:"nginx"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq" yaml:"rabbitmq" json:"rabbitmq"`
}

type MemoryConfig struct {
	Enabled   bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Interval  time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	Refresh   time.Duration `mapstructure:"refresh" yaml:"refresh" json:"refresh"`
	Threshold float64       `mapstructure:"threshold" yaml:"threshold" json:"threshold"`
	Severity  string        `mapstructure:"severity" yaml:"severity" json:"severity"`
}

type CPUConfig struct {
	Enabled   bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Interval  time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	Refresh   time.Duration `mapstructure:"refresh" yaml:"refresh" json:"refresh"`
	Threshold float64       `mapstructure:"threshold" yaml:"threshold" json:"threshold"`
	Severity  string        `mapstructure:"severity" yaml:"severity" json:"severity"`
}

type DiskConfig struct {
	Enabled   bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Interval  time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	Refresh   time.Duration `mapstructure:"refresh" yaml:"refresh" json:"refresh"`
	Path      string        `mapstructure:"path" yaml:"path" json:"path"`
	Threshold float64       `mapstructure:"threshold" yaml:"threshold" json:"threshold"`
	Severity  string        `mapstructure:"severity" yaml:"severity" json:"severity"`
}

type NginxConfig struct {
	Enabled        bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	AccessLog      string          `mapstructure:"access_log" yaml:"access_log" json:"access_log"`
	WatchStatuses  []string        `mapstructure:"watch_statuses" yaml:"watch_statuses" json:"watch_statuses"`
	Interval       time.Duration   `mapstructure:"interval" yaml:"interval" json:"interval"`
	Refresh        time.Duration   `mapstructure:"refresh" yaml:"refresh" json:"refresh"`
	Window         time.Duration   `mapstructure:"window" yaml:"window" json:"window"`
	Threshold      int             `mapstructure:"threshold" yaml:"threshold" json:"threshold"`
	Severity       string          `mapstructure:"severity" yaml:"severity" json:"severity"`
	SlowThreshold  float64         `mapstructure:"slow_threshold" yaml:"slow_threshold,omitempty" json:"slow_threshold,omitempty"`
	SlowWindow     time.Duration   `mapstructure:"slow_window" yaml:"slow_window,omitempty" json:"slow_window,omitempty"`
	SlowCount      int             `mapstructure:"slow_count" yaml:"slow_count,omitempty" json:"slow_count,omitempty"`
	MinGroupCount  int             `mapstructure:"min_group_count" yaml:"min_group_count,omitempty" json:"min_group_count,omitempty"`
	PriorityRoutes []PriorityRoute `mapstructure:"priority_routes" yaml:"priority_routes,omitempty" json:"priority_routes,omitempty"`
	IgnoreRoutes   []IgnoreRoute   `mapstructure:"ignore_routes" yaml:"ignore_routes,omitempty" json:"ignore_routes,omitempty"`
}

type IgnoreRoute struct {
	Method  string `mapstructure:"method" yaml:"method" json:"method"`
	Pattern string `mapstructure:"pattern" yaml:"pattern" json:"pattern"`
}

type PriorityRoute struct {
	Method          string   `mapstructure:"method" yaml:"method" json:"method"`
	Pattern         string   `mapstructure:"pattern" yaml:"pattern" json:"pattern"`
	MinCount        int      `mapstructure:"min_count" yaml:"min_count" json:"min_count"`
	ExcludeStatuses []string `mapstructure:"exclude_statuses" yaml:"exclude_statuses,omitempty" json:"exclude_statuses,omitempty"`
}

type RabbitMQConfig struct {
	Enabled       bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ManagementURL string          `mapstructure:"management_url" yaml:"management_url" json:"management_url"`
	Username      string          `mapstructure:"username" yaml:"username" json:"username"`
	Password      string          `mapstructure:"password" yaml:"password,omitempty" json:"password,omitempty"`
	Interval      time.Duration   `mapstructure:"interval" yaml:"interval" json:"interval"`
	Queues        []RabbitMQQueue `mapstructure:"queues" yaml:"queues" json:"queues"`
}

type RabbitMQQueue struct {
	Name             string `mapstructure:"name" yaml:"name" json:"name"`
	Threshold        int64  `mapstructure:"threshold" yaml:"threshold" json:"threshold"`
	UnackedThreshold int64  `mapstructure:"unacked_threshold" yaml:"unacked_threshold" json:"unacked_threshold"`
	Severity         string `mapstructure:"severity" yaml:"severity" json:"severity"`
}

type NotifiersConfig struct {
	Telegram TelegramConfig `mapstructure:"telegram" yaml:"telegram" json:"telegram"`
	Webhook  WebhookConfig  `mapstructure:"webhook" yaml:"webhook" json:"webhook"`
	LogPath  string         `mapstructure:"log_path" yaml:"log_path" json:"log_path"`
}

type TelegramConfig struct {
	Enabled  bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Token    string   `mapstructure:"token" yaml:"token" json:"token"`
	ChatID   int64    `mapstructure:"chat_id" yaml:"chat_id" json:"chat_id"`
	Mentions []string `mapstructure:"mentions" yaml:"mentions,omitempty" json:"mentions,omitempty"`
}

type WebhookConfig struct {
	Enabled      bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	URL          string            `mapstructure:"url" yaml:"url" json:"url"`
	Method       string            `mapstructure:"method" yaml:"method" json:"method"`
	APIKey       string            `mapstructure:"api_key" yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIKeyHeader string            `mapstructure:"api_key_header" yaml:"api_key_header,omitempty" json:"api_key_header,omitempty"`
	Headers      map[string]string `mapstructure:"headers" yaml:"headers,omitempty" json:"headers,omitempty"`
}

type HistoryConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	FilePath   string `mapstructure:"file_path" yaml:"file_path" json:"file_path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb" yaml:"max_size_mb" json:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`
}

type ServerConfig struct {
	Listen   string `mapstructure:"listen" yaml:"listen" json:"listen"`
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Username string `mapstructure:"username" yaml:"username,omitempty" json:"username,omitempty"`
	Password string `mapstructure:"password" yaml:"password,omitempty" json:"password,omitempty"`
}

// FindConfig returns the first existing config.yaml found next to the binary,
// then /etc/argus/config.yaml, then ./config.yaml as fallback.
func FindConfig() string {
	candidates := []string{"/etc/argus/config.yaml"}

	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			dir := filepath.Dir(real)
			// skip system bin dirs — config next to binary only makes sense for portable bundles
			switch dir {
			case "/usr/local/bin", "/usr/bin", "/bin":
			default:
				candidates = append(candidates, filepath.Join(dir, "config.yaml"))
			}
		}
	}

	candidates = append(candidates, "config.yaml")

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "config.yaml"
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config %q: %w", path, err)
	}
	return nil
}

func Load(path string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigFile(path)

	v.SetEnvPrefix("ARGUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("collectors.memory.enabled", true)
	v.SetDefault("collectors.memory.interval", "10m")
	v.SetDefault("collectors.memory.refresh", "5s")
	v.SetDefault("collectors.memory.threshold", 85.0)
	v.SetDefault("collectors.memory.severity", "warning")

	v.SetDefault("collectors.cpu.enabled", true)
	v.SetDefault("collectors.cpu.interval", "5m")
	v.SetDefault("collectors.cpu.refresh", "5s")
	v.SetDefault("collectors.cpu.threshold", 90.0)
	v.SetDefault("collectors.cpu.severity", "critical")

	v.SetDefault("collectors.disk.enabled", true)
	v.SetDefault("collectors.disk.interval", "30m")
	v.SetDefault("collectors.disk.refresh", "5s")
	v.SetDefault("collectors.disk.path", "/")
	v.SetDefault("collectors.disk.threshold", 80.0)
	v.SetDefault("collectors.disk.severity", "warning")

	v.SetDefault("collectors.nginx.enabled", false)
	v.SetDefault("collectors.nginx.access_log", "/var/log/nginx/access.log")
	v.SetDefault("collectors.nginx.watch_statuses", []string{"500-599"})
	v.SetDefault("collectors.nginx.interval", "1m")
	v.SetDefault("collectors.nginx.refresh", "5s")
	v.SetDefault("collectors.nginx.window", "60s")
	v.SetDefault("collectors.nginx.threshold", 10)
	v.SetDefault("collectors.nginx.severity", "critical")
	v.SetDefault("collectors.nginx.slow_threshold", 7.0)
	v.SetDefault("collectors.nginx.slow_window", "60s")
	v.SetDefault("collectors.nginx.slow_count", 5)
	v.SetDefault("collectors.nginx.min_group_count", 5)

	v.SetDefault("notifiers.webhook.method", "POST")

	v.SetDefault("history.enabled", true)
	v.SetDefault("history.file_path", "/var/log/argus/alarms.log")
	v.SetDefault("history.max_size_mb", 100)
	v.SetDefault("history.max_backups", 3)

	v.SetDefault("collectors.rabbitmq.enabled", false)
	v.SetDefault("collectors.rabbitmq.management_url", "http://localhost:15672")
	v.SetDefault("collectors.rabbitmq.username", "guest")
	v.SetDefault("collectors.rabbitmq.interval", "30s")

	v.SetDefault("name", "")
	v.SetDefault("server.listen", "127.0.0.1:8765")
	v.SetDefault("server.enabled", true)
	v.SetDefault("notifiers.log_path", "")
}

func validate(cfg *Config) error {
	valid := map[string]bool{
		"info": true, "warning": true, "critical": true,
	}

	checks := []struct {
		name, sev string
		on        bool
	}{
		{"memory", cfg.Collectors.Memory.Severity, cfg.Collectors.Memory.Enabled},
		{"cpu", cfg.Collectors.CPU.Severity, cfg.Collectors.CPU.Enabled},
		{"disk", cfg.Collectors.Disk.Severity, cfg.Collectors.Disk.Enabled},
		{"nginx", cfg.Collectors.Nginx.Severity, cfg.Collectors.Nginx.Enabled},
	}

	for _, c := range checks {
		if c.on && !valid[c.sev] {
			return fmt.Errorf("%s: invalid severity %q", c.name, c.sev)
		}
	}

	if cfg.Notifiers.Telegram.Enabled && cfg.Notifiers.Telegram.Token == "" {
		return fmt.Errorf("telegram: token is required")
	}

	anyNotifier := cfg.Notifiers.Telegram.Enabled ||
		cfg.Notifiers.Webhook.Enabled

	if !anyNotifier {
		fmt.Fprintln(os.Stderr, "warning: no notifiers enabled")
	}

	return nil
}
