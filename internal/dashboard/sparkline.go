package dashboard

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	sparkChars = "▁▂▃▄▅▆▇█"
	sparkLen   = 12
)

type Sparkline struct {
	mu     sync.Mutex
	values []float64
	min    float64
	max    float64
	label  string
	color  lipgloss.Color
}

func NewSparkline(label string, color lipgloss.Color) *Sparkline {
	return &Sparkline{
		values: make([]float64, 0, sparkLen),
		label:  label,
		color:  color,
	}
}

func (s *Sparkline) Push(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, v)
	if len(s.values) > sparkLen {
		s.values = s.values[len(s.values)-sparkLen:]
	}
	s.recalc()
}

func (s *Sparkline) recalc() {
	if len(s.values) == 0 {
		return
	}
	s.min = s.values[0]
	s.max = s.values[0]
	for _, v := range s.values {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
	if s.max-s.min < 0.001 {
		s.max = s.min + 1
	}
}

func (s *Sparkline) View() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.values) == 0 {
		return ""
	}

	chars := make([]byte, len(s.values))
	for i, v := range s.values {
		normalized := (v - s.min) / (s.max - s.min)
		idx := int(math.Round(normalized * 7))
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		chars[i] = sparkChars[idx]
	}

	sparkStyle := lipgloss.NewStyle().Foreground(s.color)
	valStr := fmt.Sprintf("%.0f", s.values[len(s.values)-1])

	return fmt.Sprintf("%s %s%s",
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(s.label+":"),
		sparkStyle.Render(string(chars)),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(" "+valStr))
}

var systemSparklines struct {
	once sync.Once
	CPU  *Sparkline
	Mem  *Sparkline
	Goro *Sparkline
}

func initSystemSparklines() {
	systemSparklines.once.Do(func() {
		systemSparklines.CPU = NewSparkline("CPU", ColorProcessing)
		systemSparklines.Mem = NewSparkline("MEM", ColorSpiral)
		systemSparklines.Goro = NewSparkline("GOR", ColorToxic)
	})
}

var (
	tickMu      sync.Mutex
	lastTickAt  time.Time
)

func measureLoad() float64 {
	tickMu.Lock()
	defer tickMu.Unlock()

	now := time.Now()
	if lastTickAt.IsZero() {
		lastTickAt = now
		return 0
	}

	elapsed := now.Sub(lastTickAt)
	lastTickAt = now

	expected := 200 * time.Millisecond
	if elapsed < expected {
		elapsed = expected
	}

	load := float64(elapsed-expected) / float64(expected)
	if load < 0 {
		load = 0
	}
	if load > 8 {
		load = 8
	}
	return load
}

func tickSparklines() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	systemSparklines.CPU.Push(measureLoad())
	systemSparklines.Mem.Push(float64(mem.Alloc / 1024 / 1024))
	systemSparklines.Goro.Push(float64(runtime.NumGoroutine()))
}

func RenderSystemSparklines() string {
	initSystemSparklines()
	return lipgloss.JoinHorizontal(lipgloss.Center,
		systemSparklines.CPU.View(),
		lipgloss.NewStyle().Foreground(ColorBorder).Render(" │ "),
		systemSparklines.Mem.View(),
		lipgloss.NewStyle().Foreground(ColorBorder).Render(" │ "),
		systemSparklines.Goro.View(),
	)
}
