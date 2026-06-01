package adapters

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

// SmartCodeAdapter is a 100% offline code generation adapter.
// It has ZERO external dependencies and works completely autonomously.
// No API keys, no cloud services, no internet required.
type SmartCodeAdapter struct {
	profile gateway.ModelProfile
}

// CodeTemplate represents a code generation template
type CodeTemplate struct {
	Name     string
	Keywords []string
	Language string
	Generate func(task string) string
}

// All available templates - comprehensive code generation for any task
var allTemplates = []CodeTemplate{
	// ═══════════════════════════════════════════════════════
	// GO TEMPLATES
	// ═══════════════════════════════════════════════════════
	{Name: "Go REST Handler", Keywords: []string{"api", "handler", "endpoint", "rest", "http", "route", "controller"}, Language: "go", Generate: goRESTHandler},
	{Name: "Go Model/Struct", Keywords: []string{"model", "struct", "entity", "type"}, Language: "go", Generate: goModel},
	{Name: "Go Middleware", Keywords: []string{"middleware", "auth", "jwt", "cors", "logging"}, Language: "go", Generate: goMiddleware},
	{Name: "Go Unit Test", Keywords: []string{"test", "spec", "testing", "unit"}, Language: "go", Generate: goTest},
	{Name: "Go CLI Command", Keywords: []string{"cli", "command", "flag", "cobra"}, Language: "go", Generate: goCLI},
	{Name: "Go Repository", Keywords: []string{"repository", "database", "db", "sql", "gorm", "sqlx"}, Language: "go", Generate: goRepository},
	{Name: "Go Error Handler", Keywords: []string{"error", "handle error", "exception"}, Language: "go", Generate: goErrorHandler},
	{Name: "Go Config", Keywords: []string{"config", "yaml", "toml", "viper", "settings"}, Language: "go", Generate: goConfig},
	{Name: "Go Worker Pool", Keywords: []string{"worker", "queue", "background", "job", "async"}, Language: "go", Generate: goWorker},
	{Name: "Go Cache", Keywords: []string{"cache", "redis", "memory"}, Language: "go", Generate: goCache},
	{Name: "Go WebSocket", Keywords: []string{"websocket", "ws", "socket", "realtime"}, Language: "go", Generate: goWebSocket},
	{Name: "Go gRPC Service", Keywords: []string{"grpc", "proto", "rpc"}, Language: "go", Generate: goGRPC},

	// ═══════════════════════════════════════════════════════
	// PYTHON TEMPLATES
	// ═══════════════════════════════════════════════════════
	{Name: "Python FastAPI", Keywords: []string{"api", "rest", "fastapi", "endpoint"}, Language: "python", Generate: pythonFastAPI},
	{Name: "Python Class/Model", Keywords: []string{"class", "model", "pydantic", "schema"}, Language: "python", Generate: pythonModel},
	{Name: "Python Test", Keywords: []string{"test", "pytest", "spec", "unittest"}, Language: "python", Generate: pythonTest},
	{Name: "Python CLI", Keywords: []string{"cli", "click", "argparse", "command"}, Language: "python", Generate: pythonCLI},
	{Name: "Python Database", Keywords: []string{"database", "sqlalchemy", "repository", "db"}, Language: "python", Generate: pythonDB},
	{Name: "Python Async Worker", Keywords: []string{"async", "asyncio", "worker", "background"}, Language: "python", Generate: pythonAsync},
	{Name: "Python Decorator", Keywords: []string{"decorator", "wrapper", "functools"}, Language: "python", Generate: pythonDecorator},

	// ═══════════════════════════════════════════════════════
	// TYPESCRIPT/JAVASCRIPT TEMPLATES
	// ═══════════════════════════════════════════════════════
	{Name: "TS/JS REST API", Keywords: []string{"api", "express", "endpoint", "route", "rest"}, Language: "typescript", Generate: tsREST},
	{Name: "TS/JS React Component", Keywords: []string{"react", "component", "jsx", "tsx", "ui"}, Language: "typescript", Generate: tsComponent},
	{Name: "TS/JS Type/Interface", Keywords: []string{"interface", "type", "typedef"}, Language: "typescript", Generate: tsType},
	{Name: "TS/JS Test", Keywords: []string{"test", "jest", "vitest", "spec"}, Language: "typescript", Generate: tsTest},
	{Name: "TS/JS API Client", Keywords: []string{"client", "fetch", "axios", "api"}, Language: "typescript", Generate: tsClient},
	{Name: "TS/JS React Hook", Keywords: []string{"hook", "useState", "useEffect"}, Language: "typescript", Generate: tsHook},

	// ═══════════════════════════════════════════════════════
	// DEVOPS TEMPLATES
	// ═══════════════════════════════════════════════════════
	{Name: "Dockerfile", Keywords: []string{"docker", "container", "dockerfile"}, Language: "dockerfile", Generate: dockerfile},
	{Name: "GitHub Action", Keywords: []string{"ci", "cd", "github action", "workflow"}, Language: "yaml", Generate: githubAction},
	{Name: "Kubernetes", Keywords: []string{"kubernetes", "k8s", "deployment", "pod"}, Language: "yaml", Generate: kubernetes},
	{Name: "README", Keywords: []string{"readme", "docs", "documentation"}, Language: "markdown", Generate: readme},
}

func NewSmartCode(profile gateway.ModelProfile) *SmartCodeAdapter {
	return &SmartCodeAdapter{profile: profile}
}

func (a *SmartCodeAdapter) Name() string { return "smartcode" }

func (a *SmartCodeAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *SmartCodeAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	// Extract user message
	userMessage := extractUserMessage(req)
	task := strings.ToLower(userMessage)

	// Detect language
	lang := detectLanguage(task)

	// Find best matching template
	template := findBestTemplate(task, lang)

	var response string
	if template != nil {
		response = template.Generate(userMessage)
	} else {
		response = generateFallback(userMessage, lang)
	}

	return &gateway.Response{
		Message: gateway.Message{
			Role:    gateway.RoleAssistant,
			Content: response,
		},
		Done: true,
	}, nil
}

func extractUserMessage(req *gateway.Prompt) string {
	for _, msg := range req.Messages {
		if msg.Role == gateway.RoleUser {
			return msg.Content
		}
	}
	return ""
}

func detectLanguage(task string) string {
	if strings.Contains(task, " go ") || strings.Contains(task, "golang") ||
		strings.Contains(task, "func ") || strings.Contains(task, "package ") {
		return "go"
	}
	if strings.Contains(task, " python ") || strings.Contains(task, "def ") {
		return "python"
	}
	if strings.Contains(task, " typescript ") || strings.Contains(task, " ts ") ||
		strings.Contains(task, "interface") || strings.Contains(task, "tsx") {
		return "typescript"
	}
	if strings.Contains(task, " javascript ") || strings.Contains(task, " js ") ||
		strings.Contains(task, " node") {
		return "javascript"
	}
	if strings.Contains(task, "rust") || strings.Contains(task, "fn ") {
		return "rust"
	}
	if strings.Contains(task, "java ") || strings.Contains(task, "public class") {
		return "java"
	}
	if strings.Contains(task, "docker") || strings.Contains(task, "container") {
		return "dockerfile"
	}
	if strings.Contains(task, "kubernetes") || strings.Contains(task, "k8s") {
		return "yaml"
	}
	if strings.Contains(task, "ci/cd") || strings.Contains(task, "github action") {
		return "yaml"
	}
	if strings.Contains(task, "readme") || strings.Contains(task, "docs") {
		return "markdown"
	}

	// Default to Go (most common for backend)
	if strings.Contains(task, "script") || strings.Contains(task, "automation") {
		return "python"
	}
	return "go"
}

func findBestTemplate(task string, lang string) *CodeTemplate {
	var bestScore = 0
	var bestTemplate *CodeTemplate

	for i := range allTemplates {
		t := &allTemplates[i]

		// Language must match or be generic
		if t.Language != lang && t.Language != "" {
			// Give partial credit for related languages
			if (lang == "typescript" || lang == "javascript") && t.Language == "typescript" {
				// OK
			} else {
				continue
			}
		}

		// Count keyword matches
		score := 0
		for _, kw := range t.Keywords {
			if strings.Contains(task, kw) {
				score += 2
			}
		}

		if score > bestScore {
			bestScore = score
			bestTemplate = t
		}
	}

	return bestTemplate
}

func extractName(task string) string {
	patterns := []string{
		`(?:create|add|make|generate)\s+(?:a\s+)?(\w+)`,
		`(\w+)\s+(?:handler|service|model|controller)`,
		`(?:for|of)\s+(\w+)`,
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p)
		matches := re.FindStringSubmatch(task)
		if len(matches) > 1 {
			name := strings.Title(matches[1])
			name = strings.TrimSuffix(name, "s")
			if len(name) > 2 {
				return name
			}
		}
	}
	return "Item"
}

func generateFallback(task, lang string) string {
	name := extractName(task)
	return fmt.Sprintf(`## Generated %s Code

Here's a complete implementation for your request:

```%s
// %s represents a %s
type %s struct {
    ID        string    `+"`"+`json:"id"`+"`"+`
    Name      string    `+"`"+`json:"name"`+"`"+`
    CreatedAt time.Time `+"`"+`json:"created_at"`+"`"+`
    UpdatedAt time.Time `+"`"+`json:"updated_at"`+"`"+`
}

// New%s creates a new %s
func New%s(name string) *%s {
    return &%s{
        ID:   uuid.New().String(),
        Name: name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}
```

This is a basic template. For more specific code generation, try:
- "create a REST API handler"
- "add unit tests"
- "implement authentication middleware"
- "create a database model"

I'm a 100%% offline code generator - no API keys or internet needed! 🚀
`, strings.Title(lang), lang, name, name, name, name, name, name, name)
}

// ═══════════════════════════════════════════════════════════════════════════
// GO TEMPLATES
// ═══════════════════════════════════════════════════════════════════════════

func goRESTHandler(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// %[1]sHandler handles HTTP requests for %[1]s
type %[1]sHandler struct {
    repo *%[1]sRepository
}

// New%[1]sHandler creates a new handler
func New%[1]sHandler(repo *%[1]sRepository) *%[1]sHandler {
    return &%[1]sHandler{repo: repo}
}

// Create handles POST /%[2]ss
func (h *%[1]sHandler) Create(c *gin.Context) {
    var req Create%[1]sRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    item, err := h.repo.Create(c.Request.Context(), req.Name)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, item)
}

// List handles GET /%[2]ss
func (h *%[1]sHandler) List(c *gin.Context) {
    items, err := h.repo.List(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, items)
}

// Get handles GET /%[2]ss/:id
func (h *%[1]sHandler) Get(c *gin.Context) {
    item, err := h.repo.Get(c.Request.Context(), c.Param("id"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, item)
}

// Update handles PUT /%[2]ss/:id
func (h *%[1]sHandler) Update(c *gin.Context) {
    var req Update%[1]sRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    item, err := h.repo.Update(c.Request.Context(), c.Param("id"), req.Name)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, item)
}

// Delete handles DELETE /%[2]ss/:id
func (h *%[1]sHandler) Delete(c *gin.Context) {
    if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.Status(http.StatusNoContent)
}

// Request types
type Create%[1]sRequest struct {
    Name string `json:"name" binding:"required"`
}

type Update%[1]sRequest struct {
    Name string `json:"name"`
}
`, name, lower)
}

func goModel(task string) string {
	name := extractName(task)

	return fmt.Sprintf(`package models

import (
    "time"
)

// %[1]s represents a %[1]s entity
type %[1]s struct {
    ID        string    %[2]sjson:"id"%[3]s
    Name      string    %[2]sjson:"name"%[3]s
    CreatedAt time.Time %[2]sjson:"created_at"%[3]s
    UpdatedAt time.Time %[2]sjson:"updated_at"%[3]s
}

// Validate checks if the %[1]s is valid
func (m *%[1]s) Validate() error {
    if m.Name == "" {
        return fmt.Errorf("name is required")
    }
    return nil
}

// BeforeCreate sets timestamps
func (m *%[1]s) BeforeCreate() {
    m.ID = uuid.New().String()
    m.CreatedAt = time.Now()
    m.UpdatedAt = time.Now()
}

// BeforeUpdate updates timestamp
func (m *%[1]s) BeforeUpdate() {
    m.UpdatedAt = time.Now()
}
`, name, "`", "`")
}

func goMiddleware(task string) string {
	lower := strings.ToLower(task)
	hasAuth := strings.Contains(lower, "auth") || strings.Contains(lower, "jwt")

	code := `package middleware

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
)

// Logger returns a logging middleware
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        log.Printf("[%d] %s %s %v",
            c.Writer.Status(), c.Request.Method, c.Request.URL.Path, time.Since(start))
    }
}

// Recovery returns a panic recovery middleware
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                c.AbortWithStatusJSON(500, gin.H{"error": "internal error"})
            }
        }()
        c.Next()
    }
}

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
`

	if hasAuth {
		code += `
// Auth returns a JWT authentication middleware
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        // Parse and validate JWT
        // TODO: Implement JWT validation with your secret
        c.Next()
    }
}

// Claims represents JWT claims
type Claims struct {
    UserID string
    Email  string
}
`
	}

	return code
}

func goTest(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`package handlers_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func setup() *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    // TODO: Register routes
    return router
}

func Test%[1]sHandler_Create(t *testing.T) {
    router := setup()

    t.Run("success", func(t *testing.T) {
        body := ` + "`" + `{"name": "test %[2]s"}` + "`" + `
        req, _ := http.NewRequest("POST", "/%[2]ss", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)
        var resp map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &resp)
        assert.NotEmpty(t, resp["id"])
    })
}

func Test%[1]sHandler_Get_NotFound(t *testing.T) {
    router := setup()
    req, _ := http.NewRequest("GET", "/%[2]ss/non-existent", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNotFound, w.Code)
}
`, name, lower)
}

func goCLI(task string) string {
	name := strings.ToLower(extractName(task))

	return fmt.Sprintf(`package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "%s",
    Short: "A brief description",
    Long:  `A longer description.`,
    Run:   run,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func run(cmd *cobra.Command, args []string) {
    if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
        fmt.Println("Running in verbose mode")
    }
    fmt.Println("Hello from %s!")
}
`, name, name)
}

func goRepository(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`package repository

import (
    "context"
    "database/sql"
    "fmt"
)

// %[1]sRepository handles data access for %[1]ss
type %[1]sRepository struct {
    db *sql.DB
}

func New%[1]sRepository(db *sql.DB) *%[1]sRepository {
    return &%[1]sRepository{db: db}
}

func (r *%[1]sRepository) Create(ctx context.Context, name string) (*%[1]s, error) {
    query := `INSERT INTO %[2]ss (id, name, created_at, updated_at)
              VALUES ($1, $2, NOW(), NOW())
              RETURNING id, name, created_at, updated_at`

    item := &%[1]s{}
    err := r.db.QueryRowContext(ctx, query, uuid.New().String(), name).
        Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("create: %%w", err)
    }
    return item, nil
}

func (r *%[1]sRepository) Get(ctx context.Context, id string) (*%[1]s, error) {
    query := `SELECT id, name, created_at, updated_at FROM %[2]ss WHERE id = $1`
    item := &%[1]s{}
    err := r.db.QueryRowContext(ctx, query, id).
        Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get: %%w", err)
    }
    return item, nil
}

func (r *%[1]sRepository) List(ctx context.Context) ([]*%[1]s, error) {
    query := `SELECT id, name, created_at, updated_at FROM %[2]ss ORDER BY created_at DESC`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("list: %%w", err)
    }
    defer rows.Close()

    var items []*%[1]s
    for rows.Next() {
        item := &%[1]s{}
        if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt); err != nil {
            return nil, fmt.Errorf("scan: %%w", err)
        }
        items = append(items, item)
    }
    return items, rows.Err()
}

func (r *%[1]sRepository) Update(ctx context.Context, id, name string) (*%[1]s, error) {
    query := `UPDATE %[2]ss SET name = $1, updated_at = NOW() WHERE id = $2
              RETURNING id, name, created_at, updated_at`
    item := &%[1]s{}
    err := r.db.QueryRowContext(ctx, query, name, id).
        Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("update: %%w", err)
    }
    return item, nil
}

func (r *%[1]sRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM %[2]ss WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete: %%w", err)
    }
    return nil
}

var ErrNotFound = fmt.Errorf("%[1]s not found")
`, name, lower)
}

func goErrorHandler(task string) string {
	return `package errors

import "fmt"

// Sentinel errors
var (
    ErrNotFound     = fmt.Errorf("resource not found")
    ErrUnauthorized = fmt.Errorf("unauthorized")
    ErrForbidden    = fmt.Errorf("forbidden")
    ErrValidation   = fmt.Errorf("validation failed")
)

// AppError represents an application error
type AppError struct {
    Code    string
    Message string
    Cause   error
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(code, msg string) *AppError { return &AppError{Code: code, Message: msg} }
func Wrap(err error, code, msg string) *AppError {
    if err == nil { return nil }
    return &AppError{Code: code, Message: msg, Cause: err}
}
`
}

func goConfig(task string) string {
	return `package config

import (
    "os"
    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

type ServerConfig struct {
    Host string
    Port int
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
}

func Load(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.AutomaticEnv()

    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("server.port", 8080)

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    // Environment overrides
    if host := os.Getenv("DB_HOST"); host != "" {
        cfg.Database.Host = host
    }

    return &cfg, nil
}
`
}

func goWorker(task string) string {
	return `package worker

import (
    "context"
    "log"
    "sync"
)

type Job func(ctx context.Context) error

type Pool struct {
    workers int
    jobs    chan Job
    wg      sync.WaitGroup
}

func NewPool(workers int, queueSize int) *Pool {
    p := &Pool{workers: workers, jobs: make(chan Job, queueSize)}
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }
    return p
}

func (p *Pool) worker(id int) {
    defer p.wg.Done()
    for job := range p.jobs {
        if err := job(context.Background()); err != nil {
            log.Printf("worker %d: %v", id, err)
        }
    }
}

func (p *Pool) Submit(job Job) { p.jobs <- job }

func (p *Pool) Shutdown() { close(p.jobs); p.wg.Wait() }
`
}

func goCache(task string) string {
	return `package cache

import (
    "sync"
    "time"
)

type Item struct {
    Value      interface{}
    Expiration time.Time
}

type Cache struct {
    items map[string]*Item
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
    c := &Cache{items: make(map[string]*Item), ttl: ttl}
    go c.cleanup()
    return c
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    if !ok || time.Now().After(item.Expiration) {
        return nil, false
    }
    return item.Value, true
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = &Item{Value: value, Expiration: time.Now().Add(c.ttl)}
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

func (c *Cache) cleanup() {
    ticker := time.NewTicker(c.ttl)
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for k, v := range c.items {
            if now.After(v.Expiration) {
                delete(c.items, k)
            }
        }
        c.mu.Unlock()
    }
}
`
}

func goWebSocket(task string) string {
	return `package websocket

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.clients[c] = true
        case c := <-h.unregister:
            if _, ok := h.clients[c]; ok {
                delete(h.clients, c)
                close(c.send)
            }
        case m := <-h.broadcast:
            for c := range h.clients {
                select {
                case c.send <- m:
                default:
                    delete(h.clients, c)
                    close(c.send)
                }
            }
        }
    }
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("upgrade: %v", err)
        return
    }
    client := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
    h.register <- client
    go client.writePump()
    go client.readPump()
}

func (c *Client) readPump() {
    defer func() { c.hub.unregister <- c; c.conn.Close() }()
    for {
        _, msg, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        c.hub.broadcast <- msg
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()
    for msg := range c.send {
        if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
            return
        }
    }
}
`
}

func goGRPC(task string) string {
	name := extractName(task)

	return fmt.Sprintf(`package service

import "context"

type %[1]sService struct {
    Unimplemented%[1]sServiceServer
    repo %[1]sRepository
}

func New%[1]sService(repo %[1]sRepository) *%[1]sService {
    return &%[1]sService{repo: repo}
}

func (s *%[1]sService) Create(ctx context.Context, req *Create%[1]sRequest) (*%[1]sResponse, error) {
    item, err := s.repo.Create(ctx, req.Name)
    if err != nil {
        return nil, err
    }
    return &%[1]sResponse{
        Item: &%[1]s{
            Id:    item.ID,
            Name:  item.Name,
        },
    }, nil
}

func (s *%[1]sService) Get(ctx context.Context, req *Get%[1]sRequest) (*%[1]sResponse, error) {
    item, err := s.repo.Get(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    return &%[1]sResponse{
        Item: &%[1]s{
            Id:   item.ID,
            Name: item.Name,
        },
    }, nil
}
`, name)
}

// ═══════════════════════════════════════════════════════════════════════════
// PYTHON TEMPLATES
// ═══════════════════════════════════════════════════════════════════════════

func pythonFastAPI(task string) string {
	name := strings.ToLower(extractName(task))

	return `from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import Optional, List
import uuid

app = FastAPI(title="API", version="1.0.0")

# Models
class Item(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    name: str
    created_at: str = ""
    updated_at: str = ""

# Storage
items: List[Item] = []

@app.post("/items", response_model=Item, status_code=201)
def create_item(item: Item):
    items.append(item)
    return item

@app.get("/items", response_model=List[Item])
def list_items(skip: int = 0, limit: int = 100):
    return items[skip:skip + limit]

@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: str):
    for item in items:
        if item.id == item_id:
            return item
    raise HTTPException(status_code=404, detail="Not found")

@app.put("/items/{item_id}", response_model=Item)
def update_item(item_id: str, item: Item):
    for i, existing in enumerate(items):
        if existing.id == item_id:
            items[i] = item
            return item
    raise HTTPException(status_code=404, detail="Not found")

@app.delete("/items/{item_id}", status_code=204)
def delete_item(item_id: str):
    for i, item in enumerate(items):
        if item.id == item_id:
            items.pop(i)
            return
    raise HTTPException(status_code=404, detail="Not found")
`
}

func pythonModel(task string) string {
	name := extractName(task)

	return fmt.Sprintf(`from pydantic import BaseModel, Field, field_validator
from typing import Optional
from datetime import datetime
from enum import Enum

class Status(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"

class %sBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None

class %sCreate(%sBase):
    pass

class %sUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1)
    status: Optional[Status] = None

class %s(%sBase):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    status: Status = Field(default=Status.ACTIVE)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    @field_validator('name')
    @classmethod
    def name_not_empty(cls, v: str) -> str:
        if not v.strip():
            raise ValueError('name cannot be empty')
        return v.strip()

    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "status": self.status.value,
            "created_at": self.created_at.isoformat(),
        }
`, name, name, name, name, name, name)
}

func pythonTest(task string) string {
	name := strings.ToLower(extractName(task))

	return fmt.Sprintf(`import pytest
from %s import %s, %sCreate

class Test%sModel:
    def test_create_%s(self):
        data = %sCreate(name="Test")
        item = %s(id="test-id", **data.model_dump())
        assert item.name == "Test"
        assert item.id == "test-id"

    def test_validation_error(self):
        with pytest.raises(ValueError):
            %sCreate(name="")

    def test_to_dict(self):
        item = %s(name="Test")
        data = item.to_dict()
        assert "id" in data
        assert data["name"] == "Test"

class Test%sAPI:
    @pytest.fixture
    def client(self):
        from fastapi.testclient import TestClient
        from main import app
        return TestClient(app)

    def test_create_%s(self, client):
        response = client.post("/%ss", json={"name": "New"})
        assert response.status_code == 201
        assert "id" in response.json()
`, name, name, name,
		name,
		name, name, name,
		name,
		name,
		name,
		name, name)
}

func pythonCLI(task string) string {
	name := strings.ToLower(extractName(task))

	return fmt.Sprintf(`#!/usr/bin/env python3
"""CLI for %s management."""

import argparse
import sys

def main():
    parser = argparse.ArgumentParser(description="%s CLI")
    subparsers = parser.add_subparsers(dest="command")

    create = subparsers.add_parser("create", help="Create")
    create.add_argument("name", help="Name")
    create.add_argument("-v", "--verbose", action="store_true")

    list_cmd = subparsers.add_parser("list", help="List")
    list_cmd.add_argument("-l", "--limit", type=int, default=10)

    get = subparsers.add_parser("get", help="Get by ID")
    get.add_argument("id", help="ID")

    args = parser.parse_args()

    if args.command == "create":
        print(f"Creating: {args.name}")
        # TODO: Implement
    elif args.command == "list":
        print(f"Listing (limit: {args.limit})")
        # TODO: Implement
    elif args.command == "get":
        print(f"Getting: {args.id}")
        # TODO: Implement
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
`, name, name)
}

func pythonDB(task string) string {
	name := strings.ToLower(extractName(task))

	return `from sqlalchemy import Column, String, DateTime, create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
import uuid
from datetime import datetime

Base = declarative_base()

class Item(Base):
    __tablename__ = "` + "`" + `"items` + "`" + `"
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String(255), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow)

    def to_dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }

class ItemRepository:
    def __init__(self, session):
        self.session = session

    def create(self, name: str):
        item = Item(name=name)
        self.session.add(item)
        self.session.commit()
        self.session.refresh(item)
        return item

    def get(self, id: str):
        return self.session.get(Item, id)

    def list(self, skip=0, limit=100):
        return self.session.query(Item).offset(skip).limit(limit).all()

    def delete(self, id: str):
        item = self.get(id)
        if item:
            self.session.delete(item)
            self.session.commit()
            return True
        return False
`
}

func pythonAsync(task string) string {
	return `import asyncio
from typing import Callable, Any
from dataclasses import dataclass, field
from datetime import datetime

@dataclass
class Task:
    id: str
    func: Callable
    args: tuple = ()
    kwargs: dict = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.utcnow)
    result: Any = None
    error: Exception = None

class AsyncWorker:
    def __init__(self, max_workers: int = 10):
        self.max_workers = max_workers
        self.tasks = {}
        self.queue = asyncio.Queue()
        self.running = False

    async def start(self):
        self.running = True
        workers = [asyncio.create_task(self._worker(i)) for i in range(self.max_workers)]
        await asyncio.gather(*workers)

    async def _worker(self, worker_id: int):
        while self.running:
            try:
                task = await asyncio.wait_for(self.queue.get(), timeout=1.0)
                try:
                    task.result = await task.func(*task.args, **task.kwargs)
                except Exception as e:
                    task.error = e
                self.tasks[task.id] = task
            except asyncio.TimeoutError:
                continue

    async def submit(self, task: Task) -> str:
        await self.queue.put(task)
        self.tasks[task.id] = task
        return task.id

    def stop(self):
        self.running = False
`
}

func pythonDecorator(task string) string {
	return `import functools
import time

def retry(max_attempts: int = 3, delay: float = 1.0):
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            last_exception = None
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    if attempt < max_attempts - 1:
                        time.sleep(delay * (2 ** attempt))
            raise last_exception
        return wrapper
    return decorator

def timing(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        start = time.perf_counter()
        result = func(*args, **kwargs)
        print(f"{func.__name__} took {time.perf_counter() - start:.4f}s")
        return result
    return wrapper

def cache(ttl: float = 300):
    def decorator(func):
        cache_store = {}
        cache_times = {}
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            key = (args, tuple(sorted(kwargs.items())))
            now = time.time()
            if key in cache_store and now - cache_times[key] < ttl:
                return cache_store[key]
            result = func(*args, **kwargs)
            cache_store[key] = result
            cache_times[key] = now
            return result
        return wrapper
    return decorator
`
}

// ═══════════════════════════════════════════════════════════════════════════
// TYPESCRIPT TEMPLATES
// ═══════════════════════════════════════════════════════════════════════════

func tsREST(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`import express, { Request, Response } from 'express';

interface %[1]s {
  id: string;
  name: string;
  createdAt: Date;
}

const router = express.Router();
const storage: %[1]s[] = [];

router.post('/%[2]ss', (req: Request, res: Response) => {
  const item: %[1]s = {
    id: crypto.randomUUID(),
    name: req.body.name,
    createdAt: new Date(),
  };
  storage.push(item);
  res.status(201).json(item);
});

router.get('/%[2]ss', (req: Request, res: Response) => {
  const { page = 1, limit = 20 } = req.query;
  const start = (Number(page) - 1) * Number(limit);
  res.json({ data: storage.slice(start, start + Number(limit)), total: storage.length });
});

router.get('/%[2]ss/:id', (req: Request, res: Response) => {
  const item = storage.find(i => i.id === req.params.id);
  if (!item) return res.status(404).json({ error: 'Not found' });
  res.json(item);
});

router.put('/%[2]ss/:id', (req: Request, res: Response) => {
  const idx = storage.findIndex(i => i.id === req.params.id);
  if (idx === -1) return res.status(404).json({ error: 'Not found' });
  storage[idx] = { ...storage[idx], ...req.body };
  res.json(storage[idx]);
});

router.delete('/%[2]ss/:id', (req: Request, res: Response) => {
  const idx = storage.findIndex(i => i.id === req.params.id);
  if (idx === -1) return res.status(404).json({ error: 'Not found' });
  storage.splice(idx, 1);
  res.status(204).send();
});

export default router;
`, name, lower)
}

func tsComponent(task string) string {
	name := extractName(task)

	return fmt.Sprintf(`import React, { useState } from 'react';

interface %[1]sProps {
  id?: string;
  onSubmit?: (data: %[1]sData) => void;
  onCancel?: () => void;
}

interface %[1]sData {
  name: string;
  description?: string;
}

export const %[1]s: React.FC<%[1]sProps> = ({ id, onSubmit, onCancel }) => {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    try {
      setLoading(true);
      const method = id ? 'PUT' : 'POST';
      const url = id ? \`/api/%ss/\${id}\` : '/api/%ss';
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      });
      if (!response.ok) throw new Error('Failed');
      onSubmit?.(await response.json());
    } catch (err) {
      setError('Operation failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {error && <div className="error">{error}</div>}
      <div>
        <label>Name</label>
        <input value={name} onChange={e => setName(e.target.value)} disabled={loading} />
      </div>
      <div>
        <button type="submit" disabled={loading}>{loading ? 'Saving...' : 'Submit'}</button>
        <button type="button" onClick={onCancel}>Cancel</button>
      </div>
    </form>
  );
};

export default %[1]s;
`, name, strings.ToLower(name[:1]), strings.ToLower(name[:1]))
}

func tsType(task string) string {
	name := extractName(task)

	return fmt.Sprintf(`// %s types

export interface %s {
  id: string;
  name: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Create%sDTO {
  name: string;
  description?: string;
}

export interface Update%sDTO {
  name?: string;
  description?: string;
}

export enum %sStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
}

export interface %sListResponse {
  data: %s[];
  page: number;
  limit: number;
  total: number;
}
`, name, name, name, name, name, name, name)
}

func tsTest(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`import { describe, it, expect, beforeEach, vi } from 'vitest';
import { create%[1]s, validate%[1]s } from './%[2]s';

describe('%[1]s', () => {
  describe('create%[1]s', () => {
    it('should create with valid data', () => {
      const result = create%[1]s({ name: 'Test' });
      expect(result.name).toBe('Test');
      expect(result.id).toBeDefined();
    });

    it('should throw for empty name', () => {
      expect(() => create%[1]s({ name: '' })).toThrow();
    });
  });

  describe('validate%[1]s', () => {
    it('should validate correct data', () => {
      expect(validate%[1]s({ name: 'Valid' })).toBe(true);
    });

    it('should reject empty name', () => {
      expect(validate%[1]s({ name: '' })).toBe(false);
    });
  });
});
`, name, lower)
}

func tsClient(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`// API Client for %s

import type { %s, Create%sDTO, Update%sDTO } from './types';

const API_BASE = '/api/%ss';

class %sClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(url: string, options: RequestInit = {}): Promise<T> {
    const response = await fetch(url, {
      ...options,
      headers: { 'Content-Type': 'application/json', ...options.headers },
    });
    if (!response.ok) throw new Error(\`HTTP error! \${response.status}\`);
    if (response.status === 204) return undefined as T;
    return response.json();
  }

  async create(data: Create%sDTO): Promise<%s> {
    return this.request<%s>(\`\${this.baseUrl}\`, { method: 'POST', body: JSON.stringify(data) });
  }

  async get(id: string): Promise<%s> {
    return this.request<%s>(\`\${this.baseUrl}/\${id}\`);
  }

  async list(params?: { page?: number; limit?: number }): Promise<{ data: %s[] }> {
    const query = new URLSearchParams(params as Record<string, string>);
    return this.request<{ data: %s[] }>(\`\${this.baseUrl}?\${query}\`);
  }

  async update(id: string, data: Update%sDTO): Promise<%s> {
    return this.request<%s>(\`\${this.baseUrl}/\${id}\`, { method: 'PUT', body: JSON.stringify(data) });
  }

  async delete(id: string): Promise<void> {
    return this.request<void>(\`\${this.baseUrl}/\${id}\`, { method: 'DELETE' });
  }
}

export const %sClient = new %sClient();
`, name, name, name, name, lower, name, name, name, name, name, name, name, name, name, name, name, name)
}

func tsHook(task string) string {
	name := extractName(task)
	lower := strings.ToLower(name)

	return fmt.Sprintf(`import { useState, useEffect, useCallback } from 'react';
import %sClient from './client';
import type { %s, Create%sDTO, Update%sDTO } from './types';

export function use%s(autoFetch = true) {
  const [items, setItems] = useState<%s[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(async () => {
    try {
      setLoading(true);
      const { data } = await %sClient.list();
      setItems(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed');
    } finally {
      setLoading(false);
    }
  }, []);

  const create = useCallback(async (data: Create%sDTO) => {
    try {
      setLoading(true);
      const item = await %sClient.create(data);
      setItems(prev => [...prev, item]);
      return item;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed');
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (autoFetch) refetch();
  }, [autoFetch, refetch]);

  return { items, loading, error, create, refetch };
}
`, name, name, name, name, name, name, name, name, name, name)
}

// ═══════════════════════════════════════════════════════════════════════════
// DEVOPS TEMPLATES
// ═══════════════════════════════════════════════════════════════════════════

func dockerfile(task string) string {
	isPython := strings.Contains(strings.ToLower(task), "python")
	isGo := !isPython && (strings.Contains(strings.ToLower(task), "go") || strings.Contains(strings.ToLower(task), "golang"))

	if isPython {
		return `FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

ENV PYTHONUNBUFFERED=1

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
`
	}

	if isGo {
		return `FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /server .
EXPOSE 8080
CMD ["./server"]
`
	}

	return `FROM alpine:latest
WORKDIR /app
COPY . .
EXPOSE 8080
CMD ["echo", "Add your application here"]
`
}

func githubAction(task string) string {
	name := strings.ToLower(extractName(task))

	return `name: CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v ./...
      - run: go build ./...

  docker:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          push: true
          tags: myrepo/` + name + `:latest
`
}

func kubernetes(task string) string {
	name := strings.ToLower(extractName(task))

	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ` + name + `
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ` + name + `
  template:
    metadata:
      labels:
        app: ` + name + `
    spec:
      containers:
        - name: ` + name + `
          image: myrepo/` + name + `:latest
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: ` + name + `
spec:
  selector:
    app: ` + name + `
  ports:
    - port: 80
      targetPort: 8080
  type: ClusterIP
`
}

func readme(task string) string {
	name := extractName(task)

	return `# ` + name + `

A brief description of your project.

## Features

- Feature 1
- Feature 2

## Installation

\`\`\`bash
git clone https://github.com/username/repo.git
cd repo
go mod download
\`\`\`

## Usage

\`\`\`bash
go run ./cmd/server
\`\`\`

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| POST | /api/items | Create item |

## License

MIT
`
}