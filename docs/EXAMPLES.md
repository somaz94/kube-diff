# Examples

Hands-on examples for testing kube-diff against a live cluster.

<br/>

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Demo](#quick-demo)
- [Example Files](#example-files)
- [Step-by-Step](#step-by-step)
- [Output Samples](#output-samples)

<br/>

## Prerequisites

- `kube-diff` installed (`brew install somaz94/tap/kube-diff`)
- `kubectl` configured with cluster access
- `helm` (for Helm examples)
- `kustomize` or `kubectl` (for Kustomize examples)

<br/>

## Quick Demo

Run the full demo with a single command:

```bash
make demo        # Deploy resources → compare → detect drift → all formats
make demo-clean  # Remove demo resources from cluster
```

The demo script will:
1. Create `kube-diff-demo` namespace with ConfigMap, Deployment, Service
2. Compare with no drift (all resources match)
3. Introduce cluster-side drift (scale replicas, patch ConfigMap)
4. Detect drift using file, Helm, and Kustomize modes
5. Show JSON, Markdown, and summary-only output formats
6. Demonstrate kind filtering

<br/>

## Example Files

```
examples/
├── file/                          # Plain YAML manifests
│   ├── namespace.yaml             # kube-diff-demo namespace
│   ├── configmap.yaml             # App configuration
│   ├── deployment.yaml            # nginx deployment (2 replicas)
│   └── service.yaml               # ClusterIP service
├── helm/
│   ├── demo-chart/                # Helm chart
│   │   ├── Chart.yaml
│   │   ├── values.yaml            # Default values (matches cluster)
│   │   └── templates/
│   │       ├── configmap.yaml
│   │       ├── deployment.yaml
│   │       └── service.yaml
│   └── values-drift.yaml          # Override values (intentional drift)
└── kustomize/
    ├── base/                      # Base manifests
    │   ├── kustomization.yaml
    │   ├── configmap.yaml
    │   ├── deployment.yaml
    │   └── service.yaml
    └── overlays/
        └── dev/                   # Dev overlay (replicas=3, image=1.26)
            └── kustomization.yaml
```

<br/>

## Step-by-Step

### 1. Deploy demo resources

```bash
kubectl apply -f examples/file/
```

### 2. Compare — no drift

```bash
kube-diff file examples/file/ -n kube-diff-demo
```

Expected: all resources show `✓ OK`.

### 3. Introduce drift

```bash
# Scale deployment (cluster-side change)
kubectl scale deploy/demo-app --replicas=5 -n kube-diff-demo

# Patch ConfigMap (cluster-side change)
kubectl patch configmap demo-config -n kube-diff-demo \
  --type merge -p '{"data":{"DEBUG_MODE":"true"}}'
```

### 4. Detect drift

```bash
# File mode
kube-diff file examples/file/ -n kube-diff-demo

# Helm mode
kube-diff helm examples/helm/demo-chart/ -r demo -n kube-diff-demo

# Kustomize mode
kube-diff kustomize examples/kustomize/base/ -n kube-diff-demo
```

### 5. Compare with intentional changes

```bash
# Helm with drift values (replicas=3, image=1.26, new config keys)
kube-diff helm examples/helm/demo-chart/ -r demo \
  -f examples/helm/values-drift.yaml -n kube-diff-demo

# Kustomize dev overlay (replicas=3, image=1.26)
kube-diff kustomize examples/kustomize/overlays/dev/ -n kube-diff-demo
```

### 6. Output formats

```bash
kube-diff file examples/file/ -n kube-diff-demo -o json
kube-diff file examples/file/ -n kube-diff-demo -o markdown
kube-diff file examples/file/ -n kube-diff-demo -o plain
kube-diff file examples/file/ -n kube-diff-demo -o table
kube-diff file examples/file/ -n kube-diff-demo -s  # summary only
```

### 7. Kind filtering

```bash
kube-diff file examples/file/ -n kube-diff-demo -k Deployment
kube-diff file examples/file/ -n kube-diff-demo -k ConfigMap,Service
```

### 8. Ignore fields

```bash
# Ignore a specific annotation added by CI
kube-diff file examples/file/ -n kube-diff-demo --ignore-field metadata.annotations.checksum/config

# Ignore replicas (managed by HPA)
kube-diff file examples/file/ -n kube-diff-demo --ignore-field spec.replicas
```

### 9. Context lines

```bash
# Show 5 lines of context around changes
kube-diff file examples/file/ -n kube-diff-demo -C 5

# Minimal context (1 line)
kube-diff file examples/file/ -n kube-diff-demo -C 1
```

### 10. Exit code control

```bash
# Report drift without failing (always exit 0)
kube-diff file examples/file/ -n kube-diff-demo --exit-code
```

### 11. Diff strategy

```bash
# Compare against live cluster state (default)
kube-diff file examples/file/ -n kube-diff-demo --diff-strategy live

# Compare against last-applied-configuration annotation
kube-diff file examples/file/ -n kube-diff-demo --diff-strategy last-applied
```

### 12. Watch mode

```bash
# Watch for changes and auto re-run
kube-diff watch file examples/file/ -n kube-diff-demo

# Watch with minimum interval
kube-diff watch file examples/file/ -n kube-diff-demo --interval 5s
```

### 13. Clean up

```bash
make demo-clean
# or
kubectl delete ns kube-diff-demo
```

<br/>

## Output Samples

### No drift

```
✓ OK     Namespace/kube-diff-demo
✓ OK     ConfigMap/demo-config (namespace: kube-diff-demo)
✓ OK     Deployment/demo-app (namespace: kube-diff-demo)
✓ OK     Service/demo-app (namespace: kube-diff-demo)

Summary: 4 resources — 4 unchanged
```

### Drift detected

```
✓ OK     Namespace/kube-diff-demo
~ CHANGED ConfigMap/demo-config (namespace: kube-diff-demo)
--- cluster
+++ local
@@ -2,4 +2,3 @@
 data:
   APP_ENV: production
-  DEBUG_MODE: "true"
   LOG_LEVEL: info
   MAX_CONNECTIONS: "100"

~ CHANGED Deployment/demo-app (namespace: kube-diff-demo)
--- cluster
+++ local
@@ -1,5 +1,5 @@
 spec:
-  replicas: 5
+  replicas: 2

✓ OK     Service/demo-app (namespace: kube-diff-demo)

Summary: 4 resources — 2 changed, 2 unchanged
```

### Table output

```
STATUS     KIND                 NAME                           NAMESPACE
------     ----                 ----                           ---------
OK         Namespace            kube-diff-demo                 -
CHANGED    ConfigMap            demo-config                    kube-diff-demo
CHANGED    Deployment           demo-app                       kube-diff-demo
OK         Service              demo-app                       kube-diff-demo

Total: 4 | Changed: 2 | New: 0 | Deleted: 0 | Unchanged: 2
```

### JSON output

```json
{
  "total": 4,
  "changed": 2,
  "new": 0,
  "deleted": 0,
  "unchanged": 2,
  "resources": [
    {
      "kind": "ConfigMap",
      "name": "demo-config",
      "namespace": "kube-diff-demo",
      "status": "changed"
    },
    {
      "kind": "Deployment",
      "name": "demo-app",
      "namespace": "kube-diff-demo",
      "status": "changed"
    }
  ]
}
```
