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
| `--summary-only` | `-s` | `false` | Show summary only, no diff details |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `markdown` |

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

### Combined filters

```bash
kube-diff file ./manifests/ -n production -k Deployment,Service
```

<br/>

## Exit Codes

| Code | Meaning | CI Usage |
|------|---------|----------|
| `0` | No changes detected | Pipeline passes |
| `1` | Changes detected (diff exists) | Can trigger review/alert |
| `2` | Error occurred (invalid input, cluster unreachable, etc.) | Pipeline fails |

### Example: CI gate

```bash
# Fail CI if any drift is detected
kube-diff file ./manifests/ -o json
if [ $? -eq 1 ]; then
  echo "Drift detected!"
  exit 1
fi
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
