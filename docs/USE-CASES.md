# Use Cases

Real-world scenarios where kube-diff and kube-diff-action help.

<br/>

## Table of Contents

- [CLI (kube-diff)](#cli-kube-diff)
  - [Pre-deploy drift check](#pre-deploy-drift-check)
  - [Post-incident audit](#post-incident-audit)
  - [GitOps sync verification](#gitops-sync-verification)
  - [Multi-cluster comparison](#multi-cluster-comparison)
  - [Helm upgrade preview](#helm-upgrade-preview)
  - [Kustomize overlay validation](#kustomize-overlay-validation)
- [GitHub Action (kube-diff-action)](#github-action-kube-diff-action)
  - [PR drift gate](#pr-drift-gate)
  - [Scheduled drift monitoring](#scheduled-drift-monitoring)
  - [Helm values change review](#helm-values-change-review)
  - [Multi-environment check](#multi-environment-check)

<br/>

## CLI (kube-diff)

### Pre-deploy drift check

Before applying changes, verify what actually differs from the cluster state.

```bash
# Check what will change before kubectl apply
kube-diff file ./manifests/ -n production

# Only check specific resource types
kube-diff file ./manifests/ -n production -k Deployment,Service
```

**When to use**: Before every `kubectl apply`, `helm upgrade`, or `kustomize build | kubectl apply` to avoid surprises.

<br/>

### Post-incident audit

After an incident, check if someone made manual changes to the cluster that deviate from the Git source of truth.

```bash
# Full audit of production namespace
kube-diff file ./manifests/ -n production -o json > drift-report.json

# Check if specific app was manually modified
kube-diff file ./manifests/ -n production -l app=payment-service
```

**When to use**: During incident review, when you suspect manual `kubectl edit` or `kubectl scale` was used.

<br/>

### GitOps sync verification

Verify that ArgoCD, Flux, or other GitOps tools have correctly synced manifests.

```bash
# Verify ArgoCD sync
kube-diff kustomize ./overlays/production -n production

# Quick summary — no diff details
kube-diff file ./manifests/ -n production -s
```

**When to use**: As a secondary check after GitOps sync, or when ArgoCD shows "Synced" but you want independent verification.

<br/>

### Multi-cluster comparison

Compare the same manifests against different clusters to ensure consistency.

```bash
# Check staging
kube-diff file ./manifests/ --context staging-cluster -n app

# Check production
kube-diff file ./manifests/ --context prod-cluster -n app

# Compare Helm chart across environments
kube-diff helm ./chart/ -f values-staging.yaml --context staging
kube-diff helm ./chart/ -f values-prod.yaml --context production
```

**When to use**: When maintaining multiple environments and need to verify consistency.

<br/>

### Helm upgrade preview

Preview what a Helm upgrade would change before actually running it.

```bash
# Current state vs new values
kube-diff helm ./my-chart/ -f values-prod.yaml -r my-release -n production

# Compare with updated values
kube-diff helm ./my-chart/ -f values-prod-new.yaml -r my-release -n production

# JSON output for scripting
kube-diff helm ./my-chart/ -f values.yaml -o json | jq '.resources[] | select(.status == "changed")'
```

**When to use**: Before `helm upgrade`, especially when updating values files or chart versions.

<br/>

### Kustomize overlay validation

Verify that Kustomize overlays produce the expected diff from the cluster.

```bash
# Check base manifests
kube-diff kustomize ./base/ -n production

# Check specific overlay
kube-diff kustomize ./overlays/production/ -n production

# Validate dev overlay won't break anything
kube-diff kustomize ./overlays/dev/ -n dev -k Deployment
```

**When to use**: When modifying Kustomize overlays, patches, or base resources.

<br/>

## GitHub Action (kube-diff-action)

### PR drift gate

Block PRs that would introduce drift, or show exactly what would change.

```yaml
name: Drift Check
on:
  pull_request:
    paths:
      - 'manifests/**'
      - 'helm/**'

jobs:
  check-drift:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubeconfig
        run: echo "${{ secrets.KUBECONFIG }}" | base64 -d > /tmp/kubeconfig
        env:
          KUBECONFIG: /tmp/kubeconfig

      - name: Check drift
        id: diff
        uses: somaz94/kube-diff-action@v1
        with:
          source: file
          path: ./manifests/
          namespace: production

      - name: Fail if drift
        if: steps.diff.outputs.has-changes == 'true'
        run: |
          echo "::error::Drift detected — review the PR comment for details"
          exit 1
```

**When to use**: On every PR that modifies Kubernetes manifests.

<br/>

### Scheduled drift monitoring

Run periodic drift checks and alert when manual changes are detected.

```yaml
name: Drift Monitor
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:

jobs:
  monitor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubeconfig
        run: echo "${{ secrets.KUBECONFIG }}" | base64 -d > /tmp/kubeconfig
        env:
          KUBECONFIG: /tmp/kubeconfig

      - name: Check drift
        id: diff
        uses: somaz94/kube-diff-action@v1
        with:
          source: file
          path: ./manifests/
          namespace: production
          output: json
          comment: 'false'

      - name: Alert on drift
        if: steps.diff.outputs.has-changes == 'true'
        run: |
          curl -X POST "${{ secrets.SLACK_WEBHOOK }}" \
            -H 'Content-Type: application/json' \
            -d '{"text": ":warning: Cluster drift detected in production!\n```\n${{ steps.diff.outputs.result }}\n```"}'
```

**When to use**: For continuous drift detection in production or critical namespaces.

<br/>

### Helm values change review

When PR modifies Helm values, show exactly what would change in the cluster.

```yaml
name: Helm Diff
on:
  pull_request:
    paths:
      - 'helm/values-*.yaml'

jobs:
  helm-diff:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubeconfig
        run: echo "${{ secrets.KUBECONFIG }}" | base64 -d > /tmp/kubeconfig
        env:
          KUBECONFIG: /tmp/kubeconfig

      - name: Diff production
        uses: somaz94/kube-diff-action@v1
        with:
          source: helm
          path: ./helm/my-chart/
          values: helm/values-prod.yaml
          release: my-release
          namespace: production
```

**When to use**: When Helm values files are part of your repo and modified via PRs.

<br/>

### Multi-environment check

Check multiple environments in a single workflow.

```yaml
name: Multi-env Drift
on:
  workflow_dispatch:

jobs:
  drift-check:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        env: [staging, production]
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubeconfig
        run: echo "${{ secrets[format('KUBECONFIG_{0}', matrix.env)] }}" | base64 -d > /tmp/kubeconfig
        env:
          KUBECONFIG: /tmp/kubeconfig

      - name: Check ${{ matrix.env }}
        uses: somaz94/kube-diff-action@v1
        with:
          source: kustomize
          path: ./overlays/${{ matrix.env }}/
          namespace: ${{ matrix.env }}
          output: markdown
          comment: 'false'
```

**When to use**: When managing multiple environments from a single repository.
