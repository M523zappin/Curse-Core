package gateway

import (
	"github.com/M523zappin/Curse-Core/internal/computer"
)

type ToolExecutor func(args map[string]interface{}) (interface{}, error)

type ToolRegistry struct {
	tools      map[string]ToolDef
	executors  map[string]ToolExecutor
	controller *computer.ComputerController
}

func NewToolRegistry(controller *computer.ComputerController) *ToolRegistry {
	tr := &ToolRegistry{
		tools:      make(map[string]ToolDef),
		executors:  make(map[string]ToolExecutor),
		controller: controller,
	}
	tr.registerBuiltins()
	return tr
}

func (tr *ToolRegistry) registerBuiltins() {
	tr.tools["browser-navigate"] = ToolDef{
		Name:        "browser-navigate",
		Description: "Navigate the browser to a URL. Opens a new page or navigates the current tab.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to navigate to",
				},
			},
			"required": []string{"url"},
		},
	}
	tr.executors["browser-navigate"] = func(args map[string]interface{}) (interface{}, error) {
		url, _ := args["url"].(string)
		return tr.controller.Navigate(url)
	}

	tr.tools["browser-click"] = ToolDef{
		Name:        "browser-click",
		Description: "Click a UI element identified by CSS selector. Performs a pre-click safety check with screenshot.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"selector": map[string]interface{}{
					"type":        "string",
					"description": "CSS selector for the element to click",
				},
			},
			"required": []string{"selector"},
		},
	}
	tr.executors["browser-click"] = func(args map[string]interface{}) (interface{}, error) {
		selector, _ := args["selector"].(string)
		return tr.controller.Click(selector)
	}

	tr.tools["browser-type"] = ToolDef{
		Name:        "browser-type",
		Description: "Type text into an input field identified by CSS selector.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"selector": map[string]interface{}{
					"type":        "string",
					"description": "CSS selector for the input element",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to type into the field",
				},
			},
			"required": []string{"selector", "text"},
		},
	}
	tr.executors["browser-type"] = func(args map[string]interface{}) (interface{}, error) {
		selector, _ := args["selector"].(string)
		text, _ := args["text"].(string)
		return tr.controller.TypeText(selector, text)
	}

	tr.tools["browser-scroll"] = ToolDef{
		Name:        "browser-scroll",
		Description: "Scroll the page or a specific element by a delta amount.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"selector": map[string]interface{}{
					"type":        "string",
					"description": "CSS selector of element to scroll (or 'window')",
				},
				"delta_y": map[string]interface{}{
					"type":        "integer",
					"description": "Vertical scroll delta in pixels (negative=up, positive=down)",
				},
			},
			"required": []string{"delta_y"},
		},
	}
	tr.executors["browser-scroll"] = func(args map[string]interface{}) (interface{}, error) {
		selector, _ := args["selector"].(string)
		if selector == "" {
			selector = "body"
		}
		deltaY, _ := args["delta_y"].(float64)
		return tr.controller.Scroll(selector, int(deltaY))
	}

	tr.tools["browser-screenshot"] = ToolDef{
		Name:        "browser-screenshot",
		Description: "Capture a full-page screenshot of the current browser state. Returns base64-encoded PNG.",
		Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
	}
	tr.executors["browser-screenshot"] = func(args map[string]interface{}) (interface{}, error) {
		return tr.controller.Screenshot()
	}

	tr.tools["desktop-launch"] = ToolDef{
		Name:        "desktop-launch",
		Description: "Launch a desktop application by name (e.g. 'notepad.exe', 'explorer', 'code').",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"app": map[string]interface{}{
					"type":        "string",
					"description": "Application name or path",
				},
				"args": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Command-line arguments",
				},
			},
			"required": []string{"app"},
		},
	}
	tr.executors["desktop-launch"] = func(args map[string]interface{}) (interface{}, error) {
		app, _ := args["app"].(string)
		argsRaw, _ := args["args"].([]interface{})
		strArgs := make([]string, len(argsRaw))
		for i, a := range argsRaw {
			strArgs[i], _ = a.(string)
		}
		return tr.controller.LaunchApp(app, strArgs)
	}

	tr.tools["desktop-terminal"] = ToolDef{
		Name:        "desktop-terminal",
		Description: "Execute a terminal/command-line command. Destructive commands require HITL confirmation.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
			},
			"required": []string{"command"},
		},
	}
	tr.executors["desktop-terminal"] = func(args map[string]interface{}) (interface{}, error) {
		cmd, _ := args["command"].(string)
		return tr.controller.Terminal(cmd)
	}

	tr.tools["desktop-file"] = ToolDef{
		Name:        "desktop-file",
		Description: "Perform a file operation (read, write, delete, list, mkdir). Destructive ops require HITL confirmation.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"read", "write", "delete", "list", "mkdir"},
					"description": "File operation to perform",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory path",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write (for write operation)",
				},
			},
			"required": []string{"operation", "path"},
		},
	}
	tr.executors["desktop-file"] = func(args map[string]interface{}) (interface{}, error) {
		op, _ := args["operation"].(string)
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		return tr.controller.FileOperation(op, path, content)
	}

	tr.tools["desktop-open-browser"] = ToolDef{
		Name:        "desktop-open-browser",
		Description: "Open a URL in the default system web browser.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to open",
				},
			},
			"required": []string{"url"},
		},
	}
	tr.executors["desktop-open-browser"] = func(args map[string]interface{}) (interface{}, error) {
		url, _ := args["url"].(string)
		return nil, tr.controller.DesktopManager().OpenBrowser(url)
	}
}

func (tr *ToolRegistry) GetTool(name string) (ToolDef, bool) {
	t, ok := tr.tools[name]
	return t, ok
}

func (tr *ToolRegistry) Execute(name string, args map[string]interface{}) (interface{}, error) {
	exec, ok := tr.executors[name]
	if !ok {
		return nil, nil
	}
	return exec(args)
}

func (tr *ToolRegistry) ListTools() []ToolDef {
	tools := make([]ToolDef, 0, len(tr.tools))
	for _, t := range tr.tools {
		tools = append(tools, t)
	}
	return tools
}

func (tr *ToolRegistry) ToolDefinitions() []Tool {
	tools := make([]Tool, 0, len(tr.tools))
	for _, def := range tr.tools {
		tools = append(tools, Tool{
			Type:     "function",
			Function: def,
		})
	}
	return tools
}
