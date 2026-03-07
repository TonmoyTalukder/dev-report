import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";
import * as fs from "fs";
import * as os from "os";

const BINARY_NAME = os.platform() === "win32" ? "dev-report.exe" : "dev-report";

/**
 * Returns the path to the bundled dev-report binary inside the extension.
 * The binary is stored in <extension>/bin/<platform>/<binary>.
 */
export function getBinaryPath(context: vscode.ExtensionContext): string {
  const platform = os.platform(); // win32 | darwin | linux
  const arch = os.arch(); // x64 | arm64

  const platformDir =
    platform === "win32"
      ? "windows-amd64"
      : platform === "darwin"
      ? arch === "arm64"
        ? "darwin-arm64"
        : "darwin-amd64"
      : arch === "arm64"
      ? "linux-arm64"
      : "linux-amd64";

  return path.join(context.extensionPath, "bin", platformDir, BINARY_NAME);
}

/**
 * Checks if the binary exists and warns the user if not.
 */
export function ensureBinary(context: vscode.ExtensionContext): boolean {
  const binPath = getBinaryPath(context);
  if (!fs.existsSync(binPath)) {
    // Try PATH as fallback (user may have installed globally via npm/brew)
    const inPath = which(BINARY_NAME);
    if (inPath) return true;

    vscode.window
      .showWarningMessage(
        "dev-report binary not found. Install it globally with `npm install -g dev-report` or via Homebrew.",
        "Install instructions"
      )
      .then((sel: string | undefined) => {
        if (sel) {
          vscode.env.openExternal(
            vscode.Uri.parse("https://github.com/dev-report/dev-report#installation")
          );
        }
      });
    return false;
  }
  return true;
}

/**
 * Runs the dev-report CLI with the given args and passes stdout to the callback.
 * Injects VS Code settings as environment variables for the subprocess.
 */
export async function runCLI(
  context: vscode.ExtensionContext,
  args: string[],
  cwd: string,
  onOutput: (output: string) => void
): Promise<void> {
  let binPath = getBinaryPath(context);
  if (!fs.existsSync(binPath)) {
    const inPath = which(BINARY_NAME);
    if (!inPath) {
      vscode.window.showErrorMessage(
        "dev-report: binary not found. Run `npm install -g dev-report` first."
      );
      return;
    }
    binPath = inPath;
  }

  const env = buildEnv();

  return new Promise((resolve) => {
    vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: "dev-report: generating report…",
        cancellable: false,
      },
      () =>
        new Promise<void>((done) => {
          let stdout = "";
          let stderr = "";

          const proc = cp.spawn(binPath, args, { cwd, env: { ...process.env, ...env } });

          proc.stdout.on("data", (d: Buffer | string) => {
            stdout += d.toString();
          });
          proc.stderr.on("data", (d: Buffer | string) => {
            stderr += d.toString();
          });

          proc.on("close", (code: number | null) => {
            done();
            resolve();
            if (code === 0) {
              onOutput(stdout);
            } else {
              vscode.window.showErrorMessage(`dev-report failed:\n${stderr || stdout}`);
            }
          });
        })
    );
  });
}

/** Builds environment variables from VS Code settings for the subprocess. */
function buildEnv(): Record<string, string> {
  const cfg = vscode.workspace.getConfiguration("devReport");
  const env: Record<string, string> = {};

  const groqKey = cfg.get<string>("groqApiKey");
  if (groqKey) env["GROQ_API_KEY"] = groqKey;

  const geminiKey = cfg.get<string>("geminiApiKey");
  if (geminiKey) env["GEMINI_API_KEY"] = geminiKey;

  const openRouterKey = cfg.get<string>("openRouterApiKey");
  if (openRouterKey) env["OPENROUTER_API_KEY"] = openRouterKey;

  const ollamaUrl = cfg.get<string>("ollamaUrl");
  if (ollamaUrl) env["OLLAMA_URL"] = ollamaUrl;

  return env;
}

/** Looks up a binary name in PATH. Returns the full path or null. */
function which(name: string): string | null {
  try {
    const result = cp.execSync(
      os.platform() === "win32" ? `where ${name}` : `which ${name}`,
      { encoding: "utf8", stdio: ["pipe", "pipe", "pipe"] }
    );
    return result.trim().split("\n")[0] || null;
  } catch {
    return null;
  }
}
