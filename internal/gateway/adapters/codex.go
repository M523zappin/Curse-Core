package adapters

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type CodexAdapter struct {
	profile   gateway.ModelProfile
	repoPath  string
	fileCache map[string]*ast.File
}

func NewCodex(profile gateway.ModelProfile, repoPath string) *CodexAdapter {
	return &CodexAdapter{
		profile:   profile,
		repoPath:  repoPath,
		fileCache: make(map[string]*ast.File),
	}
}

func (a *CodexAdapter) Name() string { return "codex" }
func (a *CodexAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *CodexAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	userMsg := ""
	for _, m := range req.Messages {
		if m.Role == gateway.RoleUser {
			userMsg = m.Content
			break
		}
	}

	resp := a.analyzeRequest(userMsg)
	return &gateway.Response{
		Message: gateway.Message{
			Role:    gateway.RoleAssistant,
			Content: resp,
		},
		Done: true,
	}, nil
}

func (a *CodexAdapter) analyzeRequest(query string) string {
	q := strings.ToLower(query)

	switch {
	case strings.Contains(q, "function") || strings.Contains(q, "func"):
		return a.listFunctions()
	case strings.Contains(q, "struct") || strings.Contains(q, "type"):
		return a.listTypes()
	case strings.Contains(q, "file") || strings.Contains(q, "module"):
		return a.listFiles()
	case strings.Contains(q, "generate") || strings.Contains(q, "create"):
		return a.generateTemplate(query)
	case strings.Contains(q, "summary") || strings.Contains(q, "overview"):
		return a.codebaseOverview()
	default:
		return a.codebaseOverview()
	}
}

func (a *CodexAdapter) listFunctions() string {
	var b strings.Builder
	b.WriteString("📋 Functions found in codebase:\n\n")
	filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
			return nil
		}
		f, err := a.parseFile(path)
		if err != nil {
			return nil
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			rel, _ := filepath.Rel(a.repoPath, path)
			b.WriteString(fmt.Sprintf("  • %s  (%s)\n", fn.Name.Name, rel))
		}
		return nil
	})
	if b.Len() == 0 {
		return "No Go functions found in codebase."
	}
	return b.String()
}

func (a *CodexAdapter) listTypes() string {
	var b strings.Builder
	b.WriteString("📋 Types/Structs found in codebase:\n\n")
	filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
			return nil
		}
		f, err := a.parseFile(path)
		if err != nil {
			return nil
		}
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				rel, _ := filepath.Rel(a.repoPath, path)
				b.WriteString(fmt.Sprintf("  • %s  (%s)\n", ts.Name.Name, rel))
			}
		}
		return nil
	})
	if b.Len() == 0 {
		return "No Go types found in codebase."
	}
	return b.String()
}

func (a *CodexAdapter) listFiles() string {
	var b strings.Builder
	b.WriteString("📁 Go source files:\n\n")
	filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
			return nil
		}
		rel, _ := filepath.Rel(a.repoPath, path)
		f, parseErr := a.parseFile(path)
		pkg := "?"
		if parseErr == nil && f.Name != nil {
			pkg = f.Name.Name
		}
		b.WriteString(fmt.Sprintf("  • %-50s  package %s\n", rel, pkg))
		return nil
	})
	if b.Len() == 0 {
		return "No Go source files found."
	}
	return b.String()
}

func (a *CodexAdapter) codebaseOverview() string {
	var funcs, types, files int
	filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
			return nil
		}
		files++
		f, parseErr := a.parseFile(path)
		if parseErr != nil {
			return nil
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				funcs++
			case *ast.GenDecl:
				if d.Tok == token.TYPE {
					types += len(d.Specs)
				}
			}
		}
		return nil
	})

	return fmt.Sprintf(`📊 Codebase Overview (AST Analysis)

  Repository: %s
  Go files:   %d
  Functions:  %d
  Types:      %d

  CURSE is running in local-only mode with its built-in code analysis engine.
  No external AI model or API key is required.

  Ask me about:
  • "list functions" — show all Go functions
  • "list types"    — show all types/structs
  • "list files"    — show all Go source files
  • "generate <X>"  — generate a code template
  • "summary"       — this overview
`, a.repoPath, files, funcs, types)
}

func (a *CodexAdapter) generateTemplate(query string) string {
	q := strings.ToLower(query)
	switch {
	case strings.Contains(q, "struct") || strings.Contains(q, "type"):
		return `Generated type template:

type Example struct {
    ID   int    ` + "`json:\"id\"`" + `
    Name string ` + "`json:\"name\"`" + `
}

func NewExample(id int, name string) *Example {
    return &Example{ID: id, Name: name}
}`
	case strings.Contains(q, "handler") || strings.Contains(q, "http"):
		return `Generated HTTP handler template:

func HandleExample(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}`
	case strings.Contains(q, "test"):
		return `Generated test template:

func TestExample(t *testing.T) {
    tests := []struct {
        name string
        want string
    }{
        {name: "case 1", want: "expected"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Example()
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}`
	default:
		return `Generated Go file template:

package main

import "fmt"

func main() {
    fmt.Println("Hello, CURSE!")
}`
	}
}

func (a *CodexAdapter) parseFile(path string) (*ast.File, error) {
	if f, ok := a.fileCache[path]; ok {
		return f, nil
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	a.fileCache[path] = f
	return f, nil
}
