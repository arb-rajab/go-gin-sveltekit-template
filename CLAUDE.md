# CLAUDE.md

Notes for Claude Code sessions working in this repository. Each item below
saves a round trip; update the file when you learn something else.

## Layout

- `backend/`: Go module `github.com/example/starterapp/backend`. Operator-facing
  routes live in `internal/operatorapi`, session logic in `internal/operatorauth`.
  Errors always go through `writeError`/`writeFieldError` so every response
  keeps the `{"error": {code, message, field}}` shape.
- `frontend/`: SvelteKit. The browser never calls the backend directly; load
  functions, form actions and `+server.ts` routes call it through
  `src/lib/server/backend.ts`. The backend enables no CORS on purpose.
- `optional/` and `backend/internal/optional/` are deletable modules. Check
  `main.go` and `.github/workflows/ci.yml` for references before deleting one.
- Only `.env.example` is committed. Never create or commit `.env`.

## One Go version, three places

The `go` line in `backend/go.mod`, the `golang:` tag in `backend/Dockerfile`
and `go-version` in the CI workflow must name the same minor release. CI uses
`go-version: "1.26"` with `check-latest`. Bump all three together. Dependabot is
told to leave the Dockerfile's Go minor alone for this reason.

## Backend tests need a real Postgres

Tests in `operatorapi` and `operatorauth` skip themselves when Postgres is not
reachable, so a green run with no database proves little. Apply
`backend/migrations` first, then run `go test ./... -race -p 1`. `-p 1` is
required: packages share one database and insert and delete `operators` rows,
and `provision-operator` refuses to run when any operator exists, so parallel
packages would interfere.

If a sandbox has no Docker, install Postgres 16 from OS packages and create
the role and database as the `postgres` OS user (`su postgres -c "psql ..."`),
never `-h 127.0.0.1 -U postgres`, which hangs on a password prompt.

`go` may download a newer toolchain than the one installed (GOTOOLCHAIN=auto).
That is expected when `go.mod` is ahead of the local Go.

## Frontend

- Svelte 5 runes only (`$props`, `$state`, `$derived`, snippets). No legacy
  `export let` or `<slot>`.
- Component tests use Testing Library and need `// @vitest-environment jsdom`
  on the first line of the file; server tests stay in the default node
  environment. The `svelteTesting()` Vite plugin handles cleanup and the
  browser export condition.
- Use `userEvent` (not `fireEvent`) to click a `disabled` control; `fireEvent`
  dispatches the event anyway.
- `svelte/no-navigation-without-resolve` flags `href` values not built with
  `resolve()`. `Pagination` takes an `href` callback, so it carries an
  `eslint-disable-next-line` with the reason on each link.
- ESLint's `no-undef` is off on purpose; TypeScript checks names, and DOM types
  such as `KeyboardEvent` are not runtime globals to ESLint.
- `npm run lint` also runs `prettier --check`. Run `npx prettier --write src`
  after hand-written files.
- `@types/node` must match the Node version CI and the Dockerfile use (22).

## Cookies and hostnames

The session cookie is `Secure`, so the browser must reach the frontend over
HTTPS (the compose `proxy` service, `https://localhost:8443`). Use `localhost`
consistently; cookies are scoped by hostname, not port.

## Docker in restricted sandboxes

Image pulls and `docker build` often fail behind an egress proxy (the build
runs `go mod download` and `npm ci` inside the container's own network). Do not
loop on it; rely on CI for image verification.

## Conventions

- Do not add real secrets anywhere. `SESSION_SIGNING_SECRET` has no default.
- `osv-scanner.toml` entries must be dated and carry a reason. Do not widen
  `.gitleaksignore` or `osv-scanner.toml` without the owner's sign-off.
- Add every new gated route to the unauthenticated-request table in
  `operatorapi/middleware_test.go`.
