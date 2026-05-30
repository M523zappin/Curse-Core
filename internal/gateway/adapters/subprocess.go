package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type SubprocessAdapter struct {
	profile gateway.ModelProfile
}

func NewSubprocess(profile gateway.ModelProfile) *SubprocessAdapter {
	return &SubprocessAdapter{profile: profile}
}

func (a *SubprocessAdapter) Name() string { return "subprocess" }

func (a *SubprocessAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *SubprocessAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	cmdPattern := a.profile.Model
	if cmdPattern == "" {
		cmdPattern = "cat"
	}

	promptText := buildSubprocessPrompt(req)

	output, err := a.execCommand(ctx, cmdPattern, promptText)
	if err != nil {
		return nil, fmt.Errorf("subprocess: %w", err)
	}

	return &gateway.Response{
		Message: gateway.Message{
			Role:    gateway.RoleAssistant,
			Content: output,
		},
		Done: true,
	}, nil
}

func (a *SubprocessAdapter) execCommand(ctx context.Context, pattern, input string) (string, error) {
	parts := splitArgs(pattern)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command pattern")
	}

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.CommandContext(ctx, parts[0])
	} else {
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("stdin pipe: %w", err)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start: %w", err)
	}

	_, writeErr := stdin.Write([]byte(input))
	stdin.Close()
	if writeErr != nil {
		cmd.Wait()
		return "", fmt.Errorf("write stdin: %w", writeErr)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return stdout.String(), fmt.Errorf("stderr: %s", strings.TrimSpace(stderr.String()))
			}
			return stdout.String(), fmt.Errorf("exit: %w", err)
		}
		return strings.TrimSpace(stdout.String()), nil
	}
}

func buildSubprocessPrompt(req *gateway.Prompt) string {
	var b strings.Builder
	if req.System != "" {
		b.WriteString(req.System)
		b.WriteString("\n\n")
	}
	for _, m := range req.Messages {
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}
	return b.String()
}

func LookupPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func splitArgs(cmd string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		switch {
		case c == '"' || c == '\'':
			if inQuote && c == quoteChar {
				inQuote = false
			} else if !inQuote {
				inQuote = true
				quoteChar = c
			} else {
				current.WriteByte(c)
			}
		case c == ' ' || c == '\t':
			if inQuote {
				current.WriteByte(c)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

var DetectableTools = []struct {
	Name     string
	CheckCmd string
	Version  string
	Default  string
}{
	{"python3", "python3 --version", "--version", `python3 -c "import sys; print(sys.stdin.read())"`},
	{"python", "python --version", "--version", `python -c "import sys; print(sys.stdin.read())"`},
	{"ollama", "ollama --version", "--version", "ollama run"},
	{"llama-cli", "llama-cli --version", "--version", "llama-cli --prompt"},
}

func DetectLocalTools() []string {
	available := make([]string, 0)
	for _, tool := range DetectableTools {
		if LookupPath(tool.CheckCmd) != "" {
			available = append(available, tool.Name)
		} else {
			parts := strings.Fields(tool.CheckCmd)
			if LookupPath(parts[0]) != "" {
				available = append(available, tool.Name)
			}
		}
	}
	return available
}

func init() {
	if _, err := exec.LookPath("python3"); err != nil {
		os.Setenv("CURSE_NO_PYTHON", "1")
	} else {
		os.Setenv("CURSE_PYTHON_PATH", LookupPath("python3"))
	}
}
