#!/usr/bin/env node
/**
 * dev-report.js — thin Node.js shim that locates and executes the
 * platform-specific Go binary.
 */

const { spawnSync } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");

const isWindows = os.platform() === "win32";
const binName = isWindows ? "dev-report.exe" : "dev-report";
const binPath = path.join(__dirname, binName);

if (!fs.existsSync(binPath)) {
  console.error(
    `\n  dev-report: binary not found at ${binPath}\n` +
      `  Try re-installing: npm install -g dev-report\n`
  );
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: false,
});

process.exit(result.status ?? 1);
