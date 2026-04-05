# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Remap Build Server is a Go HTTP server that builds QMK keyboard firmware on demand. It serves the [Remap](https://remap-keys.app) platform, accepting build requests via HTTPS, compiling QMK firmware with user-supplied source files, and uploading the resulting binary to Cloud Storage.

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

Deployed to GCP via Cloud Build. The `cloudbuild.yaml` builds and pushes a Docker image to Artifact Registry (`asia-northeast1-docker.pkg.dev`).

## Architecture

### Request Flow

1. **`main.go`** — Single HTTPS endpoint `GET /build?uid=...&taskId=...` with autocert TLS (domain: `build.remap-keys.app`)
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

The Docker image ships multiple QMK versions under `/root/versions/<version>/` (currently 0.22.14 and 0.28.3). Each firmware/project specifies its target version. Keyboard source files are written into the version-specific `keyboards/` directory, compiled, then cleaned up.

### Firestore Schema

- `build/v1/tasks/{taskId}` — Build task status and results
- `build/v1/firmwares/{firmwareId}` — Registered firmware metadata
- `build/v1/firmwares/{firmwareId}/keyboardFiles` and `keymapFiles` — Source files
- `build/v1/projects/{projectId}` — Workbench project metadata
- `build/v1/projects/{projectId}/keyboardFiles` and `keymapFiles` — Workbench source files
- `users/v1/purchases/{uid}` — User purchase info (remaining build count)
- `certificates/` — TLS certificate cache for autocert

### Key Design Decisions

- All responses return HTTP 200 even on failure, to prevent Cloud Tasks from retrying failed builds. Actual status is written to Firestore.
- TLS certificates are cached in Firestore (`web/certificate.go` implements `autocert.Cache`).
- Auth validates against Google's OIDC public keys fetched at runtime from `accounts.google.com/.well-known/openid-configuration`.
