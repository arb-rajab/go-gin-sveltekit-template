# starterapp

A full-stack starter: a Go (Gin) backend with pgx and golang-migrate, and a
SvelteKit frontend, with real password login, a single JSON error contract,
login rate limiting, Docker images, tests and CI already wired together.
Everything domain-specific has been removed; what remains is the plumbing every
project of this shape needs. Kubernetes, Terraform and Lambda examples are
included as optional folders you can delete.

## What is included

Backend (`backend/`)
- Gin server with `/health` (liveness) and `/ready` (Postgres check), graceful
  shutdown, and `slog` logging.
- Operator auth: bcrypt passwords, HMAC-signed session cookie
  (`internal/operatorauth`), `RequireOperator` and JSON-content-type (CSRF)
  middleware, per-email and per-IP failed-login limiter (`internal/operatorapi`).
- One error contract: `{"error": {"code", "message", "field"}}`
  (`operatorapi/errors.go`).
- Pagination helper: `internal/pagination` parses `?page=&per_page=` (defaults
  1 and 20, max 100, rejects bad values) and builds the list envelope
  `{items, page, per_page, total, total_pages}`; `operatorapi.PageParams(c)`
  wires it into a handler and answers 422 with the offending field.
- Agent credential helper (`internal/agentauth`) for machine-to-machine keys.
- golang-migrate migrations (`migrations/000001_create_operators`) and the
  `cmd/provision-operator` CLI that creates the first account.
- Tests against a real Postgres; `.golangci.yml`; a Dockerfile.

Frontend (`frontend/`)
- SvelteKit 2 on Svelte 5 with the node adapter. The browser only talks to
  SvelteKit; load functions and form actions call the backend with the session
  cookie (`src/lib/server/backend.ts`). Login, gated dashboard, `/health`.
- Shared components in `src/lib/components/`, each with a test:
  `Button` (loading state), `FormField` (label, hint, field errors,
  `aria-describedby`), `Alert`, `Spinner`, `EmptyState`, `Modal` (focus,
  Escape, backdrop), `Pagination` (windowed page links), `DataTable` (typed
  columns, empty and loading states). Import them from `$lib/components`.
- Root layout with skip link and design tokens (`src/app.css`, light and
  dark), and an error page (`+error.svelte`).
- Vitest with Testing Library, ESLint 10 + Prettier, `svelte-check`.

Optional modules (delete the folder to remove them)
- `backend/internal/optional/`: `livefeed` (Server-Sent Events hub), `ssrf`
  (outbound URL guard), `emailprovider` (SMTP), `pushprovider` (APNs and FCM).
  The frontend half is `src/lib/optional/` and `src/routes/(optional-sse)/`.
- `optional/k8s/`: kustomize base and production overlay.
- `optional/terraform/`: cluster-level pieces for the k8s manifests.
- `optional/serverless/example-function/`: AWS Lambda function with SAM.

Repository
- `docker-compose.yml`: Postgres, migrations, backend, frontend and a Caddy TLS
  proxy (`https://localhost:8443`).
- GitHub Actions CI with every action pinned to a commit SHA: gofmt, vet,
  golangci-lint, tests, govulncheck, migration up/down/up, frontend lint,
  type-check, build and tests, npm audit, terraform fmt, kustomize render,
  gitleaks, osv-scanner, CodeQL (Go, JavaScript/TypeScript, and the workflows).
- Dependabot, issue and PR templates, SECURITY.md, CONTRIBUTING.md, MIT LICENSE
  placeholder, empty `docs/project-memory/` scaffold.

## Getting started

1. Rename the placeholder. It must be lowercase letters and digits. Set
   `GO_OWNER` to your GitHub user or org to fix the Go module paths too:

   ```
   GO_OWNER=acme ./rename.sh acmeportal AcmePortal
   rm rename.sh
   ```

   Then run `go mod tidy` in `backend/` and in
   `optional/serverless/example-function/`.

2. Fill in the `LICENSE` copyright line and `SECURITY.md`.
3. Run the stack:

   ```
   cp .env.example .env
   # put the output of `openssl rand -base64 32` in SESSION_SIGNING_SECRET
   docker compose up --build
   echo 'a-real-password' | docker compose run --rm -T --no-deps backend provision-operator -email you@example.com
   ```

   `provision-operator` reads the password from stdin (one line). Open
   `https://localhost:8443` (the certificate is self-signed).

## Developing without Docker

Postgres 16 must be running. Tests default to
`postgres://starterapp:starterapp@localhost:5432/starterapp`; set
`DATABASE_URL` to use another database. Apply the migration first
(`migrate -path backend/migrations -database "$DATABASE_URL" up`).

```
cd backend
gofmt -l . && go vet ./... && golangci-lint run
go test ./... -race -p 1
export SESSION_SIGNING_SECRET=$(openssl rand -base64 32)
go run .
```

```
cd frontend
npm ci
npm run lint && npm run check && npm test && npm run build
BACKEND_URL=http://localhost:8080 npm run dev
```

## Adding your domain

1. Add a migration pair under `backend/migrations/` (`000002_...up.sql` and
   `.down.sql`).
2. Write handlers in a new package and register them on `routes.API` in
   `setupRouter` (`main.go`) behind `routes.Auth`. Raise errors with the
   `{error: {code, message, field}}` envelope. For lists, use
   `operatorapi.PageParams` and `pagination.New`.
3. Add every gated route to the table in
   `operatorapi/middleware_test.go` so it is proven to reject anonymous calls.
4. In the frontend, fetch from `+page.server.ts` with `backendFetch`, and build
   pages from `$lib/components`.

## Security notes

- `SESSION_SIGNING_SECRET` has no default; the backend refuses to start
  without it. Sessions are stateless, so logout only clears the cookie and a
  stolen token stays valid until it expires or the secret is rotated.
- The login limiter is in-process. With several replicas the limit is per
  replica; move it to a shared store if you scale out.
- `osv-scanner.toml` and `.gitleaksignore` ship empty. Add a dated entry only
  after reviewing a finding.

## License

MIT. Replace the placeholder copyright holder in `LICENSE`.
