# Development

Guide for building, testing, and contributing to kube-diff.

<br/>

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Build](#build)
- [Testing](#testing)
- [CI/CD Workflows](#cicd-workflows)
- [Conventions](#conventions)

<br/>

## Prerequisites

- Go 1.26+
- Make
- kubectl configured (for integration testing)
- golangci-lint (for linting)

<br/>

## Project Structure

```
.
├── cmd/
│   ├── main.go                    # Entry point
│   └── cli/
│       ├── root.go                # Root command with global flags
│       ├── file.go                # file subcommand
│       ├── helm.go                # helm subcommand
│       ├── kustomize.go           # kustomize subcommand
│       ├── version.go             # version subcommand
│       └── cli_test.go            # CLI tests
├── internal/
│   ├── source/
│   │   ├── types.go               # Resource struct & Source interface
│   │   ├── file.go                # YAML file/directory loader
│   │   ├── file_test.go           # File source tests
│   │   ├── helm.go                # helm template runner
│   │   ├── helm_test.go           # Helm source tests
│   │   ├── kustomize.go           # kustomize build runner
│   │   └── kustomize_test.go      # Kustomize source tests
│   ├── cluster/
│   │   ├── fetcher.go             # Dynamic client-go cluster fetcher
│   │   └── fetcher_test.go        # Fetcher tests (fake client)
│   ├── diff/
│   │   ├── normalize.go           # Strip cluster-managed fields
│   │   ├── normalize_test.go      # Normalization tests
│   │   ├── compare.go             # Unified diff generation
│   │   └── compare_test.go        # Comparison tests
│   └── report/
│       ├── summary.go             # Color/JSON/Markdown output
│       └── summary_test.go        # Report tests
├── docs/                          # Documentation
├── .github/
│   ├── workflows/                 # CI/CD workflows (9 files)
│   ├── dependabot.yml             # Dependency updates
│   └── release.yml                # Release note categories
├── .goreleaser.yml                # Multi-platform build + Krew + Homebrew
├── Makefile                       # Build, test, lint
├── CODEOWNERS                     # Repository ownership
└── go.mod
```

<br/>

### Key Directories

| Directory | Description |
|-----------|-------------|
| `cmd/cli/` | Cobra CLI commands and flag definitions |
| `internal/source/` | Manifest loaders: file (YAML), helm (template), kustomize (build) |
| `internal/cluster/` | Kubernetes dynamic client for fetching live resources |
| `internal/diff/` | Resource normalization and unified diff generation |
| `internal/report/` | Output formatting (color, plain, JSON, markdown) |

<br/>

## Build

```bash
make build           # Build binary → ./kube-diff
make clean           # Remove build artifacts
```

<br/>

## Testing

```bash
make test            # Run unit tests (alias)
make test-unit       # go test ./... -v -race -cover
make cover           # Generate coverage report
make cover-html      # Open coverage report in browser
make bench           # Run benchmarks
```

### Test Coverage

Current coverage: **92.6%**

| Package | Coverage |
|---------|----------|
| `cmd/cli` | 100% |
| `internal/cluster` | 96.9% |
| `internal/diff` | 91.3% |
| `internal/report` | 100% |
| `internal/source` | 88.7% |

### Test Patterns

- **Table-driven tests** for multiple scenarios (e.g., `guessResourceName`, `HasChanges`)
- **Temp directories** with `t.TempDir()` for file I/O tests
- **Fake dynamic client** (`dynamicfake.NewSimpleDynamicClient`) for cluster tests
- **Fake kubeconfig** files for `NewFetcher` tests

<br/>

## CI/CD Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `ci.yml` | push, PR, dispatch | Unit tests → Build → Version verify |
| `lint.yml` | dispatch | golangci-lint |
| `release.yml` | tag push `v*` | GoReleaser (binaries + Homebrew + Krew) |
| `changelog-generator.yml` | after release, PR merge | Auto-generate CHANGELOG.md |
| `contributors.yml` | after changelog | Auto-generate CONTRIBUTORS.md |
| `gitlab-mirror.yml` | push(main) | Backup to GitLab |
| `stale-issues.yml` | daily cron | Auto-close stale issues |
| `dependabot-auto-merge.yml` | PR (dependabot) | Auto-merge minor/patch updates |
| `issue-greeting.yml` | issue opened | Welcome message |

### Workflow Chain

```
tag push v* → Create release (GoReleaser)
                └→ Generate changelog
                      └→ Generate Contributors
```

<br/>

## Conventions

- **Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)
- **Secrets**: `PAT_TOKEN` (cross-repo ops), `GITHUB_TOKEN` (releases), `GITLAB_TOKEN` (mirror)
- **Comments**: English only in code
- **paths-ignore**: `.github/workflows/**`, `**/*.md`
