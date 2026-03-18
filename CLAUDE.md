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
- **Fetcher**: Uses client-go dynamic client to get live resources from cluster
- **Normalize**: Strips cluster-managed fields (managedFields, uid, status, etc.)
- **Compare**: Generates unified diff per resource
- **Report**: Outputs color/plain/json/markdown summary

## Exit Codes

- `0`: No changes
- `1`: Changes detected
- `2`: Error

<br/>

## Language

- Communicate with the user in Korean.
