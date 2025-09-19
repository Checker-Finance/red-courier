# Contributing to Red Courier

Thank you for considering contributing to Red Courier! We welcome contributions of all kinds — bug fixes, tests, documentation, ideas, or anything else you'd like to help with.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [How to Get Started](#how-to-get-started)
3. [Bug Reports / Feature Requests](#bug-reports--feature-requests)
4. [Pull Request Process](#pull-request-process)
5. [Coding & Style Guidelines](#coding--style-guidelines)
6. [Testing](#testing)
7. [Development Environment](#development-environment)
8. [Health Checks / Readiness](#health-checks--readiness)
9. [Issue Labels and Prioritization](#issue-labels-and-prioritization)
10. [Communication & Support](#communication--support)

---

## Code of Conduct

Please note that by participating in this project, you are expected to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md). If you see behavior that goes against it, please reach out to the maintainers.

---

## How to Get Started

- Check existing issues to see if someone else is already working on what you want to fix or add.
- If you aren’t sure where to start, issues labeled **good first issue** or **help wanted** are a great place.
- Fork the repository, create a branch, make your changes, and then submit a PR.

---

## Bug Reports & Feature Requests

**Before submitting a bug report:**

- Make sure you are using the latest version.
- Try to isolate the issue; include steps to reproduce.
- Include relevant logs, error messages, and stack traces.
- If the issue involves external dependencies (Postgres, Redis, etc.), note your version or config.

**Feature requests:**

- Describe what you want and why it would be useful.
- If possible, include examples or pseudo-code.
- Check if a similar request exists already; if so, comment there instead of creating a duplicate.

---

## Pull Request Process

1. Create a branch based off `main` (or your environment’s default branch):
```bash
git checkout -b my-feature-or-fix
```

2. Make your changes. Ensure your code builds / compiles / passes tests.
3.	Commit with a descriptive message. Use something like:
```text
feat: add health endpoint
fix: handle SIGTERM shutdown
chore: update config structure
```
4.	Open a pull request targeting main. Include:
 - Description of what has changed.
 - Reason for the change (why).
 - Any dependencies or setup required.
 - If applicable: example usage or configuration changes.

5. Wait for reviews. Address requested changes. Once approvals + CI checks pass, the PR can be merged according to the project’s branching rules.
---

## Coding & Style Guidelines
- Follow Go formatting (go fmt) and idiomatic practices.
- Keep functions reasonably short; focus on readability.
- Use context for cancellations/timeouts.
- Handle errors explicitly; don’t ignore errors unless there is a very clear reason.
- For config files / YAML, keep them consistent with existing ones.
- Write comments where code is complex or non-obvious.

--- 

## Testing
- Write unit tests for new functionality.
- Ensure existing tests pass before submitting a PR.
- Use table-driven tests where applicable.
- For integration tests, mock external dependencies when possible.

--- 

## Development Environment
- Requirements: Go version X.Y+ (adjust as needed), Postgres, Redis.
- Environment config: use provided config.example.yaml as a template.
- How to run locally:
```bash
go build ./cmd/red-courier
./red-courier --config config.example.yaml
```