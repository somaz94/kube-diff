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

- **코드/테스트 수정 후 반드시 관련 문서를 확인하고 업데이트할 것.** 새 플래그, 기능, 동작 변경이 있으면 아래 파일들을 검토:
  - `README.md` — Quick Start, 비교 테이블, Exit Codes
  - `CHANGELOG.md` — Unreleased 섹션에 변경사항 추가
  - `docs/USAGE.md` — Global Flags 테이블, Output Formats, Filtering, CI/CD 예시
  - `docs/CONFIGURATION.md` — CLI Flags 테이블, Normalized Fields
  - `docs/EXAMPLES.md` — 새 기능 관련 예시 추가
  - `CLAUDE.md` — Key Concepts, CLI Flags 테이블
  - kube-diff-action의 `action.yml`, `scripts/run.sh`, `README.md` — 새 input 반영

<br/>

## Language

- Communicate with the user in Korean.
