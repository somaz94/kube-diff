# Configuration

Reference for all kube-diff configuration options.

<br/>

## Table of Contents

- [CLI Flags](#cli-flags)
- [Environment Variables](#environment-variables)
- [Normalized Fields](#normalized-fields)
- [Supported Resource Types](#supported-resource-types)

<br/>

## CLI Flags

### Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--kubeconfig` | | string | `$KUBECONFIG` or `~/.kube/config` | Path to kubeconfig file |
| `--context` | | string | Current context | Kubernetes context to use |
| `--namespace` | `-n` | string | (all) | Filter by namespace |
| `--kind` | `-k` | []string | (all) | Filter by resource kind (comma-separated) |
| `--summary-only` | `-s` | bool | `false` | Show summary only, no diff details |
| `--output` | `-o` | string | `color` | Output format: `color`, `plain`, `json`, `markdown` |

### Helm-specific Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--values` | `-f` | []string | | Values files for `helm template` |
| `--release` | `-r` | string | `release` | Release name for `helm template` |

<br/>

## Environment Variables

| Variable | Description |
|----------|-------------|
| `KUBECONFIG` | Path to kubeconfig file (overridden by `--kubeconfig` flag) |

<br/>

## Normalized Fields

When comparing resources, kube-diff automatically strips cluster-managed fields to produce clean diffs. These fields are removed from both local and cluster resources before comparison:

### Metadata fields removed

| Field | Reason |
|-------|--------|
| `metadata.managedFields` | Server-side apply tracking |
| `metadata.resourceVersion` | Internal versioning |
| `metadata.uid` | Cluster-assigned unique ID |
| `metadata.creationTimestamp` | Object creation time |
| `metadata.generation` | Generation counter |
| `metadata.selfLink` | Deprecated API link |

### Annotations removed

| Annotation | Reason |
|-----------|--------|
| `kubectl.kubernetes.io/last-applied-configuration` | kubectl apply tracking |
| `deployment.kubernetes.io/revision` | Deployment rollout tracking |

### Top-level fields removed

| Field | Reason |
|-------|--------|
| `status` | Runtime state, not part of desired spec |

> **Note**: If all annotations are removed, the empty `annotations` map is also cleaned up.

<br/>

## Supported Resource Types

kube-diff uses the Kubernetes dynamic client, so it supports **all resource types** including CRDs. Resource names are automatically pluralized:

### Built-in pluralization

| Kind | Resource |
|------|----------|
| Deployment | deployments |
| Service | services |
| ConfigMap | configmaps |
| Secret | secrets |
| Pod | pods |
| Ingress | **ingresses** |
| NetworkPolicy | **networkpolicies** |
| StorageClass | **storageclasses** |
| IngressClass | **ingressclasses** |
| EndpointSlice | **endpointslices** |
| ResourceQuota | **resourcequotas** |
| PriorityClass | **priorityclasses** |
| RuntimeClass | **runtimeclasses** |

Bold entries use special pluralization rules. All other kinds follow the simple `lowercase + s` pattern.
