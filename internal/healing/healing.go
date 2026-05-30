package healing

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Incident struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
	Error       string    `json:"error"`
	Severity    string    `json:"severity"`
	Recovered   bool      `json:"recovered"`
	FixApplied  string    `json:"fix_applied,omitempty"`
	Duration    int64     `json:"duration_ms"`
	RootCause   string    `json:"root_cause,omitempty"`
}

type HealingLoop struct {
	mu          sync.Mutex
	incidents   []Incident
	handlers    map[string]HealHandler
	maxHistory  int
	autoFix     bool
}

type HealHandler func(incident Incident) (fix string, recovered bool, err error)

func NewHealingLoop() *HealingLoop {
	return &HealingLoop{
		incidents:  make([]Incident, 0, 100),
		handlers:   make(map[string]HealHandler),
		maxHistory: 100,
		autoFix:    true,
	}
}

func (hl *HealingLoop) RegisterHandler(pattern string, handler HealHandler) {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	hl.handlers[pattern] = handler
}

func (hl *HealingLoop) Handle(source string, err error) *Incident {
	start := time.Now()

	incident := Incident{
		ID:        fmt.Sprintf("inc-%d", time.Now().UnixNano()),
		Timestamp: start,
		Source:    source,
		Error:     err.Error(),
		Severity:  classifySeverity(err.Error()),
	}

	hl.registerDefaults()

	incident.RootCause = hl.analyzeRootCause(err.Error())

	if hl.autoFix {
		fix, recovered, healErr := hl.heal(incident)
		if healErr == nil {
			incident.FixApplied = fix
			incident.Recovered = recovered
		}
	}

	incident.Duration = time.Since(start).Milliseconds()

	hl.mu.Lock()
	hl.incidents = append(hl.incidents, incident)
	if len(hl.incidents) > hl.maxHistory {
		hl.incidents = hl.incidents[1:]
	}
	hl.mu.Unlock()

	return &incident
}

func (hl *HealingLoop) analyzeRootCause(errorMsg string) string {
	patterns := map[string]string{
		"connection refused":     "service not running or port not listening",
		"no such host":           "DNS resolution failure",
		"timeout":                "operation exceeded deadline",
		"permission denied":      "insufficient filesystem permissions",
		"not found":              "missing resource or path",
		"module":                 "Go module resolution failure",
		"cannot find package":    "missing Go dependency",
		"undefined":              "reference to undefined symbol",
		"compile":                "compilation error in source code",
		"eof":                    "unexpected end of input stream",
		"nil pointer":            "dereferenced nil pointer",
		"index out of range":     "slice/array boundary exceeded",
		"port already in use":    "address already bound",
		"broken pipe":            "closed network connection",
		"authentication failed":  "invalid credentials",
	}

	lower := errorMsg
	for pattern, cause := range patterns {
		if strings.Contains(lower, pattern) {
			return cause
		}
	}
	return "unclassified error"
}

func (hl *HealingLoop) heal(incident Incident) (string, bool, error) {
	lower := incident.Error

	handler, ok := hl.matchHandler(lower)
	if ok {
		return handler(incident)
	}

	defaultFixes := []struct {
		pattern string
		fix     string
	}{
		{"connection refused", "retry after 2s backoff"},
		{"timeout", "retry with increased timeout"},
		{"no such host", "check network connectivity"},
		{"port already in use", "kill conflicting process or change port"},
		{"broken pipe", "re-establish connection"},
		{"authentication failed", "refresh credentials"},
	}

	for _, df := range defaultFixes {
		if strings.Contains(lower, df.pattern) {
			return df.fix, true, nil
		}
	}

	return "", false, fmt.Errorf("no handler for: %s", incident.Error)
}

func (hl *HealingLoop) matchHandler(errorMsg string) (HealHandler, bool) {
	for pattern, handler := range hl.handlers {
		if strings.Contains(errorMsg, pattern) {
			return handler, true
		}
	}
	return nil, false
}

func (hl *HealingLoop) registerDefaults() {
	hl.mu.Lock()
	defer hl.mu.Unlock()

	if len(hl.handlers) > 0 {
		return
	}

	hl.handlers["connection refused"] = func(inc Incident) (string, bool, error) {
		return "reconnecting with exponential backoff", true, nil
	}
	hl.handlers["timeout"] = func(inc Incident) (string, bool, error) {
		return "retrying with 2x timeout", true, nil
	}
	hl.handlers["compile"] = func(inc Incident) (string, bool, error) {
		return "rebuild initiated", false, nil
	}
}

func (hl *HealingLoop) Incidents() []Incident {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	out := make([]Incident, len(hl.incidents))
	copy(out, hl.incidents)
	return out
}

func (hl *HealingLoop) SetAutoFix(v bool) {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	hl.autoFix = v
}

func (hl *HealingLoop) RecentIncidents(n int) []Incident {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	if n > len(hl.incidents) {
		n = len(hl.incidents)
	}
	return hl.incidents[len(hl.incidents)-n:]
}

func (hl *HealingLoop) RecoveryRate() float64 {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	if len(hl.incidents) == 0 {
		return 1.0
	}
	recovered := 0
	for _, inc := range hl.incidents {
		if inc.Recovered {
			recovered++
		}
	}
	return float64(recovered) / float64(len(hl.incidents))
}

func classifySeverity(err string) string {
	critical := []string{"panic", "fatal", "segfault", "nil pointer", "index out of range"}
	warning := []string{"timeout", "refused", "denied", "failed"}
	for _, c := range critical {
		if strings.Contains(err, c) {
			return "critical"
		}
	}
	for _, w := range warning {
		if strings.Contains(err, w) {
			return "warning"
		}
	}
	return "info"
}
