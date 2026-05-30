package computer

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type VisionEngine struct {
	screenshotDir string
}

func NewVisionEngine(screenshotDir string) *VisionEngine {
	os.MkdirAll(screenshotDir, 0755)
	return &VisionEngine{
		screenshotDir: screenshotDir,
	}
}

type ElementInfo struct {
	TagName   string  `json:"tag_name"`
	ID        string  `json:"id,omitempty"`
	Classes   string  `json:"classes,omitempty"`
	Text      string  `json:"text,omitempty"`
	Href      string  `json:"href,omitempty"`
	Src       string  `json:"src,omitempty"`
	Type      string  `json:"type,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	RectX     int     `json:"rect_x"`
	RectY     int     `json:"rect_y"`
	RectW     int     `json:"rect_w"`
	RectH     int     `json:"rect_h"`
	IsVisible bool    `json:"is_visible"`
	IsClickable bool  `json:"is_clickable"`
}

type SafetyClassification struct {
	Level        SafetyLevel  `json:"level"`
	Reason       string       `json:"reason"`
	ElementInfo  *ElementInfo  `json:"element_info,omitempty"`
	Warnings     []string     `json:"warnings,omitempty"`
	IsClickable  bool         `json:"is_clickable"`
}

func (ve *VisionEngine) SaveScreenshot(base64Data, actionID string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("decode screenshot: %w", err)
	}

	hash := sha256.Sum256(data)
	filename := fmt.Sprintf("%s_%x.png", actionID, hash[:8])
	path := filepath.Join(ve.screenshotDir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("save screenshot: %w", err)
	}
	return path, nil
}

func (ve *VisionEngine) AnalyzeScreenshot(base64Data string) (*SafetyClassification, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	class := &SafetyClassification{
		Level:  SafetySafe,
		Reason: "screenshot captured",
	}

	if len(data) == 0 {
		class.Warnings = append(class.Warnings, "screenshot appears empty")
	}

	return class, nil
}

func (ve *VisionEngine) ClassifyElement(info *ElementInfo) SafetyClassification {
	class := SafetyClassification{Level: SafetySafe}

	if info == nil {
		class.Reason = "no element info available"
		return class
	}

	class.ElementInfo = info

	switch info.TagName {
	case "button", "input", "a":
		class.IsClickable = true
		class.Reason = "clickable element"
	}

	text := strings.ToLower(info.Text + info.Placeholder + info.ID + info.Classes)

	if strings.Contains(text, "delete") || strings.Contains(text, "remove") {
		class.Level = SafetyDestructive
		class.Reason = "destructive action: delete/remove detected"
		class.Warnings = append(class.Warnings, "element appears to be a destructive control")
	}

	if strings.Contains(text, "submit") || strings.Contains(text, "confirm") {
		if strings.Contains(text, "payment") || strings.Contains(text, "purchase") || strings.Contains(text, "checkout") {
			class.Level = SafetyDestructive
			class.Reason = "financial transaction detected"
			class.Warnings = append(class.Warnings, "this action may result in a financial transaction")
		} else {
			class.Level = SafetyWarning
			class.Reason = "submission action, review recommended"
		}
	}

	if info.Type == "password" || info.Type == "email" {
		class.Level = SafetyWarning
		class.Warnings = append(class.Warnings, "interacting with sensitive input field")
	}

	if info.Href != "" {
		class.IsClickable = true
		if strings.Contains(info.Href, "logout") || strings.Contains(info.Href, "delete") {
			class.Level = SafetyDestructive
			class.Reason = "navigation to destructive endpoint"
		}
	}

	return class
}

func (ve *VisionEngine) CompareScreenshots(before, after string) (changed bool, diffPct float64, err error) {
	beforeData, err := os.ReadFile(before)
	if err != nil {
		return false, 0, err
	}
	afterData, err := os.ReadFile(after)
	if err != nil {
		return false, 0, err
	}

	beforeHash := sha256.Sum256(beforeData)
	afterHash := sha256.Sum256(afterData)

	return beforeHash != afterHash, 0, nil
}

func (ve *VisionEngine) ScreenshotDir() string {
	return ve.screenshotDir
}

func ParseElementHTML(html string) *ElementInfo {
	if html == "" {
		return nil
	}

	info := &ElementInfo{}

	html = strings.ToLower(html)

	if start := strings.Index(html, "<"); start >= 0 {
		if end := strings.Index(html[start+1:], " "); end >= 0 {
			info.TagName = html[start+1 : start+1+end]
		} else if end := strings.Index(html[start+1:], ">"); end >= 0 {
			info.TagName = html[start+1 : start+1+end]
		}
	}

	extractAttr := func(name string) string {
		search := name + `="`
		if idx := strings.Index(html, search); idx >= 0 {
			start := idx + len(search)
			if end := strings.Index(html[start:], `"`); end >= 0 {
				return html[start : start+end]
			}
		}
		return ""
	}

	info.ID = extractAttr("id")
	info.Classes = extractAttr("class")
	info.Href = extractAttr("href")
	info.Src = extractAttr("src")
	info.Type = extractAttr("type")
	info.Placeholder = extractAttr("placeholder")

	if strings.Contains(html, ">") && strings.Contains(html, "</") {
		start := strings.Index(html, ">") + 1
		end := strings.LastIndex(html, "</")
		if end > start {
			info.Text = strings.TrimSpace(html[start:end])
		}
	}

	return info
}
