package computer

import (
	"fmt"
	"strings"
)

type SafetyCheckResult struct {
	Passed          bool                `json:"passed"`
	ActionID        string              `json:"action_id"`
	ActionType      ActionType          `json:"action_type"`
	Target          string              `json:"target"`
	Classification  SafetyClassification `json:"classification"`
	RequiresReview  bool                `json:"requires_review"`
	Reviewed        bool                `json:"reviewed"`
	ReviewApproved  bool                `json:"review_approved"`
}

type SafetyChecker struct {
	engine *VisionEngine
}

func NewSafetyChecker(engine *VisionEngine) *SafetyChecker {
	return &SafetyChecker{
		engine: engine,
	}
}

func (sc *SafetyChecker) CheckClick(action *UIAction, elementHTML string) *SafetyCheckResult {
	result := &SafetyCheckResult{
		ActionID:   action.ID,
		ActionType: action.Type,
		Target:     action.Target,
	}

	selector := strings.ToLower(action.Target)
	html := strings.ToLower(elementHTML)

	classification := sc.classifyAction(selector, html)
	result.Classification = classification

	switch classification.Level {
	case SafetySafe:
		result.Passed = true
		result.RequiresReview = false
	case SafetyWarning:
		result.Passed = true
		result.RequiresReview = false
	case SafetyDestructive:
		result.Passed = false
		result.RequiresReview = true
	}

	return result
}

func (sc *SafetyChecker) CheckTerminal(command string) *SafetyCheckResult {
	result := &SafetyCheckResult{
		ActionType: ActionTerminal,
		Target:     command,
	}

	lower := strings.ToLower(command)

	destructivePatterns := []string{
		"rm ", "sudo ", "dd ", "format",
		":(){", "git push --force", "gh repo delete",
		"chmod -R 777", "rmdir /s", "del /f",
		"shutdown", "reboot", "init 0",
		"mv ", "> ", ">> ",
	}

	warningPatterns := []string{
		"git push", "git commit", "git merge",
		"npm publish", "npm run",
		"pip install", "go install",
		"docker ", "kubectl ",
		"curl ", "wget ",
	}

	for _, p := range destructivePatterns {
		if strings.Contains(lower, p) {
			result.Classification = SafetyClassification{
				Level:  SafetyDestructive,
				Reason: fmt.Sprintf("matches destructive pattern: %s", p),
			}
			result.Passed = false
			result.RequiresReview = true
			return result
		}
	}

	for _, p := range warningPatterns {
		if strings.Contains(lower, p) {
			result.Classification = SafetyClassification{
				Level:  SafetyWarning,
				Reason: fmt.Sprintf("matches warning pattern: %s", p),
			}
			result.Passed = true
			result.RequiresReview = false
			return result
		}
	}

	result.Classification = SafetyClassification{
		Level:  SafetySafe,
		Reason: "command appears safe",
	}
	result.Passed = true
	result.RequiresReview = false
	return result
}

func (sc *SafetyChecker) CheckFileOp(op, path string) *SafetyCheckResult {
	result := &SafetyCheckResult{
		ActionType: ActionFileOp,
		Target:     fmt.Sprintf("%s %s", op, path),
	}

	lower := strings.ToLower(path)

	protectedPaths := []string{
		".git", ".env", ".ssh",
		"go.sum", "go.mod",
		"CONSTITUTION.md",
		"models.json",
	}

	for _, p := range protectedPaths {
		if strings.Contains(lower, p) {
			result.Classification = SafetyClassification{
				Level:  SafetyDestructive,
				Reason: fmt.Sprintf("protected path: %s", p),
			}
			result.Passed = false
			result.RequiresReview = true
			return result
		}
	}

	if op == "delete" || op == "write" || op == "chmod" {
		result.Classification = SafetyClassification{
			Level:  SafetyDestructive,
			Reason: fmt.Sprintf("modifying operation: %s", op),
		}
		result.Passed = false
		result.RequiresReview = true
		return result
	}

	result.Classification = SafetyClassification{
		Level:  SafetySafe,
		Reason: "read operation",
	}
	result.Passed = true
	return result
}

func (sc *SafetyChecker) CheckDownload(url string) *SafetyCheckResult {
	result := &SafetyCheckResult{
		ActionType: ActionNavigate,
		Target:     url,
	}

	lower := strings.ToLower(url)

	if strings.Contains(lower, "github.com") || strings.Contains(lower, "git@") {
		if strings.Contains(lower, "/releases/download") ||
			strings.Contains(lower, "/archive/") ||
			strings.Contains(lower, ".git") {
			result.Classification = SafetyClassification{
				Level:  SafetyDestructive,
				Reason: "GitHub repository interaction",
			}
			result.Passed = false
			result.RequiresReview = true
			return result
		}
	}

	if strings.Contains(lower, "personal") || strings.Contains(lower, "private") ||
		strings.Contains(lower, "token") || strings.Contains(lower, "auth") {
		result.Classification = SafetyClassification{
			Level:  SafetyDestructive,
			Reason: "potential personal data access",
		}
		result.Passed = false
		result.RequiresReview = true
		return result
	}

	result.Classification = SafetyClassification{
		Level:  SafetySafe,
		Reason: "navigation appears safe",
	}
	result.Passed = true
	return result
}

func (sc *SafetyChecker) classifyAction(selector, html string) SafetyClassification {
	class := SafetyClassification{Level: SafetySafe, Reason: "no risk detected"}

	dangerPatterns := []struct {
		pattern string
		reason  string
		level   SafetyLevel
	}{
		{"delete", "delete action", SafetyDestructive},
		{"remove", "remove action", SafetyDestructive},
		{"destroy", "destroy action", SafetyDestructive},
		{"trash", "trash action", SafetyDestructive},
		{"unlink", "unlink action", SafetyDestructive},
		{"submit", "form submission", SafetyWarning},
		{"checkout", "checkout/purchase", SafetyDestructive},
		{"purchase", "purchase action", SafetyDestructive},
		{"confirm-payment", "payment confirmation", SafetyDestructive},
		{"fork", "repository fork", SafetyDestructive},
		{"transfer", "ownership transfer", SafetyDestructive},
		{"add-collaborator", "collaborator addition", SafetyDestructive},
		{"change-visibility", "visibility change", SafetyDestructive},
		{"password", "password field", SafetyWarning},
		{"token", "token access", SafetyDestructive},
		{".env", "environment file access", SafetyDestructive},
	}

	combined := selector + " " + html
	for _, dp := range dangerPatterns {
		if strings.Contains(combined, dp.pattern) {
			class = SafetyClassification{
				Level:  dp.level,
				Reason: dp.reason,
			}
			if dp.level == SafetyDestructive {
				break
			}
		}
	}

	return class
}
