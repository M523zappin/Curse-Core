package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// AutoDiscovery automatically finds and suggests skills based on codebase patterns
type AutoDiscovery struct {
	store         *Store
	codebaseTypes []string
	patterns      map[string]*SkillPattern
	mu            sync.RWMutex
	codebaseScanner *CodebaseScanner
}

// SkillPattern represents a discovered code pattern
type SkillPattern struct {
	Name        string
	Description string
	Pattern     string
	Language    string
	FileMatch   string
	Confidence  float64
	TimesUsed   int
	LastUsed    time.Time
	AutoGen     bool
}

// CodebaseScanner scans the codebase for patterns
type CodebaseScanner struct {
	repoPath string
	cache    map[string]*ScanResult
	mu       sync.RWMutex
}

// ScanResult holds the results of a codebase scan
type ScanResult struct {
	Files       int
	Lines       int
	Languages    map[string]int
	Frameworks   []string
	Packages     []string
	APIs        []string
	Patterns    []string
	ScannedAt   time.Time
}

// NewAutoDiscovery creates a new auto-discovery instance
func NewAutoDiscovery(store *Store, repoPath string) *AutoDiscovery {
	return &AutoDiscovery{
		store:    store,
		patterns: make(map[string]*SkillPattern),
		codebaseScanner: &CodebaseScanner{
			repoPath: repoPath,
			cache:    make(map[string]*ScanResult),
		},
	}
}

// Discover performs automatic skill discovery
func (a *AutoDiscovery) Discover(ctx context.Context) ([]*SkillPattern, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Scan codebase first
	scan, err := a.codebaseScanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Analyze for patterns
	a.analyzeCodebasePatterns(scan)

	// Generate skills from patterns
	suggestions := a.generateSuggestions(scan)

	return suggestions, nil
}

func (a *AutoDiscovery) GetScanner() *CodebaseScanner {
	return a.codebaseScanner
}

// Scan performs a comprehensive codebase scan
func (s *CodebaseScanner) Scan(ctx context.Context) (*ScanResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check cache (valid for 5 minutes)
	if cached, ok := s.cache["full"]; ok {
		if time.Since(cached.ScannedAt) < 5*time.Minute {
			return cached, nil
		}
	}

	result := &ScanResult{
		Languages:  make(map[string]int),
		Frameworks: []string{},
		Packages:   []string{},
		APIs:      []string{},
		Patterns:  []string{},
		ScannedAt: time.Now(),
	}

	// Walk the repository
	err := filepath.Walk(s.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden dirs and vendor
		if info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") ||
				path == "vendor" || path == "node_modules" || path == "__pycache__" {
				return filepath.SkipAll
			}
			return nil
		}

		// Determine language
		ext := strings.ToLower(filepath.Ext(path))
		lang := extToLanguage(ext)
		if lang != "" {
			result.Languages[lang]++
		}

		result.Files++

		// Read and analyze content for large enough files
		if info.Size() < 1<<20 { // < 1MB
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				result.Lines += strings.Count(content, "\n")

				// Detect frameworks and packages
				s.detectFrameworks(content, lang, result)
				s.detectPatterns(content, lang, result)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.cache["full"] = result
	return result, nil
}

func (s *CodebaseScanner) detectFrameworks(content, lang string, result *ScanResult) {
	frameworks := map[string]map[string]string{
		"go": {
			"github.com/gin-gonic/gin":    "Gin Web Framework",
			"github.com/gofiber/fiber":    "Fiber Web Framework",
			"github.com/labstack/echo":     "Echo Framework",
			"github.com/gorm.io/gorm":     "GORM Database",
			"github.com/jmoiron/sqlx":     "sqlx Database",
			"github.com/go-playground/validator": "Validator",
			"github.com/spf13/cobra":      "Cobra CLI",
			"github.com/charmbracelet":    "Bubble Tea TUI",
			"k8s.io/client-go":            "Kubernetes Client",
			"github.com/aws/aws-sdk-go":    "AWS SDK",
		},
		"python": {
			"flask":                   "Flask",
			"django":                  "Django",
			"fastapi":                 "FastAPI",
			"sqlalchemy":              "SQLAlchemy",
			"pydantic":                "Pydantic",
			"numpy":                   "NumPy",
			"pandas":                  "Pandas",
			"pytest":                  "Pytest",
			"torch":                   "PyTorch",
			"tensorflow":              "TensorFlow",
		},
		"javascript": {
			"express":                 "Express",
			"next":                    "Next.js",
			"react":                   "React",
			"vue":                     "Vue",
			"mongoose":                "Mongoose",
			"prisma":                  "Prisma",
			"jest":                    "Jest",
			"webpack":                 "Webpack",
		},
		"typescript": {
			"express":                 "Express",
			"next":                    "Next.js",
			"react":                   "React",
			"nest":                    "NestJS",
			"typeorm":                 "TypeORM",
			"prisma":                  "Prisma",
			"jest":                    "Jest",
			"webpack":                 "Webpack",
		},
	}

	if fw, ok := frameworks[lang]; ok {
		for pattern, name := range fw {
			if strings.Contains(content, pattern) {
				if !contains(result.Frameworks, name) {
					result.Frameworks = append(result.Frameworks, name)
				}
			}
		}
	}
}

func (s *CodebaseScanner) detectPatterns(content, lang string, result *ScanResult) {
	patterns := map[string]map[string]string{
		"go": {
			"func (.*) Handler":        "HTTP Handler Pattern",
			"func New.*\\(":              "Factory Pattern",
			"interface {}":              "Interface Definition",
			"context.Context":           "Context Propagation",
			"defer.*Close":              "Resource Cleanup",
			"go func()":                "Goroutine Spawn",
			"chan ":                     "Channel Usage",
			"sync.Mutex":               "Mutex Locking",
			"database/sql":             "Database SQL",
			"(*sqlx.DB)":               "Database Connection",
		},
		"python": {
			"@app.route":               "Flask Route",
			"@router":                  "FastAPI Router",
			"class.*\\(.*Model\\)" :     "ORM Model",
			"async def":                "Async Function",
			"@pytest.fixture":           "Pytest Fixture",
			"with.*session":             "Session Management",
			"def __init__":              "Class Constructor",
		},
	}

	if ps, ok := patterns[lang]; ok {
		for pattern, name := range ps {
			re := regexp.MustCompile(pattern)
			if re.MatchString(content) {
				if !contains(result.Patterns, name) {
					result.Patterns = append(result.Patterns, name)
				}
			}
		}
	}
}

func extToLanguage(ext string) string {
	langs := map[string]string{
		".go":     "go",
		".py":     "python",
		".js":     "javascript",
		".ts":     "typescript",
		".jsx":    "javascript",
		".tsx":    "typescript",
		".java":   "java",
		".rs":     "rust",
		".c":      "c",
		".cpp":    "cpp",
		".h":      "c",
		".rb":     "ruby",
		".php":    "php",
		".cs":     "csharp",
		".swift":  "swift",
		".kt":     "kotlin",
		".scala":  "scala",
	}
	return langs[ext]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (a *AutoDiscovery) analyzeCodebasePatterns(scan *ScanResult) {
	// Analyze detected patterns and create skill patterns
	for _, pattern := range scan.Patterns {
		a.patterns[pattern] = &SkillPattern{
			Name:        pattern,
			Description: fmt.Sprintf("Skill for %s", pattern),
			Pattern:     pattern,
			Confidence:  0.8,
			AutoGen:     true,
		}
	}

	// Analyze frameworks and create specialized patterns
	for _, fw := range scan.Frameworks {
		a.patterns[fw] = &SkillPattern{
			Name:        fw,
			Description: fmt.Sprintf("Framework skill for %s", fw),
			Pattern:     fw,
			Confidence:  0.9,
			AutoGen:     true,
		}
	}
}

func (a *AutoDiscovery) generateSuggestions(scan *ScanResult) []*SkillPattern {
	var suggestions []*SkillPattern

	// Get primary language
	primaryLang := a.detectPrimaryLanguage(scan.Languages)

	// Generate language-specific skills
	for lang := range scan.Languages {
		skills := a.generateLanguageSkills(lang)
		suggestions = append(suggestions, skills...)
	}

	// Generate framework-specific skills
	for _, fw := range scan.Frameworks {
		skill := a.generateFrameworkSkill(fw, primaryLang)
		suggestions = append(suggestions, skill)
	}

	// Generate common patterns
	commonPatterns := a.generateCommonPatterns(primaryLang)
	suggestions = append(suggestions, commonPatterns...)

	return suggestions
}

func (a *AutoDiscovery) detectPrimaryLanguage(languages map[string]int) string {
	maxCount := 0
	var primary string

	for lang, count := range languages {
		if count > maxCount {
			maxCount = count
			primary = lang
		}
	}

	return primary
}

func (a *AutoDiscovery) generateLanguageSkills(lang string) []*SkillPattern {
	var skills []*SkillPattern

	templates := map[string][]struct {
		Name, Desc, Pattern string
	}{
		"go": {
			{"REST Handler", "Create REST API handlers", "handler"},
			{"Database Model", "Create database models with GORM/sqlx", "model"},
			{"Middleware", "Create HTTP middleware", "middleware"},
			{"CLI Command", "Create Cobra CLI commands", "cli"},
			{"Test Suite", "Write comprehensive tests", "test"},
			{"Config Parser", "Parse YAML/TOML config", "config"},
			{"Error Wrapper", "Wrap errors with context", "error"},
			{"Graceful Shutdown", "Implement graceful shutdown", "shutdown"},
		},
		"python": {
			{"REST API", "Create FastAPI endpoints", "api"},
			{"SQLAlchemy Model", "Create database models", "model"},
			{"Pytest Fixtures", "Create test fixtures", "fixture"},
			{"CLI Tool", "Create Click/Argparse CLI", "cli"},
			{"Pydantic Schema", "Create validation schemas", "schema"},
			{"Async Handler", "Create async endpoints", "async"},
		},
		"typescript": {
			{"Express Route", "Create Express routes", "route"},
			{"React Component", "Create React components", "component"},
			{"TypeScript Interface", "Create TypeScript interfaces", "interface"},
			{"API Client", "Create API client", "client"},
			{"Test Suite", "Write Jest tests", "test"},
		},
	}

	if templates, ok := templates[lang]; ok {
		for _, t := range templates {
			skills = append(skills, &SkillPattern{
				Name:        t.Name,
				Description: t.Desc,
				Pattern:     t.Pattern,
				Language:    lang,
				Confidence:  0.85,
				AutoGen:     true,
			})
		}
	}

	return skills
}

func (a *AutoDiscovery) generateFrameworkSkill(fw, lang string) *SkillPattern {
	return &SkillPattern{
		Name:        fw + " Integration",
		Description: fmt.Sprintf("Skills for working with %s in %s", fw, lang),
		Pattern:     fw,
		Language:    lang,
		Confidence:  0.9,
		AutoGen:     true,
	}
}

func (a *AutoDiscovery) generateCommonPatterns(lang string) []*SkillPattern {
        return []*SkillPattern{
                {Name: "Error Handling", Description: "Best practices for error handling", Pattern: "error", Language: lang, Confidence: 0.9, AutoGen: true},
                {Name: "Logging", Description: "Structured logging patterns", Pattern: "log", Language: lang, Confidence: 0.85, AutoGen: true},
                {Name: "Configuration", Description: "Configuration management", Pattern: "config", Language: lang, Confidence: 0.85, AutoGen: true},
                {Name: "Testing", Description: "Testing best practices", Pattern: "test", Language: lang, Confidence: 0.9, AutoGen: true},
                {Name: "Documentation", Description: "Code documentation patterns", Pattern: "docs", Language: lang, Confidence: 0.8, AutoGen: true},
                {Name: "API Design", Description: "REST API design patterns", Pattern: "api", Language: lang, Confidence: 0.9, AutoGen: true},
        }
}

// GenerateSkillFromPattern creates a Skill from a SkillPattern
func (a *AutoDiscovery) GenerateSkillFromPattern(p *SkillPattern) *Skill {
	steps := a.generateStepsForPattern(p)

	return &Skill{
		ID:          sanitizeID(p.Name),
		Name:        p.Name,
		Description: p.Description,
		Steps:       steps,
		Pattern:     p.Pattern,
		Tags:        []string{p.Language, "auto-generated", "discovered"},
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"confidence":    p.Confidence,
			"auto_generated": true,
		},
	}
}

func (a *AutoDiscovery) generateStepsForPattern(p *SkillPattern) []string {
	templates := map[string]map[string]string{
		"REST Handler": {
			"Define request/response structs",
			"Create handler function with context",
			"Add input validation",
			"Implement business logic",
			"Return appropriate status codes",
			"Add error handling",
		},
		"Database Model": {
			"Define struct with fields",
			"Add JSON and DB tags",
			"Create validation methods",
			"Implement repository interface",
			"Add migration helpers",
		},
		"Test Suite": {
			"Set up test fixtures",
			"Write happy path tests",
			"Add edge case coverage",
			"Mock external dependencies",
			"Verify assertions",
		},
		"API Design": {
			"Define resource endpoints",
			"Choose HTTP methods",
			"Design request/response formats",
			"Add pagination support",
			"Implement error responses",
			"Document with OpenAPI",
		},
	}

	if steps, ok := templates[p.Name]; ok {
		return steps
	}

	return []string{
		"Understand the requirements",
		"Plan the implementation",
		"Write the code",
		"Add tests",
		"Document the changes",
	}
}

func sanitizeID(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "/", "-")
	id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
	return id
}

// RecommendationEngine suggests the best skill for a task
type RecommendationEngine struct {
	patterns map[string]*SkillPattern
	skills   map[string]*Skill
	mu       sync.RWMutex
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{
		patterns: make(map[string]*SkillPattern),
		skills:   make(map[string]*Skill),
	}
}

// Recommend returns the best matching skill for a task
func (r *RecommendationEngine) Recommend(task string) (*Skill, float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	taskLower := strings.ToLower(task)
	var bestMatch *Skill
	var bestScore float64

	// Calculate scores for all skills
	for _, skill := range r.skills {
		score := r.calculateMatchScore(taskLower, skill)
		if score > bestScore {
			bestScore = score
			bestMatch = skill
		}
	}

	return bestMatch, bestScore
}

func (r *RecommendationEngine) calculateMatchScore(task string, skill *Skill) float64 {
	score := 0.0

	// Check name match
	nameLower := strings.ToLower(skill.Name)
	if strings.Contains(task, nameLower) {
		score += 0.5
	}

	// Check tags match
	for _, tag := range skill.Tags {
		tagLower := strings.ToLower(tag)
		if strings.Contains(task, tagLower) {
			score += 0.2
		}
	}

	// Check description match
	descLower := strings.ToLower(skill.Description)
	if strings.Contains(task, descLower) {
		score += 0.3
	}

	// Normalize by skill age (prefer recently used)
	age := time.Since(skill.UpdatedAt).Hours()
	if age < 24 {
		score *= 1.2 // Boost recently updated skills
	}

	return score
}

// AddSkill adds a skill to the recommendation engine
func (r *RecommendationEngine) AddSkill(skill *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.ID] = skill
}

// GetAllSkills returns all skills
func (r *RecommendationEngine) GetAllSkills() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills
}

// SerializeState serializes the recommendation engine state
func (r *RecommendationEngine) SerializeState() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state := struct {
		Patterns map[string]*SkillPattern
		Skills   map[string]*Skill
	}{
		Patterns: r.patterns,
		Skills:   r.skills,
	}

	return json.Marshal(state)
}

// LoadState loads the recommendation engine state
func (r *RecommendationEngine) LoadState(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := struct {
		Patterns map[string]*SkillPattern
		Skills   map[string]*Skill
	}{}

	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	r.patterns = state.Patterns
	r.skills = state.Skills
	return nil
}