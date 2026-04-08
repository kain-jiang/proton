# Offline App Cache Dir Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `--cache-dir` to `proton-cli offline-package app export` so repeated exports can reuse a persistent OCI image layout.

**Architecture:** Keep the current export flow and temp work directory, but route image export through a selectable image layout root. When `--cache-dir` is present, use `<cache-dir>/images`; otherwise preserve the current temporary `images/` behavior.

**Tech Stack:** Go, Cobra, ORAS OCI layout, Go unit tests

---

### Task 1: Add failing tests for image cache directory selection

**Files:**
- Modify: `cmd/proton-cli/cmd/offline_package/app_test.go`
- Modify: `cmd/proton-cli/cmd/offline_package/app_export.go`

**Step 1: Write the failing test**

Add tests that cover:

- default image directory remains `<workdir>/images`
- configured cache directory resolves to `<cache-dir>/images`

Prefer testing a small helper that computes the export image directory rather than driving the whole export command.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/proton-cli/cmd/offline_package -run 'TestResolveAppExportImagesDir'`

Expected: FAIL because the helper does not exist yet.

**Step 3: Write minimal implementation**

Introduce the helper with the smallest logic needed to satisfy the tests.

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/proton-cli/cmd/offline_package -run 'TestResolveAppExportImagesDir'`

Expected: PASS

**Step 5: Commit**

```bash
git add cmd/proton-cli/cmd/offline_package/app_export.go cmd/proton-cli/cmd/offline_package/app_test.go
git commit -m "test: cover app export image cache dir selection"
```

### Task 2: Wire `--cache-dir` into app export

**Files:**
- Modify: `cmd/proton-cli/cmd/offline_package/app_export.go`
- Test: `cmd/proton-cli/cmd/offline_package/app_test.go`

**Step 1: Write the failing test**

Add a test that exercises the new option structure and verifies a configured cache directory is turned into the persistent `images` OCI layout path.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/proton-cli/cmd/offline_package -run 'TestResolveAppExportImagesDir|TestAppExportOptions'`

Expected: FAIL because `appExportOptions` does not expose `cacheDir` and the command does not use it.

**Step 3: Write minimal implementation**

- Add `cacheDir string` to `appExportOptions`
- Add the `--cache-dir` flag
- Use the new helper in `runAppExport()`
- Ensure the chosen image directory exists before export

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/proton-cli/cmd/offline_package -run 'TestResolveAppExportImagesDir|TestAppExportOptions'`

Expected: PASS

**Step 5: Commit**

```bash
git add cmd/proton-cli/cmd/offline_package/app_export.go cmd/proton-cli/cmd/offline_package/app_test.go
git commit -m "feat: add cache dir support to app export"
```

### Task 3: Verify the offline package test slice stays green

**Files:**
- Test: `cmd/proton-cli/cmd/offline_package/app_test.go`

**Step 1: Run focused package tests**

Run: `go test ./cmd/proton-cli/cmd/offline_package`

Expected: PASS

**Step 2: Spot-check command help**

Run: `go test ./cmd/proton-cli/cmd/offline_package -run TestResolveAppExportImagesDir -v`

Expected: PASS and no regressions in the new path logic.

**Step 3: Commit**

```bash
git add cmd/proton-cli/cmd/offline_package/app_export.go cmd/proton-cli/cmd/offline_package/app_test.go
git commit -m "test: verify offline app export cache dir flow"
```
