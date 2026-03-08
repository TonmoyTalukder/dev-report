package constants

import "time"

var SupportedProviders = []string{
	ProviderGroq,
	ProviderGemini,
	ProviderOllama,
	ProviderOpenRouter,
}

var OutputFormats = []string{
	OutputTable,
	OutputMarkdown,
	OutputExcel,
	OutputJSON,
}

var TaskGranularities = []string{
	TaskGranularityBalanced,
	TaskGranularityDetailed,
	TaskGranularityGranular,
}

var GroqFreeModels = []string{
	DefaultGroqModel,
	"llama3-8b-8192",
	"gemma2-9b-it",
	"mixtral-8x7b-32768",
}

var OpenRouterFreeModels = []string{
	DefaultOpenRouterModel,
	"google/gemma-3-27b-it:free",
	"mistralai/mistral-7b-instruct:free",
}

var GitCommitTypes = []string{
	"feat",
	"fix",
	"chore",
	"refactor",
	"test",
	"docs",
	"style",
	"build",
	"ci",
	"perf",
	"revert",
	"wip",
	"merge",
	"hotfix",
}

var ComplexityBoost = map[string]float64{
	".sql":       3.0,
	".migration": 3.0,
	".proto":     2.5,
	".graphql":   2.5,
	".yaml":      1.5,
	".yml":       1.5,
	".json":      1.2,
	".ts":        1.3,
	".tsx":       1.3,
	".go":        1.3,
	".py":        1.3,
	".java":      1.3,
	".kt":        1.3,
	".js":        1.1,
	".jsx":       1.1,
	".md":        0.8,
	".txt":       0.5,
}

var ImportanceMultiplier = map[string]float64{
	"feat":     1.4,
	"fix":      1.0,
	"refactor": 1.2,
	"perf":     1.3,
	"build":    1.1,
	"chore":    0.7,
	"docs":     0.6,
	"style":    0.5,
	"test":     0.8,
	"ci":       0.7,
}

var GitCommitPrefixes = []string{
	"feat:",
	"fix:",
	"chore:",
	"refactor:",
	"test:",
	"docs:",
	"style:",
	"build:",
	"ci:",
	"perf:",
	"revert:",
	"wip:",
	"merge:",
	"hotfix:",
}

const (
	AppName        = "dev-report"
	RepoURL        = "https://github.com/dev-report/dev-report"
	DocsInstallURL = RepoURL + "#install"
	ConfigFileName = "dev-report.config.json"

	ProviderGroq       = "groq"
	ProviderGemini     = "gemini"
	ProviderOllama     = "ollama"
	ProviderOpenRouter = "openrouter"

	OutputTable    = "table"
	OutputMarkdown = "markdown"
	OutputExcel    = "excel"
	OutputJSON     = "json"

	TaskGranularityBalanced = "balanced"
	TaskGranularityDetailed = "detailed"
	TaskGranularityGranular = "granular"

	DefaultAIProvider      = ProviderGroq
	DefaultOutput          = OutputTable
	DefaultTaskGranularity = TaskGranularityBalanced
	DefaultGroqModel       = "llama-3.3-70b-versatile"
	DefaultGeminiModel     = "gemini-1.5-flash"
	DefaultOpenRouterModel = "meta-llama/llama-3.1-8b-instruct:free"
	DefaultOllamaURL       = "http://localhost:11434"
	DefaultOllamaModel     = "llama3"
	DefaultTaskEstimate    = 30 * time.Minute

	GroqConsoleURL    = "https://console.groq.com"
	GeminiConsoleURL  = "https://aistudio.google.com/app/apikey"
	OpenRouterKeysURL = "https://openrouter.ai/keys"

	EnvGroqAPIKey       = "GROQ_API_KEY"
	EnvGeminiAPIKey     = "GEMINI_API_KEY"
	EnvOpenRouterAPIKey = "OPENROUTER_API_KEY"
	EnvOllamaURL        = "OLLAMA_URL"
	EnvOllamaModel      = "OLLAMA_MODEL"
	EnvGroqModel        = "GROQ_MODEL"
	EnvDevReportAPIKey  = "DEV_REPORT_API_KEY"
	EnvDevReportAI      = "DEV_REPORT_AI"
	EnvDevReportOutput  = "DEV_REPORT_OUTPUT"
)
