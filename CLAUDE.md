# CLAUDE.md - kube-diff

CLI tool to compare local Kubernetes manifests against live cluster state.

<br/>

## Commit Guidelines

- Do not include `Co-Authored-By` lines in commit messages.
- Use Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)

<br/>

## Project Structure

```
cmd/main.go                  # Entry point
cmd/cli/                     # Cobra CLI (root, file, helm, kustomize, version)
internal/source/             # Manifest loaders (file, helm, kustomize)
internal/cluster/            # K8s dynamic client fetcher
internal/diff/               # Normalization + unified diff
internal/report/             # Color/JSON/Markdown output
Makefile                     # Build, test, lint, cover, bench
.goreleaser.yml              # Multi-platform build + Krew
```

## Build & Test

```bash
make build           # Build binary
make test            # Run unit tests (alias for test-unit)
make test-unit       # go test ./... -v -race -cover
make cover           # Generate coverage report
make cover-html      # Open coverage in browser
make bench           # Run benchmarks
make lint            # golangci-lint
make fmt             # go fmt
make vet             # go vet
```

<br/>

## Key Concepts

- **Source**: Loads manifests from file/helm/kustomize
- **Fetcher**: Uses client-go dynamic client to get live resources from cluster (`ResourceFetcher` interface)
- **Normalize**: Strips cluster-managed fields (managedFields, uid, status, etc.)
- **RemoveFields**: Removes user-specified field paths via `--ignore-field` (dot notation)
- **Compare**: Generates unified diff per resource, accepts `CompareOptions` for context lines, ignore fields, and diff strategy
- **ExtractLastApplied**: Parses `kubectl.kubernetes.io/last-applied-configuration` annotation for `--diff-strategy last-applied`
- **Report**: Outputs color/plain/json/markdown/table summary
- **Watch**: fsnotify-based file watcher for auto re-run on changes

## CLI Flags

### Global Flags
| Flag | Short | Description |
|------|-------|-------------|
| `--kubeconfig` | | Path to kubeconfig file |
| `--context` | | Kubernetes context to use |
| `--namespace` | `-n` | Filter by namespace |
| `--kind` | `-k` | Filter by resource kind |
| `--selector` | `-l` | Filter by label selector |
| `--summary-only` | `-s` | Show summary only |
| `--output` | `-o` | Output format: color, plain, json, markdown, table |
| `--ignore-field` | | Field paths to ignore in diff (dot notation, repeatable) |
| `--context-lines` | `-C` | Number of context lines in diff (default: 3) |
| `--exit-code` | | Always exit 0 even with changes |
| `--diff-strategy` | | Comparison strategy: live or last-applied |

### Helm Flags
| Flag | Short | Description |
|------|-------|-------------|
| `--values` | `-f` | Values files (repeatable) |
| `--release` | `-r` | Release name (default: release) |

## Exit Codes

- `0`: No changes (or always 0 with `--exit-code`)
- `1`: Changes detected
- `2`: Error

<br/>

## Important Rules

- **After modifying code or tests, always review and update the related documentation.** If there are new flags, features, or behavior changes, check the following files:
  - `README.md` — Quick Start, comparison table, Exit Codes
  - `CHANGELOG.md` — Add changes to the Unreleased section
  - `docs/USAGE.md` — Global Flags table, Output Formats, Filtering, CI/CD examples
  - `docs/CONFIGURATION.md` — CLI Flags table, Normalized Fields
  - `docs/EXAMPLES.md` — Add examples for new features
  - `CLAUDE.md` — Key Concepts, CLI Flags table
  - kube-diff-action's `action.yml`, `scripts/run.sh`, `README.md` — Reflect new inputs

<br/>

## Language

- Communicate with the user in Korean.
