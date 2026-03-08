# Publishing Guide — dev-report

This guide covers how to publish dev-report across all distribution channels:
npm, Homebrew, VS Code Marketplace, and Open VSX.

---

## Prerequisites

Have these ready before publishing:

| What | Used for | Where to get it |
|------|----------|-----------------|
| [GoReleaser](https://goreleaser.com/install/) | Building cross-platform binaries | `brew install goreleaser` or [goreleaser.com](https://goreleaser.com/install/) |
| GitHub Personal Access Token | Creating GitHub Releases | [github.com/settings/tokens](https://github.com/settings/tokens) — scope: `repo` |
| npm account | Publishing the npm package | [npmjs.com/signup](https://www.npmjs.com/signup) |
| Azure DevOps Personal Access Token | Publishing to VS Code Marketplace | [dev.azure.com](https://dev.azure.com) |
| Open VSX access token | Publishing to Open VSX | [open-vsx.org](https://open-vsx.org/) |

---

## Release Order

> ⚠ **Always follow this order.** The npm installer downloads binaries from GitHub Releases — if you publish npm first, the install will fail.

```
Step 1 → GitHub Release  (GoReleaser)
Step 2 → npm
Step 3 → Homebrew
Step 4 → VS Code Marketplace
Step 5 → Open VSX
```

---

## Step 1 — GitHub Release (GoReleaser)

GoReleaser builds cross-platform binaries and uploads them as GitHub Release assets.

**Platforms built:**
- Windows x64
- macOS Intel (amd64)
- macOS Apple Silicon (arm64)
- Linux x64
- Linux ARM64

### How to release

```bash
# 1. Run tests first
go test ./...
go vet ./...

# 2. Tag the release
git tag v0.1.0
git push origin main --tags

# 3. Run GoReleaser
export GITHUB_TOKEN=ghp_your_token_here
goreleaser release --clean
```

### Verify

Go to [github.com/dev-report/dev-report/releases](https://github.com/dev-report/dev-report/releases) and confirm the release includes:

```
checksums.txt
dev-report_0.1.0_darwin_amd64.tar.gz
dev-report_0.1.0_darwin_arm64.tar.gz
dev-report_0.1.0_linux_amd64.tar.gz
dev-report_0.1.0_linux_arm64.tar.gz
dev-report_0.1.0_windows_amd64.zip
```

---

## Step 2 — npm Package

The npm package (`npm/`) downloads the correct binary for the user's OS during `postinstall`.

### How to publish

```bash
# 1. Update version in npm/package.json to match the release tag
#    e.g. "version": "0.1.0"

# 2. Log in to npm
npm login

# 3. Publish
npm publish ./npm --access public
```

### Version must match

The `npm/package.json` version **must exactly match** the GitHub Release tag. The install script builds the binary download URL from the version:

```
https://github.com/dev-report/dev-report/releases/download/v{VERSION}/dev-report_{VERSION}_{OS}_{ARCH}.tar.gz
```

### Verify

```bash
npm install -g dev-report
dev-report version
```

---

## Step 3 — Homebrew

Homebrew requires a **tap repository** — a separate GitHub repo that contains the formula.

### One-time setup

1. Create a new GitHub repository named `homebrew-dev-report`
   - Full name: `github.com/dev-report/homebrew-dev-report`

2. Inside that repo, create a file: `Formula/dev-report.rb`

### Formula template

Get the SHA256 checksums from `checksums.txt` in the GitHub Release, then fill in the formula:

```ruby
class DevReport < Formula
  desc "AI-powered developer work report generator"
  homepage "https://github.com/dev-report/dev-report"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dev-report/dev-report/releases/download/v0.1.0/dev-report_0.1.0_darwin_arm64.tar.gz"
      sha256 "PASTE_SHA256_FROM_checksums.txt"
    else
      url "https://github.com/dev-report/dev-report/releases/download/v0.1.0/dev-report_0.1.0_darwin_amd64.tar.gz"
      sha256 "PASTE_SHA256_FROM_checksums.txt"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dev-report/dev-report/releases/download/v0.1.0/dev-report_0.1.0_linux_arm64.tar.gz"
      sha256 "PASTE_SHA256_FROM_checksums.txt"
    else
      url "https://github.com/dev-report/dev-report/releases/download/v0.1.0/dev-report_0.1.0_linux_amd64.tar.gz"
      sha256 "PASTE_SHA256_FROM_checksums.txt"
    end
  end

  def install
    bin.install "dev-report"
  end

  test do
    assert_match "dev-report", shell_output("#{bin}/dev-report version")
  end
end
```

### For each new release

1. Update the version number in the formula
2. Update all SHA256 values from the new `checksums.txt`
3. Update all download URLs to the new version
4. Commit and push to `homebrew-dev-report`

### Verify

```bash
brew tap dev-report/dev-report
brew install dev-report
dev-report version
```

---

## Step 4 — VS Code Marketplace

### One-time setup: Create a publisher

1. Go to [marketplace.visualstudio.com/manage](https://marketplace.visualstudio.com/manage)
2. Sign in with a Microsoft account
3. Click **Create Publisher**
4. Set display name and ID (e.g. `dev-report`)
5. Update `"publisher"` in `vscode-extension/package.json` to match

### One-time setup: Create a Personal Access Token

You need a PAT with permission to publish extensions:

1. Go to [dev.azure.com](https://dev.azure.com) → sign in
2. Click your profile icon → **Personal Access Tokens**
3. Click **New Token**
4. Set:
   - **Name:** `vsce-publish`
   - **Organization:** All accessible organizations
   - **Scopes:** Custom → check **Marketplace → Manage**
5. Create and copy the token

### How to publish

```bash
cd vscode-extension

# 1. Install dependencies (first time only)
npm install

# 2. Compile TypeScript
npm run compile

# 3. Log in using your PAT
npx vsce login dev-report
# Paste your Personal Access Token when prompted

# 4. Publish
npx vsce publish

# Or publish a specific version
npx vsce publish 0.1.0
```

### For each new release

- Update `"version"` in `vscode-extension/package.json`
- Re-run the publish steps above

### Verify

After publishing, search **"Dev Report"** in the VS Code Extensions panel. It usually appears within a few minutes.

---

## Step 5 — Open VSX

Open VSX is the extension registry for VS Code-compatible editors (VSCodium, Gitpod, Eclipse Theia, etc.).

### One-time setup: Get an access token

1. Go to [open-vsx.org](https://open-vsx.org/) → sign in with GitHub
2. Click your profile → **Access Tokens**
3. Click **Create Token** → name it `ovsx-publish`
4. Copy the token

### How to publish

```bash
cd vscode-extension

# 1. Compile (if not already done)
npm install
npm run compile

# 2. Publish using your token
export OVSX_PAT=your_token_here
npm run publish-ovsx
# This runs: ovsx publish
```

### Verify

Search for **dev-report** on [open-vsx.org](https://open-vsx.org/).

---

## Checklist for Every Release

Use this checklist when releasing a new version:

```
[ ] Go tests pass:  go test ./...
[ ] No vet issues:  go vet ./...
[ ] Version bumped in: npm/package.json, vscode-extension/package.json
[ ] Git tagged:     git tag vX.Y.Z && git push origin main --tags
[ ] GoReleaser run: goreleaser release --clean
[ ] GitHub Release confirmed (all 5 archives + checksums.txt visible)
[ ] npm published:  npm publish ./npm --access public
[ ] Homebrew formula updated with new version + SHA256 values
[ ] VS Code extension published: npx vsce publish
[ ] Open VSX extension published: npm run publish-ovsx
```
