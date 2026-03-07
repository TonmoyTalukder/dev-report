import * as vscode from "vscode";
import { runCLI } from "./runner";

interface SidebarPayload {
  user: string;
  checkin: string;
  checkout: string;
  date: string;
  last: string;
  adjust: string;
  ai: string;
}

export class ReportSidebarProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = "dev-report.panel";

  constructor(private readonly context: vscode.ExtensionContext) {}

  public resolveWebviewView(webviewView: vscode.WebviewView): void | Thenable<void> {
    webviewView.webview.options = { enableScripts: true };
    webviewView.webview.html = this.getHtml();

    webviewView.webview.onDidReceiveMessage(async (msg: { command: string; data: SidebarPayload }) => {
      if (msg.command !== "generate") {
        return;
      }

      const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspaceFolder) {
        webviewView.webview.postMessage({ command: "error", text: "No workspace folder open." });
        return;
      }

      const args = ["generate", "--output=markdown"];
      if (msg.data.user) args.push(`--user=${msg.data.user}`);
      if (msg.data.date) args.push(`--date=${msg.data.date}`);
      if (msg.data.checkin) args.push(`--checkin=${msg.data.checkin}`);
      if (msg.data.checkout) args.push(`--checkout=${msg.data.checkout}`);
      if (msg.data.last) args.push(`--last=${msg.data.last}`);
      if (msg.data.adjust) args.push(`--adjust=${msg.data.adjust}`);
      if (msg.data.ai) args.push(`--ai=${msg.data.ai}`);

      webviewView.webview.postMessage({ command: "loading" });

      await runCLI(this.context, args, workspaceFolder, (output: string) => {
        webviewView.webview.postMessage({ command: "result", markdown: output });
      });
    });
  }

  private getHtml(): string {
    const cfg = vscode.workspace.getConfiguration("devReport");
    const defaultUser = cfg.get<string>("user") || "";
    const defaultAI = cfg.get<string>("aiProvider") || "groq";
    const today = new Date().toISOString().slice(0, 10);

    return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Dev Report</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: var(--vscode-font-family);
    font-size: var(--vscode-font-size);
    color: var(--vscode-foreground);
    background: var(--vscode-editor-background);
    padding: 12px;
  }
  h2 { font-size: 14px; font-weight: 600; margin-bottom: 14px; }
  .section { margin-bottom: 12px; }
  label { display: block; font-size: 11px; color: var(--vscode-descriptionForeground); margin-bottom: 4px; }
  input, select {
    width: 100%;
    padding: 6px 8px;
    background: var(--vscode-input-background);
    color: var(--vscode-input-foreground);
    border: 1px solid var(--vscode-input-border, #555);
    border-radius: 3px;
  }
  .row { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
  button {
    width: 100%;
    padding: 8px;
    background: var(--vscode-button-background);
    color: var(--vscode-button-foreground);
    border: none;
    border-radius: 3px;
    cursor: pointer;
    font-weight: 600;
  }
  button:hover { background: var(--vscode-button-hoverBackground); }
  button:disabled { opacity: 0.6; cursor: not-allowed; }
  #status { font-size: 11px; margin-top: 8px; min-height: 16px; color: var(--vscode-descriptionForeground); }
  #result {
    margin-top: 12px;
    padding: 10px;
    border-radius: 4px;
    background: var(--vscode-textCodeBlock-background);
    border: 1px solid var(--vscode-panel-border);
    white-space: pre-wrap;
    font-family: var(--vscode-editor-font-family);
    max-height: 380px;
    overflow-y: auto;
  }
</style>
</head>
<body>
<h2>Dev Report</h2>
<div class="section">
  <label>Git Author Name</label>
  <input id="user" type="text" value="${defaultUser}" placeholder="leave blank for all authors" />
</div>
<div class="section">
  <label>Mode</label>
  <select id="mode" onchange="onModeChange()">
    <option value="range">Time range</option>
    <option value="date">Specific date</option>
    <option value="last">Last N commits</option>
  </select>
</div>
<div id="rangeFields" class="section">
  <div class="row">
    <div>
      <label>Check-in</label>
      <input id="checkin" type="time" value="09:00" />
    </div>
    <div>
      <label>Check-out</label>
      <input id="checkout" type="time" value="18:00" />
    </div>
  </div>
  <div style="margin-top:8px">
    <label>Adjust</label>
    <input id="adjust" type="text" placeholder="35min or 1h30m" />
  </div>
</div>
<div id="dateFields" class="section" style="display:none">
  <label>Date</label>
  <input id="date" type="date" value="${today}" />
</div>
<div id="lastFields" class="section" style="display:none">
  <label>Commit Count</label>
  <input id="last" type="number" value="10" min="1" max="200" />
</div>
<div class="section">
  <label>AI Provider</label>
  <select id="ai">
    <option value="groq" ${defaultAI === "groq" ? "selected" : ""}>Groq</option>
    <option value="gemini" ${defaultAI === "gemini" ? "selected" : ""}>Gemini</option>
    <option value="ollama" ${defaultAI === "ollama" ? "selected" : ""}>Ollama</option>
    <option value="openrouter" ${defaultAI === "openrouter" ? "selected" : ""}>OpenRouter</option>
  </select>
</div>
<button id="generateBtn" onclick="generate()">Generate Report</button>
<div id="status"></div>
<div id="result" style="display:none"></div>
<script>
  const vscode = acquireVsCodeApi();
  function onModeChange() {
    const mode = document.getElementById('mode').value;
    document.getElementById('rangeFields').style.display = mode === 'range' ? '' : 'none';
    document.getElementById('dateFields').style.display = mode === 'date' ? '' : 'none';
    document.getElementById('lastFields').style.display = mode === 'last' ? '' : 'none';
  }
  function generate() {
    const mode = document.getElementById('mode').value;
    const data = {
      user: document.getElementById('user').value,
      ai: document.getElementById('ai').value,
      checkin: mode === 'range' ? document.getElementById('checkin').value : '',
      checkout: mode === 'range' ? document.getElementById('checkout').value : '',
      adjust: mode === 'range' ? document.getElementById('adjust').value : '',
      date: mode === 'date' ? document.getElementById('date').value : '',
      last: mode === 'last' ? document.getElementById('last').value : ''
    };
    vscode.postMessage({ command: 'generate', data });
  }
  window.addEventListener('message', (event) => {
    const msg = event.data;
    const btn = document.getElementById('generateBtn');
    const status = document.getElementById('status');
    const result = document.getElementById('result');
    if (msg.command === 'loading') {
      btn.disabled = true;
      status.textContent = 'Generating report...';
      result.style.display = 'none';
    } else if (msg.command === 'result') {
      btn.disabled = false;
      status.textContent = 'Done';
      result.style.display = '';
      result.textContent = msg.markdown;
    } else if (msg.command === 'error') {
      btn.disabled = false;
      status.textContent = msg.text;
    }
  });
</script>
</body>
</html>`;
  }
}
