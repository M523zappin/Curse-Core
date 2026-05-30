package gateway

import "context"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function ToolFunction  `json:"function"`
}

type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string      `json:"type"`
	Function ToolDef     `json:"function"`
}

type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Prompt struct {
	Messages  []Message
	System    string
	Tools     []Tool
	MaxTokens int
	Model     string
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Response struct {
	Message   Message   `json:"message"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     Usage     `json:"usage"`
	Done      bool      `json:"done"`
}

type Adapter interface {
	Name() string
	Send(ctx context.Context, req *Prompt) (*Response, error)
	ModelInfo() ModelProfile
}
