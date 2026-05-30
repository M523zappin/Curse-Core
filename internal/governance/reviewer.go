package governance

import (
	"fmt"
)

type ReviewResult struct {
	Rule      Rule
	Pass      bool
	Message   string
}

type ReviewReport struct {
	Passed    bool
	Results   []ReviewResult
	Blocks    int
	Warnings  int
}

type Reviewer struct {
	constitution *Constitution
}

func NewReviewer(c *Constitution) *Reviewer {
	return &Reviewer{constitution: c}
}

func (r *Reviewer) Review(action string, context map[string]string) *ReviewReport {
	report := &ReviewReport{Passed: true}
	for _, rule := range r.constitution.Rules {
		result := r.evaluate(rule, action, context)
		report.Results = append(report.Results, result)
		if !result.Pass {
			if rule.Severity == SeverityBlock {
				report.Blocks++
				report.Passed = false
			} else {
				report.Warnings++
			}
		}
	}
	return report
}

func (r *Reviewer) evaluate(rule Rule, action string, context map[string]string) ReviewResult {
	switch rule.ID {
	case "S-001":
		for _, v := range context {
			if containsSecret(v) {
				return ReviewResult{Rule: rule, Pass: false, Message: "secret detected in context"}
			}
		}
	case "S-002":
		if containsSecret(action) {
			return ReviewResult{Rule: rule, Pass: false, Message: "secret detected in action string"}
		}
	case "G-001":
		return ReviewResult{Rule: rule, Pass: true, Message: "constitution review performed"}
	case "E-002":
		if !isStaged(action) {
			return ReviewResult{Rule: rule, Pass: false, Message: "write not staged through sandbox"}
		}
	case "R-001":
		if isDirectAction(action) && !hasLogEntry(action) {
			return ReviewResult{Rule: rule, Pass: false, Message: "action not logged"}
		}
	default:
		return ReviewResult{Rule: rule, Pass: true, Message: "no violation detected"}
	}
	return ReviewResult{Rule: rule, Pass: true, Message: "no violation detected"}
}

func containsSecret(s string) bool {
	lower := s
	indicators := []string{"sk-", "api_key", "api-key", "secret", "token", "password", "bearer "}
	for _, ind := range indicators {
		if stringsContains(lower, ind) {
			return true
		}
	}
	return false
}

func isStaged(action string) bool {
	return stringsContains(action, ".curse/staging/")
}

func isDirectAction(action string) bool {
	return stringsContains(action, "write") || stringsContains(action, "edit") || stringsContains(action, "delete")
}

func hasLogEntry(action string) bool {
	return stringsContains(action, "event.log")
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && containsString(s, substr)
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (r *ReviewReport) String() string {
	s := fmt.Sprintf("Review: passed=%v  blocks=%d  warnings=%d\n", r.Passed, r.Blocks, r.Warnings)
	for _, res := range r.Results {
		status := "PASS"
		if !res.Pass {
			if res.Rule.Severity == SeverityBlock {
				status = "BLOCK"
			} else {
				status = "WARN"
			}
		}
		s += fmt.Sprintf("  [%s] %s: %s\n", status, res.Rule.ID, res.Message)
	}
	return s
}
