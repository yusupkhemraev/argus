package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/yusupkhemraev/argus/internal/alarm"
	"github.com/yusupkhemraev/argus/internal/bus"
	"github.com/yusupkhemraev/argus/internal/collector"
	"github.com/yusupkhemraev/argus/internal/config"
	"github.com/yusupkhemraev/argus/internal/history"
	"github.com/yusupkhemraev/argus/internal/notifier"
	"github.com/yusupkhemraev/argus/internal/server"
	"github.com/yusupkhemraev/argus/internal/service"
	"github.com/yusupkhemraev/argus/web"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "start":
			if err := service.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("argus started")
			return
		case "stop":
			if err := service.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("argus stopped")
			return
		case "restart":
			if err := service.Restart(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("argus restarted")
			return
		case "service":
			out, err := service.Status()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(out)
			return
		case "status":
			runStatus(os.Args[2:])
			return
		case "history":
			runHistory(os.Args[2:])
			return
		case "logs":
			runLogs(os.Args[2:])
			return
		case "--version", "version":
			fmt.Printf("argus %s\n", version)
			return
		}
	}
	runDaemon(os.Args[1:])
}

func runDaemon(args []string) {
	fset := flag.NewFlagSet("argus", flag.ExitOnError)
	configPath := fset.String("config", "", "path to config file")
	noWeb := fset.Bool("no-web", false, "disable web server")
	webAddr := fset.String("web", "", "web server listen address (overrides config)")
	fset.Parse(args)

	if *configPath == "" {
		*configPath = config.FindConfig()
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *webAddr != "" {
		cfg.Server.Listen = *webAddr
	}

	b := bus.New()

	logPublish := func(entry notifier.LogEntry) {
		data, _ := json.Marshal(entry)
		b.Publish(bus.Event{Type: "log", Data: data})
	}

	var (
		mu      sync.Mutex
		manager *alarm.Manager
		srv     *server.Server
	)

	resolveServerName := func(name string) string {
		if name != "" {
			return name
		}
		h, _ := os.Hostname()
		if h == "" {
			return "unknown"
		}
		return h
	}

	buildNotifiers := func(ncfg config.NotifiersConfig, serverName string) []notifier.Notifier {
		var nn []notifier.Notifier

		wrap := func(n notifier.Notifier) notifier.Notifier {
			if ncfg.LogPath != "" {
				return notifier.NewNotificationLogger(n, ncfg.LogPath, logPublish)
			}
			return n
		}

		if ncfg.Telegram.Enabled {
			nn = append(nn, wrap(notifier.NewTelegramNotifier(ncfg.Telegram, serverName)))
		}
		if ncfg.Webhook.Enabled {
			nn = append(nn, wrap(notifier.NewWebhookNotifier(ncfg.Webhook, serverName)))
		}
		return nn
	}

	buildCollectors := func(ccfg config.CollectorsConfig) []collector.Collector {
		cc := []collector.Collector{
			collector.NewMemoryCollector(ccfg.Memory),
			collector.NewCPUCollector(ccfg.CPU),
			collector.NewDiskCollector(ccfg.Disk),
		}
		if ccfg.Nginx.Enabled {
			if nc, err := collector.NewNginxCollector(ccfg.Nginx); err == nil {
				cc = append(cc, nc)
			}
		}
		if ccfg.RabbitMQ.Enabled {
			cc = append(cc, collector.NewRabbitMQCollector(ccfg.RabbitMQ))
		}
		return cc
	}

	onReload := func(newCfg config.Config) error {
		mu.Lock()
		defer mu.Unlock()

		if manager != nil {
			manager.Stop()
		}

		nn := buildNotifiers(newCfg.Notifiers, resolveServerName(newCfg.Name))
		cc := buildCollectors(newCfg.Collectors)

		hist, _ := history.New(newCfg.History)

		manager = alarm.NewManager(newCfg, cc, nn, hist, b)
		manager.Start()

		if srv != nil {
			srv.SetNotifiers(nn)
			srv.SetCollectors(cc)
		}

		return nil
	}

	notifiers := buildNotifiers(cfg.Notifiers, resolveServerName(cfg.Name))
	collectors := buildCollectors(cfg.Collectors)

	hist, err := history.New(cfg.History)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: history disabled: %v\n", err)
		hist = nil
	}

	manager = alarm.NewManager(*cfg, collectors, notifiers, hist, b)
	manager.Start()
	fmt.Printf("argus %s started (%d collectors)\n", version, len(collectors))

	if !*noWeb && cfg.Server.Enabled {
		staticFS, _ := fs.Sub(web.StaticFiles, "dist")
		srv = server.New(*cfg, *configPath, b, staticFS, notifiers, collectors, onReload)
		go func() {
			if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
				fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			}
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nshutting down...")
	if srv != nil {
		srv.ShutdownGraceful()
	}
	mu.Lock()
	manager.Stop()
	mu.Unlock()
	if hist != nil {
		hist.Close()
	}
}

func runStatus(args []string) {
	v, _ := mem.VirtualMemory()
	cpuPct, _ := cpu.Percent(time.Second, false)
	d, _ := disk.Usage("/")

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "METRIC\tVALUE\tDETAILS\n")
	fmt.Fprintf(tw, "memory\t%.1f%%\t%.1f / %.1f GB\n", v.UsedPercent, float64(v.Used)/1e9, float64(v.Total)/1e9)
	if len(cpuPct) > 0 {
		fmt.Fprintf(tw, "cpu\t%.1f%%\t\n", cpuPct[0])
	}
	fmt.Fprintf(tw, "disk\t%.1f%%\t%.1f / %.1f GB\n", d.UsedPercent, float64(d.Used)/1e9, float64(d.Total)/1e9)
	tw.Flush()
}

func runHistory(args []string) {
	fset := flag.NewFlagSet("history", flag.ExitOnError)
	n := fset.Int("n", 10, "number of entries")
	severity := fset.String("severity", "", "filter by severity (info/warning/critical)")
	configPath := fset.String("config", "", "path to config file")
	fset.Parse(args)

	if *configPath == "" {
		*configPath = config.FindConfig()
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	records, err := history.ReadRecords(cfg.History.FilePath, *n, *severity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading history: %v\n", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "TIME\tSEVERITY\tCOLLECTOR\tMESSAGE\tVALUE\n")
	for _, r := range records {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%.1f\n", r.Timestamp, r.Severity, r.Collector, r.Message, r.Value)
	}
	tw.Flush()
}

func runLogs(args []string) {
	fset := flag.NewFlagSet("logs", flag.ExitOnError)
	n := fset.Int("n", 20, "number of entries")
	status := fset.String("status", "", "filter by status (sent/error)")
	configPath := fset.String("config", "", "path to config file")
	fset.Parse(args)

	if *configPath == "" {
		*configPath = config.FindConfig()
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	logPath := cfg.Notifiers.LogPath
	if logPath == "" {
		fmt.Fprintln(os.Stderr, "notification log path not configured")
		os.Exit(1)
	}

	entries, err := notifier.ReadLogEntries(logPath, *n, *status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading logs: %v\n", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "TIME\tNOTIFIER\tALARM\tSTATUS\tERROR\n")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", e.Timestamp, e.Notifier, e.AlarmID, e.Status, e.Error)
	}
	tw.Flush()
}
