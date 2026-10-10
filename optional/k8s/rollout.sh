#!/usr/bin/env bash
# Ordered rollout for the Kubernetes target.
#
# Enforces what Compose gets from `depends_on: condition:
# service_completed_successfully` on `migrate`: migrations run to completion
# BEFORE the new backend/frontend images roll out, not concurrently. A
# Kubernetes Job cannot block a Deployment, so this script does it explicitly.
#
# Usage:
#   ./rollout.sh <namespace> <backend-image> <frontend-image>
# Example:
#   ./rollout.sh starterapp ghcr.io/OWNER/starterapp-backend:abc1234 ghcr.io/OWNER/starterapp-frontend:abc1234
#
# Prerequisites:
#   - the namespace, starterapp-app-secrets, ingress-nginx and cert-manager
#     exist (optional/terraform creates them)
#   - overlays/production/kustomization.yaml's Ingress host patch is edited
#   - the standalone `kustomize` CLI is installed (`kubectl kustomize` renders
#     but has no `edit` subcommand), plus python3 with PyYAML

set -euo pipefail

NAMESPACE="${1:?namespace required}"
BACKEND_IMAGE="${2:?backend image required, e.g. ghcr.io/OWNER/starterapp-backend:TAG}"
FRONTEND_IMAGE="${3:?frontend image required, e.g. ghcr.io/OWNER/starterapp-frontend:TAG}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OVERLAY_DIR="${SCRIPT_DIR}/overlays/production"

echo "==> Setting image tags for this rollout"
(
  cd "${OVERLAY_DIR}"
  kustomize edit set image "starterapp-backend=${BACKEND_IMAGE}" "starterapp-frontend=${FRONTEND_IMAGE}"
)

echo "==> Rendering manifests (--load-restrictor LoadRestrictionsNone is required: the kustomization references backend/migrations/*.sql from outside this directory)"
RENDERED="$(mktemp)"
trap 'rm -f "${RENDERED}"' EXIT
kubectl kustomize --load-restrictor LoadRestrictionsNone "${OVERLAY_DIR}" > "${RENDERED}"

echo "==> Phase 1: migration Job"
# A Job's pod template is immutable, so delete any prior one before applying.
kubectl -n "${NAMESPACE}" delete job starterapp-migrate --ignore-not-found
python3 - "${RENDERED}" <<'PY' | kubectl -n "${NAMESPACE}" apply -f -
import sys, yaml
docs = list(yaml.safe_load_all(open(sys.argv[1])))
for d in docs:
    if d and d.get("kind") in ("Job", "ConfigMap") and d["metadata"]["name"].startswith(("starterapp-migrate", "starterapp-migrations")):
        print("---")
        print(yaml.safe_dump(d))
PY

echo "==> Waiting for migration Job to complete (bounded: 120s)"
kubectl -n "${NAMESPACE}" wait --for=condition=complete job/starterapp-migrate --timeout=120s

echo "==> Phase 2: applying the rest of the stack"
kubectl -n "${NAMESPACE}" apply -f "${RENDERED}"

echo "==> Waiting for backend rollout"
kubectl -n "${NAMESPACE}" rollout status deployment/backend --timeout=180s

echo "==> Waiting for frontend rollout"
kubectl -n "${NAMESPACE}" rollout status deployment/frontend --timeout=180s

echo "==> Done. backend=${BACKEND_IMAGE} frontend=${FRONTEND_IMAGE}"
