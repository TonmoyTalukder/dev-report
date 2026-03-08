#!/usr/bin/env node
/**
 * install.js — downloads the correct pre-built Go binary for the current
 * OS and architecture from GitHub Releases.
 *
 * Supports: Windows (amd64), macOS (amd64, arm64), Linux (amd64, arm64)
 */

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const os = require("os");
const zlib = require("zlib");

const VERSION = require("./package.json").version;
const REPO = "TonmoyTalukder/dev-report";
const BIN_DIR = path.join(__dirname, "bin");

function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  const osMap = { win32: "windows", darwin: "darwin", linux: "linux" };
  const archMap = { x64: "amd64", arm64: "arm64", ia32: "386" };

  const goos = osMap[platform];
  const goarch = archMap[arch];

  if (!goos) throw new Error(`Unsupported OS: ${platform}`);
  if (!goarch) throw new Error(`Unsupported arch: ${arch}`);

  const ext = goos === "windows" ? ".zip" : ".tar.gz";
  const binName = goos === "windows" ? "dev-report.exe" : "dev-report";
  const archiveName = `dev-report_${VERSION}_${goos}_${goarch}${ext}`;

  return { goos, goarch, ext, binName, archiveName };
}

function extractZip(zipPath, extractDir, binName, finalDest) {
  return new Promise((resolve, reject) => {
    const yauzl = require("yauzl");
    
    yauzl.open(zipPath, { lazyEntries: true }, (err, zipfile) => {
      if (err) return reject(err);
      
      zipfile.readEntry();
      zipfile.on("entry", (entry) => {
        if (/\/$/.test(entry.fileName)) {
          zipfile.readEntry();
        } else {
          zipfile.openReadStream(entry, (err, readStream) => {
            if (err) return reject(err);
            
            const outputPath = path.join(extractDir, path.basename(entry.fileName));
            const writeStream = fs.createWriteStream(outputPath);
            
            readStream.pipe(writeStream);
            writeStream.on("close", () => {
              if (path.basename(entry.fileName) === binName) {
                fs.copyFileSync(outputPath, finalDest);
              }
              zipfile.readEntry();
            });
          });
        }
      });
      
      zipfile.on("end", () => {
        resolve();
      });
    });
  });
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    const get = (url) => {
      https
        .get(url, { headers: { "User-Agent": "dev-report-installer" } }, (res) => {
          if (res.statusCode === 301 || res.statusCode === 302) {
            return get(res.headers.location);
          }
          if (res.statusCode !== 200) {
            return reject(new Error(`Download failed: HTTP ${res.statusCode} — ${url}`));
          }
          res.pipe(file);
          file.on("finish", () => file.close(resolve));
        })
        .on("error", reject);
    };
    get(url);
  });
}

async function main() {
  const { goos, binName, archiveName } = getPlatformInfo();
  const downloadURL = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
  const tmpFile = path.join(os.tmpdir(), archiveName);

  if (!fs.existsSync(BIN_DIR)) {
    fs.mkdirSync(BIN_DIR, { recursive: true });
  }

  const binDest = path.join(BIN_DIR, binName);

  if (fs.existsSync(binDest)) {
    fs.unlinkSync(binDest);
  }

  console.log(`  dev-report: downloading ${archiveName}…`);
  try {
    await downloadFile(downloadURL, tmpFile);
  } catch (err) {
    console.error(`\n  ⚠  Binary download failed: ${err.message}`);
    console.error(`  You can download it manually from:`);
    console.error(`  https://github.com/${REPO}/releases/tag/v${VERSION}\n`);
    process.exit(1);
  }

  // Extract binary
  console.log(`  dev-report: extracting…`);
  if (goos === "windows") {
    // Use Node.js to extract zip on Windows
    const extractDir = path.join(os.tmpdir(), "dev-report-extract");
    if (!fs.existsSync(extractDir)) {
      fs.mkdirSync(extractDir, { recursive: true });
    }
    
    try {
      await extractZip(tmpFile, extractDir, binName, binDest);
    } catch (err) {
      // Fallback to tar if yauzl is not available
      try {
        execSync(`tar -xf "${tmpFile}" -C "${extractDir}"`, { stdio: "pipe" });
        const extracted = path.join(extractDir, binName);
        fs.copyFileSync(extracted, binDest);
      } catch (tarErr) {
        throw new Error(`Failed to extract zip: ${err.message}`);
      }
    }
  } else {
    execSync(`tar -xzf "${tmpFile}" -C "${BIN_DIR}" "${binName}"`, { stdio: "pipe" });
    fs.chmodSync(binDest, 0o755);
  }

  fs.unlinkSync(tmpFile);
  console.log(`  ✅ dev-report installed to ${binDest}`);
}

main().catch((err) => {
  console.error(`  ❌ Install error: ${err.message}`);
  process.exit(1);
});
