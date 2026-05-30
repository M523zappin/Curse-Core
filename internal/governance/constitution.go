package governance

import (
	"fmt"
	"os"
	"strings"
)

type Severity int

const (
	SeverityBlock Severity = iota
	SeverityWarn
)

type Rule struct {
	ID          string
	Check       string
	Severity    Severity
	Description string
}

type Constitution struct {
	Principles []string
	Rules      []Rule
}

func Parse(path string) (*Constitution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read constitution: %w", err)
	}
	content := string(data)
	c := &Constitution{}
	lines := strings.Split(content, "\n")
	inPrinciples := false
	inRules := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Principles") {
			inPrinciples = true
			inRules = false
			continue
		}
		if strings.HasPrefix(trimmed, "## Guardrails") {
			inPrinciples = false
			inRules = true
			continue
		}
		if strings.HasPrefix(trimmed, "##") {
			inPrinciples = false
			inRules = false
		}
		if inPrinciples && strings.HasPrefix(trimmed, "1.") || inPrinciples && strings.HasPrefix(trimmed, "2.") ||
			inPrinciples && strings.HasPrefix(trimmed, "3.") || inPrinciples && strings.HasPrefix(trimmed, "4.") ||
			inPrinciples && strings.HasPrefix(trimmed, "5.") || inPrinciples && strings.HasPrefix(trimmed, "6.") ||
			inPrinciples && strings.HasPrefix(trimmed, "7.") || inPrinciples && strings.HasPrefix(trimmed, "8.") {
			c.Principles = append(c.Principles, trimmed)
		}
		if inRules && strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "|---") && !strings.HasPrefix(trimmed, "| Rule") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 5 {
				id := strings.TrimSpace(parts[1])
				check := strings.TrimSpace(parts[2])
				sevStr := strings.TrimSpace(parts[3])
				desc := strings.TrimSpace(parts[4])
				sev := SeverityWarn
				if sevStr == "block" || sevStr == "`block`" {
					sev = SeverityBlock
				}
				if id != "" && id != "-" {
					c.Rules = append(c.Rules, Rule{
						ID: id, Check: check,
						Severity: sev, Description: desc,
					})
				}
			}
		}
	}
	if len(c.Rules) == 0 {
		return nil, fmt.Errorf("no guardrail rules parsed from constitution")
	}
	return c, nil
}

func (c *Constitution) FindRule(id string) *Rule {
	for _, r := range c.Rules {
		if r.ID == id {
			return &r
		}
	}
	return nil
}
