package adapters

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type SystemAdapter struct {
	profile gateway.ModelProfile
}

func NewSystem(profile gateway.ModelProfile) *SystemAdapter {
	return &SystemAdapter{profile: profile}
}

func (a *SystemAdapter) Name() string { return "system" }
func (a *SystemAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *SystemAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	q := ""
	for _, m := range req.Messages {
		if m.Role == gateway.RoleUser {
			q = strings.ToLower(m.Content)
			break
		}
	}

	var b strings.Builder

	switch {
	case strings.Contains(q, "cpu"), strings.Contains(q, "processor"), strings.Contains(q, "hardware"):
		b.WriteString(fmt.Sprintf("🧠 CPU: %s (%d cores)\n", runtime.GOARCH, runtime.NumCPU()))
		b.WriteString(fmt.Sprintf("   OS: %s/%s\n", runtime.GOOS, runtime.GOARCH))
		b.WriteString(fmt.Sprintf("   Compiler: %s\n", runtime.Compiler))
		b.WriteString(fmt.Sprintf("   Go Version: %s\n", runtime.Version()))

	case strings.Contains(q, "memory"), strings.Contains(q, "ram"):
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		b.WriteString(fmt.Sprintf("💾 Memory Stats:\n"))
		b.WriteString(fmt.Sprintf("   Alloc:      %s\n", formatBytes(m.Alloc)))
		b.WriteString(fmt.Sprintf("   TotalAlloc: %s\n", formatBytes(m.TotalAlloc)))
		b.WriteString(fmt.Sprintf("   Sys:        %s\n", formatBytes(m.Sys)))
		b.WriteString(fmt.Sprintf("   HeapInuse:  %s\n", formatBytes(m.HeapInuse)))
		b.WriteString(fmt.Sprintf("   StackInuse: %s\n", formatBytes(m.StackInuse)))
		b.WriteString(fmt.Sprintf("   Goroutines: %d\n", runtime.NumGoroutine()))
		b.WriteString(fmt.Sprintf("   CGo Calls:  %d\n", runtime.NumCgoCall()))

	case strings.Contains(q, "disk"), strings.Contains(q, "storage"), strings.Contains(q, "file"):
		hostname, _ := os.Hostname()
		cwd, _ := os.Getwd()
		b.WriteString(fmt.Sprintf("💿 System:\n"))
		b.WriteString(fmt.Sprintf("   Hostname: %s\n", hostname))
		b.WriteString(fmt.Sprintf("   CWD:      %s\n", cwd))
		b.WriteString(fmt.Sprintf("   TempDir:  %s\n", os.TempDir()))
		b.WriteString(fmt.Sprintf("   Uptime:   %s\n", formatUptime()))

	case strings.Contains(q, "go"), strings.Contains(q, "golang"):
		b.WriteString(fmt.Sprintf("🔵 Go Runtime:\n"))
		b.WriteString(fmt.Sprintf("   Version:  %s\n", runtime.Version()))
		b.WriteString(fmt.Sprintf("   GOROOT:   %s\n", runtime.GOROOT()))
		b.WriteString(fmt.Sprintf("   GOOS:     %s\n", runtime.GOOS))
		b.WriteString(fmt.Sprintf("   GOARCH:   %s\n", runtime.GOARCH))
		b.WriteString(fmt.Sprintf("   Compiler: %s\n", runtime.Compiler))
		b.WriteString(fmt.Sprintf("   NumCPU:   %d\n", runtime.NumCPU()))
		b.WriteString(fmt.Sprintf("   NumGoroutine: %d\n", runtime.NumGoroutine()))

	case strings.Contains(q, "network"), strings.Contains(q, "ip"), strings.Contains(q, "dns"):
		hostname, _ := os.Hostname()
		b.WriteString(fmt.Sprintf("🌐 Network:\n"))
		b.WriteString(fmt.Sprintf("   Hostname: %s\n", hostname))
		b.WriteString(fmt.Sprintf("   TempDir:  %s\n", os.TempDir()))

	default:
		hostname, _ := os.Hostname()
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		b.WriteString(fmt.Sprintf("🖥️  System Report — %s\n\n", time.Now().Format(time.RFC1123)))
		b.WriteString(fmt.Sprintf("   Host:     %s\n", hostname))
		b.WriteString(fmt.Sprintf("   Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH))
		b.WriteString(fmt.Sprintf("   Go:       %s (%s)\n", runtime.Version(), runtime.Compiler))
		b.WriteString(fmt.Sprintf("   CPU Cores: %d\n", runtime.NumCPU()))
		b.WriteString(fmt.Sprintf("   Goroutines: %d\n", runtime.NumGoroutine()))
		b.WriteString(fmt.Sprintf("   Memory:   %s allocated / %s total / %s system\n",
			formatBytes(mem.Alloc), formatBytes(mem.TotalAlloc), formatBytes(mem.Sys)))
	}

	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: b.String()},
		Done:    true,
	}, nil
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

var startTime = time.Now()

func formatUptime() string {
	return time.Since(startTime).Round(time.Second).String()
}
