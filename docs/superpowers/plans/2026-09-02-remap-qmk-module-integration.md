# remap-qmk-module Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Pre-install the `remap-qmk-module` QMK Community Module into every Community-Modules-capable QMK version tree inside the build image, so that keyboards whose `keymap.json` references `"modules": ["remap"]` can be built.

**Architecture:** Docker-image-only change. The Dockerfile clones `github.com/remap-keys/remap-qmk-module` at a pinned tag into `modules/remap/` under `/root/versions/0.28.3` and `/root/versions/0.32.8`. Version 0.22.14 is untouched (Community Modules were introduced in QMK 0.28.0). A `test -f` smoke step at the end of the QMK-setup block fails the image build if placement is broken. No Go code changes, no Firestore schema changes, no changes to the HTTP API. Mode selection (VIA vs. module) is handled entirely on the Remap client side via the source files it uploads.

**Tech Stack:** Docker (Debian-bookworm-based golang:1.21 base), git, QMK CLI 0.28.3 / 0.32.8, `remap-qmk-module` v0.1.0 (upstream tag).

**Spec:** [docs/superpowers/specs/2026-09-02-remap-qmk-module-integration-design.md](../specs/2026-09-02-remap-qmk-module-integration-design.md)

## Global Constraints

- Pinned upstream module version: `REMAP_QMK_MODULE_VERSION=v0.1.0` (git tag on `remap-keys/remap-qmk-module`, commit `1fc93f6e7eea06b79309d747608d61b853d3903d`). Expose it as a Docker `ARG` so future bumps are a one-line change.
- Community-Modules-capable QMK versions in this image today: `0.28.3`, `0.32.8`. `0.22.14` does NOT get the module.
- Shallow clone only (`--depth 1 --branch <tag>`). No submodules, no runtime network access to GitHub.
- No Go source, Firestore, or HTTP API changes. If a task appears to need one, stop and re-check the spec.
- Client owns `keymap.json` (`"modules": ["remap"]`) and `rules.mk` (`VIA_ENABLE=no`). Server performs no rewriting or validation.
- Commit messages: English, single line, matching the repo's existing style (see `git log --oneline`).
- Do **not** push to remote or open a PR unless explicitly asked by the user (per project CLAUDE.md).

---

## File Structure

| File | Change | Responsibility |
|------|--------|----------------|
| `Dockerfile` | Modify | Add `REMAP_QMK_MODULE_VERSION` ARG; add `git clone` steps for 0.28.3 and 0.32.8; add end-of-QMK-setup smoke test. |
| `CLAUDE.md` | Modify | Document that Community-Modules-capable versions ship the `remap-qmk-module` at `modules/remap/`, pinned via the `REMAP_QMK_MODULE_VERSION` build arg. |

No new files. No files deleted. `cloudbuild.yaml`, `docker-compose.yml`, all Go sources, and Firestore schema are intentionally untouched.

---

## Task 1: Add `remap-qmk-module` to the Docker image

**Files:**
- Modify: `Dockerfile` (currently 53 lines; see lines 26-34 for the two Community-Modules-capable version blocks)

**Interfaces:**
- Consumes: `github.com/remap-keys/remap-qmk-module` git tag `v0.1.0` (external — verified live via `gh api repos/remap-keys/remap-qmk-module/git/refs/tags/v0.1.0`).
- Produces: An image where `/root/versions/0.28.3/modules/remap/qmk_module.json` and `/root/versions/0.32.8/modules/remap/qmk_module.json` both exist. Task 2 references this location in `CLAUDE.md`.

**Test strategy note:** There is no unit-test surface for a Dockerfile change. The verification is a full `docker build`, whose success depends on the two `test -f` smoke lines added at the end of the QMK-setup block. To keep the feedback loop tight we introduce those smoke lines *before* the `git clone` lines are added: build → observe failure → add the `git clone` lines → build → observe success.

- [ ] **Step 1: Introduce the ARG and the smoke test (expected to fail the build)**

Edit `Dockerfile`. After the existing line:

```dockerfile
RUN python3 -m pip install --user qmk
```

insert one blank line and then:

```dockerfile
ARG REMAP_QMK_MODULE_VERSION=v0.1.0
```

Then, immediately before the `COPY go.* ./` line, insert:

```dockerfile
# Smoke test: fail the image build if remap-qmk-module is not in place
# for every Community-Modules-capable QMK version.
RUN test -f /root/versions/0.28.3/modules/remap/qmk_module.json \
 && test -f /root/versions/0.32.8/modules/remap/qmk_module.json
```

Do NOT add the `git clone` lines yet — this step deliberately leaves the smoke test as our failing test.

- [ ] **Step 2: Run `docker build` and confirm it fails at the smoke test**

Run:

```bash
docker build -t remap-build-server:plan-t1-step2 .
```

Expected: the build proceeds until the new `RUN test -f ...` line and fails there (non-zero exit code, message about missing file). This confirms the smoke test is wired up and would catch a broken placement.

If the build fails *earlier* (e.g. syntax error in the Dockerfile) or *later* (e.g. the smoke test passes when it should not), stop and fix the Dockerfile before proceeding.

- [ ] **Step 3: Add the `git clone` lines for both QMK versions**

Edit `Dockerfile`. In the `0.28.3` block, after the line:

```dockerfile
RUN echo "{}" > /root/versions/0.28.3/data/mappings/keyboard_aliases.hjson
```

append:

```dockerfile
RUN git clone --depth 1 --branch ${REMAP_QMK_MODULE_VERSION} \
    https://github.com/remap-keys/remap-qmk-module.git \
    /root/versions/0.28.3/modules/remap
```

In the `0.32.8` block, after the line:

```dockerfile
RUN echo "{}" > /root/versions/0.32.8/data/mappings/keyboard_aliases.hjson
```

append:

```dockerfile
RUN git clone --depth 1 --branch ${REMAP_QMK_MODULE_VERSION} \
    https://github.com/remap-keys/remap-qmk-module.git \
    /root/versions/0.32.8/modules/remap
```

Do NOT touch the `0.22.14` block — Community Modules are unsupported there and the module must not be placed under it.

- [ ] **Step 4: Run `docker build` and confirm it succeeds**

Run:

```bash
docker build -t remap-build-server:plan-t1-step4 .
```

Expected: build succeeds through every step including the smoke test. Total wall time is dominated by the pre-existing `qmk setup` layers (~10–20 min from cold cache; a few seconds on incremental rebuild).

- [ ] **Step 5: Inspect the resulting image and confirm module layout**

Run:

```bash
docker run --rm remap-build-server:plan-t1-step4 \
    bash -c 'ls /root/versions/0.28.3/modules/remap && \
             ls /root/versions/0.32.8/modules/remap && \
             ! test -e /root/versions/0.22.14/modules/remap'
```

Expected: both `ls` commands list at least `qmk_module.json`, `remap.c`, `remap.h`, `rules.mk`; the final negated `test -e` confirms 0.22.14 does not have the module (exits 0).

- [ ] **Step 6: Verify the pinned tag actually landed**

Run:

```bash
docker run --rm remap-build-server:plan-t1-step4 \
    bash -c 'cd /root/versions/0.28.3/modules/remap && \
             git describe --tags --exact-match HEAD'
```

Expected output: `v0.1.0`.

- [ ] **Step 7: Verify the final Dockerfile diff is exactly what the spec calls for**

Run:

```bash
git diff Dockerfile
```

Expected: one `ARG REMAP_QMK_MODULE_VERSION=v0.1.0` line added; two `RUN git clone --depth 1 --branch ${REMAP_QMK_MODULE_VERSION} ...` blocks added (one under each of 0.28.3 and 0.32.8); one final `RUN test -f ... && test -f ...` smoke block. No other changes. No changes to the `0.22.14` block.

- [ ] **Step 8: Commit**

Run:

```bash
git add Dockerfile
git commit -m "Pre-install remap-qmk-module into Community Modules-capable QMK versions."
```

---

## Task 2: Document the module in `CLAUDE.md`

**Files:**
- Modify: `CLAUDE.md` (project instructions checked into the repo)

**Interfaces:**
- Consumes: image layout produced by Task 1 (`modules/remap/` under 0.28.3 and 0.32.8; pinned via `REMAP_QMK_MODULE_VERSION` build arg).
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Add a paragraph under the "QMK Firmware Versions" heading**

In `CLAUDE.md`, locate the section:

```markdown
### QMK Firmware Versions

The Docker image ships multiple QMK versions under `/root/versions/<version>/` (currently 0.22.14, 0.28.3, and 0.32.8). Each firmware/project specifies its target version. Keyboard source files are written into the version-specific `keyboards/` directory, compiled, then cleaned up.
```

Append the following new paragraph immediately after that existing paragraph, inside the same section (do not touch the existing paragraph):

```markdown
Community-Modules-capable versions (0.28.3 and 0.32.8) additionally ship the [`remap-qmk-module`](https://github.com/remap-keys/remap-qmk-module) QMK Community Module pre-installed at `/root/versions/<version>/modules/remap/`. The module version is pinned in the Dockerfile via the `REMAP_QMK_MODULE_VERSION` build arg (default `v0.1.0`). Keyboards opt in by including `"modules": ["remap"]` in their `keymap.json` (and setting `VIA_ENABLE=no`); the server itself does not rewrite or validate those files.
```

- [ ] **Step 2: Verify the diff**

Run:

```bash
git diff CLAUDE.md
```

Expected: one paragraph added under `### QMK Firmware Versions`. No other lines changed.

- [ ] **Step 3: Commit**

Run:

```bash
git add CLAUDE.md
git commit -m "Document remap-qmk-module pre-installation in CLAUDE.md."
```

---

## Task 3: End-to-end sanity check with `docker-compose`

**Files:**
- Read-only: `docker-compose.yml`

**Interfaces:**
- Consumes: Task 1's image.
- Produces: nothing.

**Purpose:** Final smoke check that the composed local dev environment still boots against the new image before handing the branch back to the user. No code changes, no commit.

- [ ] **Step 1: Bring the container up**

Run:

```bash
docker-compose up --build -d
```

Expected: the image rebuilds (fast — Task 1's layers are cached), the container starts, and `docker-compose ps` shows it as `Up`.

- [ ] **Step 2: Confirm the server is listening**

Run:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost:8080/build
```

Expected: `200` (the endpoint returns 200 even for missing query params — see `main.go:104-110`). Any connection error means the container did not start.

- [ ] **Step 3: Tear down**

Run:

```bash
docker-compose down
```

- [ ] **Step 4: (No commit — this task is verification only)**

If the branch is being finalized, stop here and hand back to the user. The user's project CLAUDE.md forbids pushing / opening PRs without an explicit ask.

---

## Self-Review Notes

- **Spec coverage:** every section of the spec (§4.1 pre-bake source, §4.2 tag pinning via ARG, §4.3 0.28.3+0.32.8 only, §4.4 client owns keymap.json / VIA_ENABLE, §4.5 smoke test) maps to a concrete step in Task 1 or Task 2. §7 error-handling table needs no task (all failure paths are already covered by the existing server behavior — that is the finding, not a to-do). §9 rollout step 4 (manual E2E from the Remap client) requires the client and is out of scope for this repo's plan; Task 3 is the closest server-side proxy for it.
- **Placeholder scan:** no `TBD`, no `TODO`, no "similar to Task N", no "add error handling" without code. All code blocks contain the exact strings to add.
- **Type consistency:** the ARG name `REMAP_QMK_MODULE_VERSION` and the paths `/root/versions/0.28.3/modules/remap` / `/root/versions/0.32.8/modules/remap` and the tag `v0.1.0` appear identically across Task 1, Task 2, and the spec.
