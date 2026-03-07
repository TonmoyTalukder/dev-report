import * as vscode from "vscode";
import { ReportPanel } from "./panel";
import { ReportSidebarProvider } from "./sidebar";
import { runCLI, ensureBinary } from "./runner";

export function activate(context: vscode.ExtensionContext) {
  // Ensure the Go binary is available on first use
  ensureBinary(context);

  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(
      ReportSidebarProvider.viewType,
      new ReportSidebarProvider(context)
    )
  );

  // Command: open the sidebar panel
  context.subscriptions.push(
    vscode.commands.registerCommand("dev-report.openPanel", () => {
      ReportPanel.createOrShow(context);
    })
  );

  // Command: quick generate with defaults (today, current user from config)
  context.subscriptions.push(
    vscode.commands.registerCommand("dev-report.generate", async () => {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspaceFolder) {
        vscode.window.showErrorMessage("dev-report: No workspace folder open.");
        return;
      }

      const cfg = vscode.workspace.getConfiguration("devReport");
      const user = cfg.get<string>("user") || "";
      const aiProvider = cfg.get<string>("aiProvider") || "groq";

      const args = buildArgs({ user, aiProvider, workDir: workspaceFolder });

      await runCLI(context, args, workspaceFolder, (output: string) => {
        showOutput(output);
      });
    })
  );

  // Command: generate with options (quick-input prompts)
  context.subscriptions.push(
    vscode.commands.registerCommand("dev-report.generateWithOptions", async () => {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspaceFolder) {
        vscode.window.showErrorMessage("dev-report: No workspace folder open.");
        return;
      }

      const options = await promptOptions();
      if (!options) return; // user cancelled

      const cfg = vscode.workspace.getConfiguration("devReport");
      const aiProvider = options.ai || cfg.get<string>("aiProvider") || "groq";

      const args = buildArgs({ ...options, aiProvider, workDir: workspaceFolder });

      await runCLI(context, args, workspaceFolder, (output: string) => {
        showOutput(output);
      });
    })
  );

  // Command: open VS Code settings for dev-report
  context.subscriptions.push(
    vscode.commands.registerCommand("dev-report.configure", () => {
      vscode.commands.executeCommand(
        "workbench.action.openSettings",
        "@ext:dev-report.dev-report"
      );
    })
  );
}

export function deactivate() {}

// ── Helpers ──────────────────────────────────────────────────────────────────

interface GenerateOptions {
  user?: string;
  checkin?: string;
  checkout?: string;
  date?: string;
  last?: string;
  adjust?: string;
  ai?: string;
  aiProvider?: string;
  output?: string;
  workDir?: string;
}

function buildArgs(opts: GenerateOptions): string[] {
  const args = ["generate"];

  if (opts.user) args.push(`--user=${opts.user}`);
  if (opts.date) args.push(`--date=${opts.date}`);
  if (opts.checkin) args.push(`--checkin=${opts.checkin}`);
  if (opts.checkout) args.push(`--checkout=${opts.checkout}`);
  if (opts.last) args.push(`--last=${opts.last}`);
  if (opts.adjust) args.push(`--adjust=${opts.adjust}`);
  if (opts.aiProvider) args.push(`--ai=${opts.aiProvider}`);
  args.push("--output=markdown");

  return args;
}

async function promptOptions(): Promise<GenerateOptions | undefined> {
  const mode = await vscode.window.showQuickPick(
    [
      { label: "Time range (check-in → check-out)", value: "range" },
      { label: "Specific date", value: "date" },
      { label: "Last N commits", value: "last" },
    ],
    { placeHolder: "How do you want to select commits?" }
  );
  if (!mode) return undefined;

  const opts: GenerateOptions = {};

  opts.user = await vscode.window.showInputBox({
    prompt: "Git author name (leave empty for all authors)",
    placeHolder: "e.g. john",
  });

  if (mode.value === "range") {
    opts.checkin = await vscode.window.showInputBox({
      prompt: "Check-in time (HH:MM)",
      placeHolder: "09:00",
      validateInput: (v: string) => (/^\d{2}:\d{2}$/.test(v) ? null : "Format: HH:MM"),
    });
    if (!opts.checkin) return undefined;

    opts.checkout = await vscode.window.showInputBox({
      prompt: "Check-out time (HH:MM)",
      placeHolder: "18:00",
      validateInput: (v: string) => (/^\d{2}:\d{2}$/.test(v) ? null : "Format: HH:MM"),
    });
    if (!opts.checkout) return undefined;

    opts.adjust = await vscode.window.showInputBox({
      prompt: "Time to subtract (breaks/meetings) — leave empty to skip",
      placeHolder: "e.g. 35min or 1h30m",
    });
  } else if (mode.value === "date") {
    opts.date = await vscode.window.showInputBox({
      prompt: "Date (YYYY-MM-DD)",
      placeHolder: new Date().toISOString().slice(0, 10),
      validateInput: (v: string) =>
        /^\d{4}-\d{2}-\d{2}$/.test(v) ? null : "Format: YYYY-MM-DD",
    });
    if (!opts.date) return undefined;
  } else {
    const last = await vscode.window.showInputBox({
      prompt: "Number of last commits",
      placeHolder: "10",
      validateInput: (v: string) => (/^\d+$/.test(v) ? null : "Enter a number"),
    });
    if (!last) return undefined;
    opts.last = last;
  }

  return opts;
}

function showOutput(markdown: string) {
  const doc = vscode.workspace.openTextDocument({
    content: markdown,
    language: "markdown",
  });
  doc.then((d: vscode.TextDocument) => {
    vscode.window.showTextDocument(d, vscode.ViewColumn.Beside);
    vscode.commands.executeCommand("markdown.showPreview", d.uri);
  });
}
