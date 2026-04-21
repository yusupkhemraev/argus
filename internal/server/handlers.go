package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
	"github.com/yusupkhemraev/argus/internal/history"
	"github.com/yusupkhemraev/argus/internal/notifier"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	snap := s.bus.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap.Metrics)
}

func (s *Server) handleAlarms(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 20)
	severity := r.URL.Query().Get("severity")

	records, err := history.ReadRecords(s.cfg.History.FilePath, limit, severity)
	if err != nil {
		records = []history.Record{}
	}

	type alarmResponse struct {
		Collector   string  `json:"collector"`
		Message     string  `json:"message"`
		Severity    string  `json:"severity"`
		Value       float64 `json:"value"`
		Threshold   float64 `json:"threshold"`
		TriggeredAt string  `json:"triggered_at"`
	}

	out := make([]alarmResponse, len(records))
	for i, r := range records {
		out[i] = alarmResponse{
			Collector:   r.Collector,
			Message:     r.Message,
			Severity:    r.Severity,
			Value:       r.Value,
			Threshold:   r.Threshold,
			TriggeredAt: r.Timestamp,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 20)
	status := r.URL.Query().Get("status")
	logPath := s.cfg.Notifiers.LogPath

	w.Header().Set("Content-Type", "application/json")

	if logPath == "" {
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	entries, err := notifier.ReadLogEntries(logPath, limit, status)
	if err != nil {
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	json.NewEncoder(w).Encode(entries)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		config.Config
		ConfigPath string `json:"config_path"`
	}{
		Config:     s.cfg,
		ConfigPath: s.configPath,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := config.Save(s.configPath, &cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.cfg = cfg

	if s.onReload != nil {
		if err := s.onReload(cfg); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "saved", "reload_error": err.Error()})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTestNotification(w http.ResponseWriter, r *http.Request) {
	if len(s.notifiers) == 0 {
		http.Error(w, "no notifiers configured", http.StatusBadRequest)
		return
	}

	testAlarm := collector.Alarm{
		ID:        "test",
		Collector: "test",
		Message:   "Test notification from Argus dashboard",
		Severity:  collector.SeverityWarning,
		Value:     42.0,
		Threshold: 80.0,
		Timestamp: time.Now(),
	}

	var errors []string
	for _, n := range s.notifiers {
		if err := n.Send(testAlarm); err != nil {
			errors = append(errors, n.Name()+": "+err.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if len(errors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"status": "error", "errors": errors})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleResetAlarms(w http.ResponseWriter, r *http.Request) {
	s.bus.Reset()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTestCollector(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}

	var target collector.Collector
	for _, c := range s.collectors {
		if c.Name() == name {
			target = c
			break
		}
	}
	if target == nil {
		http.Error(w, "collector not found: "+name, http.StatusNotFound)
		return
	}

	metric, err := target.Collect()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}

	alarm := target.Check(metric)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"metric":  metric,
		"alarm":   alarm,
		"has_alarm": alarm != nil,
	})
}

func parseLimit(r *http.Request, defaultLimit int) int {
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return defaultLimit
}
