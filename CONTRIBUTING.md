# Contributing

## Workflow

- **Branching:** trunk-based. Branch from `main` as `feat/<issue>-<slug>`,
  `fix/...`, `chore/...`, or `sec/...`.
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `docs:`, `test:`, `ci:`, `refactor:`, `chore:`, `perf:`).
- **Pull requests:** link an issue (`Closes #N`), pass CI (lint, type-check,
  tests, security scans) and use the PR template.
- Open an issue before a large PR; small fixes can go straight to a PR.

## Development setup

```
docker compose up
```

Backend (Go, `backend/go.mod`), from `backend/`: `gofmt -l .`, `go vet ./...`,
`golangci-lint run`, `go test ./... -race -p 1` (needs Postgres, see
`DATABASE_URL`).

Frontend (SvelteKit, `frontend/package.json`), from `frontend/`:
`npm run lint`, `npm run check`, `npm run build`, `npm test`.

## Reporting security issues

See [`SECURITY.md`](SECURITY.md). Do not use public issues for vulnerabilities.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
