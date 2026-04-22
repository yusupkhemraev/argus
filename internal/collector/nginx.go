package collector

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nxadm/tail"

	"github.com/yusupkhemraev/argus/internal/config"
)

type requestEntry struct {
	Method string
	Path   string
	Status int
	Times  []time.Time
}

func (e *requestEntry) count(now time.Time, window time.Duration) int {
	cutoff := now.Add(-window)
	c := 0
	for _, t := range e.Times {
		if t.After(cutoff) {
			c++
		}
	}
	return c
}

func (e *requestEntry) countSince(since time.Time) int {
	c := 0
	for _, t := range e.Times {
		if t.After(since) {
			c++
		}
	}
	return c
}

func (e *requestEntry) prune(now time.Time, window time.Duration) {
	e.Times = pruneOld(e.Times, now, window)
}

type statusMatcher struct {
	low, high int
}

type compiledRoute struct {
	method          string
	re              *regexp.Regexp
	minCount        int
	excludeStatuses []statusMatcher
}

type parsedLine struct {
	method string
	path   string
	status int
	rt     float64
	urt    float64
}

type compiledIgnore struct {
	method string
	re     *regexp.Regexp
}

type NginxCollector struct {
	cfg config.NginxConfig

	mu         sync.Mutex
	httpErrors map[string]*requestEntry
	slowReqs   map[string]*requestEntry

	lastCheck      *time.Time
	lineRegexp     *regexp.Regexp
	rtRegexp       *regexp.Regexp
	urtRegexp      *regexp.Regexp
	idxMethod      int
	idxPath        int
	idxStatus      int
	idxRT          int
	idxURT         int
	statusMatchers []statusMatcher
	priorityRoutes []compiledRoute
	ignoreRoutes   []compiledIgnore
}

func NewNginxCollector(cfg config.NginxConfig) (*NginxCollector, error) {
	c := &NginxCollector{
		cfg:        cfg,
		httpErrors: make(map[string]*requestEntry),
		slowReqs:   make(map[string]*requestEntry),
		idxRT:      -1,
		idxURT:     -1,
	}

	// Permissive pattern: finds request line + status regardless of surrounding fields.
	// Timing is extracted via rtRegexp/urtRegexp fallback below.
	c.lineRegexp = regexp.MustCompile(
		`"(?P<method>\w+)\s+(?P<path>[^\s"]+)\s+HTTP/[\d.]+"\s+(?P<status>\d{3})`,
	)

	c.idxMethod = c.lineRegexp.SubexpIndex("method")
	c.idxPath = c.lineRegexp.SubexpIndex("path")
	c.idxStatus = c.lineRegexp.SubexpIndex("status")

	c.rtRegexp = regexp.MustCompile(`(?:rt[=:]|request_time[=:])(\d+\.\d+)`)
	c.urtRegexp = regexp.MustCompile(`(?:urt[=:]|upstream_response_time[=:])(\d+[\d.]*)`)

	for _, s := range cfg.WatchStatuses {
		sm, err := parseStatusMatcher(s)
		if err != nil {
			continue
		}
		c.statusMatchers = append(c.statusMatchers, sm)
	}

	for _, pr := range cfg.PriorityRoutes {
		pattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pr.Pattern), `\*`, `[^?]*`) + "$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		minCount := pr.MinCount
		if minCount <= 0 {
			minCount = 1
		}
		var excludes []statusMatcher
		for _, es := range pr.ExcludeStatuses {
			sm, err := parseStatusMatcher(es)
			if err != nil {
				continue
			}
			excludes = append(excludes, sm)
		}
		c.priorityRoutes = append(c.priorityRoutes, compiledRoute{
			method:          strings.ToUpper(pr.Method),
			re:              re,
			minCount:        minCount,
			excludeStatuses: excludes,
		})
	}

	for _, ir := range cfg.IgnoreRoutes {
		pattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(ir.Pattern), `\*`, `[^?]*`) + "$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		c.ignoreRoutes = append(c.ignoreRoutes, compiledIgnore{
			method: strings.ToUpper(ir.Method),
			re:     re,
		})
	}

	return c, nil
}

func (n *NginxCollector) Name() string { return "nginx" }

func (n *NginxCollector) Start() {
	if n.cfg.AccessLog != "" {
		go n.tailAccessLog()
	}
}

func (n *NginxCollector) tailAccessLog() {
	t, err := tail.TailFile(n.cfg.AccessLog, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
	})
	if err != nil {
		return
	}
	defer t.Cleanup()

	for line := range t.Lines {
		if line.Err != nil {
			continue
		}
		n.processLine(line.Text)
	}
}

func (n *NginxCollector) parseLine(line string) (parsedLine, bool) {
	matches := n.lineRegexp.FindStringSubmatch(line)
	if matches == nil {
		return parsedLine{}, false
	}

	if n.idxMethod < 0 || n.idxPath < 0 || n.idxStatus < 0 {
		return parsedLine{}, false
	}

	status, err := strconv.Atoi(matches[n.idxStatus])
	if err != nil {
		return parsedLine{}, false
	}

	p := parsedLine{
		method: matches[n.idxMethod],
		path:   stripQuery(matches[n.idxPath]),
		status: status,
	}

	if n.idxRT >= 0 && matches[n.idxRT] != "" && matches[n.idxRT] != "-" {
		p.rt, _ = strconv.ParseFloat(matches[n.idxRT], 64)
	}
	if n.idxURT >= 0 && matches[n.idxURT] != "" && matches[n.idxURT] != "-" {
		p.urt, _ = strconv.ParseFloat(matches[n.idxURT], 64)
	}

	if p.rt == 0 && p.urt == 0 {
		p.rt = n.searchFloat(n.rtRegexp, line)
		p.urt = n.searchFloat(n.urtRegexp, line)
	}

	return p, true
}

func (n *NginxCollector) isIgnored(method, path string) bool {
	for _, ir := range n.ignoreRoutes {
		if (ir.method == "*" || ir.method == "" || ir.method == method) && ir.re.MatchString(path) {
			return true
		}
	}
	return false
}

func (n *NginxCollector) processLine(line string) {
	p, ok := n.parseLine(line)
	if !ok {
		return
	}

	if n.isIgnored(p.method, p.path) {
		return
	}

	now := time.Now()

	isSlow := (n.cfg.SlowThreshold > 0) &&
		((p.rt > 0 && p.rt > n.cfg.SlowThreshold) || (p.urt > 0 && p.urt > n.cfg.SlowThreshold))

	if isSlow {
		n.recordSlow(p.method, p.path, p.status, now)
	}

	isWatched := n.matchesStatus(p.status)
	isPriorityError := p.status >= 400 && n.isPriority(p.method, p.path, p.status)

	if isWatched || isPriorityError {
		n.recordHTTPError(p.method, p.path, p.status, now)
	}
}

func (n *NginxCollector) searchFloat(re *regexp.Regexp, line string) float64 {
	m := re.FindStringSubmatch(line)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	return v
}

const maxMapEntries = 10000

func (n *NginxCollector) recordHTTPError(method, path string, status int, now time.Time) {
	key := fmt.Sprintf("%s %s %d", method, path, status)
	n.mu.Lock()
	defer n.mu.Unlock()

	entry, ok := n.httpErrors[key]
	if !ok {
		if len(n.httpErrors) >= maxMapEntries {
			return
		}
		entry = &requestEntry{Method: method, Path: path, Status: status}
		n.httpErrors[key] = entry
	}
	entry.Times = append(entry.Times, now)
}

func (n *NginxCollector) recordSlow(method, path string, status int, now time.Time) {
	key := fmt.Sprintf("%s %s", method, path)
	n.mu.Lock()
	defer n.mu.Unlock()

	entry, ok := n.slowReqs[key]
	if !ok {
		if len(n.slowReqs) >= maxMapEntries {
			return
		}
		entry = &requestEntry{Method: method, Path: path, Status: status}
		n.slowReqs[key] = entry
	}
	entry.Times = append(entry.Times, now)
}

func (n *NginxCollector) isExcludedStatus(pr compiledRoute, status int) bool {
	for _, m := range pr.excludeStatuses {
		if status >= m.low && status <= m.high {
			return true
		}
	}
	return false
}

func (n *NginxCollector) isPriority(method, path string, status int) bool {
	for _, pr := range n.priorityRoutes {
		if (pr.method == "*" || pr.method == method) && pr.re.MatchString(path) {
			return !n.isExcludedStatus(pr, status)
		}
	}
	return false
}

func (n *NginxCollector) priorityMinCount(method, path string, status int) int {
	for _, pr := range n.priorityRoutes {
		if (pr.method == "*" || pr.method == method) && pr.re.MatchString(path) {
			if n.isExcludedStatus(pr, status) {
				return 0
			}
			return pr.minCount
		}
	}
	return 0
}

func (n *NginxCollector) Collect() (Metric, error) {
	totalErrors := 0
	totalSlow := 0

	lines := readLastLines(n.cfg.AccessLog, 10000)
	for _, line := range lines {
		p, ok := n.parseLine(line)
		if !ok {
			continue
		}

		isWatched := n.matchesStatus(p.status)
		isPriority := p.status >= 400 && n.isPriority(p.method, p.path, p.status)
		if isWatched || isPriority {
			totalErrors++
		}

		if n.cfg.SlowThreshold > 0 {
			if (p.rt > 0 && p.rt > n.cfg.SlowThreshold) || (p.urt > 0 && p.urt > n.cfg.SlowThreshold) {
				totalSlow++
			}
		}
	}

	return Metric{
		Collector: n.Name(),
		Value:     float64(totalErrors),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"http_errors":   fmt.Sprintf("%d", totalErrors),
			"slow_requests": fmt.Sprintf("%d", totalSlow),
		},
	}, nil
}

func readLastLines(path string, maxLines int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > maxLines {
			lines = lines[1:]
		}
	}
	return lines
}

// Check only counts entries that arrived AFTER the last alarm.
// This prevents re-alerting on the same data.
func (n *NginxCollector) Check(metric Metric) *Alarm {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now()

	since := now.Add(-n.cfg.Window)
	if n.lastCheck != nil {
		if n.lastCheck.After(since) {
			since = *n.lastCheck
		}
	}

	newErrors := 0
	for _, entry := range n.httpErrors {
		newErrors += entry.countSince(since)
	}

	newPriority := false
	for _, entry := range n.httpErrors {
		minC := n.priorityMinCount(entry.Method, entry.Path, entry.Status)
		if minC > 0 && entry.countSince(since) >= minC {
			newPriority = true
			break
		}
	}

	newSlow := 0
	for _, entry := range n.slowReqs {
		newSlow += entry.countSince(since)
	}
	slowTriggered := n.cfg.SlowCount > 0 && newSlow >= n.cfg.SlowCount

	httpTriggered := n.cfg.Threshold > 0 && newErrors >= n.cfg.Threshold

	if !httpTriggered && !newPriority && !slowTriggered {
		return nil
	}

	n.lastCheck = &now

	alarmID := "nginx_errors"
	sev := Severity(n.cfg.Severity)
	if newPriority {
		alarmID = "nginx_priority"
		sev = SeverityCritical
	} else if slowTriggered && !httpTriggered {
		alarmID = "nginx_slow"
		sev = SeverityWarning
	}

	msg := n.buildMessageSince(since)

	return &Alarm{
		ID:        alarmID,
		Collector: n.Name(),
		Message:   msg,
		Severity:  sev,
		Value:     float64(newErrors),
		Threshold: float64(n.cfg.Threshold),
		Timestamp: now,
	}
}

func (n *NginxCollector) countAllSlow(now time.Time, window time.Duration) int {
	total := 0
	for _, entry := range n.slowReqs {
		total += entry.count(now, window)
	}
	return total
}

type endpointStat struct {
	label string
	count int
}

func (n *NginxCollector) buildMessageSince(since time.Time) string {
	minGroup := n.cfg.MinGroupCount
	if minGroup <= 0 {
		minGroup = 1
	}

	var lines []endpointStat

	for _, entry := range n.httpErrors {
		c := entry.countSince(since)
		if c == 0 {
			continue
		}
		isPrio := n.priorityMinCount(entry.Method, entry.Path, entry.Status) > 0
		if !isPrio && c < minGroup {
			continue
		}
		label := fmt.Sprintf("%s %s - %d", entry.Method, entry.Path, entry.Status)
		lines = append(lines, endpointStat{label: label, count: c})
	}

	for _, entry := range n.slowReqs {
		c := entry.countSince(since)
		if c == 0 || c < minGroup {
			continue
		}
		label := fmt.Sprintf("🐢 %s %s (slow)", entry.Method, entry.Path)
		lines = append(lines, endpointStat{label: label, count: c})
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].count > lines[j].count
	})

	// minGroup filtered everything out — show all non-zero entries so message is never empty
	if len(lines) == 0 {
		for _, entry := range n.httpErrors {
			c := entry.countSince(since)
			if c == 0 {
				continue
			}
			label := fmt.Sprintf("%s %s - %d", entry.Method, entry.Path, entry.Status)
			lines = append(lines, endpointStat{label: label, count: c})
		}
		for _, entry := range n.slowReqs {
			c := entry.countSince(since)
			if c == 0 {
				continue
			}
			label := fmt.Sprintf("🐢 %s %s (slow)", entry.Method, entry.Path)
			lines = append(lines, endpointStat{label: label, count: c})
		}
		sort.Slice(lines, func(i, j int) bool {
			return lines[i].count > lines[j].count
		})
	}

	if len(lines) == 0 {
		return "nginx errors detected"
	}

	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		parts = append(parts, fmt.Sprintf("%s: %d", l.label, l.count))
	}
	return strings.Join(parts, "\n")
}

func (n *NginxCollector) matchesStatus(status int) bool {
	for _, m := range n.statusMatchers {
		if status >= m.low && status <= m.high {
			return true
		}
	}
	return false
}

func parseStatusMatcher(s string) (statusMatcher, error) {
	s = strings.TrimSpace(s)
	if parts := strings.SplitN(s, "-", 2); len(parts) == 2 {
		low, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		high, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return statusMatcher{}, fmt.Errorf("invalid range %q", s)
		}
		return statusMatcher{low: low, high: high}, nil
	}
	code, err := strconv.Atoi(s)
	if err != nil {
		return statusMatcher{}, fmt.Errorf("invalid code %q", s)
	}
	return statusMatcher{low: code, high: code}, nil
}


func stripQuery(path string) string {
	if i := strings.IndexByte(path, '?'); i >= 0 {
		return path[:i]
	}
	return path
}

func pruneOld(events []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	i := 0
	for i < len(events) && events[i].Before(cutoff) {
		i++
	}
	return events[i:]
}
