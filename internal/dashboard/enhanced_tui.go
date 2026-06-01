package dashboard

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Enhanced TUI Components for a Superior Terminal Experience

// ═══════════════════════════════════════════════════════════════════════════
// COMMAND PALETTE (Ctrl+K) - VSCode-style fuzzy finder
// ═══════════════════════════════════════════════════════════════════════════

type CommandPalette struct {
	visible     bool
	query       string
	selectedIdx int
	commands    []CommandItem
	filtered    []CommandItem
	width       int
	height      int
}

type CommandItem struct {
	ID          string
	Name        string
	Description string
	Category    string
	Shortcut    string
	Action      func() tea.Msg
}

func NewCommandPalette() *CommandPalette {
	cmds := []CommandItem{
		{"new_file", "New File", "Create a new file", "File", "⌘N"},
		{"open_file", "Open File", "Open an existing file", "File", "⌘O"},
		{"save", "Save", "Save current file", "File", "⌘S"},
		{"find", "Find in Files", "Search across codebase", "Edit", "⌘F"},
		{"replace", "Find & Replace", "Find and replace text", "Edit", "⌘H"},
		{"format", "Format Code", "Auto-format the code", "Edit", "⇧⌘F"},
		{"comment", "Toggle Comment", "Comment/uncomment lines", "Edit", "⌘/"},
		
		{"git_status", "Git Status", "Show git status", "Git", "⌘G"},
		{"git_commit", "Git Commit", "Commit changes", "Git", "⌘⏎"},
		{"git_push", "Git Push", "Push to remote", "Git", ""},
		{"git_pull", "Git Pull", "Pull from remote", "Git", ""},
		{"git_diff", "Git Diff", "Show changes", "Git", "⌘D"},
		{"git_branch", "Git Branch", "Manage branches", "Git", ""},
		
		{"run_test", "Run Tests", "Execute test suite", "Run", "⌘T"},
		{"run_build", "Build", "Build the project", "Run", "⌘B"},
		{"run_debug", "Debug", "Start debugging", "Run", "⌘⌥D"},
		{"stop", "Stop", "Stop current operation", "Run", "⌘."},
		
		{"model_list", "List Models", "Show available AI models", "AI", "/list"},
		{"model_switch", "Switch Model", "Change active model", "AI", "/model"},
		{"model_info", "Model Info", "Show current model details", "AI", "/stats"},
		
		{"settings", "Settings", "Open settings", "View", "⌘,"},
		{"toggle_terminal", "Toggle Terminal", "Show/hide terminal", "View", "⌘`"},
		{"toggle_sidebar", "Toggle Sidebar", "Show/hide file tree", "View", "⌘B"},
		{"split_view", "Split View", "Toggle split view", "View", "⌘\\"},
		
		{"help", "Help", "Show help and shortcuts", "Help", "F1"},
		{"about", "About", "Show about info", "Help", ""},
		{"check_updates", "Check Updates", "Check for updates", "Help", ""},
	}

	return &CommandPalette{
		commands:    cmds,
		filtered:    cmds,
		visible:     false,
		selectedIdx: 0,
	}
}

func (cp *CommandPalette) Toggle() {
	cp.visible = !cp.visible
	cp.query = ""
	cp.selectedIdx = 0
	cp.filter()
}

func (cp *CommandPalette) UpdateQuery(q string) {
	cp.query = q
	cp.filter()
}

func (cp *CommandPalette) filter() {
	if cp.query == "" {
		cp.filtered = cp.commands
		return
	}

	lower := strings.ToLower(cp.query)
	cp.filtered = nil
	for _, cmd := range cp.commands {
		if strings.Contains(strings.ToLower(cmd.Name), lower) ||
			strings.Contains(strings.ToLower(cmd.Description), lower) ||
			strings.Contains(strings.ToLower(cmd.Category), lower) {
			cp.filtered = append(cp.filtered, cmd)
		}
	}
}

func (cp *CommandPalette) MoveUp() {
	if cp.selectedIdx > 0 {
		cp.selectedIdx--
	}
}

func (cp *CommandPalette) MoveDown() {
	if cp.selectedIdx < len(cp.filtered)-1 {
		cp.selectedIdx++
	}
}

func (cp *CommandPalette) Selected() *CommandItem {
	if cp.selectedIdx < len(cp.filtered) {
		return &cp.filtered[cp.selectedIdx]
	}
	return nil
}

func (cp *CommandPalette) View() string {
	if !cp.visible {
		return ""
	}

	var buf strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Width(cp.width - 2)

	buf.WriteString(headerStyle.Render("┌" + strings.Repeat("─", cp.width-2) + "┐") + "\n")
	buf.WriteString(headerStyle.Render("│ 🔍 Command Palette") + "\n")
	buf.WriteString(headerStyle.Render("├" + strings.Repeat("─", cp.width-2) + "┤") + "\n")

	// Search input
	inputStyle := lipgloss.NewStyle().
		Foreground(ColorFg).
		Width(cp.width - 4)
	buf.WriteString(fmt.Sprintf("│ %s│\n", inputStyle.Render("> " + cp.query + "_")))

	buf.WriteString(headerStyle.Render("├" + strings.Repeat("─", cp.width-2) + "┤") + "\n")

	// Results
	maxItems := cp.height - 10
	if maxItems > len(cp.filtered) {
		maxItems = len(cp.filtered)
	}

	for i := 0; i < maxItems; i++ {
		cmd := cp.filtered[i]
		isSelected := i == cp.selectedIdx

		var line string
		if isSelected {
			selected := lipgloss.NewStyle().
				Background(ColorAccent).
				Foreground(ColorBg).
				Bold(true)
			line = fmt.Sprintf("│ ▸ %s", selected.Render(cmd.Name))
			if cmd.Shortcut != "" {
				line += " " + lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("(" + cmd.Shortcut + ")")
			}
		} else {
			line = fmt.Sprintf("│   %s", cmd.Name)
			if cmd.Shortcut != "" {
				line += " " + lipgloss.NewStyle().Foreground(ColorFgInactive).Render("(" + cmd.Shortcut + ")")
			}
		}

		// Pad to width
		line = fmt.Sprintf("%-"+fmt.Sprintf("%d", cp.width-2)+"s", line)
		buf.WriteString(line + "│\n")
	}

	buf.WriteString(headerStyle.Render("├" + strings.Repeat("─", cp.width-2) + "┤") + "\n")
	
	// Footer
	hint := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("↑↓ navigate · enter select · esc close")
	buf.WriteString(fmt.Sprintf("│ %-"+fmt.Sprintf("%d", cp.width-4)+"s│\n", hint))
	buf.WriteString(headerStyle.Render("└" + strings.Repeat("─", cp.width-2) + "┘"))

	return buf.String()
}

// ═══════════════════════════════════════════════════════════════════════════
// FILE BROWSER - Interactive tree with icons and git status
// ═══════════════════════════════════════════════════════════════════════════

type FileBrowser struct {
	visible     bool
	root        string
	expanded    map[string]bool
	selected    string
	showHidden  bool
	width       int
	height      int
	gitStatus   map[string]string // file -> status (M, A, D, etc.)
}

type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	Depth    int
}

func NewFileBrowser(root string) *FileBrowser {
	return &FileBrowser{
		visible:   true,
		root:       root,
		expanded:  make(map[string]bool),
		gitStatus: make(map[string]string),
	}
}

func (fb *FileBrowser) Toggle() {
	fb.visible = !fb.visible
}

func (fb *FileBrowser) ToggleHidden() {
	fb.showHidden = !fb.showHidden
}

func (fb *FileBrowser) Expand(path string) {
	fb.expanded[path] = true
}

func (fb *FileBrowser) Collapse(path string) {
	delete(fb.expanded, path)
}

func (fb *FileBrowser) ToggleExpand(path string) {
	if fb.expanded[path] {
		delete(fb.expanded, path)
	} else {
		fb.expanded[path] = true
	}
}

func (fb *FileBrowser) Select(path string) {
	fb.selected = path
}

func (fb *FileBrowser) SetGitStatus(path, status string) {
	fb.gitStatus[path] = status
}

func (fb *FileBrowser) View() string {
	if !fb.visible {
		return ""
	}

	var buf strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	buf.WriteString(headerStyle.Render("┌─ Explorer ─────────────┐") + "\n")

	// Build tree
	files := fb.buildTree(fb.root, 0)
	for _, f := range files {
		line := fb.renderFileNode(f)
		buf.WriteString(line + "\n")
	}

	buf.WriteString(headerStyle.Render("└" + strings.Repeat("─", 26) + "┘"))

	return buf.String()
}

func (fb *FileBrowser) buildTree(dir string, depth int) []*FileNode {
	var nodes []*FileNode

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nodes
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && !fb.showHidden {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		node := &FileNode{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
			Depth: depth,
		}

		if entry.IsDir() {
			node.Children = fb.buildTree(fullPath, depth+1)
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (fb *FileBrowser) renderFileNode(node *FileNode) string {
	indent := strings.Repeat("  ", node.Depth)
	
	var icon string
	if node.IsDir {
		expanded := fb.expanded[node.Path]
		if expanded {
			icon = "📂 "
		} else {
			icon = "📁 "
		}
	} else {
		icon = fb.getFileIcon(node.Name)
	}

	// Git status indicator
	status := fb.gitStatus[node.Path]
	var statusIcon string
	switch status {
	case "M":
		statusIcon = " ✏️"
	case "A":
		statusIcon = " ✅"
	case "D":
		statusIcon = " ❌"
	case "?":
		statusIcon = " ◌"
	default:
		statusIcon = ""
	}

	selected := ""
	if fb.selected == node.Path {
		selected = lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorBg).
			Render(" ")
	}

	name := node.Name
	if fb.selected == node.Path {
		name = lipgloss.NewStyle().
			Foreground(ColorBg).
			Bold(true).
			Render(name)
	} else if node.IsDir {
		name = lipgloss.NewStyle().
			Foreground(ColorToxic).
			Render(name)
	} else {
		name = lipgloss.NewStyle().
			Foreground(ColorFg).
			Render(name)
	}

	expand := ""
	if node.IsDir {
		expanded := fb.expanded[node.Path]
		if expanded {
			expand = "▼"
		} else {
			expand = "▶"
		}
	}

	return fmt.Sprintf("%s%s %s%s%s%s", selected, indent, expand, icon, name, statusIcon)
}

func (fb *FileBrowser) getFileIcon(name string) string {
	ext := filepath.Ext(name)
	switch strings.ToLower(ext) {
	case ".go":
		return "🐹 "
	case ".py":
		return "🐍 "
	case ".js", ".ts", ".jsx", ".tsx":
		return "⚡ "
	case ".rs":
		return "🦀 "
	case ".java":
		return "☕ "
	case ".md":
		return "📝 "
	case ".json":
		return "{ }"
	case ".yaml", ".yml":
		return "⚙️ "
	case ".toml":
		return "📦 "
	case ".sh":
		return "🔧 "
	case ".css":
		return "🎨 "
	case ".html":
		return "🌐 "
	case ".sql":
		return "🗄️ "
	default:
		return "📄 "
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// SPLIT VIEW - Code and chat side by side
// ═══════════════════════════════════════════════════════════════════════════

type SplitView struct {
	enabled        bool
	leftPane       string
	rightPane      string
	leftWidth      int
	rightWidth     int
	leftPaneTitle  string
	rightPaneTitle string
}

func NewSplitView() *SplitView {
	return &SplitView{
		enabled:    false,
		leftWidth:  50,
		rightWidth: 50,
	}
}

func (sv *SplitView) Toggle() {
	sv.enabled = !sv.enabled
}

func (sv *SplitView) SetLeftPane(title, content string) {
	sv.leftPaneTitle = title
	sv.leftPane = content
}

func (sv *SplitView) SetRightPane(title, content string) {
	sv.rightPaneTitle = title
	sv.rightPane = content
}

func (sv *SplitView) AdjustRatio(delta int) {
	sv.leftWidth += delta
	sv.rightWidth -= delta
	if sv.leftWidth < 30 {
		sv.leftWidth = 30
		sv.rightWidth = 70
	}
	if sv.leftWidth > 70 {
		sv.leftWidth = 70
		sv.rightWidth = 30
	}
}

func (sv *SplitView) View(totalWidth int) string {
	if !sv.enabled {
		return ""
	}

	leftW := totalWidth * sv.leftWidth / 100
	rightW := totalWidth - leftW - 4

	leftPanel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorAccent).
		Width(leftW).
		Render(sv.leftPaneTitle + "\n" + sv.leftPane)

	rightPanel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorSpiral).
		Width(rightW).
		Render(sv.rightPaneTitle + "\n" + sv.rightPane)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
}

// ═══════════════════════════════════════════════════════════════════════════
// DIFF VIEWER - Side-by-side code comparison
// ═══════════════════════════════════════════════════════════════════════════

type DiffViewer struct {
	visible     bool
	leftContent string
	rightContent string
	leftTitle    string
	rightTitle   string
	width        int
}

func NewDiffViewer() *DiffViewer {
	return &DiffViewer{
		visible: false,
	}
}

func (dv *DiffViewer) Toggle() {
	dv.visible = !dv.visible
}

func (dv *DiffViewer) SetDiff(left, right, leftTitle, rightTitle string) {
	dv.leftContent = left
	dv.rightContent = right
	dv.leftTitle = leftTitle
	dv.rightTitle = rightTitle
}

type DiffLine struct {
	LineNum int
	Content string
	Type    string // "added", "removed", "context"
}

func (dv *DiffViewer) View() string {
	if !dv.visible {
		return ""
	}

	var buf strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	buf.WriteString("┌─ Diff ──────────────────────────────┐\n")

	// Simplified diff display
	lines := strings.Split(dv.leftContent, "\n")
	for i, line := range lines {
		if i > 20 {
			buf.WriteString(lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("  ... (truncated)"))
			break
		}

		var lineStyle lipgloss.Style
		if strings.HasPrefix(line, "+") {
			lineStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Background("#0a2f0a")
		} else if strings.HasPrefix(line, "-") {
			lineStyle = lipgloss.NewStyle().Foreground(ColorError).Background("#2f0a0a")
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(ColorFg)
		}

		buf.WriteString(fmt.Sprintf("  %s\n", lineStyle.Render(fmt.Sprintf("%3d %s", i+1, line))))
	}

	buf.WriteString("└" + strings.Repeat("─", 37) + "┘")

	return buf.String()
}

// ═══════════════════════════════════════════════════════════════════════════
// ANIMATED PROCESSING - Better progress indicators
// ═══════════════════════════════════════════════════════════════════════════

type AnimatedSpinner struct {
	frames []string
	frame  int
}

func NewAnimatedSpinner() *AnimatedSpinner {
	return &AnimatedSpinner{
		frames: []string{
			"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		},
	}
}

func (as *AnimatedSpinner) Next() string {
	as.frame = (as.frame + 1) % len(as.frames)
	return as.frames[as.frame]
}

type ProgressBar struct {
	total     int
	current   int
	barWidth  int
	completed string
	remaining string
	animated  bool
}

func NewProgressBar(total, width int) *ProgressBar {
	return &ProgressBar{
		total:     total,
		barWidth:  width,
		completed: "█",
		remaining: "░",
	}
}

func (pb *ProgressBar) SetProgress(current int) {
	pb.current = current
}

func (pb *ProgressBar) View() string {
	if pb.total == 0 {
		return lipgloss.NewStyle().Foreground(ColorFgInactive).Render("[░░░░░░░░░░]")
	}

	progress := float64(pb.current) / float64(pb.total)
	completed := int(progress * float64(pb.barWidth))
	remaining := pb.barWidth - completed

	bar := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Render(strings.Repeat(pb.completed, completed))

	bar += lipgloss.NewStyle().
		Foreground(ColorFgInactive).
		Render(strings.Repeat(pb.remaining, remaining))

	pct := int(progress * 100)

	return fmt.Sprintf("[%s] %d%%", bar, pct)
}

// MultiStepProgress - Shows multiple progress steps
type MultiStepProgress struct {
	steps    []StepInfo
	current  int
	width    int
	frame    int
}

type StepInfo struct {
	Name      string
	Status    string // "pending", "running", "done", "error"
	Progress  int
	Message   string
}

func NewMultiStepProgress(steps []string, width int) *MultiStepProgress {
	stepInfos := make([]StepInfo, len(steps))
	for i, name := range steps {
		stepInfos[i] = StepInfo{Name: name, Status: "pending", Progress: 0}
	}
	return &MultiStepProgress{steps: stepInfos, width: width}
}

func (msp *MultiStepProgress) SetStepStatus(idx int, status string, message string) {
	if idx < len(msp.steps) {
		msp.steps[idx].Status = status
		msp.steps[idx].Message = message
	}
	msp.current = idx
}

func (msp *MultiStepProgress) SetStepProgress(idx, progress int) {
	if idx < len(msp.steps) {
		msp.steps[idx].Progress = progress
	}
}

func (msp *MultiStepProgress) View() string {
	var buf strings.Builder

	for i, step := range msp.steps {
		var icon, color lipgloss.Color
		switch step.Status {
		case "done":
			icon = "✓"
			color = ColorSuccess
		case "running":
			icon = "●"
			color = PulseColor(msp.frame)
		case "error":
			icon = "✗"
			color = ColorError
		default:
			icon = "○"
			color = ColorFgInactive
		}

		stepName := lipgloss.NewStyle().Foreground(color).Render(icon + " " + step.Name)
		
		var progressStr string
		if step.Status == "running" && step.Progress > 0 {
			progressStr = lipgloss.NewStyle().Foreground(ColorFgSubtle).
				Render(fmt.Sprintf(" %d%%", step.Progress))
		}

		var message string
		if step.Message != "" {
			message = lipgloss.NewStyle().Foreground(ColorFgSubtle).
				Render(" — " + step.Message)
		}

		buf.WriteString(fmt.Sprintf("  %s%s%s\n", stepName, progressStr, message))
	}

	return buf.String()
}

// ═══════════════════════════════════════════════════════════════════════════
// CODE SYNTAX HIGHLIGHTING - Simple token-based highlighter
// ═══════════════════════════════════════════════════════════════════════════

type SyntaxHighlighter struct {
	theme map[string]lipgloss.Color
}

func NewSyntaxHighlighter() *SyntaxHighlighter {
	return &SyntaxHighlighter{
		theme: map[string]lipgloss.Color{
			"keyword":    ColorToxic,     // func, return, if, else
			"type":       ColorSpiral,    // int, string, bool
			"string":     ColorSuccess,   // "hello"
			"comment":    ColorFgInactive, // // comment
			"number":     ColorWarning,    // 123, 45.67
			"function":   ColorAccent,     // functionName
			"operator":   ColorFgSubtle,   // +, -, *, /
			"punctuation": ColorFg,        // (), {}, []
		},
	}
}

func (sh *SyntaxHighlighter) Highlight(code, language string) string {
	var result strings.Builder

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		result.WriteString(sh.highlightLine(line, language))
		result.WriteString("\n")
	}

	return result.String()
}

func (sh *SyntaxHighlighter) highlightLine(line, language string) string {
	switch language {
	case "go":
		return sh.highlightGo(line)
	case "python":
		return sh.highlightPython(line)
	case "javascript", "typescript":
		return sh.highlightJS(line)
	default:
		return line
	}
}

func (sh *SyntaxHighlighter) highlightGo(line string) string {
	// Keywords
	keywords := []string{"func", "return", "if", "else", "for", "switch", "case", "default",
		"break", "continue", "type", "struct", "interface", "package", "import",
		"const", "var", "map", "chan", "go", "defer", "select", "fallthrough", "goto"}

	for _, kw := range keywords {
		line = highlightKeyword(line, kw, sh.theme["keyword"])
	}

	// Types
	types := []string{"string", "int", "int64", "int32", "float64", "float32", "bool", "byte", "rune", "error", "any"}
	for _, t := range types {
		line = highlightKeyword(line, t, sh.theme["type"])
	}

	return line
}

func (sh *SyntaxHighlighter) highlightPython(line string) string {
	keywords := []string{"def", "class", "return", "if", "elif", "else", "for", "while",
		"try", "except", "finally", "with", "as", "import", "from", "True", "False", "None",
		"async", "await", "lambda", "yield", "pass", "break", "continue"}

	for _, kw := range keywords {
		line = highlightKeyword(line, kw, sh.theme["keyword"])
	}

	return line
}

func (sh *SyntaxHighlighter) highlightJS(line string) string {
	keywords := []string{"function", "const", "let", "var", "return", "if", "else", "for",
		"while", "switch", "case", "default", "class", "extends", "import", "export",
		"async", "await", "try", "catch", "finally", "throw", "new", "this", "true", "false"}

	for _, kw := range keywords {
		line = highlightKeyword(line, kw, sh.theme["keyword"])
	}

	return line
}

func highlightKeyword(line, keyword string, color lipgloss.Color) string {
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	// Simple replacement - in production you'd want word-boundary matching
	replaced := strings.ReplaceAll(line, keyword, style.Render(keyword))
	return replaced
}

// ═══════════════════════════════════════════════════════════════════════════
// ENHANCED STATUS BAR - Better system information display
// ═══════════════════════════════════════════════════════════════════════════

type EnhancedStatusBar struct {
	cpuHistory    []float64
	memHistory    []float64
	networkHistory []float64
	width         int
}

func NewEnhancedStatusBar(width int) *EnhancedStatusBar {
	return &EnhancedStatusBar{
		cpuHistory:     make([]float64, 0, 60),
		memHistory:     make([]float64, 0, 60),
		networkHistory: make([]float64, 0, 60),
		width:          width,
	}
}

func (esb *EnhancedStatusBar) UpdateMetrics(cpu, mem float64) {
	esb.cpuHistory = append(esb.cpuHistory, cpu)
	esb.memHistory = append(esb.memHistory, mem)
	
	if len(esb.cpuHistory) > 60 {
		esb.cpuHistory = esb.cpuHistory[1:]
	}
	if len(esb.memHistory) > 60 {
		esb.memHistory = esb.memHistory[1:]
	}
}

func (esb *EnhancedStatusBar) MiniSparkline(history []float64, width int) string {
	if len(history) == 0 {
		return ""
	}

	var buf strings.Builder
	max := history[0]
	min := history[0]
	for _, v := range history {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}

	for _, v := range history {
		normalized := (v - min) / range_
		height := int(normalized * 4)
		if height < 1 {
			height = 1
		}

		switch height {
		case 1:
			buf.WriteString("▁")
		case 2:
			buf.WriteString("▂")
		case 3:
			buf.WriteString("▃")
		case 4:
			buf.WriteString("▄")
		default:
			buf.WriteString("░")
		}
	}

	return buf.String()
}

func (esb *EnhancedStatusBar) View() string {
	var buf strings.Builder

	// CPU Line
	cpuSpark := esb.MiniSparkline(esb.cpuHistory, 40)
	cpuLabel := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("CPU")
	cpuVal := esb.getCurrentOrAvg(esb.cpuHistory)
	cpuBar := esb.renderMiniBar(cpuVal, 10)
	buf.WriteString(fmt.Sprintf("  %s %s %s %s\n", cpuLabel, cpuBar, cpuSpark, 
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf("%.1f%%", cpuVal))))

	// Memory Line
	memSpark := esb.MiniSparkline(esb.memHistory, 40)
	memLabel := lipgloss.NewStyle().Foreground(ColorSpiral).Bold(true).Render("MEM")
	memVal := esb.getCurrentOrAvg(esb.memHistory)
	memBar := esb.renderMiniBar(memVal, 10)
	buf.WriteString(fmt.Sprintf("  %s %s %s %s\n", memLabel, memBar, memSpark,
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf("%.0fMB", memVal))))

	return buf.String()
}

func (esb *EnhancedStatusBar) getCurrentOrAvg(history []float64) float64 {
	if len(history) == 0 {
		return 0
	}
	if len(history) < 5 {
		sum := 0.0
		for _, v := range history {
			sum += v
		}
		return sum / float64(len(history))
	}
	return history[len(history)-1]
}

func (esb *EnhancedStatusBar) renderMiniBar(value float64, width int) string {
	filled := int(value / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := lipgloss.NewStyle().Foreground(ColorSuccess).Render(strings.Repeat("█", filled))
	bar += lipgloss.NewStyle().Foreground(ColorFgInactive).Render(strings.Repeat("░", width-filled))
	return "[" + bar + "]"
}

// ═══════════════════════════════════════════════════════════════════════════
// NOTIFICATION SYSTEM - Non-blocking toast notifications
// ═══════════════════════════════════════════════════════════════════════════

type Notification struct {
	ID      string
	Type    string // "info", "success", "warning", "error"
	Title   string
	Message string
	Time    time.Time
	Duration time.Duration
}

type NotificationManager struct {
	notifications []Notification
	maxVisible    int
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		notifications: make([]Notification, 0),
		maxVisible:    5,
	}
}

func (nm *NotificationManager) Add(notif Notification) {
	notif.Time = time.Now()
	nm.notifications = append(nm.notifications, notif)
	
	if len(nm.notifications) > 20 {
		nm.notifications = nm.notifications[1:]
	}
}

func (nm *NotificationManager) Info(title, message string) {
	nm.Add(Notification{Type: "info", Title: title, Message: message})
}

func (nm *NotificationManager) Success(title, message string) {
	nm.Add(Notification{Type: "success", Title: title, Message: message})
}

func (nm *NotificationManager) Warning(title, message string) {
	nm.Add(Notification{Type: "warning", Title: title, Message: message})
}

func (nm *NotificationManager) Error(title, message string) {
	nm.Add(Notification{Type: "error", Title: title, Message: message})
}

func (nm *NotificationManager) View() string {
	var buf strings.Builder

	visible := nm.notifications
	if len(visible) > nm.maxVisible {
		visible = visible[len(visible)-nm.maxVisible:]
	}

	for _, n := range visible {
		var color lipgloss.Color
		var icon string
		switch n.Type {
		case "success":
			color = ColorSuccess
			icon = "✓"
		case "warning":
			color = ColorWarning
			icon = "⚠"
		case "error":
			color = ColorError
			icon = "✗"
		default:
			color = ColorAccent
			icon = "ℹ"
		}

		notifStyle := lipgloss.NewStyle().
			Foreground(color)

		titleStyle := notifStyle.Bold(true)

		buf.WriteString(fmt.Sprintf("  %s %s: %s\n",
			notifStyle.Render(icon),
			titleStyle.Render(n.Title),
			lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(n.Message)))
	}

	return buf.String()
}

