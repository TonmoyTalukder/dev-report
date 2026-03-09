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

# 2. Commit and tag the release
git add -A && git commit -m "chore: bump version to vX.Y.Z"
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z

# 3. Run GoReleaser (uses gh CLI token automatically if you are logged in)
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```

> **Note:** GoReleaser requires a clean git state — no uncommitted files. Always commit everything before tagging.

### Verify

Go to [github.com/TonmoyTalukder/dev-report/releases](https://github.com/TonmoyTalukder/dev-report/releases) and confirm the release includes:

```
checksums.txt
dev-report_X.Y.Z_darwin_amd64.tar.gz
dev-report_X.Y.Z_darwin_arm64.tar.gz
dev-report_X.Y.Z_linux_amd64.tar.gz
dev-report_X.Y.Z_linux_arm64.tar.gz
dev-report_X.Y.Z_windows_amd64.zip
```

---

## Step 2 — npm Package

The npm package (`npm/`) downloads the correct binary for the user's OS during `postinstall`.

There are now two npm distribution targets:

- **npmjs public package** — `dev-report`
- **GitHub Packages mirror** — `@tonmoytalukder/dev-report`

Publishing to npmjs does **not** make the repository's **Packages** tab show a package. That page only lists packages published to **GitHub Packages**.

### ⚠ Required: Create an npm Automation Token

npmjs.com has 2FA enabled by default. Publishing from the CLI requires an **Automation Token** (not a regular login) to bypass the 2FA prompt.

**One-time setup:**

1. Go to [npmjs.com](https://www.npmjs.com/) → click your profile → **Access Tokens**
2. Click **Generate New Token** → choose **Granular Access Token**
3. Configure:
   - **Token name:** `dev-report-publish`
   - **Expiration:** your preference
   - **Packages and scopes:** Select `dev-report` → Permission: **Read and write**
   - **Organization:** None required
   - Check **Allow publishing with two-factor authentication bypass** (this is the key setting)
4. Click **Generate Token** and copy it
5. Add it to your shell profile so it persists:
   ```bash
   echo 'export NPM_TOKEN=npm_your_token_here' >> ~/.zshrc
   source ~/.zshrc
   ```

### How to publish

```bash
# 1. Update version in npm/package.json to match the release tag
#    e.g. "version": "0.1.0"

# 2. Set your Granular Access Token (one-time per machine)
npm config set //registry.npmjs.org/:_authToken npm_your_token_here

# 3. Publish
npm publish ./npm --access public
```

> **Why `npm config set`?** Setting `NPM_TOKEN=xxx` in the shell alone doesn't work — npm only reads it if `~/.npmrc` has `${NPM_TOKEN}` as a template. Using `npm config set` writes it directly to `~/.npmrc` so it works every time.

### Version must match

The `npm/package.json` version **must exactly match** the GitHub Release tag. The install script builds the binary download URL from the version:

```
https://github.com/TonmoyTalukder/dev-report/releases/download/v{VERSION}/dev-report_{VERSION}_{OS}_{ARCH}.tar.gz
```

### Verify

```bash
npm install -g dev-report
dev-report version
```

### GitHub Packages mirror

The repo includes a GitHub Actions workflow at `.github/workflows/publish-github-package.yml`.
On each published GitHub release, it prepares a scoped package from `npm/` and publishes:

```text
@tonmoytalukder/dev-report
```

to:

```text
https://npm.pkg.github.com
```

#### Manual publish steps

```bash
# 1. Prepare the scoped GitHub package mirror
npm --prefix ./npm run prepare:github-package

# 2. Authenticate to GitHub Packages
npm login --scope=@tonmoytalukder --auth-type=legacy --registry=https://npm.pkg.github.com

# 3. Publish the scoped mirror package
npm publish ./npm/.github-package
```

#### Verify

- Open `https://github.com/TonmoyTalukder/dev-report/packages`
- Confirm `@tonmoytalukder/dev-report` is listed

---

## Step 3 — Homebrew

Homebrew requires a **tap repository** — a separate GitHub repo that contains the formula.

### One-time setup

1. Create a new GitHub repository named `homebrew-dev-report`
   - Full name: `github.com/TonmoyTalukder/homebrew-dev-report`

2. Inside that repo, create a file: `Formula/dev-report.rb`

### Formula template (v0.1.0 — real checksums)

```ruby
class DevReport < Formula
  desc "AI-powered developer work report generator"
  homepage "https://github.com/TonmoyTalukder/dev-report"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/TonmoyTalukder/dev-report/releases/download/v0.1.0/dev-report_0.1.0_darwin_arm64.tar.gz"
      sha256 "693913d8fc43cff04df82d4deab4579acdf78c599ae23ec9bdcf841024ac2b5a"
    else
      url "https://github.com/TonmoyTalukder/dev-report/releases/download/v0.1.0/dev-report_0.1.0_darwin_amd64.tar.gz"
      sha256 "54119f33ce7c7c43db1e964214fd4efe661699a3301dbd6265991438e91bf225"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/TonmoyTalukder/dev-report/releases/download/v0.1.0/dev-report_0.1.0_linux_arm64.tar.gz"
      sha256 "3442d483bad9f216729693d064950a8a1a18906a10fa914f3d006904442c9e99"
    else
      url "https://github.com/TonmoyTalukder/dev-report/releases/download/v0.1.0/dev-report_0.1.0_linux_amd64.tar.gz"
      sha256 "e67e95ca8bd6233a52a3436f1fdcecfc11803ced25260b13f9191c148de119c4"
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

> **Tip:** For each new release, get the SHA256 values from the `checksums.txt` file attached to the GitHub Release.

### For each new release

1. Update the version number in the formula
2. Update all SHA256 values from the new `checksums.txt`
3. Update all download URLs to the new version
4. Commit and push to `homebrew-dev-report`

### Verify

```bash
brew tap TonmoyTalukder/homebrew-dev-report
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
[ ] Go tests pass:        go test ./...
[ ] No vet issues:        go vet ./...
[ ] Version bumped in:    npm/package.json, vscode-extension/package.json
[ ] Committed + tagged:   git tag vX.Y.Z && git push origin main && git push origin vX.Y.Z
[ ] GoReleaser run:       GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
[ ] GitHub Release confirmed (all 5 archives + checksums.txt visible)
[ ] npm Automation Token ready (npmjs.com → Access Tokens → Granular)
[ ] npm published:        NPM_TOKEN=npm_... npm publish ./npm --access public
[ ] GitHub Packages mirror visible: @tonmoytalukder/dev-report on the repo Packages tab
[ ] Homebrew formula updated with new version + real SHA256 from checksums.txt
[ ] VS Code extension published: npx vsce publish
[ ] Open VSX extension published: npm run publish-ovsx
```
