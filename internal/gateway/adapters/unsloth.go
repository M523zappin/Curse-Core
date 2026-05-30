package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type unslothRequest struct {
	Prompt   string `json:"prompt"`
	System   string `json:"system,omitempty"`
	Model    string `json:"model"`
	MaxTokens int   `json:"max_tokens"`
}

type unslothResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

type UnslothAdapter struct {
	profile  gateway.ModelProfile
	curseDir string

	mu       sync.Mutex
	cmd      *exec.Cmd
	stdin    *bufio.Writer
	stdout   *bufio.Scanner
	stderr   *bufio.Scanner
	procAlive bool
	modelName string
}

func NewUnsloth(profile gateway.ModelProfile, curseDir string) *UnslothAdapter {
	return &UnslothAdapter{
		profile:  profile,
		curseDir: curseDir,
	}
}

func (a *UnslothAdapter) Name() string { return "unsloth" }
func (a *UnslothAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *UnslothAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.ensureProc(ctx); err != nil {
		return nil, fmt.Errorf("unsloth: %w", err)
	}

	prompt := buildUnslothPrompt(req)
	model := a.profile.Model
	if model == "" {
		model = "unsloth/Llama-3.2-1B-Instruct"
	}

	ureq := unslothRequest{
		Prompt:    prompt,
		System:    req.System,
		Model:     model,
		MaxTokens: req.MaxTokens,
	}
	if ureq.MaxTokens <= 0 {
		ureq.MaxTokens = 2048
	}

	data, err := json.Marshal(ureq)
	if err != nil {
		a.kill()
		return nil, fmt.Errorf("unsloth marshal: %w", err)
	}
	if _, err := a.stdin.WriteString(string(data) + "\n"); err != nil {
		a.kill()
		return nil, fmt.Errorf("unsloth write: %w", err)
	}
	a.stdin.Flush()

	respCh := make(chan *unslothResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		if a.stdout.Scan() {
			var resp unslothResponse
			if err := json.Unmarshal([]byte(a.stdout.Text()), &resp); err != nil {
				errCh <- fmt.Errorf("unsloth decode: %w", err)
				return
			}
			respCh <- &resp
		} else {
			errCh <- fmt.Errorf("unsloth: process exited (stderr: %s)", readStderr(a.stderr))
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		a.kill()
		return nil, err
	case resp := <-respCh:
		if resp.Error != "" {
			return nil, fmt.Errorf("unsloth: %s", resp.Error)
		}
		return &gateway.Response{
			Message: gateway.Message{Role: gateway.RoleAssistant, Content: resp.Content},
			Done:    true,
		}, nil
	}
}

func (a *UnslothAdapter) ensureProc(ctx context.Context) error {
	if a.procAlive {
		return nil
	}

	helperPath := filepath.Join(a.curseDir, "unsloth_helper.py")
	if err := a.writeHelper(helperPath); err != nil {
		return fmt.Errorf("write helper: %w", err)
	}

	model := a.profile.Model
	if model == "" {
		model = "unsloth/Llama-3.2-1B-Instruct"
	}
	a.modelName = model

	pythonPath := findPython()
	if pythonPath == "" {
		return fmt.Errorf("python3 not found — install Python + unsloth: pip install unsloth")
	}

	cmd := exec.CommandContext(ctx, pythonPath, helperPath)
	cmd.Env = append(os.Environ(),
		"CURSE_UNSLOTH_MODEL="+model,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start unsloth: %w", err)
	}

	a.cmd = cmd
	a.stdin = bufio.NewWriter(stdin)
	a.stdout = bufio.NewScanner(stdout)
	a.stderr = bufio.NewScanner(stderr)
	a.procAlive = true

	// Wait for readiness signal from the Python process
	readyCh := make(chan error, 1)
	go func() {
		if a.stdout.Scan() {
			var readyResp unslothResponse
			if err := json.Unmarshal([]byte(a.stdout.Text()), &readyResp); err != nil {
				readyCh <- fmt.Errorf("readiness decode: %w", err)
				return
			}
			if readyResp.Error != "" {
				readyCh <- fmt.Errorf("readiness error: %s", readyResp.Error)
				return
			}
			if readyResp.Content != "__ready__" {
				readyCh <- fmt.Errorf("unexpected readiness signal: %s", readyResp.Content)
				return
			}
			readyCh <- nil
		} else {
			errMsg := readStderr(a.stderr)
			readyCh <- fmt.Errorf("process exited before readiness (stderr: %s)", errMsg)
		}
	}()
	select {
	case <-ctx.Done():
		a.kill()
		return ctx.Err()
	case err := <-readyCh:
		if err != nil {
			a.kill()
			return fmt.Errorf("unsloth readiness: %w", err)
		}
	}

	go func() {
		cmd.Wait()
		a.mu.Lock()
		a.procAlive = false
		a.mu.Unlock()
	}()

	return nil
}

func (a *UnslothAdapter) writeHelper(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create helper dir: %w", err)
	}

	script := `#!/usr/bin/env python3
"""CURSE Unsloth helper — persistent LLM inference process.
Reads JSON requests from stdin, writes JSON responses to stdout.
Model is loaded once and kept alive for the lifetime of the process.
"""

import json
import os
import sys
import traceback

MODEL_NAME = os.environ.get("CURSE_UNSLOTH_MODEL", "unsloth/Llama-3.2-1B-Instruct")

model = None
tokenizer = None

def load_model():
    global model, tokenizer
    try:
        from unsloth import FastLanguageModel
        import torch
    except ImportError:
        # Try loading without unsloth (use plain transformers)
        try:
            from transformers import AutoModelForCausalLM, AutoTokenizer
            model = AutoModelForCausalLM.from_pretrained(
                MODEL_NAME,
                device_map="auto",
                torch_dtype="auto",
            )
            tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
            return
        except ImportError:
            json.dump({"error": "neither unsloth nor transformers found — pip install unsloth"}, sys.stdout)
            sys.stdout.flush()
            sys.exit(1)

    try:
        model, tokenizer = FastLanguageModel.from_pretrained(
            model_name=MODEL_NAME,
            max_seq_length=8192,
            dtype=None,
            load_in_4bit=True,
            device_map="auto",
        )
        FastLanguageModel.for_inference(model)
    except Exception as e:
        json.dump({"error": f"load model: {e}"}, sys.stdout)
        sys.stdout.flush()
        sys.exit(1)

def generate(prompt, system="", max_tokens=2048):
    global model, tokenizer
    if model is None:
        load_model()

    messages = []
    if system:
        messages.append({"role": "system", "content": system})
    messages.append({"role": "user", "content": prompt})

    try:
        from unsloth import FastLanguageModel
        inputs = tokenizer.apply_chat_template(
            messages,
            tokenize=True,
            add_generation_prompt=True,
            return_tensors="pt",
        ).to(model.device)

        outputs = model.generate(
            inputs,
            max_new_tokens=max_tokens,
            temperature=0.3,
            top_p=0.9,
            repetition_penalty=1.1,
            do_sample=True,
        )
        response = tokenizer.decode(outputs[0][inputs.shape[1]:], skip_special_tokens=True)
        return response.strip()
    except NameError:
        # transformers-only fallback
        inputs = tokenizer.apply_chat_template(messages, return_tensors="pt")
        outputs = model.generate(inputs, max_new_tokens=max_tokens)
        response = tokenizer.decode(outputs[0][inputs.shape[1]:], skip_special_tokens=True)
        return response.strip()
    except Exception as e:
        return f"[unsloth error: {e}]"

if __name__ == "__main__":
    # Signal readiness
    json.dump({"content": "__ready__"}, sys.stdout)
    sys.stdout.flush()

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
            content = generate(
                prompt=req.get("prompt", ""),
                system=req.get("system", ""),
                max_tokens=req.get("max_tokens", 2048),
            )
            json.dump({"content": content}, sys.stdout)
        except Exception as e:
            json.dump({"error": f"{e}: {traceback.format_exc()}"}, sys.stdout)
        sys.stdout.flush()
`
	return os.WriteFile(path, []byte(script), 0644)
}

func (a *UnslothAdapter) kill() {
	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
		a.cmd.Wait()
	}
	a.procAlive = false
}

func buildUnslothPrompt(req *gateway.Prompt) string {
	var b strings.Builder
	for _, m := range req.Messages {
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}
	return b.String()
}

func readStderr(s *bufio.Scanner) string {
	var b strings.Builder
	for s.Scan() {
		b.WriteString(s.Text() + "\n")
	}
	return strings.TrimSpace(b.String())
}

func findPython() string {
	for _, name := range []string{"python3", "python"} {
		path, err := exec.LookPath(name)
		if err == nil {
			return path
		}
	}
	return ""
}

func DetectUnsloth(ctx context.Context) (bool, string) {
	pythonPath := findPython()
	if pythonPath == "" {
		return false, ""
	}
	cmd := exec.CommandContext(ctx, pythonPath, "-c", "import unsloth; print(unsloth.__version__)")
	out, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(out))
}

func listUnslothModels(ctx context.Context) ([]string, error) {
	return []string{
		"unsloth/Llama-3.2-1B-Instruct",
		"unsloth/Llama-3.2-3B-Instruct",
		"unsloth/Mistral-7B-Instruct-v0.3",
		"unsloth/Qwen2.5-1.5B-Instruct",
		"unsloth/Qwen2.5-7B-Instruct",
		"unsloth/gemma-2-2b-it",
		"unsloth/Phi-3.5-mini-instruct",
	}, nil
}
