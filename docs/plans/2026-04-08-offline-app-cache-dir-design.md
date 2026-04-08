# Offline App Cache Dir Design

**Problem**

`proton-cli offline-package app export` always exports images into a temporary OCI layout under a temporary work directory and removes it at the end of the run. Repeated exports therefore re-pull image layers even when the same images were already exported locally.

**Goal**

Add `--cache-dir` to `offline-package app export` so callers can persist the OCI image layout across runs and let repeated exports reuse already downloaded image blobs.

**Non-goals**

- Do not change `-o/--output` semantics.
- Do not add chart caching in this change.
- Do not change import behavior.

**Recommended Approach**

Add a new optional flag, `--cache-dir`, that only affects the images OCI layout path used during export. When unset, the command keeps using the existing temporary `images/` directory. When set, image export writes into `<cache-dir>/images`, while the rest of the export continues to use the per-run temporary work directory.

This keeps the behavior surface small:

- current users see no change
- image blob reuse becomes opt-in
- tar output assembly still works from the normal export work directory

**Data Flow**

1. Parse `--cache-dir`.
2. Create the normal temporary work directory.
3. Use temporary `charts/`.
4. Use:
   - `<tmp>/images` when `--cache-dir` is empty
   - `<cache-dir>/images` when `--cache-dir` is set
5. Export images into the chosen OCI layout.
6. Package `manifest.yaml`, `charts/`, and the chosen `images/` layout into the output tar file.

**Behavior Details**

- `--cache-dir` is treated as a directory root owned by the feature.
- The actual OCI layout location is `<cache-dir>/images` so the package layout stays aligned with the existing `images/` convention.
- The command should create the cache directory tree when needed.
- Relative cache paths should resolve consistently via `filepath.Abs`, same style as other path resolution in the package.

**Why This Should Reuse Layers**

The export path already uses `oras-go` with a local OCI store. ORAS checks whether the target descriptor already exists and skips copies for existing nodes, while the OCI storage is content-addressed by digest. Persisting the same OCI layout directory is therefore enough to let repeated runs reuse previously downloaded blobs.

**Risks**

- Multi-platform exports may still need follow-up work if repeated tagging/index reuse exposes existing edge cases. This change is intentionally scoped to adding the cache path first.
- Reusing a persistent OCI layout means the local cache may accumulate blobs over time. That is acceptable for an explicit cache directory.

**Testing Strategy**

- Add unit tests for image directory selection with and without `--cache-dir`.
- Keep tests local to `cmd/proton-cli/cmd/offline_package`.
- Run the focused offline package test suite after implementation.
