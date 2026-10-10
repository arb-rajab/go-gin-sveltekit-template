# Dependabot status

_Last updated: 2026-10-10. Starter template extracted from `pulsewatch`; update when the state changes._

## Configuration

- Ecosystems covered: gomod (`/backend`, `/optional/serverless/example-function`), npm (`/frontend`), docker (`/backend`, `/frontend`), terraform (`/optional/terraform`), docker-compose (`/`), github-actions (`/`).
- Grouping: minor and patch updates grouped as `minor-and-patch` everywhere except terraform.
- Schedule: weekly.
- Ignore rules: `@types/node` majors in `/frontend`; the `golang` image's major and minor (the Go version is kept in step across `go.mod`, the Dockerfile and CI by hand); `node` image majors; docker-compose image majors.

## State at last update

- Open Dependabot PRs: 0 (each merged or closed only after reading its checks).
- Default-branch CI: green at last check.
- Repository settings (read live 2026-10-10): public, `is_template` true, description set, default branch `main` (the only branch kept), secret scanning, push protection and Dependabot security updates enabled, dependency graph on (SBOM endpoint 200). Non-provider secret patterns and validity checks stay off: GitHub does not offer them on this user-owned public repo.
- Set by the repo owner through the API on 2026-10-10 (the session's proxy blocks repository-settings writes) and read back by the session the same day: private vulnerability reporting enabled (PUT 204), `delete_branch_on_merge` on (PATCH 200), repository topics set (PUT 200). Fork-PR workflow approval is `all_external_contributors` (read with the owner's token). Stale branches left over from squash-merged PRs were deleted after a fresh diff against `main` showed their content already on `main`.
- Alerts read 2026-10-10 with the repo owner's token: No Dependabot, code-scanning or secret-scanning alerts.

## Time-limited exemptions

- None.

## Notes

- Merge policy (deliberate choice by the repo owner, the same blanket policy as the application repos): every PR, major-version dependency bumps included, is squash-merged as soon as all of its required checks are green, confirmed per PR. This repo is a code showcase with no business or sensitive dependency, so green checks are the only gate. Red, pending or conflicted PRs are fixed or closed instead.
- `main` protection (2026-10-10, read back): pull request required (0 approvals), enforced for admins, force pushes and deletions blocked, "require branches to be up to date" off (`strict` false) on purpose. Nobody pushes to `main` directly; status-log updates go through a PR too.
- Required contexts: `Backend (lint, vet, tests, vuln scan, migrations)`, `Frontend (lint, type-check, build, tests)`, `Serverless example (vet, tests)`, `Terraform fmt and Kubernetes manifest build (offline)`, `Secret scanning (gitleaks)`. Nothing else is required on purpose: osv-scanner, npm audit, dependency review and CodeQL depend on a moving advisory database or run only on some events.
- Closed unmerged on 2026-10-10, each with an explanatory comment: #2 (`@sveltejs/adapter-node` 6), #3 (TypeScript 7), #4 (SvelteKit 3), #5 and #6 (the `hashicorp/helm` and `hashicorp/kubernetes` Terraform providers 3.x). If Dependabot recreates one, apply the merge policy; the closure is not re-raised. Dependabot does not re-propose a major after its PR is closed, so a later upgrade is a manual change.

## Deferred (not re-raised each pass)

- Alert reads (Dependabot, code scanning, secret scanning) and Actions settings (fork-PR approval) are blocked for Claude sessions by the proxy; they are read and written with the repo owner's token on their own machine.
- The owner keeps the token's Administration, Contents and Secret scanning permissions and does not rotate it (deliberate, 2026-10-09, reconfirmed 2026-10-10 for the template repos).
