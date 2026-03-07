# dev-report — AI Developer Work Report Generator

---

## Project Overview

`dev-report` is a developer productivity tool that reads Git commits and generates a structured daily work report automatically. No manual writing needed.

The report has these columns:

| # | Task | Module | Description | Time Spent | Status |
|---|------|--------|-------------|------------|--------|

The tool is delivered as:
- A **VS Code extension** (publishable to [open-vsx.org](https://open-vsx.org/))
- A **global CLI tool** usable from any project folder terminal (Windows & Mac)

---

## Delivery Targets

### 1. VS Code Extension
- Works inside VS Code and any Open VSX compatible editor (e.g. VSCodium, Gitpod, Eclipse Theia)
- Publishable to `open-vsx.org`
- Commands accessible via Command Palette (`Ctrl+Shift+P`)
- Sidebar panel to fill inputs and view/export the report

### 2. CLI (Global Install)
- Install globally: `npm install -g dev-report`
- Open terminal inside any Git project folder and run commands directly
- Works on Windows and Mac

Example CLI usage:
```
dev-report generate --user=john --checkin=09:00 --checkout=18:00
dev-report generate --user=john --date=2026-03-07 --adjust=35min
dev-report generate --user=john --last=10
```

---

## Input Options

### Mode 1 — Date Based
```
dev-report generate --user=john --date=2026-03-07
```
Fetches all commits by the user on that date.

### Mode 2 — Commit Count Based
```
dev-report generate --user=john --last=10
```
Fetches last N commits by the user.

### Mode 3 — Time Range Based (Check-in / Check-out)
```
dev-report generate --user=john --checkin=09:00 --checkout=18:00
```
Fetches commits that fall within that time window on today's date.

### Optional: Time Adjust
```
--adjust=35min
--adjust=1h40m
```
Represents time that was NOT spent on tasks (breaks, meetings, admin, etc.).

If `--checkin=09:00 --checkout=18:00` → total window = **9 hrs**
And `--adjust=35min` → adjusted task time budget = **8 hrs 25 mins**

The sum of all `Time Spent` values in the report will equal the adjusted task time budget.

If no `--adjust` is given, the full check-in to check-out duration is used.

---

## Time Distribution Logic

Time is distributed across tasks — not guessed randomly. The system weighs each task using these factors:

| Factor | Description |
|--------|-------------|
| **Commit count** | Tasks with more commits get more time |
| **Commit spread** | Gap between earliest and latest commit for a task |
| **File count changed** | More files touched = more time |
| **Lines changed** | Larger diffs = more time |
| **Complexity signals** | File types like migrations, configs get a weight boost |
| **Importance signals** | Commit prefixes like `feat:`, `fix:` affect weight |

Each task gets a proportional share of the total available time budget.

Example:
- Total time budget: 8 hrs 25 mins
- Task A weight: 3 → gets ~50% → 4 hrs 10 mins
- Task B weight: 2 → gets ~33% → 2 hrs 50 mins
- Task C weight: 1 → gets ~17% → 1 hr 25 mins

All values are rounded to nearest 5 min. Total always equals the time budget.

---

## Task Language Rules

Task names and descriptions must be:
- Short and clear — plain English, no jargon
- Business-friendly, not technical
- Simple sentences, not code terms

**Good examples:**
- `Added export to Excel for MR list`
- `Fixed input bug in suggestion screen`
- `Updated doctor summary with return amount`

**Bad examples:**
- `feat: add return_amount col in doctorSummary.tsx`
- `Refactored DI container and fixed null pointer`

---

## Data Extraction from Git

For each commit the system reads:
- Commit hash
- Author name
- Timestamp
- Commit message
- Changed files (paths)
- Module (detected from folder name)

Example parsed commit:
```
hash:    a1b9f3
author:  dev1
time:    11:25
message: feat: add return amount column in doctor summary
files:   hospital/doctorSummary.tsx, hospital/service.ts
```

---

## Pre-Processing

Before sending to AI:

1. **Clean messages** — strip prefixes like `feat:`, `fix:`, `chore:`, `refactor:`
2. **Detect modules** — from top-level folder name (`hospital/` → `Hospital`)
3. **Group commits** — related commits merged into one task
4. **Compute weights** — each task assigned a time weight based on the factors above

---

## AI Processing

Processed commits + time budget → sent to AI model.

The AI receives:
- Grouped commit messages
- Changed files and modules
- Total available time budget (after adjustment)
- Time weights per task group

The AI returns structured tasks with:
- Task title (simple words)
- Module
- Description (one plain sentence)
- Time spent (calculated from weights, not guessed by AI)
- Status

Example AI prompt instruction:
```
You are writing a developer work report.
Write task titles and descriptions in plain English, simple sentences.
Do not use technical terms. Write like you are explaining to a manager.
One commit group = one task row.
Use the time values provided — do not change them.
```

---

## AI Output (JSON)

```json
[
  {
    "task": "Added return amount to doctor summary",
    "module": "Hospital",
    "description": "Return amount column added to the doctor summary report",
    "timeSpent": "1h 25m",
    "status": "Completed"
  },
  {
    "task": "Excel export for MR list",
    "module": "Hospital",
    "description": "Users can now export the MR list to Excel",
    "timeSpent": "2h 50m",
    "status": "Completed"
  }
]
```

---

## Report Output Formats

### Markdown
```markdown
# Daily Work Report
Date: 2026-03-07 | Developer: dev1 | Check-in: 09:00 | Check-out: 18:00 | Adjusted: 35 min

| # | Task | Module | Description | Time Spent | Status |
|---|------|--------|-------------|------------|--------|
| 1 | Added return amount to doctor summary | Hospital | Return amount column added | 1h 25m | Completed |
| 2 | Excel export for MR list | Hospital | Users can export MR list to Excel | 2h 50m | Completed |
```

### Plain Table (terminal)
Printed in terminal as a formatted table.

### Excel Export
Saved as `work_report_2026-03-07.xlsx`

### PDF (optional, future)

---

## Full System Flow

```
1. User runs command (CLI or VS Code panel)
   └─ Inputs: user, date/range/count, checkin, checkout, adjust

2. Git log is fetched
   └─ Filtered by author + date/time range

3. Commit parser extracts messages, files, timestamps

4. Pre-processor:
   ├─ Cleans commit messages
   ├─ Detects modules from folder names
   ├─ Groups related commits into task groups
   └─ Calculates time weights per group

5. Time budget is computed
   └─ Budget = (checkout - checkin) - adjust

6. Time is distributed across task groups (proportional to weights)

7. AI receives grouped commits + time values
   └─ Returns task title, module, description, status

8. JSON → formatted report

9. Output: Markdown / Table / Excel / PDF
```

---

## Example Final Output

```
Daily Work Report
Date: 2026-03-07 | Developer: dev1
Check-in: 09:00 | Check-out: 18:00 | Adjusted: 35 min | Total Task Time: 8h 25m

#  Task                                  Module    Description                              Time Spent  Status
1  Added return amount to doctor summary  Hospital  Return amount column added to report     1h 25m      Completed
2  Excel export for MR list               Hospital  Users can export MR list to Excel        2h 50m      Completed
3  Fixed input bug in suggestion screen   Store     Text input issue in suggestion screen    1h 10m      Completed
4  Updated product listing filters        Store     Filters now work correctly on listing    3h 00m      Completed
```
Total: 8h 25m ✓

---

## Configuration File (optional)

A `dev-report.config.json` in the project root can store defaults:
```json
{
  "user": "john",
  "aiProvider": "openai",
  "apiKey": "sk-...",
  "defaultOutput": "markdown"
}
```

API key can also be set via environment variable: `DEV_REPORT_API_KEY`

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Core language | **Go** (CLI binary, Git parsing, AI, report engine) |
| VS Code extension | **TypeScript** (required by VS Code API — thin shell calling Go binary) |
| AI providers | Groq, Google Gemini, Ollama, OpenRouter, Hugging Face (all free tiers) |
| Excel export | `tealeg/xlsx` or `qax-os/excelize` (Go) |
| Packaging | GoReleaser (cross-compile win/mac/linux) |
| npm distribution | npm package wrapping Go binary (esbuild pattern — installs correct binary for OS) |
| brew distribution | Custom Homebrew tap (`brew tap user/dev-report && brew install dev-report`) |
| VS Code registry | open-vsx.org + VS Code Marketplace |
| npm registry | npmjs.com (`npm install -g dev-report`) |

---

## Free AI Providers

The tool supports multiple free AI providers. User picks one via config or `--ai` flag.

| Provider | Free Tier | Notes |
|----------|-----------|-------|
| **Groq** | Yes — fast, generous free tier | Recommended default |
| **Google Gemini** | Yes — Gemini 1.5 Flash free | Good quality |
| **Ollama** | Free — runs locally | No internet needed |
| **OpenRouter** | Free models available | Multiple models via one API |
| **Hugging Face** | Free Inference API | Slower but free |
| **OpenAI** | Paid — optional fallback | GPT-4o etc. |

Default provider: **Groq** (fastest free option).
User sets API key in config file or env var per provider.

---

## Distribution

### npm (Windows + Mac + Linux)
```
npm install -g dev-report
```
- The npm package detects OS + arch and downloads the correct pre-built Go binary
- Pattern used by esbuild, Prisma, etc.
- Works on Windows, macOS (Intel + Apple Silicon), Linux

### Homebrew (Mac)
```
brew tap user/dev-report
brew install dev-report
```
- Custom Homebrew tap pointing to GitHub releases
- Mac-native install, no Node.js required

### Direct binary
- GitHub Releases: download `dev-report-windows-amd64.exe`, `dev-report-darwin-arm64`, etc.
- Built automatically via GoReleaser in CI

---

## Future Enhancements

- Weekly report: `dev-report generate --week`
- GitHub / GitLab API integration (no local Git needed)
- Slack bot: auto-post report to team channel
- Jira integration: map tasks to Jira tickets

---

## Implementation Plan

### Phase 1 — Core Engine
- [ ] Git commit fetcher (by date, count, time range)
- [ ] Commit parser (hash, author, time, message, files)
- [ ] Module detector from file paths
- [ ] Commit grouper (related commits → one task)
- [ ] Time weight calculator (commits, files, lines, complexity)
- [ ] Time distributor (proportional, respects budget + adjust)

### Phase 2 — AI Integration
- [ ] AI provider interface (pluggable)
- [ ] Groq integration (default free provider)
- [ ] Google Gemini integration (free)
- [ ] Ollama integration (local/free)
- [ ] OpenRouter integration (free models)
- [ ] AI prompt builder (task language rules enforced)
- [ ] Response parser and validator
- [ ] Fallback if AI fails (use cleaned commit text directly)

### Phase 3 — CLI (Go)
- [ ] Cobra CLI with all flags (`generate`, `init`, `version`)
- [ ] Config file support (`dev-report.config.json`)
- [ ] Output formats: Markdown, terminal table, Excel (.xlsx)
- [ ] Cross-compile via GoReleaser (win/mac/linux)
- [ ] npm wrapper package (OS-aware binary installer)
- [ ] Homebrew tap formula

### Phase 4 — VS Code Extension (TypeScript)
- [ ] Extension scaffold (`yo code`)
- [ ] Bundles or downloads Go binary on first use
- [ ] Command palette: `Dev Report: Generate`, `Dev Report: Configure`
- [ ] Sidebar webview panel (inputs form + rendered report)
- [ ] Open VSX + VS Code Marketplace publish setup

---
