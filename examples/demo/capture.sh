#!/usr/bin/env bash
# Runs the README demo for real on a throwaway kind cluster and saves each
# kube-diff output under examples/demo/captured/, which replay.sh prints back.
# Usage: make build && examples/demo/capture.sh   (needs kind, kubectl, helm, kustomize)
set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
ROOT_DIR="$(cd "${DEMO_DIR}/../.." && pwd)"
OUT="${DEMO_DIR}/captured"
CLUSTER="kube-diff-demo"
NS="kube-diff-demo"

for tool in kind kubectl helm kustomize; do
    command -v "${tool}" >/dev/null 2>&1 || { echo "${tool} not found" >&2; exit 1; }
done
[ -x "${ROOT_DIR}/kube-diff" ] || { echo "kube-diff not found; run make build" >&2; exit 1; }
if kind get clusters 2>/dev/null | grep -qx "${CLUSTER}"; then
    echo "kind cluster ${CLUSTER} already exists; delete it first" >&2
    exit 1
fi

WORK="$(mktemp -d)"
# kind would otherwise switch the current context of the user's own kubeconfig.
export KUBECONFIG="${WORK}/kubeconfig"
cleanup() {
    kind delete cluster --name "${CLUSTER}" --kubeconfig "${KUBECONFIG}" >/dev/null 2>&1 || true
    rm -rf "${WORK:?}"
}
trap cleanup EXIT

kind create cluster --name "${CLUSTER}" --kubeconfig "${KUBECONFIG}" --wait 60s

cd "${ROOT_DIR}"
kubectl apply -f examples/file/namespace.yaml
helm template demo examples/helm/demo-chart/ | kubectl apply -f -
# Drift made on the cluster side, as someone running kubectl by hand would.
kubectl scale deploy/demo-app --replicas=5 -n "${NS}"
kubectl patch configmap demo-config -n "${NS}" --type merge -p '{"data":{"DEBUG_MODE":"true"}}'
kubectl delete service demo-svc -n "${NS}"
# A Deployment the chart no longer renders, still labeled as the release's.
kubectl create deployment demo-worker --image=nginx:1.25 -n "${NS}"
kubectl label deployment demo-worker app=demo --overwrite -n "${NS}"

rm -rf "${OUT}"
mkdir -p "${OUT}"
./kube-diff helm examples/helm/demo-chart/ -r demo -n "${NS}" > "${OUT}/01-kube-diff" 2>&1 || true
./kube-diff helm examples/helm/demo-chart/ -r demo -n "${NS}" -l app=demo --deleted -o table > "${OUT}/02-kube-diff" 2>&1 || true

echo "captured into ${OUT}"
