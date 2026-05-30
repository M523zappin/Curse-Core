package computer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ActionType string

const (
	ActionNavigate  ActionType = "navigate"
	ActionClick     ActionType = "click"
	ActionTypeText  ActionType = "type"
	ActionScroll    ActionType = "scroll"
	ActionScreenshot ActionType = "screenshot"
	ActionLaunch    ActionType = "launch_app"
	ActionFileOp    ActionType = "file_operation"
	ActionTerminal  ActionType = "terminal"
)

type SafetyLevel int

const (
	SafetySafe      SafetyLevel = iota
	SafetyWarning
	SafetyDestructive
)

type UIAction struct {
	ID          string      `json:"id"`
	Type        ActionType  `json:"type"`
	Target      string      `json:"target"`
	Value       string      `json:"value,omitempty"`
	Coordinates struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"coordinates,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
	Screenshot  string      `json:"screenshot,omitempty"`
	ElementHTML string      `json:"element_html,omitempty"`
	SafetyLevel SafetyLevel `json:"safety_level"`
	Reviewed    bool        `json:"reviewed"`
	Confirmed   bool        `json:"confirmed"`
	Error       string      `json:"error,omitempty"`
}

type ReviewRequest struct {
	Action    UIAction
	Channel   chan ReviewDecision
	CreatedAt time.Time
}

type ReviewDecision struct {
	Approved bool          `json:"approved"`
	Reason   string        `json:"reason"`
	Scope    ApprovalScope `json:"scope,omitempty"`
}

type VisionSnapshot struct {
	Timestamp   time.Time `json:"timestamp"`
	Screenshot  string    `json:"screenshot"`
	ActionID    string    `json:"action_id"`
	PageTitle   string    `json:"page_title,omitempty"`
	URL         string    `json:"url,omitempty"`
	ElementInfo string    `json:"element_info,omitempty"`
}

type ComputerController struct {
	mu             sync.RWMutex
	browserMgr     *BrowserManager
	desktopMgr     *DesktopManager
	visionBuffer   []VisionSnapshot
	maxBufferSize  int
	pendingReview  *ReviewRequest
	reviewCallback func(ReviewRequest)
	ctx            context.Context
	cancel         context.CancelFunc
}

func New() *ComputerController {
	ctx, cancel := context.WithCancel(context.Background())
	return &ComputerController{
		browserMgr:    NewBrowserManager(),
		desktopMgr:    NewDesktopManager(),
		visionBuffer:  make([]VisionSnapshot, 0, 100),
		maxBufferSize: 100,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (cc *ComputerController) StartBrowser() error {
	return cc.browserMgr.Start(cc.ctx)
}

func (cc *ComputerController) StopBrowser() error {
	cc.cancel()
	return cc.browserMgr.Stop()
}

func (cc *ComputerController) Navigate(url string) (*UIAction, error) {
	action := cc.newAction(ActionNavigate, url, "")
	action.SafetyLevel = SafetySafe

	shot, err := cc.browserMgr.Navigate(cc.ctx, url)
	if err != nil {
		action.Error = err.Error()
		cc.recordVision(action)
		return action, fmt.Errorf("navigate: %w", err)
	}
	action.Screenshot = shot
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) Click(selector string) (*UIAction, error) {
	action := cc.newAction(ActionClick, selector, "")

	screenshot, html, err := cc.browserMgr.PreClickSnapshot(cc.ctx, selector)
	if err == nil {
		action.Screenshot = screenshot
		action.ElementHTML = html
	}

	action.SafetyLevel = cc.classifyClick(selector, html)

	if action.SafetyLevel == SafetyDestructive {
		approved, err := cc.requestReview(action)
		if err != nil || !approved {
			action.Error = "blocked by review"
			cc.recordVision(action)
			return action, fmt.Errorf("click blocked: review rejected")
		}
		action.Reviewed = true
		action.Confirmed = true
	}

	if err := cc.browserMgr.Click(cc.ctx, selector); err != nil {
		action.Error = err.Error()
		cc.recordVision(action)
		return action, fmt.Errorf("click: %w", err)
	}

	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) TypeText(selector, text string) (*UIAction, error) {
	action := cc.newAction(ActionTypeText, selector, text)
	action.SafetyLevel = SafetySafe

	if err := cc.browserMgr.Type(cc.ctx, selector, text); err != nil {
		action.Error = err.Error()
		cc.recordVision(action)
		return action, fmt.Errorf("type: %w", err)
	}
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) Scroll(selector string, deltaY int) (*UIAction, error) {
	action := cc.newAction(ActionScroll, selector, fmt.Sprintf("%d", deltaY))
	action.SafetyLevel = SafetySafe

	if err := cc.browserMgr.Scroll(cc.ctx, selector, deltaY); err != nil {
		action.Error = err.Error()
		return action, fmt.Errorf("scroll: %w", err)
	}
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) Screenshot() (*UIAction, error) {
	action := cc.newAction(ActionScreenshot, "full_page", "")
	action.SafetyLevel = SafetySafe

	shot, err := cc.browserMgr.Screenshot(cc.ctx)
	if err != nil {
		return action, fmt.Errorf("screenshot: %w", err)
	}
	action.Screenshot = shot
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) LaunchApp(appName string, args []string) (*UIAction, error) {
	action := cc.newAction(ActionLaunch, appName, "")
	action.SafetyLevel = SafetyWarning

	if err := cc.desktopMgr.Launch(appName, args); err != nil {
		action.Error = err.Error()
		return action, fmt.Errorf("launch: %w", err)
	}
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) FileOperation(op, path, content string) (*UIAction, error) {
	action := cc.newAction(ActionFileOp, path, content)
	action.SafetyLevel = SafetyDestructive

	approved, err := cc.requestReview(action)
	if err != nil || !approved {
		action.Error = "blocked by review"
		cc.recordVision(action)
		return action, fmt.Errorf("file op blocked: review rejected")
	}
	action.Reviewed = true
	action.Confirmed = true

	if err := cc.desktopMgr.FileOp(op, path, content); err != nil {
		action.Error = err.Error()
		return action, fmt.Errorf("file op: %w", err)
	}
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) Terminal(command string) (*UIAction, error) {
	action := cc.newAction(ActionTerminal, command, "")
	action.SafetyLevel = cc.classifyTerminal(command)

	if action.SafetyLevel == SafetyDestructive {
		approved, err := cc.requestReview(action)
		if err != nil || !approved {
			action.Error = "blocked by review"
			return action, fmt.Errorf("terminal blocked: review rejected")
		}
		action.Reviewed = true
		action.Confirmed = true
	}

	out, err := cc.desktopMgr.RunCommand(command)
	if err != nil {
		action.Error = err.Error()
		return action, fmt.Errorf("terminal: %w", err)
	}
	_ = out
	cc.recordVision(action)
	return action, nil
}

func (cc *ComputerController) SetReviewCallback(cb func(ReviewRequest)) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.reviewCallback = cb
}

func (cc *ComputerController) PendingReview() *ReviewRequest {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.pendingReview
}

func (cc *ComputerController) ResolveReview(decision ReviewDecision) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.pendingReview != nil {
		cc.pendingReview.Channel <- decision
		cc.pendingReview = nil
	}
}

func (cc *ComputerController) VisionBuffer() []VisionSnapshot {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	out := make([]VisionSnapshot, len(cc.visionBuffer))
	copy(out, cc.visionBuffer)
	return out
}

func (cc *ComputerController) ClearVisionBuffer() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.visionBuffer = make([]VisionSnapshot, 0, cc.maxBufferSize)
}

func (cc *ComputerController) BrowserManager() *BrowserManager {
	return cc.browserMgr
}

func (cc *ComputerController) DesktopManager() *DesktopManager {
	return cc.desktopMgr
}

func (cc *ComputerController) newAction(typ ActionType, target, value string) *UIAction {
	return &UIAction{
		ID:        fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Type:      typ,
		Target:    target,
		Value:     value,
		Timestamp: time.Now().UTC(),
	}
}

func (cc *ComputerController) recordVision(action *UIAction) {
	ss := VisionSnapshot{
		Timestamp:  action.Timestamp,
		Screenshot: action.Screenshot,
		ActionID:   action.ID,
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.visionBuffer = append(cc.visionBuffer, ss)
	if len(cc.visionBuffer) > cc.maxBufferSize {
		cc.visionBuffer = cc.visionBuffer[1:]
	}
}

func (cc *ComputerController) classifyClick(selector, html string) SafetyLevel {
	destructive := []string{"delete", "remove", "destroy", "trash", "unlink",
		"submit", "checkout", "purchase", "confirm-payment",
		"fork", "transfer", "add-collaborator", "change-visibility"}
	forbidden := []string{".env", ".git", "token", "secret", "password"}

	for _, d := range destructive {
		if contains(selector, d) || contains(html, d) {
			return SafetyDestructive
		}
	}
	for _, f := range forbidden {
		if contains(selector, f) || contains(html, f) {
			return SafetyDestructive
		}
	}
	return SafetySafe
}

func (cc *ComputerController) classifyTerminal(command string) SafetyLevel {
	destructive := []string{"rm ", "sudo ", "dd ", "format ",
		":(){ :|:& };:", "git push --force", "gh repo delete",
		"chmod -R 777", "rmdir /s", "del /f /s"}

	upper := command
	for _, d := range destructive {
		if contains(upper, d) {
			return SafetyDestructive
		}
	}
	return SafetySafe
}

func (cc *ComputerController) requestReview(action *UIAction) (bool, error) {
	ch := make(chan ReviewDecision, 1)
	req := ReviewRequest{
		Action:    *action,
		Channel:   ch,
		CreatedAt: time.Now(),
	}

	cc.mu.Lock()
	cc.pendingReview = &req
	cb := cc.reviewCallback
	cc.mu.Unlock()

	if cb != nil {
		cb(req)
	}

	select {
	case decision := <-ch:
		return decision.Approved, nil
	case <-time.After(5 * time.Minute):
		return false, fmt.Errorf("review timeout")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	lower := s
	for i := 0; i <= len(lower)-len(substr); i++ {
		if lower[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
