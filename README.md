# kube-diff

[![CI](https://github.com/somaz94/kube-diff/actions/workflows/ci.yml/badge.svg)](https://github.com/somaz94/kube-diff/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Latest Tag](https://img.shields.io/github/v/tag/somaz94/kube-diff)](https://github.com/somaz94/kube-diff/tags)
[![Top Language](https://img.shields.io/github/languages/top/somaz94/kube-diff)](https://github.com/somaz94/kube-diff)

A CLI tool that compares local Kubernetes manifests (plain YAML, Helm charts, Kustomize overlays) against the actual state in your cluster, providing clear, colorized diffs with a summary report.

> For detailed documentation, see the [docs/](docs/) folder:
>
> [Usage](docs/USAGE.md) |
> [Configuration](docs/CONFIGURATION.md) |
> [Examples](docs/EXAMPLES.md) |
> [Deployment](docs/DEPLOYMENT.md) |
> [Development](docs/DEVELOPMENT.md) |
> [Use Cases](docs/USE-CASES.md)

<br/>

## Why kube-diff?

| | `kubectl diff` | `kube-diff` |
|---|---|---|
| **Input** | YAML files only | Helm / Kustomize / plain YAML |
| **Output** | Raw unified diff | Per-resource colorized diff + summary |
| **New resources** | Full content dump | **NEW** label |
| **Deleted detection** | Not supported | Detects resources only in cluster |
| **CI integration** | Exit code only | JSON / Markdown / Table report output |
| **Filtering** | None | Namespace, kind, label selector filter |
| **Field ignore** | Not supported | `--ignore-field` to exclude specific paths |
| **Context lines** | Fixed | Configurable `--context-lines` |
| **Diff strategy** | Live only | `--diff-strategy live\|last-applied` |
| **Watch mode** | Not supported | `kube-diff watch` for auto re-run on file change |

<br/>

## Quick Start

### Install

```bash
# Homebrew
brew install somaz94/tap/kube-diff

# Scoop (Windows)
scoop bucket add somaz94 https://github.com/somaz94/scoop-bucket
scoop install kube-diff

# Krew (kubectl plugin)
kubectl krew install diff2

# Binary (latest)
curl -sL https://github.com/somaz94/kube-diff/releases/latest/download/kube-diff_linux_amd64.tar.gz | tar xz
sudo mv kube-diff /usr/local/bin/

# Binary (specific version)
curl -sL https://github.com/somaz94/kube-diff/releases/download/v0.1.0/kube-diff_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv kube-diff /usr/local/bin/

# From source
go install github.com/somaz94/kube-diff/cmd@latest
```

### Upgrade

```bash
# Homebrew
brew update && brew upgrade kube-diff

# Scoop
scoop update kube-diff

# Krew
kubectl krew upgrade diff2

# From source
go install github.com/somaz94/kube-diff/cmd@latest
```

### Uninstall

```bash
# Homebrew
brew uninstall kube-diff

# Scoop
scoop uninstall kube-diff

# Krew
kubectl krew uninstall diff2

# Manual
sudo rm /usr/local/bin/kube-diff
```

### Basic Usage

```bash
# Compare YAML manifests against cluster
kube-diff file ./manifests/

# Compare Helm chart
kube-diff helm ./my-chart --values values-prod.yaml --release my-release

# Compare Kustomize overlay
kube-diff kustomize ./overlays/production

# Ignore specific fields in diff
kube-diff file ./manifests/ --ignore-field metadata.annotations.checksum

# Table output with custom context lines
kube-diff file ./manifests/ -o table -C 5

# Don't fail CI on changes
kube-diff file ./manifests/ --exit-code

# Compare against last-applied-configuration instead of live state
kube-diff file ./manifests/ --diff-strategy last-applied

# Watch mode — auto re-run on file changes
kube-diff watch file ./manifests/
kube-diff watch helm ./my-chart/ -f values.yaml
```

### Example Output

```
★ NEW    ConfigMap/app-config (namespace: production)
~ CHANGED Deployment/web-app (namespace: production)
--- cluster
+++ local
@@ -1,5 +1,5 @@
 spec:
-  replicas: 2
+  replicas: 3
✓ OK     Service/web-svc (namespace: production)

Summary: 3 resources — 1 changed, 1 new, 1 unchanged
```

<br/>

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No changes detected |
| `1` | Changes detected (use `--exit-code` to always exit 0) |
| `2` | Error occurred |

<br/>

## Project Structure

```
cmd/                    # CLI entry point & Cobra commands
internal/
  source/               # Manifest loaders (file, helm, kustomize)
  cluster/              # K8s dynamic client fetcher
  diff/                 # Normalization & unified diff
  report/               # Color/JSON/Markdown output
```

<br/>

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

<br/>

## License

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.
