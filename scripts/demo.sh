#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLES_DIR="$ROOT_DIR/examples"
NS="kube-diff-demo"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

info()  { echo -e "${CYAN}▶ $*${RESET}"; }
step()  { echo -e "\n${GREEN}═══════════════════════════════════════${RESET}"; echo -e "${GREEN}  $*${RESET}"; echo -e "${GREEN}═══════════════════════════════════════${RESET}\n"; }
warn()  { echo -e "${YELLOW}⚠ $*${RESET}"; }

# ─── Prerequisites ───────────────────────────────────────────────
command -v kube-diff >/dev/null 2>&1 || { echo "kube-diff not found. Install: brew install somaz94/tap/kube-diff"; exit 1; }
command -v kubectl >/dev/null 2>&1   || { echo "kubectl not found"; exit 1; }

# ─── Phase 1: Setup ─────────────────────────────────────────────
step "Phase 1: Deploy demo resources to cluster"

info "Creating namespace $NS"
kubectl apply -f "$EXAMPLES_DIR/file/namespace.yaml"

info "Applying manifests (ConfigMap, Deployment, Service)"
kubectl apply -f "$EXAMPLES_DIR/file/configmap.yaml"
kubectl apply -f "$EXAMPLES_DIR/file/deployment.yaml"
kubectl apply -f "$EXAMPLES_DIR/file/service.yaml"

info "Waiting for deployment to be ready..."
kubectl rollout status deploy/demo-app -n "$NS" --timeout=60s

# ─── Phase 2: No drift (identical) ──────────────────────────────
step "Phase 2: Compare — no drift expected"

info "kube-diff file examples/file/ -n $NS"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" || true
echo ""

# ─── Phase 3: Introduce drift ───────────────────────────────────
step "Phase 3: Introduce cluster drift"

info "Scaling deployment to 5 replicas (cluster-side change)"
kubectl scale deploy/demo-app --replicas=5 -n "$NS"

info "Updating ConfigMap in cluster (add DEBUG_MODE)"
kubectl patch configmap demo-config -n "$NS" --type merge -p '{"data":{"DEBUG_MODE":"true"}}'

# ─── Phase 4: Detect drift ──────────────────────────────────────
step "Phase 4: Detect drift — file mode"

info "kube-diff file examples/file/ -n $NS"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" || true
echo ""

# ─── Phase 5: Helm comparison ───────────────────────────────────
step "Phase 5: Helm chart comparison"

info "kube-diff helm examples/helm/demo-chart/ -r demo -n $NS"
kube-diff helm "$EXAMPLES_DIR/helm/demo-chart/" -r demo -n "$NS" || true
echo ""

info "kube-diff helm with drift values (replicas=3, image=1.26)"
kube-diff helm "$EXAMPLES_DIR/helm/demo-chart/" -r demo -f "$EXAMPLES_DIR/helm/values-drift.yaml" -n "$NS" || true
echo ""

# ─── Phase 6: Kustomize comparison ──────────────────────────────
step "Phase 6: Kustomize comparison"

info "kube-diff kustomize examples/kustomize/base/ -n $NS"
kube-diff kustomize "$EXAMPLES_DIR/kustomize/base/" -n "$NS" || true
echo ""

info "kube-diff kustomize with dev overlay (replicas=3, image=1.26)"
kube-diff kustomize "$EXAMPLES_DIR/kustomize/overlays/dev/" -n "$NS" || true
echo ""

# ─── Phase 7: Output formats ────────────────────────────────────
step "Phase 7: Output formats"

info "JSON output"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -o json || true
echo ""

info "Markdown output"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -o markdown || true
echo ""

info "Table output"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -o table || true
echo ""

info "Summary only"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -s || true
echo ""

# ─── Phase 8: Filtering ─────────────────────────────────────────
step "Phase 8: Kind filtering"

info "kube-diff file -n $NS -k Deployment"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -k Deployment || true
echo ""

info "kube-diff file -n $NS -k ConfigMap,Service"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -k ConfigMap,Service || true
echo ""

# ─── Phase 9: Advanced features ─────────────────────────────────
step "Phase 9: Advanced features"

info "Ignore field: --ignore-field spec.replicas (ignore replicas diff)"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" --ignore-field spec.replicas || true
echo ""

info "Context lines: -C 1 (minimal context)"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" -C 1 || true
echo ""

info "Exit code: --exit-code (always exit 0)"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" --exit-code
echo "Exit code: $?"
echo ""

info "Diff strategy: --diff-strategy last-applied"
kube-diff file "$EXAMPLES_DIR/file/" -n "$NS" --diff-strategy last-applied || true
echo ""

# ─── Done ────────────────────────────────────────────────────────
step "Demo complete!"
echo "To clean up:  make demo-clean"
echo ""
