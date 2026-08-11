# Keim (🌱)

[![CI](https://github.com/Petox03/Keim/actions/workflows/ci.yml/badge.svg)](https://github.com/Petox03/Keim/actions/workflows/ci.yml)
[![Version](https://img.shields.io/github/v/tag/Petox03/Keim?label=version)](https://github.com/Petox03/Keim/tags)

> Keim (German for germ, sprout or seed) is a CLI that scaffolds a reproducible Go development environment isolated in Docker, with no Go installation required on the host.

---

## What it is

Keim removes the initial friction of starting Go projects in containerized environments. It generates a reproducible, efficient, modern Docker-based development setup ready to code with a single command.

## Philosophy

- **Zero Manual Configuration:** eliminates repetitive boilerplate when starting a Go project in Docker.
- **Total Isolation:** the project runs, compiles and downloads dependencies exclusively inside Docker. You don't need Go on your machine.
- **Cache Persistence:** Go caches (`GOCACHE`, `GOMODCACHE`) persist in named volumes to avoid redundant downloads and recompilations.
- **Non-Maleficence:** Keim never destroys, overwrites or alters pre-existing files in the working directory.

## Target user

A developer learning Go who wants to experiment without installing Go locally. Comfortable with Docker as the standard tool in their social or corporate environment.

## What it generates

Keim injects 6 files into the target directory (7 with `--devcontainer`):

```
[Project Directory]
 ├── .dockerignore
 ├── .gitignore
 ├── compose.yml
 ├── Dockerfile
 ├── go.mod
 ├── main.go
 └── .devcontainer/devcontainer.json   (only with --devcontainer)
```

- `go.mod` — initializes the Go module with the folder name and detected version.
- `main.go` — basic entry point at the module root.
- `Dockerfile` — `golang:alpine` image with caches redirected to persistable volumes.
- `compose.yml` — `app` service sleeping (`sleep infinity`), bind mount for code, named volumes for cache.
- `.gitignore` — excludes temporary binaries and local cache.
- `.dockerignore` — excludes irrelevant files from the Docker build context.
- `.devcontainer/devcontainer.json` — devcontainer configuration for VS Code / compatible IDEs (only with `--devcontainer`).

## Usage

```
keim init                                # Scaffold in current directory, default cascade (host → internet → manual)
keim init my-project                     # Create folder and scaffold there
keim init --detect host my-project       # Host detection only
keim init --detect manual=1.26 my-project  # Explicit fixed version
keim init --devcontainer my-project      # Generate devcontainer configuration
```

> **Important:** `--detect` goes before the project name, not after
> (`keim init my-project --detect host` does not work).

**Available detection strategies:**

- `host` — detects Go installed on the machine.
- `internet` — queries the latest stable Go version online.
- `manual=X.Y` — explicit fixed version.
- `manual` (no version) — interactive prompt via stdin.

The default cascade is `host → internet → manual`.

After generation, the development workflow is:

```
cd my-project
docker compose up -d
docker compose exec app go run .
```

## Status

Version `0.2.1`. `keim init` is fully functional end-to-end with complete detection cascade (`host → internet → manual`), `--devcontainer` support and a Lipgloss-based visual interface.
