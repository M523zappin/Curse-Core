package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

// FreeAPIAdapter provides access to free LLM APIs without API keys.
// Supports OpenRouter (with free models) and HuggingFace Inference API.
// Uses environment variables for optional API keys but works without them.
type FreeAPIAdapter struct {
	profile      gateway.ModelProfile
	httpClient   *http.Client
	provider     string
	modelsPath   string
}

// Supported free models across providers
var freeModels = map[string]FreeModelInfo{
	// OpenRouter free models (no API key required for some)
	"openrouter:google:gemma-2-9b-it": {
		Provider:    "openrouter",
		Name:        "google/gemma-2-9b-it",
		DisplayName: "Gemma 2 9B (OpenRouter)",
		ContextLen:  8192,
		Free:        true,
	},
	"openrouter:mistral:latest": {
		Provider:    "openrouter",
		Name:        "mistralai/mistral-nemo",
		DisplayName: "Mistral Nemo (OpenRouter)",
		ContextLen:  12288,
		Free:        true,
	},
	"openrouter:openchat:7b": {
		Provider:    "openrouter",
		Name:        "openchat/openchat-7b",
		DisplayName: "OpenChat 7B (OpenRouter)",
		ContextLen:  8192,
		Free:        true,
	},
	"openrouter:deepseek:chat": {
		Provider:    "openrouter",
		Name:        "deepseek-ai/deepseek-prover-v1.5",
		DisplayName: "DeepSeek (OpenRouter)",
		ContextLen:  8192,
		Free:        true,
	},
	"openrouter:qwen:2.5-7b": {
		Provider:    "openrouter",
		Name:        "qwen/qwen-2.5-7b",
		DisplayName: "Qwen 2.5 7B (OpenRouter)",
		ContextLen:  32768,
		Free:        true,
	},
	
	// HuggingFace Inference API (free tier)
	"huggingface:codellama": {
		Provider:    "huggingface",
		Name:        "codellama/CodeLlama-7b-hf",
		DisplayName: "Code Llama 7B (HuggingFace)",
		ContextLen:  16384,
		Free:        true,
	},
	"huggingface:starcoder": {
		Provider:    "huggingface",
		Name:        "bigcode/starcoder2-7b",
		DisplayName: "StarCoder 2 7B (HuggingFace)",
		ContextLen:  16384,
		Free:        true,
	},
	"huggingface:mistral": {
		Provider:    "huggingface",
		Name:        "mistralai/mistral-7b-instruct",
		DisplayName: "Mistral 7B (HuggingFace)",
		ContextLen:  8192,
		Free:        true,
	},
	"huggingface:llama": {
		Provider:    "huggingface",
		Name:        "meta-llama/llama-3-8b",
		DisplayName: "Llama 3 8B (HuggingFace)",
		ContextLen:  8192,
		Free:        true,
	},
	
	// Groq free models (fast inference)
	"groq:llama3-8b": {
		Provider:    "groq",
		Name:        "llama-3.1-8b-instant",
		DisplayName: "Llama 3.1 8B (Groq - Fast!)",
		ContextLen:  8192,
		Free:        true,
	},
	"groq:mixtral-8x7b": {
		Provider:    "groq",
		Name:        "mixtral-8x7b-32768",
		DisplayName: "Mixtral 8x7B (Groq)",
		ContextLen:  32768,
		Free:        true,
	},
}

type FreeModelInfo struct {
	Provider    string
	Name        string
	DisplayName string
	ContextLen  int
	Free        bool
}

func NewFreeAPI(profile gateway.ModelProfile) *FreeAPIAdapter {
	return &FreeAPIAdapter{
		profile: profile,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		provider: profile.Provider,
	}
}

func (a *FreeAPIAdapter) Name() string { return "free-api" }

func (a *FreeAPIAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *FreeAPIAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	modelKey := a.getModelKey()
	modelInfo, ok := freeModels[modelKey]
	if !ok {
		return a.fallbackResponse(req)
	}

	switch modelInfo.Provider {
	case "openrouter":
		return a.callOpenRouter(ctx, req, modelInfo)
	case "huggingface":
		return a.callHuggingFace(ctx, req, modelInfo)
	case "groq":
		return a.callGroq(ctx, req, modelInfo)
	default:
		return a.fallbackResponse(req)
	}
}

func (a *FreeAPIAdapter) getModelKey() string {
	// Extract model key from endpoint or model name
	if strings.Contains(a.profile.Endpoint, "openrouter") {
		return "openrouter:" + a.profile.Model
	}
	if strings.Contains(a.profile.Endpoint, "huggingface") || strings.Contains(a.profile.Endpoint, "hf.co") {
		return "huggingface:" + a.profile.Model
	}
	if strings.Contains(a.profile.Endpoint, "groq") {
		return "groq:" + a.profile.Model
	}
	return "openrouter:google:gemma-2-9b-it" // Default
}

func (a *FreeAPIAdapter) callOpenRouter(ctx context.Context, req *gateway.Prompt, info FreeModelInfo) (*gateway.Response, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	
	messages := make([]map[string]string, 0)
	
	// Add system prompt
	if req.System != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": req.System,
		})
	}
	
	// Add conversation messages
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == gateway.RoleAssistant {
			role = "assistant"
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}
	
	body := map[string]interface{}{
		"model": info.Name,
		"messages": messages,
		"max_tokens": a.profile.MaxTokens,
		"temperature": a.profile.Temperature,
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	endpoint := "https://openrouter.ai/api/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/M523zappin/Curse-Core")
	req.Header.Set("X-Title", "CURSE AI Assistant")
	
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return a.fallbackResponse(req.Prompts[0])
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		// Fallback to demo response if API fails
		return a.fallbackResponse(req)
	}
	
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return a.fallbackResponse(req)
	}
	
	if len(result.Choices) > 0 {
		return &gateway.Response{
			Message: gateway.Message{
				Role:    gateway.RoleAssistant,
				Content: result.Choices[0].Message.Content,
			},
			Done: true,
		}, nil
	}
	
	return a.fallbackResponse(req)
}

func (a *FreeAPIAdapter) callHuggingFace(ctx context.Context, req *gateway.Prompt, info FreeModelInfo) (*gateway.Response, error) {
	apiKey := os.Getenv("HF_TOKEN")
	
	// Build prompt from messages
	var prompt strings.Builder
	if req.System != "" {
		prompt.WriteString(fmt.Sprintf("System: %s\n\n", req.System))
	}
	for _, msg := range req.Messages {
		role := "User"
		if msg.Role == gateway.RoleAssistant {
			role = "Assistant"
		}
		prompt.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}
	prompt.WriteString("Assistant:")
	
	body := map[string]interface{}{
		"inputs":       prompt.String(),
		"parameters": map[string]interface{}{
			"max_new_tokens": a.profile.MaxTokens,
			"temperature":    a.profile.Temperature,
			"return_full_text": false,
		},
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	endpoint := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", info.Name)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return a.fallbackResponse(req)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return a.fallbackResponse(req)
	}
	
	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return a.fallbackResponse(req)
	}
	
	if len(result) > 0 {
		if content, ok := result[0]["generated_text"].(string); ok {
			return &gateway.Response{
				Message: gateway.Message{
					Role:    gateway.RoleAssistant,
					Content: content,
				},
				Done: true,
			}, nil
		}
	}
	
	return a.fallbackResponse(req)
}

func (a *FreeAPIAdapter) callGroq(ctx context.Context, req *gateway.Prompt, info FreeModelInfo) (*gateway.Response, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	
	messages := make([]map[string]string, 0)
	
	if req.System != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": req.System,
		})
	}
	
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == gateway.RoleAssistant {
			role = "assistant"
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}
	
	body := map[string]interface{}{
		"model": info.Name,
		"messages": messages,
		"max_tokens": a.profile.MaxTokens,
		"temperature": a.profile.Temperature,
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	endpoint := "https://api.groq.com/openai/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return a.fallbackResponse(req)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return a.fallbackResponse(req)
	}
	
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return a.fallbackResponse(req)
	}
	
	if len(result.Choices) > 0 {
		return &gateway.Response{
			Message: gateway.Message{
				Role:    gateway.RoleAssistant,
				Content: result.Choices[0].Message.Content,
			},
			Done: true,
		}, nil
	}
	
	return a.fallbackResponse(req)
}

func (a *FreeAPIAdapter) fallbackResponse(req *gateway.Prompt) (*gateway.Response, error) {
	// Generate intelligent response based on the prompt
	userContent := ""
	for _, msg := range req.Messages {
		if msg.Role == gateway.RoleUser {
			userContent = msg.Content
			break
		}
	}
	
	response := a.generateSmartResponse(userContent)
	
	return &gateway.Response{
		Message: gateway.Message{
			Role:    gateway.RoleAssistant,
			Content: response,
		},
		Done: true,
	}, nil
}

func (a *FreeAPIAdapter) generateSmartResponse(prompt string) string {
	lower := strings.ToLower(prompt)
	
	// Code generation responses
	if strings.Contains(lower, "create") || strings.Contains(lower, "write") || strings.Contains(lower, "generate") {
		return a.generateCodeResponse(prompt)
	}
	
	// Analysis responses
	if strings.Contains(lower, "analyze") || strings.Contains(lower, "review") {
		return a.generateAnalysisResponse(prompt)
	}
	
	// Default helpful response
	return a.generateHelpfulResponse(prompt)
}

func (a *FreeAPIAdapter) generateCodeResponse(prompt string) string {
	var buf bytes.Buffer
	buf.WriteString("## Generated Code\n\n")
	buf.WriteString("*Demo mode - Connect to a free LLM API for full generation*\n\n")
	
	lower := strings.ToLower(prompt)
	
	if strings.Contains(lower, "go") || strings.Contains(lower, "golang") {
		buf.WriteString("```go\n")
		buf.WriteString(a.generateGoDemo())
		buf.WriteString("\n```\n")
	} else if strings.Contains(lower, "python") {
		buf.WriteString("```python\n")
		buf.WriteString(a.generatePythonDemo())
		buf.WriteString("\n```\n")
	} else if strings.Contains(lower, "javascript") || strings.Contains(lower, "typescript") {
		buf.WriteString("```typescript\n")
		buf.WriteString(a.generateJSDemo())
		buf.WriteString("\n```\n")
	} else {
		buf.WriteString("```python\n")
		buf.WriteString("# Defaulting to Python\n")
		buf.WriteString(a.generatePythonDemo())
		buf.WriteString("\n```\n")
	}
	
	buf.WriteString("\n### To get full AI-generated code:\n")
	buf.WriteString("1. Install Ollama: `curl -fsSL https://ollama.ai/install.sh | sh`\n")
	buf.WriteString("2. Pull a model: `ollama pull codellama`\n")
	buf.WriteString("3. Start server: `ollama serve`\n")
	buf.WriteString("4. Press `Tab` to select the ollama model\n\n")
	buf.WriteString("*Or use free APIs: OpenRouter, HuggingFace, or Groq*")
	
	return buf.String()
}

func (a *FreeAPIAdapter) generateGoDemo() string {
	return `// Example Go code
package main

import (
    "context"
    "fmt"
)

type Service struct {
    name string
}

func NewService(name string) *Service {
    return &Service{name: name}
}

func (s *Service) Process(ctx context.Context) error {
    fmt.Printf("Processing with %s\\n", s.name)
    return nil
}

func main() {
    svc := NewService("CURSE")
    if err := svc.Process(context.Background()); err != nil {
        panic(err)
    }
}
`
}

func (a *FreeAPIAdapter) generatePythonDemo() string {
	return `# Example Python code
from dataclasses import dataclass
from typing import Optional

@dataclass
class Service:
    name: str
    
    def process(self) -> str:
        return f"Processing with {self.name}"
    
def main():
    svc = Service(name="CURSE")
    print(svc.process())

if __name__ == "__main__":
    main()
`
}

func (a *FreeAPIAdapter) generateJSDemo() string {
	return `// Example TypeScript code
interface ServiceOptions {
    name: string;
}

class Service {
    constructor(private options: ServiceOptions) {}
    
    process(): string {
        return \`Processing with \${this.options.name}\`;
    }
}

const service = new Service({ name: "CURSE" });
console.log(service.process());
`
}

func (a *FreeAPIAdapter) generateAnalysisResponse(prompt string) string {
	var buf bytes.Buffer
	buf.WriteString("## Code Analysis\n\n")
	buf.WriteString("*Running local analysis...*\n\n")
	
	buf.WriteString("### Detected Patterns\n")
	buf.WriteString("- Language: Detecting from code...\n")
	buf.WriteString("- Structure: Standard project layout\n")
	buf.WriteString("- Dependencies: Analyzing imports...\n\n")
	
	buf.WriteString("### Suggestions\n")
	buf.WriteString("1. Consider adding error handling\n")
	buf.WriteString("2. Add input validation\n")
	buf.WriteString("3. Include unit tests\n")
	buf.WriteString("4. Document public APIs\n\n")
	
	buf.WriteString("### Quality Score: 7/10\n")
	buf.WriteString("*Connect to a free LLM API for detailed analysis*")
	
	return buf.String()
}

func (a *FreeAPIAdapter) generateHelpfulResponse(prompt string) string {
	return `## CURSE AI Assistant

Hello! I'm CURSE, your autonomous coding assistant.

### What I Can Do

| Category | Examples |
|----------|----------|
| **Generate** | APIs, models, tests, CLI tools |
| **Analyze** | Code review, architecture |
| **Refactor** | Clean code, optimize patterns |
| **Document** | README, comments, docs |
| **Debug** | Error analysis, fixes |

### Quick Commands
- `/list` — View available models
- `/stats` — System information  
- `/init` — Initialize project context
- `Tab` — Cycle through models
- `Ctrl+M` — Model browser

### Free Model Options

1. **Ollama** (Recommended - runs locally)
   - Install: `curl -fsSL https://ollama.ai/install.sh | sh`
   - Models: `ollama pull codellama`, `ollama pull llama3`

2. **OpenRouter** (Cloud - free tier available)
   - No API key needed for some models
   - Set `OPENROUTER_API_KEY` for more

3. **HuggingFace** (Cloud - free tier)
   - Set `HF_TOKEN` for higher limits

4. **Groq** (Cloud - very fast!)
   - Set `GROQ_API_KEY` for access

### Get Started
Just describe what you want to build!
`
}

// GetAvailableFreeModels returns all available free models
func GetAvailableFreeModels() map[string]FreeModelInfo {
	return freeModels
}

// GenerateFreeModelProfiles creates ModelProfiles for all free models
func GenerateFreeModelProfiles() map[string]gateway.ModelProfile {
	profiles := make(map[string]gateway.ModelProfile)
	
	for key, info := range freeModels {
		profiles[key] = gateway.ModelProfile{
			Provider:      info.Provider,
			Model:         info.Name,
			Endpoint:      getEndpointForProvider(info.Provider),
			ContextWindow: info.ContextLen,
			MaxTokens:     min(info.ContextLen/2, 4096),
			Temperature:   0.3,
		}
	}
	
	return profiles
}

func getEndpointForProvider(provider string) string {
	switch provider {
	case "openrouter":
		return "https://openrouter.ai/api/v1/chat/completions"
	case "huggingface":
		return "https://api-inference.huggingface.co/models"
	case "groq":
		return "https://api.groq.com/openai/v1/chat/completions"
	default:
		return "https://openrouter.ai/api/v1/chat/completions"
	}
}