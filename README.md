# Recipes Monorepo

This repository contains a small recipe application built as a monorepo:

- `database`: Postgres image with the recipe schema and entrypoint.
- `backend`: Go workspace for the API, CLI, repository, and shared types.
- `agent`: Go recipe copilot service backed by Gemini and the backend CLI.
- `web`: React Router web app.
- `helm`: Kubernetes chart for Postgres, API, agent, and web.

The root `Taskfile.yml` is the main entrypoint for day-to-day commands.

## Prerequisites

Install these tools before working on the project:

- [Task](https://taskfile.dev/)
- Docker
- Go 1.26 or newer
- Node.js 20 or newer
- npm
- kubectl, Helm, and DevSpace for Kubernetes development/deploys

For agent features, copy the example environment file and set a Gemini key:

```bash
cp agent/.env.example agent/.env
```

Set `GEMINI_API_KEY` in `agent/.env`. DevSpace also reads this value when creating the `recipes-agent` Kubernetes Secret.

## Getting Started

Install dependencies:

```bash
task install
```

Start the local Postgres container in one shell:

```bash
task build:image:db
task db:up
```

Start the API and web dev servers in another shell:

```bash
task dev:local
```

Useful local commands:

```bash
task test
task build:backend
task build:agent
task build:web
task clean
```

Run `task -l` to see all available tasks.

## Container Images

Images are named from `DOCKER_USER`, which defaults to `juancavallotti`:

- `recipes-db`
- `recipes-api`
- `recipes-agent`
- `recipes-web`

Build all images locally:

```bash
task build:images
```

Push the current `TAG` for all images:

```bash
DOCKER_USER=your-dockerhub-user TAG=latest task push:images
```

Build and push release images with both a version tag and `latest`:

```bash
DOCKER_USER=your-dockerhub-user VERSION=x.y.z task release:images
```

If `VERSION` is omitted, the release image task reads the chart version from `helm/Chart.yaml`.

## Kubernetes Development

The DevSpace workflow builds images locally, deploys the Helm chart, creates the agent Secret from `GEMINI_API_KEY`, and starts file sync/dev containers.

```bash
export GEMINI_API_KEY=your-key
task dev
```

DevSpace forwards:

- API: `localhost:4000`
- Agent: `localhost:4100`
- Web: `localhost:3000`

For a deploy without file sync:

```bash
export GEMINI_API_KEY=your-key
task deploy
```

## Helm

Lint and render the chart:

```bash
task helm:lint
task helm:template
```

Package the chart into `dist/`:

```bash
task helm:package
```

Push the packaged chart to Docker Hub OCI after logging in with Helm:

```bash
helm registry login registry-1.docker.io
DOCKER_USER=your-dockerhub-user task helm:push
```

Runtime image repositories and tags live in `helm/values.yaml`. DevSpace overrides them during local Kubernetes development.

## Release Process

Releases are managed with [release-please](https://github.com/googleapis/release-please). The workflow lives at `.github/workflows/release-please.yml` and runs when changes land on `main`; it can also be started manually.

If you want CI workflows to run on release-please PRs, add a `RELEASE_PLEASE_TOKEN` secret with a personal access token. Without it, the workflow falls back to `GITHUB_TOKEN`, which can open the PR but does not trigger other workflows from that PR.

Use Conventional Commit messages for changes that should appear in the release PR and changelog:

- `fix: ...` creates a patch release.
- `feat: ...` creates a minor release.
- `feat!: ...` or a `BREAKING CHANGE:` footer creates a major release.
- `deps: ...`, `perf: ...`, `revert: ...`, `chore: ...`, and `refactor: ...` are included in the changelog sections configured in `release-please-config.json`.
- `docs: ...`, `style: ...`, `test: ...`, `build: ...`, and `ci: ...` are tracked but hidden from the changelog.

The release PR title pattern is:

```text
chore${scope}: release${component} ${version}
```

For the main branch and `recipes` component, that produces titles like `chore(main): release recipes v0.3.0`.

To cut a release, push Conventional Commit changes to `main` or run the workflow manually, then review and merge the release PR that release-please opens. The PR updates `CHANGELOG.md`, `.release-please-manifest.json`, and `helm/Chart.yaml`. After the release PR merges, the workflow runs again on `main` and creates the GitHub release/tag.

## Notes

- Release-please tracks the last released version in `.release-please-manifest.json` and updates `helm/Chart.yaml` in release PRs.
- The root release-please package uses the Go strategy for changelog generation and treats `helm/Chart.yaml` as an extra release file.
- The workflow creates release PRs and GitHub releases/tags. It does not publish container images; run `task release:images` when images should be pushed for a release.
- Local DevSpace image builds intentionally skip pushing images.
