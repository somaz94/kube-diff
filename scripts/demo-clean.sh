#!/usr/bin/env bash
set -euo pipefail

NS="kube-diff-demo"

echo "Cleaning up demo resources..."
kubectl delete ns "$NS" --ignore-not-found
echo "Done. Namespace $NS deleted."
