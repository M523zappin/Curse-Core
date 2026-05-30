package lsp

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
	"time"
)

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source"`
	Code     string `json:"code,omitempty"`
}

type CompletionItem struct {
	Label       string `json:"label"`
	Kind        int    `json:"kind"`
	Detail      string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

type SymbolInfo struct {
	Name          string `json:"name"`
	Kind          int    `json:"kind"`
	ContainerName string `json:"containerName,omitempty"`
	Detail        string `json:"detail,omitempty"`
	FilePath      string `json:"filePath,omitempty"`
}

type Client struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	stdin       *bufio.Writer
	stdout      *bufio.Scanner
	stderr      *bufio.Scanner
	connected   bool
	serverPath  string
	rootURI     string
	msgID       int
	capabilities map[string]interface{}
}

func NewClient(serverPath, workspaceRoot string) *Client {
	return &Client{
		serverPath: serverPath,
		rootURI:    fmt.Sprintf("file://%s", strings.ReplaceAll(workspaceRoot, "\\", "/")),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := exec.CommandContext(ctx, c.serverPath)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	c.cmd = cmd
	c.stdin = bufio.NewWriter(stdin)
	c.stdout = bufio.NewScanner(bufio.NewReaderSize(stdout, 65536))
	c.stderr = bufio.NewScanner(stderr)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start LSP server: %w", err)
	}

	c.connected = true

	go func() {
		for c.stderr.Scan() {
			_ = c.stderr.Text()
		}
	}()

	if err := c.initialize(); err != nil {
		c.connected = false
		return fmt.Errorf("LSP initialize: %w", err)
	}

	return nil
}

func (c *Client) initialize() error {
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   c.rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport": true,
					},
				},
				"diagnostics": true,
				"definition":  true,
				"references":  true,
				"symbol":      true,
				"hover":       true,
			},
		},
	}

	resp, err := c.sendRequest("initialize", params)
	if err != nil {
		return err
	}

	if caps, ok := resp["capabilities"].(map[string]interface{}); ok {
		c.capabilities = caps
	}

	c.sendNotification("initialized", map[string]interface{}{})
	return nil
}

func (c *Client) OpenDocument(filePath string, content string) error {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": c.detectLanguage(filePath),
			"version":    1,
			"text":       content,
		},
	}
	return c.sendNotification("textDocument/didOpen", params)
}

func (c *Client) ChangeDocument(filePath string, content string, version int) error {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": []map[string]interface{}{
			{"text": content},
		},
	}
	return c.sendNotification("textDocument/didChange", params)
}

func (c *Client) GetDiagnostics(filePath string) ([]Diagnostic, error) {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	resp, err := c.sendRequest("textDocument/diagnostic", params)
	if err != nil {
		return nil, err
	}

	diagnostics := c.extractDiagnostics(resp)
	return diagnostics, nil
}

func (c *Client) GetCompletions(filePath string, line, character int) ([]CompletionItem, error) {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
		"context": map[string]interface{}{
			"triggerKind": 1,
		},
	}

	resp, err := c.sendRequest("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	return c.extractCompletions(resp), nil
}

func (c *Client) GetSymbols(filePath string) ([]SymbolInfo, error) {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}

	resp, err := c.sendRequest("textDocument/documentSymbol", params)
	if err != nil {
		return nil, err
	}

	return c.extractSymbols(resp, filePath), nil
}

func (c *Client) GetDefinition(filePath string, line, character int) ([]SymbolInfo, error) {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	resp, err := c.sendRequest("textDocument/definition", params)
	if err != nil {
		return nil, err
	}

	return c.extractLocations(resp), nil
}

func (c *Client) Hover(filePath string, line, character int) (string, error) {
	uri := fmt.Sprintf("file://%s", strings.ReplaceAll(filePath, "\\", "/"))

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	resp, err := c.sendRequest("textDocument/hover", params)
	if err != nil {
		return "", err
	}

	return c.extractHover(resp), nil
}

func (c *Client) Shutdown() error {
	_, err := c.sendRequest("shutdown", nil)
	if err != nil {
		return err
	}
	c.sendNotification("exit", nil)
	c.connected = false
	return c.cmd.Wait()
}

func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

func (c *Client) sendRequest(method string, params interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	c.msgID++
	id := c.msgID
	c.mu.Unlock()

	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	c.mu.Lock()
	if c.stdin == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("LSP not connected")
	}
	c.stdin.WriteString(header)
	c.stdin.Write(data)
	c.stdin.Flush()
	c.mu.Unlock()

	return c.readResponse(id)
}

func (c *Client) sendNotification(method string, params interface{}) error {
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	c.mu.Lock()
	if c.stdin != nil {
		c.stdin.WriteString(header)
		c.stdin.Write(data)
		c.stdin.Flush()
	}
	c.mu.Unlock()
	return nil
}

func (c *Client) readResponse(id int) (map[string]interface{}, error) {
	timeout := time.After(30 * time.Second)
	done := make(chan map[string]interface{}, 1)
	errs := make(chan error, 1)

	go func() {
		for c.stdout.Scan() {
			line := c.stdout.Text()

			if strings.HasPrefix(line, "Content-Length:") || line == "" {
				continue
			}

			var resp map[string]interface{}
			if err := json.Unmarshal([]byte(line), &resp); err != nil {
				continue
			}

			respID, _ := resp["id"].(float64)
			if int(respID) == id {
				if result, ok := resp["result"]; ok {
					if resultMap, ok := result.(map[string]interface{}); ok {
						done <- resultMap
						return
					}
					done <- map[string]interface{}{"result": result}
					return
				}
				if errData, ok := resp["error"]; ok {
					errMap := errData.(map[string]interface{})
					errs <- fmt.Errorf("LSP error: %v", errMap["message"])
					return
				}
			}
		}
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errs:
		return nil, err
	case <-timeout:
		return nil, fmt.Errorf("LSP response timeout")
	}
}

func (c *Client) extractDiagnostics(resp map[string]interface{}) []Diagnostic {
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return nil
	}
	items, ok := result["items"].([]interface{})
	if !ok {
		return nil
	}
	var diags []Diagnostic
	for _, item := range items {
		diagMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		diag := Diagnostic{
			Message: getStr(diagMap, "message"),
			Source:  getStr(diagMap, "source"),
			Code:    getStr(diagMap, "code"),
		}
		if severity, ok := diagMap["severity"].(float64); ok {
			diag.Severity = int(severity)
		}
		if rng, ok := diagMap["range"].(map[string]interface{}); ok {
			if start, ok := rng["start"].(map[string]interface{}); ok {
				diag.Range.Start.Line = int(getFloat(start, "line"))
				diag.Range.Start.Character = int(getFloat(start, "character"))
			}
			if end, ok := rng["end"].(map[string]interface{}); ok {
				diag.Range.End.Line = int(getFloat(end, "line"))
				diag.Range.End.Character = int(getFloat(end, "character"))
			}
		}
		diags = append(diags, diag)
	}
	return diags
}

func (c *Client) extractCompletions(resp map[string]interface{}) []CompletionItem {
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return nil
	}
	items, ok := result["items"].([]interface{})
	if !ok {
		if itemsArr, ok := result["result"].([]interface{}); ok {
			items = itemsArr
		} else {
			return nil
		}
	}
	var completions []CompletionItem
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		comp := CompletionItem{
			Label:  getStr(itemMap, "label"),
			Detail: getStr(itemMap, "detail"),
			Documentation: getStr(itemMap, "documentation"),
		}
		if kind, ok := itemMap["kind"].(float64); ok {
			comp.Kind = int(kind)
		}
		completions = append(completions, comp)
	}
	return completions
}

func (c *Client) extractSymbols(resp map[string]interface{}, filePath string) []SymbolInfo {
	result, ok := resp["result"].([]interface{})
	if !ok {
		return nil
	}
	var symbols []SymbolInfo
	for _, item := range result {
		symMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		sym := SymbolInfo{
			Name:          getStr(symMap, "name"),
			ContainerName: getStr(symMap, "containerName"),
			Detail:        getStr(symMap, "detail"),
			FilePath:      filePath,
		}
		if kind, ok := symMap["kind"].(float64); ok {
			sym.Kind = int(kind)
		}
		symbols = append(symbols, sym)
	}
	return symbols
}

func (c *Client) extractLocations(resp map[string]interface{}) []SymbolInfo {
	result, ok := resp["result"].([]interface{})
	if !ok {
		if single, ok := resp["result"].(map[string]interface{}); ok {
			result = []interface{}{single}
		} else {
			return nil
		}
	}
	var symbols []SymbolInfo
	for _, item := range result {
		locMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		uri := getStr(locMap, "uri")
		filePath := strings.TrimPrefix(uri, "file://")
		sym := SymbolInfo{FilePath: filePath}

		if rng, ok := locMap["range"].(map[string]interface{}); ok {
			if start, ok := rng["start"].(map[string]interface{}); ok {
				_ = start
			}
		}
		symbols = append(symbols, sym)
	}
	return symbols
}

func (c *Client) extractHover(resp map[string]interface{}) string {
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return ""
	}
	contents, ok := result["contents"]
	if !ok {
		return ""
	}

	switch c := contents.(type) {
	case map[string]interface{}:
		return getStr(c, "value")
	case []interface{}:
		var parts []string
		for _, item := range c {
			if m, ok := item.(map[string]interface{}); ok {
				parts = append(parts, getStr(m, "value"))
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func (c *Client) detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".ts":
		return strings.TrimPrefix(ext, ".")
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	default:
		return "plaintext"
	}
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func FindLSServer(language string) string {
	servers := map[string]string{
		"go":   "gopls",
		"ts":   "typescript-language-server",
		"js":   "typescript-language-server",
		"py":   "pylsp",
		"rust": "rust-analyzer",
	}
	if s, ok := servers[language]; ok {
		if path, err := exec.LookPath(s); err == nil {
			return path
		}
	}
	return ""
}
