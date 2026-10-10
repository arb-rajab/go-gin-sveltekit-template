# Security Policy

## Supported versions

Describe which versions receive security fixes (for example, "the latest
minor release").

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Use GitHub's private vulnerability reporting (Security tab, then "Report a
vulnerability").

Please include:
- A description of the vulnerability and its potential impact
- Steps to reproduce
- Affected version or commit

## Disclosure process

1. Acknowledgement within 5 business days.
2. Assessment and severity rating.
3. Fix developed on a private branch where feasible.
4. Coordinated disclosure once a fix is released, with credit to the reporter
   unless anonymity is requested.

## Scope

This policy covers the application code in this repository, its Docker
images, and its GitHub Actions workflows. It does not cover third-party
dependencies (report those upstream) or self-hosted deployments the
maintainers do not operate.

## Security tooling in this repository

- Secret scanning: gitleaks in CI. Known false positives go in
  `.gitleaksignore`, one fingerprint per line, each with a comment saying why.
- Dependency scanning: osv-scanner in CI. Exceptions go in `osv-scanner.toml`
  as dated entries (see the format comment in that file).
- Static analysis: CodeQL for the application languages and for the GitHub
  Actions workflows.
