import * as vscode from "vscode";
import { runCLI } from "./runner";

/**
 * ReportPanel manages a persistent WebView panel in the VS Code sidebar.
 * It renders an input form and displays the generated report.
 */
export class ReportPanel {
  public static currentPanel: ReportPanel | undefined;
  private readonly _panel: vscode.WebviewPanel;
  private readonly _context: vscode.ExtensionContext;
  private _disposables: vscode.Disposable[] = [];

  public static createOrShow(context: vscode.ExtensionContext) {
    if (ReportPanel.currentPanel) {
      ReportPanel.currentPanel._panel.reveal(vscode.ViewColumn.Beside);
      return;
    }
    const panel = vscode.window.createWebviewPanel(
      "devReportPanel",
      "Dev Report",
      vscode.ViewColumn.Beside,
      { enableScripts: true, retainContextWhenHidden: true }
    );
    ReportPanel.currentPanel = new ReportPanel(panel, context);
  }

  private constructor(panel: vscode.WebviewPanel, context: vscode.ExtensionContext) {
    this._panel = panel;
    this._context = context;
    this._panel.webview.html = this._getHtml();

    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    this._panel.webview.onDidReceiveMessage(
      async (msg: { command: string; data: { user: string; checkin: string; checkout: string; date: string; last: string; adjust: string; ai: string; output: string } }) => {
        if (msg.command === "generate") {
          await this._generate(msg.data);
        }
      },
      null,
      this._disposables
    );
  }

  private async _generate(data: {
    user: string;
    checkin: string;
    checkout: string;
    date: string;
    last: string;
    adjust: string;
    ai: string;
    output: string;
  }) {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
    if (!workspaceFolder) {
      this._panel.webview.postMessage({ command: "error", text: "No workspace folder open." });
      return;
    }

    const args = ["generate", "--output=markdown"];
    if (data.user) args.push(`--user=${data.user}`);
    if (data.date) args.push(`--date=${data.date}`);
    if (data.checkin) args.push(`--checkin=${data.checkin}`);
    if (data.checkout) args.push(`--checkout=${data.checkout}`);
    if (data.last) args.push(`--last=${data.last}`);
    if (data.adjust) args.push(`--adjust=${data.adjust}`);
    if (data.ai) args.push(`--ai=${data.ai}`);

    this._panel.webview.postMessage({ command: "loading" });

    await runCLI(this._context, args, workspaceFolder, (output) => {
      this._panel.webview.postMessage({ command: "result", markdown: output });
    });
  }

  private _getHtml(): string {
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
    padding: 16px;
  }
  h2 { font-size: 14px; font-weight: 600; margin-bottom: 14px; color: var(--vscode-titleBar-activeForeground); }
  .section { margin-bottom: 14px; }
  label { display: block; font-size: 11px; color: var(--vscode-descriptionForeground); margin-bottom: 3px; }
  input, select {
    width: 100%;
    padding: 5px 8px;
    background: var(--vscode-input-background);
    color: var(--vscode-input-foreground);
    border: 1px solid var(--vscode-input-border, #555);
    border-radius: 3px;
    font-size: 13px;
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
    font-size: 13px;
    font-weight: 600;
    margin-top: 4px;
  }
  button:hover { background: var(--vscode-button-hoverBackground); }
  button:disabled { opacity: 0.5; cursor: not-allowed; }
  #status { font-size: 11px; color: var(--vscode-descriptionForeground); margin-top: 8px; min-height: 16px; }
  #result {
    margin-top: 16px;
    padding: 12px;
    background: var(--vscode-textCodeBlock-background);
    border-radius: 4px;
    white-space: pre-wrap;
    font-family: var(--vscode-editor-font-family);
    font-size: 12px;
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid var(--vscode-panel-border);
  }
  .divider { border: none; border-top: 1px solid var(--vscode-panel-border); margin: 12px 0; }
</style>
</head>
<body>
<h2>⚡ Dev Report Generator</h2>

<div class="section">
  <label>Git Author Name</label>
  <input id="user" type="text" placeholder="leave blank for all authors" value="${defaultUser}" />
</div>

<hr class="divider" />

<div class="section">
  <label>Mode</label>
  <select id="mode" onchange="onModeChange()">
    <option value="range">Time range (check-in → check-out)</option>
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
    <label>Adjust (breaks/meetings — optional)</label>
    <input id="adjust" type="text" placeholder="e.g. 35min or 1h30m" />
  </div>
</div>

<div id="dateFields" class="section" style="display:none">
  <label>Date</label>
  <input id="date" type="date" value="${today}" />
</div>

<div id="lastFields" class="section" style="display:none">
  <label>Number of commits</label>
  <input id="last" type="number" value="10" min="1" max="200" />
</div>

<hr class="divider" />

<div class="section">
  <label>AI Provider</label>
  <select id="ai">
    <option value="groq" ${defaultAI === "groq" ? "selected" : ""}>Groq (free — recommended)</option>
    <option value="gemini" ${defaultAI === "gemini" ? "selected" : ""}>Google Gemini (free)</option>
    <option value="ollama" ${defaultAI === "ollama" ? "selected" : ""}>Ollama (local, free)</option>
    <option value="openrouter" ${defaultAI === "openrouter" ? "selected" : ""}>OpenRouter (free models)</option>
  </select>
</div>

<button id="genBtn" onclick="generate()">Generate Report</button>
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
      last: mode === 'last' ? document.getElementById('last').value : '',
    };
    vscode.postMessage({ command: 'generate', data });
  }

  window.addEventListener('message', (event) => {
    const msg = event.data;
    const btn = document.getElementById('genBtn');
    const status = document.getElementById('status');
    const result = document.getElementById('result');

    if (msg.command === 'loading') {
      btn.disabled = true;
      status.textContent = 'Generating report…';
      result.style.display = 'none';
    } else if (msg.command === 'result') {
      btn.disabled = false;
      status.textContent = '✅ Report generated.';
      result.style.display = '';
      result.textContent = msg.markdown;
    } else if (msg.command === 'error') {
      btn.disabled = false;
      status.textContent = '❌ ' + msg.text;
    }
  });
</script>
</body>
</html>`;
  }

  public dispose() {
    ReportPanel.currentPanel = undefined;
    this._panel.dispose();
    while (this._disposables.length) {
      const x = this._disposables.pop();
      if (x) x.dispose();
    }
  }
}
