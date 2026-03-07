# dev-report

AI-powered developer work report generator. Reads your Git commits and generates a structured daily work report — automatically.

```
#  Task                                  Module    Description                        Time Spent  Status
1  Added return amount to doctor summary  Hospital  Return amount column added          1h 25m      Completed
2  Excel export for MR list               Hospital  Users can export MR list to Excel   2h 50m      Completed
3  Fixed input bug in suggestion screen   Store     Text input fixed in suggestion      1h 10m      Completed
```

---

## Install

### npm (Windows + Mac + Linux)
```bash
npm install -g dev-report
```

### Homebrew (Mac)
```bash
brew tap dev-report/dev-report
brew install dev-report
```

### VS Code / Open VSX Extension
Search **"Dev Report"** in the Extensions panel, or install from [open-vsx.org](https://open-vsx.org/).

You can also install the packaged VSIX manually:

```bash
code --install-extension vscode-extension/dev-report-0.1.0.vsix
```

---

## Quick Start

```bash
# Set up your config (one-time)
dev-report init

# Generate report for today (time range)
dev-report generate --user=john --checkin=09:00 --checkout=18:00

# With break adjustment (35 min of non-task time subtracted)
dev-report generate --user=john --checkin=09:00 --checkout=18:00 --adjust=35min

# Specific date
dev-report generate --user=john --date=2026-03-07 --checkin=09:00 --checkout=17:30

# Last N commits
dev-report generate --user=john --last=10

# Export to Excel
dev-report generate --user=john --checkin=09:00 --checkout=18:00 --output=excel

# Save as Markdown file
dev-report generate --user=john --checkin=09:00 --checkout=18:00 --output=markdown --out=report.md
```

---

## How Time Is Distributed

When you provide `--checkin` and `--checkout`, the tool calculates a **time budget**:

```
Budget = (checkout - checkin) - adjust
```

Each task gets a proportional share of that budget based on:

| Factor | Weight |
|--------|--------|
| Number of commits in the task | High |
| Time gap between first and last commit | Medium |
| Number of files changed | Medium |
| Lines added/deleted | Medium |
| File type complexity (migrations, configs) | Low boost |
| Commit prefix (feat > fix > chore) | Multiplier |

The sum of all `Time Spent` values always equals the budget exactly.

---

## AI Providers (all free)

| Provider | Free Tier | How to get key |
|----------|-----------|----------------|
| **Groq** *(default)* | Yes — fast | [console.groq.com](https://console.groq.com) |
| **Google Gemini** | Yes — Gemini Flash | [aistudio.google.com](https://aistudio.google.com/app/apikey) |
| **Ollama** | Free — local | [ollama.com](https://ollama.com) |
| **OpenRouter** | Free models | [openrouter.ai/keys](https://openrouter.ai/keys) |

Switch provider with `--ai=gemini` or set it in config.

### API Key Setup

You can configure providers in any of these places:

1. `dev-report.config.json`
2. environment variables
3. VS Code settings for the extension

#### Environment variables

```bash
export GROQ_API_KEY=gsk_...
export GEMINI_API_KEY=AIza...
export OPENROUTER_API_KEY=sk-or-...
export OLLAMA_URL=http://localhost:11434
export OLLAMA_MODEL=llama3
export GROQ_MODEL=llama-3.3-70b-versatile
export DEV_REPORT_AI=groq
export DEV_REPORT_OUTPUT=table
```

PowerShell:

```powershell
$env:GROQ_API_KEY = "gsk_..."
$env:GEMINI_API_KEY = "AIza..."
$env:OPENROUTER_API_KEY = "sk-or-..."
$env:OLLAMA_URL = "http://localhost:11434"
$env:OLLAMA_MODEL = "llama3"
$env:GROQ_MODEL = "llama-3.3-70b-versatile"
$env:DEV_REPORT_AI = "groq"
$env:DEV_REPORT_OUTPUT = "table"
```

#### Config file

```json
{
  "user": "john",
  "aiProvider": "groq",
  "groqApiKey": "gsk_...",
  "groqModel": "llama-3.3-70b-versatile",
  "geminiApiKey": "",
  "openRouterApiKey": "",
  "ollamaUrl": "http://localhost:11434",
  "ollamaModel": "llama3",
  "defaultOutput": "table"
}
```

#### VS Code extension settings

Set these in VS Code Settings UI or `settings.json`:

```json
{
  "devReport.user": "john",
  "devReport.aiProvider": "groq",
  "devReport.groqApiKey": "gsk_...",
  "devReport.geminiApiKey": "",
  "devReport.openRouterApiKey": "",
  "devReport.ollamaUrl": "http://localhost:11434",
  "devReport.defaultOutput": "markdown"
}
```

Do not commit API keys to Git.

---

## Configuration

Run `dev-report init` to create a `dev-report.config.json` in your project:

```json
{
  "user": "john",
  "aiProvider": "groq",
  "groqApiKey": "gsk_...",
  "groqModel": "llama-3.3-70b-versatile",
  "defaultOutput": "table"
}
```

Or use environment variables:

```bash
export GROQ_API_KEY=gsk_...
export GEMINI_API_KEY=AI...
export OPENROUTER_API_KEY=sk-or-...
```

Environment variables override values from `dev-report.config.json`.

---

## All Flags

```
dev-report generate [flags]

Flags:
  --user       Git author name filter (empty = all authors)
  --date       Date to report on, YYYY-MM-DD (default: today)
  --checkin    Work start time, HH:MM
  --checkout   Work end time, HH:MM
  --adjust     Non-task time to subtract (e.g. 35min, 1h40m)
  --last       Use last N commits instead of date/time filter
  --ai         AI provider: groq, gemini, ollama, openrouter
  --output     Output format: table, markdown, excel, json (default: table)
  --out        Output file path (for markdown/excel/json)
```

---

## Project Structure

```
dev-report/
├── main.go                      # Entry point
├── cmd/                         # CLI commands (Cobra)
│   ├── root.go
│   ├── generate.go
│   ├── init.go
│   └── version.go
├── internal/
│   ├── types/types.go           # Shared data types
│   ├── config/config.go         # Config loading
│   ├── constants/constants.go   # Shared defaults and constant values
│   ├── git/
│   │   ├── fetcher.go           # git log execution
│   │   └── parser.go            # git output parsing
│   ├── processor/
│   │   ├── module.go            # Module detection from file paths
│   │   ├── grouper.go           # Related commit grouping
│   │   └── time.go              # Time weight + distribution
│   ├── ai/
│   │   ├── provider.go          # AI interface + factory
│   │   ├── openai_compat.go     # Shared OpenAI-compat client (Groq/Ollama/OpenRouter)
│   │   ├── groq.go              # Groq constants
│   │   ├── gemini.go            # Gemini API client
│   │   ├── openrouter.go        # OpenRouter provider
│   │   └── prompt.go            # Prompt builder + response parser
│   ├── report/
│   │   ├── markdown.go          # Markdown output
│   │   ├── table.go             # Terminal table output
│   │   └── excel.go             # Excel (.xlsx) export
│   └── engine/engine.go         # Full pipeline orchestrator
├── npm/                         # npm wrapper package
│   ├── package.json
│   ├── install.js               # OS-aware binary downloader
│   └── bin/dev-report.js        # Node shim → Go binary
├── vscode-extension/            # VS Code / Open VSX extension
│   ├── src/
│   │   ├── extension.ts         # Activation + commands
│   │   ├── runner.ts            # Binary execution
│   │   ├── panel.ts             # Command-opened WebView panel
│   │   └── sidebar.ts           # Activity bar sidebar WebView provider
│   ├── resources/
│   │   └── icon-small.svg
│   └── package.json
└── .goreleaser.yml              # Cross-platform build + GitHub release
```

---

## Build from Source

```bash
git clone https://github.com/dev-report/dev-report
cd dev-report
go build -o dev-report .

# Windows
go build -o dev-report.exe .
```

## Verification

```bash
go test ./...
go vet ./...
go build ./...
npm --prefix vscode-extension run compile
npm --prefix vscode-extension run package
```

---

## Publishing and Deployment

### 1. Create a GitHub Release with GoReleaser

This project uses `.goreleaser.yml` to build release archives for:

- Windows amd64
- macOS amd64
- macOS arm64
- Linux amd64
- Linux arm64

Typical release flow:

```bash
git tag v0.1.0
git push origin main --tags
goreleaser release --clean
```

This produces GitHub Release assets that the npm installer downloads.

### 2. Publish the npm package

The npm package lives in `npm/` and downloads the correct release binary during `postinstall`.

Typical flow:

```bash
# update npm/package.json version first
npm publish ./npm --access public
```

Requirements:

- npm account with publish rights for the `dev-report` package
- GitHub Release for the same version must already exist
- `npm/package.json` version must match the release tag version

### 3. Publish/update the Homebrew formula

Homebrew installation expects a tap such as:

```bash
brew tap dev-report/dev-report
brew install dev-report
```

For each release you should:

- update the formula in the tap repository
- point it to the new release archive URL
- update the SHA256 checksum
- test with `brew install dev-report`

### 4. Publish the VS Code extension

From `vscode-extension/`:

```bash
npm install
npm run compile
npm run package
```

To publish to the Microsoft Marketplace you need a publisher account and Personal Access Token for `vsce`.

```bash
npx vsce login <publisher>
npx vsce publish
```

### 5. Publish to Open VSX

From `vscode-extension/`:

```bash
npm install
npm run compile
npm run publish-ovsx
```

You need an Open VSX access token configured for `ovsx`.

### Release Order

Use this order to avoid broken installs:

1. publish GitHub Release binaries
2. publish npm package
3. update Homebrew tap
4. publish VS Code Marketplace extension
5. publish Open VSX extension

If you publish npm before GitHub Releases are live, the installer in `npm/install.js` will fail to download the binary.

---

## License

MIT
