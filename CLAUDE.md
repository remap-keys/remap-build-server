# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Remap Build Server is a Go HTTP server that builds QMK keyboard firmware on demand. It serves the [Remap](https://remap-keys.app) platform, accepting build requests from Cloud Tasks, compiling QMK firmware with user-supplied source files, and uploading the resulting binary to Cloud Storage. Deployed to Cloud Run; TLS is terminated by the Cloud Run frontend.

## Build & Development Commands

```bash
# Build the Go binary
go build -mod=readonly -v -o server

# Run all tests
go test -v ./...

# Run tests for a specific package
go test -v ./parameter/
go test -v ./build/

# Build Docker image (includes QMK firmware toolchain)
docker build -t remap-build-server .

# Run with docker-compose (local dev on port 8080)
docker-compose up
```

## Deployment

Deployed to Cloud Run (region `asia-northeast1`, service `remap-build-server`) via Cloud Build. The `cloudbuild.yaml` builds the Docker image, pushes it to Artifact Registry (`asia-northeast1-docker.pkg.dev`), and deploys to Cloud Run with memory=4Gi, cpu=2, timeout=3600s, concurrency=1.

The server listens on the port specified by the `PORT` env var (defaults to 8080). TLS is terminated by Cloud Run — the Go process always speaks plain HTTP.

## Architecture

### Request Flow

1. **`main.go`** — Single endpoint `GET /build?uid=...&taskId=...` served as plain HTTP on `$PORT` (Cloud Run fronts it with TLS)
2. **`web/parser.go`** — Extracts `uid` and `taskId` query parameters
3. **`auth/token.go`** — Validates Google OIDC JWT Bearer token (must be from the specific GCP service account)
4. **`database/firestore.go`** — Fetches task, firmware/project metadata, and source files from Firestore
5. **`parameter/parameters.go`** — Replaces `<remap name="..." />` template tags in source files with user-provided values; supports v1 (simple key-value) and v2 (with `type: "code"` for full code replacement)
6. **`build/qmk.go`** — Writes source files to the QMK firmware tree, invokes `qmk compile`, and cleans up
7. **`database/uploader.go`** — Uploads compiled firmware to Cloud Storage (`remap-b2d08.appspot.com`)

### Two Build Paths

- **Registered firmware** (`task.FirmwareId` set): Builds from keyboard owner's pre-registered source files. No build count limit.
- **Workbench firmware** (`task.ProjectId` set): Builds from user-created Workbench project files. Requires and decrements `remainingBuildCount` from user's purchase record.

### QMK Firmware Versions

The Docker image ships multiple QMK versions under `/root/versions/<version>/` (currently 0.22.14, 0.28.3, and 0.32.8). Each firmware/project specifies its target version. Keyboard source files are written into the version-specific `keyboards/` directory, compiled, then cleaned up.

Community-Modules-capable versions (0.28.3 and 0.32.8) additionally ship the [`remap-qmk-module`](https://github.com/remap-keys/remap-qmk-module) QMK Community Module pre-installed at `/root/versions/<version>/modules/remap/`. The module version is pinned in the Dockerfile via the `REMAP_QMK_MODULE_VERSION` build arg (default `v0.1.0`). Keyboards opt in by including `"modules": ["remap"]` in their `keymap.json` (and setting `VIA_ENABLE=no`); the server itself does not rewrite or validate those files.

### Firestore Schema

- `build/v1/tasks/{taskId}` — Build task status and results
- `build/v1/firmwares/{firmwareId}` — Registered firmware metadata
- `build/v1/firmwares/{firmwareId}/keyboardFiles` and `keymapFiles` — Source files
- `build/v1/projects/{projectId}` — Workbench project metadata
- `build/v1/projects/{projectId}/keyboardFiles` and `keymapFiles` — Workbench source files
- `users/v1/purchases/{uid}` — User purchase info (remaining build count)

### Key Design Decisions

- All responses return HTTP 200 even on failure, to prevent Cloud Tasks from retrying failed builds. Actual status is written to Firestore.
- Auth validates against Google's OIDC public keys fetched at runtime from `accounts.google.com/.well-known/openid-configuration`.
