package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"io/fs"
	"net/http"
	"time"

	"github.com/yusupkhemraev/argus/internal/bus"
	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
	"github.com/yusupkhemraev/argus/internal/notifier"
)

type ReloadFunc func(cfg config.Config) error

type Server struct {
	httpServer *http.Server
	bus        *bus.Bus
	cfg        config.Config
	configPath string
	staticFS   fs.FS
	notifiers  []notifier.Notifier
	collectors []collector.Collector
	onReload   ReloadFunc
}

func New(cfg config.Config, configPath string, b *bus.Bus, staticFS fs.FS, notifiers []notifier.Notifier, collectors []collector.Collector, onReload ReloadFunc) *Server {
	s := &Server{
		bus:        b,
		cfg:        cfg,
		configPath: configPath,
		staticFS:   staticFS,
		notifiers:  notifiers,
		collectors: collectors,
		onReload:   onReload,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/alarms", s.handleAlarms)
	mux.HandleFunc("GET /api/logs", s.handleLogs)
	mux.HandleFunc("GET /api/events", s.handleSSE)
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("PUT /api/config", s.handleSaveConfig)
	mux.HandleFunc("POST /api/test-notification", s.handleTestNotification)
	mux.HandleFunc("POST /api/reset-alarms", s.handleResetAlarms)
	mux.HandleFunc("POST /api/test-collector", s.handleTestCollector)

	if staticFS != nil {
		mux.Handle("GET /", http.FileServerFS(staticFS))
	}

	var handler http.Handler = mux
	if cfg.Server.Username != "" && cfg.Server.Password != "" {
		handler = s.basicAuth(mux)
	}

	s.httpServer = &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: handler,
	}

	return s
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) SetNotifiers(notifiers []notifier.Notifier) {
	s.notifiers = notifiers
}

func (s *Server) SetCollectors(collectors []collector.Collector) {
	s.collectors = collectors
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="argus"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userHash := sha256.Sum256([]byte(user))
		passHash := sha256.Sum256([]byte(pass))
		expectedUser := sha256.Sum256([]byte(s.cfg.Server.Username))
		expectedPass := sha256.Sum256([]byte(s.cfg.Server.Password))

		userMatch := subtle.ConstantTimeCompare(userHash[:], expectedUser[:]) == 1
		passMatch := subtle.ConstantTimeCompare(passHash[:], expectedPass[:]) == 1

		if !userMatch || !passMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="argus"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) ShutdownGraceful() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Shutdown(ctx)
}
