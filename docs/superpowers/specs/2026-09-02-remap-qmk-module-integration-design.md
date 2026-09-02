# Design: Integrate `remap-qmk-module` into the Build Image

- **Date**: 2026-09-02
- **Status**: Approved (design), awaiting implementation
- **Author**: Yoichiro Tanaka (with Claude Code assistance)
- **Scope**: `remap-build-server` (this repository)
- **Depends on**: [`remap-keys/remap-qmk-module`](https://github.com/remap-keys/remap-qmk-module) tag `v0.1.0` (already published)

## 1. Background & Motivation

`remap-build-server` currently builds QMK firmware for Remap by compiling
user-supplied keyboard/keymap source files against a stock QMK tree with
`VIA_ENABLE=yes` and the `-DBUILD_ON_REMAP` compile flag. Remap-specific
behavior is provided through the VIA protocol.

[`remap-qmk-module`](https://github.com/remap-keys/remap-qmk-module) is a
new QMK Community Module (introduced in QMK 0.28.0) that turns any QMK
keyboard into a self-describing Remap-compatible device without requiring
pre-registration in the Remap catalog. It vendors the VIA protocol subset
Remap needs under a new `0x80+` command namespace and streams the
keyboard's definition JSON over Raw HID. It is **mutually exclusive** with
`VIA_ENABLE=yes` (both handle `raw_hid_receive`).

We want the build server to be able to produce firmware that uses this
module, so keyboards can adopt the module-based Remap integration path.

## 2. Non-Goals

- Deprecating the existing VIA-based build path. Both paths coexist.
- Introducing a server-side flag or Firestore schema change to switch
  between VIA and module modes. Mode selection is entirely a client-side
  concern (see §4).
- Supporting `remap-qmk-module` on QMK versions without Community Modules
  (i.e. QMK 0.22.14). Attempts will fail at `qmk compile` time.
- Auto-injecting `"modules": ["remap"]` or disabling `VIA_ENABLE` on the
  server. The Remap client is responsible for shipping source files that
  are internally consistent.

## 3. Requirements Summary

- The module must be available inside the build image at
  `modules/remap/` under every QMK version tree that supports Community
  Modules (currently 0.28.3 and 0.32.8).
- The exact module revision must be reproducible across image rebuilds.
- Image build must fail fast if the module cannot be placed correctly.
- No runtime network access to GitHub during a build request.
- No changes to the Go source code, Firestore schema, or HTTP API.

## 4. Design Decisions

### 4.1 Module source

**Decision**: Pre-bake the module into the Docker image at image build
time via `git clone`.

Alternatives considered:
- Runtime `git clone` per build request — rejected: adds network
  dependency and latency to every build, and complicates caching /
  version pinning.
- Runtime clone + local cache — rejected: overkill for a module that is
  effectively static per image release.

### 4.2 Version pinning

**Decision**: Pin to an upstream git tag, exposed via a Docker `ARG`
(`REMAP_QMK_MODULE_VERSION`), defaulted to `v0.1.0`. `cloudbuild.yaml`
may override the arg on later releases.

Alternatives considered:
- Commit SHA pin — rejected in favor of tag pinning for readability.
  The upstream repository publishes tags starting with `v0.1.0`.
- Track `main` — rejected: not reproducible.

### 4.3 QMK version coverage

**Decision**: Install the module into every QMK version tree that
supports Community Modules. That is 0.28.3 and 0.32.8 today.
0.22.14 does not get the module.

Rationale: makes the module usable everywhere Community Modules exist
without a per-build flag, and follows the existing Dockerfile pattern of
explicit per-version setup. When a new QMK version is added to the image
in the future, the same pattern applies.

### 4.4 Client responsibility for `keymap.json` and `VIA_ENABLE`

**Decision**: The Remap client is the source of truth for the
`"modules": ["remap"]` entry in `keymap.json` and for setting
`VIA_ENABLE=no` in `rules.mk`. The server performs no rewriting and no
consistency validation.

Rationale: keeps the server side trivially simple. Any inconsistency
(e.g. `VIA_ENABLE=yes` alongside the module, or `"modules": ["remap"]`
on a QMK version that predates Community Modules) is caught by the QMK
toolchain itself and surfaces to the user through the existing
stderr → Firestore `task` document path.

### 4.5 Smoke test in the image

**Decision**: Add a final `RUN test -f .../modules/remap/qmk_module.json`
per Community-Modules-capable QMK version. If the module was not placed
correctly for any reason, the Docker build fails and the image is never
pushed.

## 5. Concrete Changes

### 5.1 `Dockerfile`

Add the module version arg near the top of the QMK setup block, then a
`git clone` step per Community-Modules-capable version, then a smoke
test. 0.22.14 remains unchanged.

```dockerfile
ARG REMAP_QMK_MODULE_VERSION=v0.1.0

# 0.22.14 — unchanged; Community Modules not supported here.

# 0.28.3
RUN mkdir -p /root/versions/0.28.3
RUN qmk setup --yes --home /root/versions/0.28.3 --branch 0.28.3
RUN rm -rf /root/versions/0.28.3/keyboards/*
RUN echo "{}" > /root/versions/0.28.3/data/mappings/keyboard_aliases.hjson
RUN git clone --depth 1 --branch ${REMAP_QMK_MODULE_VERSION} \
    https://github.com/remap-keys/remap-qmk-module.git \
    /root/versions/0.28.3/modules/remap

# 0.32.8
RUN mkdir -p /root/versions/0.32.8
RUN qmk setup --yes --home /root/versions/0.32.8 --branch 0.32.8
RUN rm -rf /root/versions/0.32.8/keyboards/*
RUN echo "{}" > /root/versions/0.32.8/data/mappings/keyboard_aliases.hjson
RUN git clone --depth 1 --branch ${REMAP_QMK_MODULE_VERSION} \
    https://github.com/remap-keys/remap-qmk-module.git \
    /root/versions/0.32.8/modules/remap

# Smoke test: fail the image build if the module is not in place.
RUN test -f /root/versions/0.28.3/modules/remap/qmk_module.json \
 && test -f /root/versions/0.32.8/modules/remap/qmk_module.json
```

### 5.2 `CLAUDE.md`

Add a short note under **Architecture → QMK Firmware Versions** stating
that Community-Modules-capable versions have `remap-qmk-module` installed
at `modules/remap/`, pinned via the `REMAP_QMK_MODULE_VERSION` Docker
build arg.

### 5.3 Files intentionally unchanged

- All Go source under `build/`, `main.go`, `common/`, `database/`,
  `parameter/`, `web/`, `auth/`.
- `docker-compose.yml` — picks up the updated image automatically.
- `cloudbuild.yaml` — may optionally receive a `--build-arg
  REMAP_QMK_MODULE_VERSION=...` in a later change; not required now.
- Firestore schema — no new fields.

## 6. Request/Data Flow (unchanged)

```
Client (Remap web)                Server                       QMK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Firestore write:                 GET /build?uid=&taskId=
  keymap.json contains           ────────────────────────→
    "modules": ["remap"]         Firestore fetch files
  rules.mk sets                  ↓
    VIA_ENABLE=no                build.CreateFiles()
                                 ↓
                                 build.BuildQmkFirmware()
                                 → qmk compile -kb X -km remap
                                                              ↓
                                                     QMK detects
                                                     modules/remap
                                                     → compile & link
```

## 7. Error Handling

The server needs no new error paths. Inconsistencies are caught by the
QMK toolchain and reported through the existing Firestore `task`
stdout / stderr fields.

| Bad input from client | Detected by | Result |
|-----------------------|-------------|--------|
| `modules: ["remap"]` with QMK 0.22.14 | `qmk compile` (no modules dir) | Build fails, stderr in Firestore |
| `modules: ["remap"]` + `VIA_ENABLE=yes` | `modules/remap/rules.mk` CATASTROPHIC_ERROR | Build fails, stderr in Firestore |
| `remap-qmk-module` Python generator failure | `qmk compile` | Build fails, stderr in Firestore |

## 8. Testing

- **Unit tests**: none added — no Go changes.
- **Image build-time**: the `test -f` smoke test ensures the module is
  present in both 0.28.3 and 0.32.8 before the image is pushed.
- **Local**: `docker-compose up` (or `docker build`) verifies the
  Dockerfile compiles end-to-end.
- **Manual E2E**: from a Remap client, submit one build with a
  `keymap.json` containing `"modules": ["remap"]` and confirm the
  produced firmware.

## 9. Rollout

1. Confirm `remap-keys/remap-qmk-module` tag `v0.1.0` exists (done).
2. Open a PR against `main` in `remap-build-server` with the Dockerfile
   and `CLAUDE.md` changes.
3. Merge. Cloud Build automatically builds the new image and deploys to
   Cloud Run (`asia-northeast1`, service `remap-build-server`).
4. Verify with one manual E2E build from the Remap client.

## 10. Future work (out of scope)

- Bumping `REMAP_QMK_MODULE_VERSION` when new module tags ship.
- If more Community-Modules-capable QMK versions are added, replicate the
  three-line `git clone` block. If we ever exceed ~4 versions,
  consolidate into a small install script (see the "Approach B"
  alternative that was discussed and deferred).
- Optional: expose `REMAP_QMK_MODULE_VERSION` in `cloudbuild.yaml` so
  module version bumps do not require a Dockerfile edit.
