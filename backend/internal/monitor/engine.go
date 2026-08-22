// Package monitor provides metric collection (local + synthetic) and the
// alert evaluation engine.
package monitor

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/logger"
	"github.com/aicenter/aicenter/internal/repository"
	"go.uber.org/zap"
)

// Collector produces a batch of samples each tick. The default collector
// synthesizes system metrics for demo servers; a real agent feed can be
// plugged in later via Ingest.
type Collector interface {
	Collect() []models.Metric
}

// Engine runs the collection loop and evaluates alert rules against new data.
type Engine struct {
	repo      *repository.MonitorRepository
	collector Collector
	interval  time.Duration
	retention time.Duration

	mu    sync.Mutex
	state map[string][]sample // ruleKey -> recent samples for duration checks
	stop  chan struct{}

	// notifier is invoked when an alert fires. Optional; nil disables linkage.
	notifier func(eventType, title, severity, message string, data map[string]string, channelIDs []string)
}

// SetNotifier wires an external notifier (Phase 7) so fired alerts dispatch
// notifications. The callback receives the rule's bound channel ids.
func (e *Engine) SetNotifier(fn func(eventType, title, severity, message string, data map[string]string, channelIDs []string)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.notifier = fn
}

type sample struct {
	Value float64
	At    time.Time
}

func NewEngine(repo *repository.MonitorRepository, collector Collector, interval time.Duration) *Engine {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Engine{
		repo:      repo,
		collector: collector,
		interval:  interval,
		retention: 7 * 24 * time.Hour,
		state:     map[string][]sample{},
		stop:      make(chan struct{}),
	}
}

// Start launches the background loop; call Stop to end it.
func (e *Engine) Start() {
	go e.loop()
}

func (e *Engine) Stop() { close(e.stop) }

func (e *Engine) loop() {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	e.tick() // collect immediately on start
	for {
		select {
		case <-ticker.C:
			e.tick()
		case <-e.stop:
			return
		}
	}
}

func (e *Engine) tick() {
	log := logger.Get()
	var metrics []models.Metric
	if e.collector != nil {
		metrics = e.collector.Collect()
	}
	now := time.Now().UTC()
	for i := range metrics {
		if metrics[i].CollectedAt.IsZero() {
			metrics[i].CollectedAt = now
		}
		if err := e.repo.InsertMetric(&metrics[i]); err != nil {
			log.Warn("monitor: insert metric failed", zap.Error(err))
		}
	}
	if err := e.EvaluateRules(now); err != nil {
		log.Warn("monitor: evaluate rules failed", zap.Error(err))
	}
	// Retention pruning (cheap enough per tick; only actually deletes rarely).
	if _, err := e.repo.DeleteMetricsBefore(now.Add(-e.retention)); err != nil {
		log.Debug("monitor: retention prune failed", zap.Error(err))
	}
}

// Ingest accepts externally collected samples (e.g. from server agents),
// stores them and runs the alert engine over them.
func (e *Engine) Ingest(metrics []models.Metric) error {
	for i := range metrics {
		if err := e.repo.InsertMetric(&metrics[i]); err != nil {
			return err
		}
	}
	return e.EvaluateRules(time.Now().UTC())
}

// EvaluateRules loads enabled rules and latest metrics and fires alerts when
// conditions hold for the configured duration (respecting cooldown).
func (e *Engine) EvaluateRules(now time.Time) error {
	rules, err := e.repo.ListRules(true)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	latest, err := e.repo.LatestMetrics("")
	if err != nil {
		return err
	}
	// index latest by "serverID|metric"
	type key struct{ server, metric string }
	idx := map[key]models.Metric{}
	for _, m := range latest {
		idx[key{m.ServerID, m.MetricName}] = m
	}

	for _, rule := range rules {
		targets := []string{""} // global rule applies per known server? MVP: aggregate "" bucket
		_ = targets
		serverIDs := []string{""}
		if rule.ServerID != nil && *rule.ServerID != "" {
			serverIDs = []string{*rule.ServerID}
		} else if rule.ServerID == nil {
			// global rule: evaluate per server that has the metric
			seen := map[string]bool{}
			serverIDs = nil
			for _, m := range latest {
				if m.MetricName == rule.MetricName && !seen[m.ServerID] {
					seen[m.ServerID] = true
					serverIDs = append(serverIDs, m.ServerID)
				}
			}
			if len(serverIDs) == 0 {
				continue
			}
		}

		for _, sid := range serverIDs {
			m, ok := idx[key{sid, rule.MetricName}]
			if !ok {
				continue
			}
			rkey := rule.ID + "|" + sid
			breach := checkCondition(m.Value, rule.Condition, rule.Threshold)

			e.mu.Lock()
			hist := e.state[rkey]
			if breach {
				hist = append(hist, sample{Value: m.Value, At: now})
			} else {
				hist = nil // condition broken → reset duration window
			}
			// keep last hour of history max
			cutoff := now.Add(-time.Hour)
			kept := hist[:0]
			for _, s := range hist {
				if s.At.After(cutoff) {
					kept = append(kept, s)
				}
			}
			e.state[rkey] = kept
			durationOK := true
			if rule.Duration > 0 {
				need := now.Add(-time.Duration(rule.Duration) * time.Second)
				durationOK = len(kept) > 0
				for _, s := range kept {
					if s.At.Before(need) {
						durationOK = false
						break
					}
				}
			}
			e.mu.Unlock()

			if !breach || !durationOK {
				continue
			}

			// cooldown check
			lastFired, fired, err := e.repo.LastFiringForRule(rule.ID, sid)
			if err != nil {
				return fmt.Errorf("cooldown lookup: %w", err)
			}
			cool := time.Duration(rule.Cooldown) * time.Second
			if cool <= 0 {
				cool = 5 * time.Minute
			}
			if fired && now.Sub(lastFired) < cool {
				continue
			}

			ev := &models.AlertEvent{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				ServerID:   sid,
				MetricName: rule.MetricName,
				Value:      m.Value,
				Threshold:  rule.Threshold,
				Condition:  rule.Condition,
				Severity:   rule.Severity,
				Message: fmt.Sprintf("%s: %s=%.2f%s %s %.2f (rule %q)",
					sevLabel(rule.Severity), rule.MetricName, m.Value,
					unitSuffix(m.Unit), condSymbol(rule.Condition), rule.Threshold, rule.Name),
				Status: "firing",
			}
			if ev.Severity == "" {
				ev.Severity = "warning"
			}
			if err := e.repo.CreateEvent(ev); err != nil {
				return fmt.Errorf("create alert event: %w", err)
			}

			// Phase 7: dispatch notifications for this alert (honouring the
			// rule's bound channel ids; empty → dispatcher picks by template).
			e.mu.Lock()
			notifyFn := e.notifier
			e.mu.Unlock()
			if notifyFn != nil {
				var channelIDs []string
				channelIDs = repository.ParseChannelIDs(rule.NotificationChannels)
				data := map[string]string{
					"rule_id":     rule.ID,
					"rule_name":   rule.Name,
					"server_id":   sid,
					"metric_name": rule.MetricName,
					"value":       trimFloat(m.Value),
					"threshold":   trimFloat(rule.Threshold),
				}
				notifyFn("alert.fired", "告警触发: "+rule.Name, ev.Severity, ev.Message, data, channelIDs)
			}
		}
	}
	return nil
}

func checkCondition(v float64, cond string, threshold float64) bool {
	switch strings.ToLower(cond) {
	case "gt":
		return v > threshold
	case "gte":
		return v >= threshold
	case "lt":
		return v < threshold
	case "lte":
		return v <= threshold
	}
	return false
}

func condSymbol(c string) string {
	switch strings.ToLower(c) {
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "lt":
		return "<"
	case "lte":
		return "<="
	}
	return c
}

func sevLabel(s string) string {
	switch s {
	case "critical":
		return "[CRITICAL]"
	case "info":
		return "[INFO]"
	default:
		return "[WARNING]"
	}
}

func trimFloat(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	return s
}

func unitSuffix(u string) string {
	if u == "" {
		return ""
	}
	return u
}

// ---- Default synthetic collector ----

// SynthCollector generates plausible host metrics for every server that has
// reported before, so the monitor pipeline is demonstrable without real agents.
type SynthCollector struct{}

var synthNames = []struct {
	name string
	base float64
	jit  float64
	unit string
}{
	{"cpu.usage", 42, 25, "%"},
	{"memory.usage", 61, 15, "%"},
	{"disk.usage", 55, 8, "%"},
	{"load.1", 1.2, 1.0, ""},
}

// Collect fabricates samples for up to three pseudo-servers.
func (SynthCollector) Collect() []models.Metric {
	out := make([]models.Metric, 0, 12)
	servers := []string{"synth-web-01", "synth-web-02"}
	now := time.Now().UTC()
	for si, sid := range servers {
		for mi, m := range synthNames {
			v := m.base + (rand.Float64()-0.5)*2*m.jit + float64(si)*3
			if v < 0 {
				v = -v
			}
			if m.name == "cpu.usage" && si == 1 && rand.Float64() < 0.12 {
				v = 92 + rand.Float64()*6 // occasionally trip a >90% alert
			}
			out = append(out, models.Metric{
				ServerID:    sid,
				MetricName:  m.name,
				Value:       v,
				Unit:        m.unit,
				Labels:      map[string]string{"source": "synthetic"},
				CollectedAt: now.Add(time.Duration(mi) * 100 * time.Millisecond),
			})
		}
	}
	return out
}
