# Deployment

Guide for releasing and distributing kube-diff across multiple channels.

<br/>

## Release Flow

A single tag push triggers the entire release pipeline automatically:

```
git tag v1.0.0 && git push origin v1.0.0
    └→ GitHub Actions (release.yml)
        └→ GoReleaser
            ├→ GitHub Releases (linux/darwin/windows × amd64/arm64)
            ├→ Homebrew tap update (somaz94/homebrew-tap)
            └→ Krew manifest update (somaz94/krew-index)
```

<br/>

## Distribution Channels

### 1. GitHub Releases (Default)

Automatically built by GoReleaser. No additional setup required.

```bash
# Download binary
curl -sL https://github.com/somaz94/kube-diff/releases/latest/download/kube-diff_linux_amd64.tar.gz | tar xz
sudo mv kube-diff /usr/local/bin/
```

**Supported platforms:**

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS (Darwin) | amd64, arm64 |
| Windows | amd64, arm64 |

<br/>

### 2. Homebrew (macOS / Linux)

GoReleaser automatically commits a formula to the `somaz94/homebrew-tap` repository.

```bash
brew install somaz94/tap/kube-diff
```

**Prerequisites:**
- Create `somaz94/homebrew-tap` repository on GitHub
- `PAT_TOKEN` secret must have write access to that repository

<br/>

### 3. Krew (kubectl plugin)

GoReleaser automatically updates the plugin manifest in `somaz94/krew-index`.

```bash
kubectl krew install diff2
kubectl diff2 file ./manifests/
```

**Prerequisites:**
- Create `somaz94/krew-index` repository on GitHub
- `PAT_TOKEN` secret must have write access to that repository

<br/>

### 4. Docker (Future — Phase 4)

```bash
docker run --rm \
  -v ~/.kube:/root/.kube \
  -v $(pwd):/work \
  somaz94/kube-diff:latest file /work/manifests/
```

<br/>

### 5. GitHub Action (Future — Phase 5)

```yaml
- uses: somaz94/kube-diff-action@v1
  with:
    source: helm
    chart: ./my-chart
    values: ./values-prod.yaml
```

<br/>

## Priority

| Priority | Channel | Status | Notes |
|----------|---------|--------|-------|
| 1 | GitHub Releases | Ready | GoReleaser auto-build |
| 2 | Homebrew tap | Ready | Requires `somaz94/homebrew-tap` repo |
| 3 | Krew | Ready | Requires `somaz94/krew-index` repo |
| 4 | Docker | Planned | Phase 4 |
| 5 | GitHub Action | Planned | Phase 5 |

<br/>

## Secrets Configuration

All secrets are configured **only on the `kube-diff` repository**. No secrets needed on `homebrew-tap` or `krew-index` repositories.

GoReleaser runs in the `kube-diff` release workflow and uses `PAT_TOKEN` to push to external repositories.

| Secret | Purpose | Scope |
|--------|---------|-------|
| `PAT_TOKEN` | Cross-repo write access (Homebrew tap, Krew index, changelog) | Must have write access to `somaz94/homebrew-tap` and `somaz94/krew-index` |
| `GITHUB_TOKEN` | GoReleaser release creation | Auto-provided by GitHub Actions |
| `GITLAB_TOKEN` | GitLab mirror backup | GitLab personal access token |

> **Note**: If `PAT_TOKEN` already has write access to all `somaz94` repositories, no additional secret configuration is needed.

<br/>

## GoReleaser Configuration

The `.goreleaser.yml` in the project root handles all distribution:

- **builds**: Multi-platform binary with ldflags (version, commit, date)
- **archives**: `.tar.gz` for Linux/macOS, `.zip` for Windows
- **checksum**: SHA256 checksums file
- **brews**: Homebrew formula auto-generation
- **krews**: Krew plugin manifest auto-generation
- **release**: GitHub Release with auto-generated changelog

<br/>

## Step-by-Step: First Release

1. Ensure all tests pass:
   ```bash
   make test
   make build
   ```

2. Create and push a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. GitHub Actions triggers `release.yml` → GoReleaser runs automatically

4. Verify:
   - Check [GitHub Releases](https://github.com/somaz94/kube-diff/releases) for binaries
   - Check `somaz94/homebrew-tap` for formula commit (if repo exists)
   - Check `somaz94/krew-index` for manifest update (if repo exists)

> **Tip**: If `homebrew-tap` or `krew-index` repositories don't exist yet, GoReleaser will skip those steps without failing. Create them when needed.
