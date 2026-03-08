<p align="center">
  <h1 align="center">dev-report</h1>
  <p align="center">
    <strong>AI-powered developer work report generator.</strong><br>
    Reads your Git commits → generates a structured daily work report. Automatically.
  </p>
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#setup">Setup (5 min)</a> •
  <a href="#api-keys">API Keys</a> •
  <a href="#usage">Usage</a> •
  <a href="#vs-code-extension">VS Code</a> •
  <a href="PUBLISHING.md">Publishing Guide</a>
</p>

---

```
#  Task                                  Module    Description                        Time Spent  Status
1  Added return amount to doctor summary  Hospital  Return amount column added          1h 25m      Completed
2  Excel export for MR list               Hospital  Users can export MR list to Excel   2h 50m      Completed
3  Fixed input bug in suggestion screen   Store     Text input fixed in suggestion      1h 10m      Completed
```

---

## Install

### npm — Windows, Mac, Linux

```bash
npm install -g dev-report
```

The npm package automatically downloads the correct pre-built binary for your OS.

### Homebrew — Mac only

```bash
brew tap dev-report/dev-report
brew install dev-report
```

### VS Code / Open VSX Extension

Search **"Dev Report"** in the Extensions panel, or install from [open-vsx.org](https://open-vsx.org/).

### Direct Binary

Download for your platform from [GitHub Releases](https://github.com/dev-report/dev-report/releases), unzip, and put the binary on your PATH.

---

## Setup

This is a one-time setup. Run it once per project (or globally).

### Step 1 — Run the setup wizard

Open a terminal in your project folder and run:

```bash
dev-report init
```

This starts an interactive wizard that asks:

- **Git author name** — your name as it appears in `git log` (e.g. `TonmoyTalukder`)
- **AI provider** — which AI to use (default: `groq`)
- **API key** — for the provider you chose
- **Default output format** — `table`, `markdown`, `excel`, or `json`

When done, a `dev-report.config.json` file is created in your project.

### Step 2 — Review your config

Open `dev-report.config.json` — it looks like this:

```json
{
  "user": "TonmoyTalukder",
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

Fill in only what you need. You just need the key for the provider you picked.

### Step 3 — Add config to .gitignore

Your config file contains your API key. **Add it to `.gitignore`** so it never gets committed:

```bash
echo "dev-report.config.json" >> .gitignore
```

> **No one else will ever see your API key.** Each developer has their own `dev-report.config.json` with their own key. The tool does not bundle any API keys.

### Step 4 — Generate your first report

```bash
dev-report generate --checkin=09:00 --checkout=18:00
```

Done! ✅

---

## API Keys

> All supported AI providers offer a **free tier**. No credit card required.

### Groq *(default — fastest)*

1. Go to [console.groq.com](https://console.groq.com) → sign up (free)
2. Click **API Keys** → **Create API Key**
3. Copy the key (starts with `gsk_...`)
4. Paste it in `dev-report.config.json`:
   ```json
   {
     "aiProvider": "groq",
     "groqApiKey": "gsk_your_key_here"
   }
   ```

**Free models you can use:**
| Model | Speed | Quality |
|-------|-------|---------|
| `llama-3.3-70b-versatile` *(default)* | Fast | High |
| `llama3-8b-8192` | Very fast | Good |
| `gemma2-9b-it` | Fast | Good |

---

### Google Gemini

1. Go to [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey) → sign in with Google
2. Click **Create API Key** → select a project
3. Copy the key (starts with `AIza...`)
4. Paste it in `dev-report.config.json`:
   ```json
   {
     "aiProvider": "gemini",
     "geminiApiKey": "AIza_your_key_here"
   }
   ```

---

### Ollama *(runs locally — no API key needed)*

1. Download and install Ollama from [ollama.com](https://ollama.com)
2. Pull a model:
   ```bash
   ollama pull llama3
   ```
3. Add to `dev-report.config.json`:
   ```json
   {
     "aiProvider": "ollama",
     "ollamaUrl": "http://localhost:11434",
     "ollamaModel": "llama3"
   }
   ```

No API key, no internet required once the model is downloaded.

---

### OpenRouter *(access to many free models)*

1. Go to [openrouter.ai/keys](https://openrouter.ai/keys) → create account
2. Click **Create Key** → copy it (starts with `sk-or-...`)
3. Paste it in `dev-report.config.json`:
   ```json
   {
     "aiProvider": "openrouter",
     "openRouterApiKey": "sk-or-your_key_here"
   }
   ```

**Free models:**
| Model |
|-------|
| `meta-llama/llama-3.1-8b-instruct:free` *(default)* |
| `google/gemma-3-27b-it:free` |
| `mistralai/mistral-7b-instruct:free` |

---

## Usage

### Basic commands

```bash
# Generate report for today (time range)
dev-report generate --checkin=09:00 --checkout=18:00

# With break time deducted (35 min break → subtracted from task time)
dev-report generate --checkin=09:00 --checkout=18:00 --adjust=35min

# Specific date
dev-report generate --date=2026-03-07 --checkin=09:00 --checkout=17:30

# Last N commits (no time range needed)
dev-report generate --last=10

# Filter by Git author
dev-report generate --user=TonmoyTalukder --checkin=09:00 --checkout=18:00

# Use a different AI provider for this run only
dev-report generate --checkin=09:00 --checkout=18:00 --ai=gemini
```

### Output formats

```bash
# Print table in terminal (default)
dev-report generate --checkin=09:00 --checkout=18:00

# Print as Markdown
dev-report generate --checkin=09:00 --checkout=18:00 --output=markdown

# Save as Markdown file
dev-report generate --checkin=09:00 --checkout=18:00 --output=markdown --out=report.md

# Export to Excel
dev-report generate --checkin=09:00 --checkout=18:00 --output=excel

# Save as JSON
dev-report generate --checkin=09:00 --checkout=18:00 --output=json
```

### All flags

| Flag | Description | Example |
|------|-------------|---------|
| `--user` | Git author name filter | `--user=TonmoyTalukder` |
| `--date` | Date (default: today) | `--date=2026-03-07` |
| `--checkin` | Work start time | `--checkin=09:00` |
| `--checkout` | Work end time | `--checkout=18:00` |
| `--adjust` | Break time to subtract | `--adjust=35min`, `--adjust=1h30m` |
| `--last` | Last N commits | `--last=10` |
| `--ai` | AI provider override | `--ai=gemini` |
| `--output` | Output format | `--output=excel` |
| `--out` | Save to file | `--out=report.md` |

### How time is distributed

When you provide `--checkin` and `--checkout`, the tool calculates your **total task time budget**:

```
Budget = (checkout − checkin) − adjust
```

Each task gets a proportional share based on:

- **Commit count** — tasks with more commits get more time
- **Time spread** — gap between first and last commit in the task
- **Files changed** — more files = more time
- **Lines changed** — larger diffs = more time
- **File complexity** — migrations, configs get a small boost
- **Commit type** — `feat:` > `fix:` > `chore:` as a multiplier

The sum of all time values always equals your budget exactly.

---

## VS Code Extension

The VS Code extension lets you generate reports without leaving your editor.

### Install

Search **"Dev Report"** in the Extensions sidebar, or install from [open-vsx.org](https://open-vsx.org/).

### Commands

Open the Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`) and search for:

| Command | What it does |
|---------|-------------|
| `Dev Report: Generate Work Report` | Generate with default settings |
| `Dev Report: Generate Work Report (with options)` | Opens a panel to set options |
| `Dev Report: Open Dev Report Panel` | Opens the sidebar panel |
| `Dev Report: Configure Dev Report` | Opens settings |

### Configure in VS Code settings

If you want one global configuration that applies to **all projects**, configure through VS Code instead of per-project `dev-report.config.json`:

1. Open VS Code Settings (`Ctrl+,` / `Cmd+,`)
2. Search for **Dev Report**
3. Fill in your values

Or in `settings.json`:

```json
{
  "devReport.user": "TonmoyTalukder",
  "devReport.aiProvider": "groq",
  "devReport.groqApiKey": "gsk_...",
  "devReport.geminiApiKey": "",
  "devReport.openRouterApiKey": "",
  "devReport.ollamaUrl": "http://localhost:11434",
  "devReport.defaultOutput": "markdown"
}
```

> **Tip:** VS Code settings apply globally across all your projects. Per-project `dev-report.config.json` overrides VS Code settings for that specific project.

---

## Config Priority

Configuration is applied in this order (later overrides earlier):

```
VS Code settings
    ↓
dev-report.config.json  (in your project folder)
    ↓
CLI flags  (e.g. --ai=gemini)
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
│   ├── constants/constants.go   # Defaults and constants
│   ├── git/
│   │   ├── fetcher.go           # git log execution
│   │   └── parser.go            # git output parsing
│   ├── processor/
│   │   ├── module.go            # Module detection from file paths
│   │   ├── grouper.go           # Commit grouping
│   │   └── time.go              # Time weight + distribution
│   ├── ai/
│   │   ├── provider.go          # AI interface + factory
│   │   ├── openai_compat.go     # Shared client for Groq/Ollama/OpenRouter
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
├── vscode-extension/            # VS Code / Open VSX extension
└── PUBLISHING.md                # How to publish (npm, Homebrew, VS Code, Open VSX)
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

---

## License

MIT
