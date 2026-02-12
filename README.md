# jave-spotless

Go-based wrapper around Spotless for Java projects.

## What It Does

- Default flow (`check` / `apply`) runs only Java files changed vs `--base-branch`.
- Whole-project flow is explicit with `--all`.
- Runner parity:
  - CI (`CI` env var set): direct Maven
  - Local: Docker Compose
- Local runs default to offline mode (`--offline=true`).

## Project Structure

- `cmd/spotless`: CLI entrypoint
- `pkg/spotless`: reusable package and runner logic
- `dev/scripts/entrypoint.spotless.sh`: guardrail entrypoint for Spotless containers
- `dev/docker/Dockerfile.spotless`: Java/Maven image used by Spotless compose services
- `docker-compose.spotless.yml`: Spotless check/apply local services
- `Makefile`: compatibility wrapper targets
- `Dockerfile`: multi-stage Alpine image for the `spotless` Go binary

## CLI Usage

```bash
# Changed-files mode (default)
spotless check --repo-dir /path/to/repo --base-branch main
spotless apply --repo-dir /path/to/repo --base-branch main

# Whole-project mode
spotless check --repo-dir /path/to/repo --all
spotless apply --repo-dir /path/to/repo --all
```

Common flags:

- `--repo-dir` (default `.`)
- `--offline` (default `true`)
- `--runner` (`auto|docker|maven`)
- `--remote` (default `origin`)

Validation:

- `--all` and `--base-branch` are mutually exclusive.
- `--base-branch` is required unless `--all` is used.

## Make Targets

```bash
# Changed-files mode
make spotless-check TARGET_BASE_BRANCH=main
make spotless-apply TARGET_BASE_BRANCH=main

# Whole-project mode
make spotless-check-all
make spotless-apply-all

# Help and image
make spotless-help
make spotless-build
```

Optional variables:

- `REPO_DIR` (default `.`)
- `SPOTLESS_OFFLINE` (default `true`)

Shortcuts:

- `make spotless-main`
- `make spotless-staging`
- `make spotless-qa`
- `make spotless-apply-main`
- `make spotless-apply-staging`
- `make spotless-apply-qa`
- `make spotless-online TARGET_BASE_BRANCH=<branch>`
- `make spotless-offline TARGET_BASE_BRANCH=<branch>`
- `make spotless-apply-online TARGET_BASE_BRANCH=<branch>`
- `make spotless-apply-offline TARGET_BASE_BRANCH=<branch>`

## Docker

### Spotless formatter services (Java/Maven)

These are used for local formatting runs:

- `docker-compose.spotless.yml`
- `dev/docker/Dockerfile.spotless`

Build with:

```bash
make spotless-build
```

### Go CLI image (multi-stage, Alpine)

Build the `spotless` CLI image:

```bash
docker build -t jave-spotless:latest .
```

Example:

```bash
docker run --rm jave-spotless:latest help
```

## Troubleshooting

- If Docker Compose is missing, install `docker-compose` or `docker compose`.
- If `CODE_STYLE_ENV_ENABLED` errors appear, use Make targets (they set it automatically).
- If Maven offline mode misses dependencies, rerun with `SPOTLESS_OFFLINE=false` once.
- If `origin/<branch>` is missing locally, changed-files mode may resolve to no files; fetch remotes first when needed.
