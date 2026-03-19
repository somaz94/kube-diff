# Usage

Complete guide for using kube-diff CLI.

<br/>

## Table of Contents

- [Commands](#commands)
- [Global Flags](#global-flags)
- [File Command](#file-command)
- [Helm Command](#helm-command)
- [Kustomize Command](#kustomize-command)
- [Output Formats](#output-formats)
- [Filtering](#filtering)
- [Diff Strategy](#diff-strategy)
- [Watch Mode](#watch-mode)
- [Exit Codes](#exit-codes)
- [CI/CD Integration](#cicd-integration)

<br/>

## Commands

| Command | Description |
|---------|-------------|
| `kube-diff file <path>` | Compare plain YAML manifests against cluster |
| `kube-diff helm <chart-path>` | Compare Helm chart template output against cluster |
| `kube-diff kustomize <path>` | Compare Kustomize build output against cluster |
| `kube-diff version` | Print version information |

<br/>

## Global Flags

These flags are available on all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--kubeconfig` | | `$KUBECONFIG` or `~/.kube/config` | Path to kubeconfig file |
| `--context` | | Current context | Kubernetes context to use |
| `--namespace` | `-n` | All | Filter by namespace |
| `--kind` | `-k` | All | Filter by resource kind (comma-separated) |
| `--name` | `-N` | All | Filter by resource name (comma-separated) |
| `--selector` | `-l` | All | Filter by label selector (e.g., `app=nginx,env=prod`) |
| `--summary-only` | `-s` | `false` | Show summary only, no diff details |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `markdown`, `table` |
| `--ignore-field` | | (none) | Field paths to ignore in diff (dot notation, repeatable) |
| `--context-lines` | `-C` | `3` | Number of context lines in unified diff output |
| `--exit-code` | | `false` | Always exit 0 even when changes are detected |
| `--diff-strategy` | | `live` | Comparison strategy: `live` or `last-applied` |

<br/>

## File Command

Compare plain YAML manifests against the live cluster state.

```bash
# Single file
kube-diff file ./manifests/deployment.yaml

# Entire directory (recursive, .yaml and .yml files)
kube-diff file ./manifests/

# With namespace filter
kube-diff file ./manifests/ -n production

# With kind filter
kube-diff file ./manifests/ -k Deployment,Service
```

### Supported file types

- `.yaml` and `.yml` extensions
- Multi-document YAML (separated by `---`)
- Nested directories (recursive walk)
- Documents without `kind` are silently skipped

<br/>

## Helm Command

Compare the output of `helm template` against the live cluster state.

```bash
# Basic
kube-diff helm ./my-chart

# With values file
kube-diff helm ./my-chart --values values-prod.yaml

# Multiple values files
kube-diff helm ./my-chart -f values.yaml -f values-prod.yaml

# Custom release name
kube-diff helm ./my-chart --release my-release

# Combined with global flags
kube-diff helm ./my-chart -f values-prod.yaml -n production -k Deployment -o json
```

### Helm Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--values` | `-f` | | Values files (can specify multiple) |
| `--release` | `-r` | `release` | Release name for `helm template` |

### How it works

1. Runs `helm template <release> <chart> [-f values...]`
2. Parses the rendered YAML output
3. Fetches corresponding resources from cluster
4. Compares and generates diff report

> **Note**: Requires `helm` CLI installed and accessible in `$PATH`.

<br/>

## Kustomize Command

Compare the output of `kustomize build` against the live cluster state.

```bash
# Basic
kube-diff kustomize ./overlays/production

# With filters
kube-diff kustomize ./overlays/staging -n staging -k Deployment

# JSON output for CI
kube-diff kustomize ./overlays/production -o json
```

### How it works

1. Runs `kustomize build <path>`
2. Falls back to `kubectl kustomize <path>` if `kustomize` CLI is not installed
3. Parses the rendered YAML output
4. Fetches corresponding resources from cluster
5. Compares and generates diff report

> **Note**: Requires either `kustomize` or `kubectl` CLI installed.

<br/>

## Output Formats

### Color (default)

Human-readable output with ANSI color codes. Best for terminal use.

```bash
kube-diff file ./manifests/ -o color
```

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
✗ DELETED Secret/old-secret (namespace: production)

Summary: 4 resources — 1 changed, 1 new, 1 deleted, 1 unchanged
```

| Symbol | Color | Meaning |
|--------|-------|---------|
| `★ NEW` | Green | Resource exists locally but not in cluster |
| `~ CHANGED` | Yellow | Resource differs between local and cluster |
| `✓ OK` | Gray | Resource is identical |
| `✗ DELETED` | Red | Resource exists in cluster but not locally |

<br/>

### Plain

Same as color but without ANSI escape codes. For piping or log files.

```bash
kube-diff file ./manifests/ -o plain
```

<br/>

### JSON

Machine-readable JSON output. Best for CI/CD pipelines.

```bash
kube-diff file ./manifests/ -o json
```

```json
{
  "total": 4,
  "changed": 1,
  "new": 1,
  "deleted": 1,
  "unchanged": 1,
  "resources": [
    {
      "kind": "Deployment",
      "name": "web-app",
      "namespace": "production",
      "status": "changed"
    },
    {
      "kind": "ConfigMap",
      "name": "app-config",
      "namespace": "production",
      "status": "new"
    }
  ]
}
```

<br/>

### Markdown

Markdown-formatted output. Best for PR comments.

```bash
kube-diff file ./manifests/ -o markdown
```

<br/>

### Table

Compact tabular output showing status, kind, name, and namespace.

```bash
kube-diff file ./manifests/ -o table
```

```
STATUS     KIND                 NAME                           NAMESPACE
------     ----                 ----                           ---------
NEW        ConfigMap            app-config                     production
CHANGED    Deployment           web-app                        production
OK         Service              web-svc                        production

Total: 3 | Changed: 1 | New: 1 | Deleted: 0 | Unchanged: 1
```

<br/>

## Filtering

### By namespace

Only compare resources in a specific namespace:

```bash
kube-diff file ./manifests/ -n production
```

### By resource kind

Only compare specific resource types:

```bash
# Single kind
kube-diff file ./manifests/ -k Deployment

# Multiple kinds (comma-separated)
kube-diff file ./manifests/ -k Deployment,Service,ConfigMap
```

### By resource name

Only compare specific resources by name:

```bash
# Single name
kube-diff file ./manifests/ -N my-app

# Multiple names (comma-separated)
kube-diff file ./manifests/ -N my-app,my-config
```

### By label selector

Only compare resources with specific labels:

```bash
# Single label
kube-diff file ./manifests/ -l app=nginx

# Multiple labels (comma-separated, AND logic)
kube-diff file ./manifests/ -l app=nginx,env=prod
```

### Combined filters

```bash
kube-diff file ./manifests/ -n production -k Deployment,Service -l app=web
```

### Ignore specific fields

Exclude fields from diff comparison using dot notation:

```bash
# Ignore a specific annotation
kube-diff file ./manifests/ --ignore-field metadata.annotations.checksum/config

# Ignore multiple fields
kube-diff file ./manifests/ --ignore-field metadata.annotations.checksum --ignore-field spec.replicas

# Useful for fields that vary between environments
kube-diff helm ./chart/ --ignore-field metadata.labels.chart --ignore-field metadata.labels.heritage
```

### Custom context lines

Control the number of context lines shown around changes:

```bash
# Show 5 lines of context (default: 3)
kube-diff file ./manifests/ -C 5

# Minimal context
kube-diff file ./manifests/ -C 1
```

<br/>

## Diff Strategy

Control what kube-diff compares your local manifests against.

### Live (default)

Compares against the current live state of resources in the cluster:

```bash
kube-diff file ./manifests/ --diff-strategy live
```

### Last Applied

Compares against the `kubectl.kubernetes.io/last-applied-configuration` annotation. This shows only what changed since the last `kubectl apply`, ignoring any runtime modifications (e.g., HPA scaling, operator mutations):

```bash
kube-diff file ./manifests/ --diff-strategy last-applied
```

> **Note**: If a resource doesn't have the `last-applied-configuration` annotation (e.g., created with `kubectl create` instead of `kubectl apply`), kube-diff falls back to comparing against the live state.

<br/>

## Watch Mode

Monitor files for changes and automatically re-run kube-diff:

```bash
# Watch plain YAML manifests
kube-diff watch file ./manifests/

# Watch Helm chart with values
kube-diff watch helm ./my-chart/ -f values-prod.yaml

# Watch Kustomize overlay
kube-diff watch kustomize ./overlays/production/

# Set minimum interval between re-runs
kube-diff watch file ./manifests/ --interval 10s
```

Watch mode:
- Monitors `.yaml`, `.yml`, and `.json` files recursively
- Debounces rapid changes (500ms)
- Skips hidden directories (e.g., `.git`)
- Automatically sets `--exit-code` to prevent exiting on changes
- Press `Ctrl+C` to stop

### Watch Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `0` | Minimum interval between re-runs (e.g., `5s`, `1m`). `0` means run on every change |

<br/>

## Exit Codes

| Code | Meaning | CI Usage |
|------|---------|----------|
| `0` | No changes detected | Pipeline passes |
| `1` | Changes detected (diff exists) | Can trigger review/alert |
| `2` | Error occurred (invalid input, cluster unreachable, etc.) | Pipeline fails |

> **Tip**: Use `--exit-code` to always exit 0 even when changes are detected. This is useful in CI pipelines where you want to report drift without failing the pipeline.

### Example: CI gate

```bash
# Fail CI if any drift is detected
kube-diff file ./manifests/ -o json
if [ $? -eq 1 ]; then
  echo "Drift detected!"
  exit 1
fi

# Report drift without failing
kube-diff file ./manifests/ -o json --exit-code
```

<br/>

## CI/CD Integration

### GitHub Actions

```yaml
- name: Check for drift
  run: |
    kube-diff file ./manifests/ -o json > drift-report.json
    if [ $? -eq 1 ]; then
      echo "::warning::Cluster drift detected"
    fi

- name: Comment PR with diff
  if: github.event_name == 'pull_request'
  run: |
    DIFF=$(kube-diff file ./manifests/ -o markdown)
    gh pr comment ${{ github.event.number }} --body "$DIFF"
```

### GitLab CI

```yaml
drift-check:
  stage: validate
  script:
    - kube-diff file ./manifests/ -o json
  allow_failure: true
  artifacts:
    when: always
    paths:
      - drift-report.json
```

<br/>

## Kubeconfig

kube-diff uses the standard Kubernetes client configuration:

1. `--kubeconfig` flag (highest priority)
2. `$KUBECONFIG` environment variable
3. `~/.kube/config` (default)

### Using specific context

```bash
# List available contexts
kubectl config get-contexts

# Use a specific context
kube-diff file ./manifests/ --context prod-cluster
```

### Multiple clusters

```bash
# Compare against staging
kube-diff file ./manifests/ --context staging-cluster

# Compare against production
kube-diff file ./manifests/ --context prod-cluster
```
