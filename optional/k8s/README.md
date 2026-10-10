# Optional module: Kubernetes (kustomize)

A kustomize base plus a `production` overlay for running the stack on a
Kubernetes cluster you control: Postgres (StatefulSet), a migration Job,
backend, frontend, an Ingress with cert-manager TLS, and a NetworkPolicy that
limits Postgres to the backend and the migration Job. Delete this directory if
you deploy another way.

Nothing here has been applied to a live cluster by the template authors. The
manifests render offline (`kubectl kustomize optional/k8s/overlays/production`)
and CI checks that; treat the first real apply as the real test.

## Layout

- `base/`: workloads, services, ingress, network policy. No `namespace:` field;
  `rollout.sh` passes `-n <namespace>` to every kubectl call, so the namespace
  name lives in one place (`optional/terraform` or your own tooling).
- `overlays/production/`: image names and the Ingress hostname. The committed
  values are placeholders (`ghcr.io/OWNER/...`, `starterapp.example.com`); edit
  them before applying.
- `rollout.sh`: ordered rollout. It runs the migration Job to completion
  before touching the Deployments, because a Kubernetes Job cannot block a
  Deployment the way Compose's `service_completed_successfully` does.

## Prerequisites

- A cluster with an ingress-nginx controller and cert-manager, and a
  ClusterIssuer named `starterapp-letsencrypt` (`optional/terraform` creates all
  three).
- A Secret named `starterapp-app-secrets` in the target namespace with
  `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DATABASE_URL` and
  `SESSION_SIGNING_SECRET` (`optional/terraform` creates it from variables, or
  create it yourself with `kubectl create secret generic`).
- The standalone `kustomize` CLI (for `kustomize edit set image`), `kubectl`,
  and Python 3 with PyYAML (used by `rollout.sh` to split out the migration Job).

## Roll out

```
./optional/k8s/rollout.sh <namespace> ghcr.io/OWNER/starterapp-backend:TAG ghcr.io/OWNER/starterapp-frontend:TAG
```

## Keeping migrations in sync

`base/kustomization.yaml` lists every migration file explicitly in the
`starterapp-migrations` ConfigMap. Add a line there whenever you add a
migration pair to `backend/migrations/`.

The kustomization references `backend/migrations/` from outside this directory,
so rendering needs `--load-restrictor LoadRestrictionsNone`.
