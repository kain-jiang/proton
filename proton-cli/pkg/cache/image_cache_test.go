package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"oras.land/oras-go/v2/registry"
)

func TestNewImageCache(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(
		WithCacheDir(tmpDir),
		WithCacheEnabled(true),
	)
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if cache.cacheDir != tmpDir {
		t.Errorf("expected cache dir %s, got %s", tmpDir, cache.cacheDir)
	}

	if !cache.enableCache {
		t.Error("expected cache to be enabled")
	}
}

func TestNewImageCacheDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(
		WithCacheDir(tmpDir),
		WithCacheEnabled(false),
	)
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if cache.enableCache {
		t.Error("expected cache to be disabled")
	}
}

func TestNewImageCacheWithXDGEnv(t *testing.T) {
	tmpDir := t.TempDir()

	original := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if original != "" {
			os.Setenv("XDG_CACHE_HOME", original)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tmpDir)

	cache, err := NewImageCache()
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	expectedDir := filepath.Join(tmpDir, CacheDirName, ImageCacheSubDir)
	if cache.cacheDir != expectedDir {
		t.Errorf("expected cache dir %s, got %s", expectedDir, cache.cacheDir)
	}
}

func TestNewImageCacheWithUserHomeDir(t *testing.T) {
	original := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if original != "" {
			os.Setenv("XDG_CACHE_HOME", original)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Unsetenv("XDG_CACHE_HOME")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get user home dir")
	}

	expectedDir := filepath.Join(home, ".cache", CacheDirName, ImageCacheSubDir)

	cache, err := NewImageCache()
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if cache.cacheDir != expectedDir {
		t.Errorf("expected cache dir %s, got %s", expectedDir, cache.cacheDir)
	}
}

func TestGetDefaultCacheDir(t *testing.T) {
	tests := []struct {
		name        string
		xdgHome     string
		wantContain string
	}{
		{
			name:        "with XDG_CACHE_HOME",
			xdgHome:     "/custom/cache",
			wantContain: "/custom/cache/proton-cli/image-cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("XDG_CACHE_HOME")
			defer func() {
				if original != "" {
					os.Setenv("XDG_CACHE_HOME", original)
				} else {
					os.Unsetenv("XDG_CACHE_HOME")
				}
			}()

			if tt.xdgHome != "" {
				os.Setenv("XDG_CACHE_HOME", tt.xdgHome)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}

			got := getDefaultCacheDir()
			if got != tt.wantContain {
				t.Errorf("getDefaultCacheDir() = %v, want contain %v", got, tt.wantContain)
			}
		})
	}
}

func TestCacheDir(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if cache.CacheDir() != tmpDir {
		t.Errorf("CacheDir() = %s, want %s", cache.CacheDir(), tmpDir)
	}
}

func TestIsEnabled(t *testing.T) {
	cache, err := NewImageCache(WithCacheEnabled(true))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if !cache.IsEnabled() {
		t.Error("IsEnabled() = false, want true")
	}

	cacheDisabled, err := NewImageCache(WithCacheEnabled(false))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cacheDisabled.Close()

	if cacheDisabled.IsEnabled() {
		t.Error("IsEnabled() = true, want false")
	}
}

func TestCacheMiss(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	store, err := cache.Get(ctx, "test.example.com/image:latest", "amd64", nil)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
	if store != nil {
		t.Error("expected nil store on cache miss")
	}
}

func TestFindAndAddEntry(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	entry := CacheEntry{
		Reference:    "test.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   filepath.Join(tmpDir, "test-layout"),
	}

	cache.addEntry(entry)

	found := cache.findEntry("test.example.com/image:latest", "amd64")
	if found == nil {
		t.Fatal("expected to find entry")
	}
	if found.Reference != entry.Reference {
		t.Errorf("expected reference %s, got %s", entry.Reference, found.Reference)
	}
	if found.Digest != entry.Digest {
		t.Errorf("expected digest %s, got %s", entry.Digest, found.Digest)
	}
}

func TestRemoveEntry(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	layoutPath := filepath.Join(tmpDir, "test-layout")
	if err := os.MkdirAll(layoutPath, 0755); err != nil {
		t.Fatalf("failed to create layout dir: %v", err)
	}

	entry := CacheEntry{
		Reference:    "test.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   layoutPath,
	}

	cache.addEntry(entry)

	if !cache.Exists("test.example.com/image:latest", "amd64") {
		t.Error("expected entry to exist")
	}

	cache.removeEntry("test.example.com/image:latest", "amd64")

	if cache.Exists("test.example.com/image:latest", "amd64") {
		t.Error("expected entry to be removed")
	}

	if _, err := os.Stat(layoutPath); !os.IsNotExist(err) {
		t.Error("expected layout directory to be removed")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	layoutPath1 := filepath.Join(tmpDir, "layout1")
	layoutPath2 := filepath.Join(tmpDir, "layout2")

	if err := os.MkdirAll(layoutPath1, 0755); err != nil {
		t.Fatalf("failed to create layout dir: %v", err)
	}
	if err := os.MkdirAll(layoutPath2, 0755); err != nil {
		t.Fatalf("failed to create layout dir: %v", err)
	}

	entry1 := CacheEntry{
		Reference:    "test1.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   layoutPath1,
	}
	entry2 := CacheEntry{
		Reference:    "test2.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:def456",
		CachedAt:     time.Now(),
		Size:         2048,
		LayoutPath:   layoutPath2,
	}

	cache.addEntry(entry1)
	cache.addEntry(entry2)

	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if len(cache.List()) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(cache.List()))
	}
}

func TestStats(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	entry := CacheEntry{
		Reference:    "test.example.com/image:latest",
		Architecture: runtimeGOARCH(),
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   filepath.Join(tmpDir, "layout"),
	}
	cache.addEntry(entry)

	stats := cache.Stats()

	if stats.TotalImages != 1 {
		t.Errorf("expected TotalImages = 1, got %d", stats.TotalImages)
	}
	if stats.TotalSize != 1024 {
		t.Errorf("expected TotalSize = 1024, got %d", stats.TotalSize)
	}
	if !stats.Enabled {
		t.Error("expected Enabled = true")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	entry1 := CacheEntry{
		Reference:    "test1.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   filepath.Join(tmpDir, "layout1"),
	}
	entry2 := CacheEntry{
		Reference:    "test2.example.com/image:latest",
		Architecture: "arm64",
		Digest:       "sha256:def456",
		CachedAt:     time.Now(),
		Size:         2048,
		LayoutPath:   filepath.Join(tmpDir, "layout2"),
	}

	cache.addEntry(entry1)
	cache.addEntry(entry2)

	list := cache.List()
	if len(list) != 2 {
		t.Errorf("expected 2 entries, got %d", len(list))
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	if cache.Exists("test.example.com/image:latest", "amd64") {
		t.Error("expected entry to not exist initially")
	}

	entry := CacheEntry{
		Reference:    "test.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   filepath.Join(tmpDir, "layout"),
	}
	cache.addEntry(entry)

	if !cache.Exists("test.example.com/image:latest", "amd64") {
		t.Error("expected entry to exist after adding")
	}
}

func TestParseReference(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{
			name:    "valid reference with tag",
			ref:     "docker.io/library/nginx:latest",
			wantErr: false,
		},
		{
			name:    "valid reference with digest",
			ref:     "ghcr.io/user/image@sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4",
			wantErr: false,
		},
		{
			name:    "valid reference with oci prefix",
			ref:     "oci://docker.io/library/nginx:latest",
			wantErr: false,
		},
		{
			name:    "valid reference with docker prefix",
			ref:     "docker://ghcr.io/user/image:latest",
			wantErr: false,
		},
		{
			name:    "invalid reference",
			ref:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseReference(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ref.Reference == "" {
				t.Error("expected non-empty reference")
			}
		})
	}
}

func TestNormalizeImageReference(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
	}{
		{
			name: "oci prefix removed",
			ref:  "oci://docker.io/library/nginx:latest",
			want: "docker.io/library/nginx:latest",
		},
		{
			name: "docker prefix removed",
			ref:  "docker://ghcr.io/user/image:latest",
			want: "ghcr.io/user/image:latest",
		},
		{
			name: "latest tag added",
			ref:  "docker.io/library/nginx",
			want: "docker.io/library/nginx:latest",
		},
		{
			name: "whitespace trimmed",
			ref:  "  docker.io/library/nginx:latest  ",
			want: "docker.io/library/nginx:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeImageReference(tt.ref)
			if got != tt.want {
				t.Errorf("NormalizeImageReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashReference(t *testing.T) {
	hash1 := hashReference("test.example.com/image:latest", "amd64")
	hash2 := hashReference("test.example.com/image:latest", "amd64")
	hash3 := hashReference("test.example.com/image:latest", "arm64")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different input should produce different hash")
	}

	if len(hash1) != 64 {
		t.Errorf("expected 64 character hash (sha256), got %d", len(hash1))
	}
}

func TestGetImageLayoutPath(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	path := cache.getImageLayoutPath("test.example.com/image:latest", "amd64")

	if path == "" {
		t.Error("expected non-empty path")
	}

	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}

	expectedHash := hashReference("test.example.com/image:latest", "amd64")
	if filepath.Base(path) != expectedHash {
		t.Errorf("expected path to end with hash %s, got %s", expectedHash, filepath.Base(path))
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()

	if !cfg.Enabled {
		t.Error("expected Enabled = true")
	}

	if cfg.MaxSize != 50*1024*1024*1024 {
		t.Errorf("expected MaxSize = 50GB, got %d", cfg.MaxSize)
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}

	entry := CacheEntry{
		Reference:    "test.example.com/image:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   filepath.Join(tmpDir, "layout"),
	}
	cache.addEntry(entry)

	if err := cache.Close(); err != nil {
		t.Fatalf("failed to close cache: %v", err)
	}

	cache2, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to reload image cache: %v", err)
	}
	defer cache2.Close()

	if !cache2.Exists("test.example.com/image:latest", "amd64") {
		t.Error("expected entry to exist after reload")
	}
}

func TestCacheDisabledOperations(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir), WithCacheEnabled(false))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	store, err := cache.Get(ctx, "test.example.com/image:latest", "amd64", nil)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss for disabled cache, got %v", err)
	}
	if store != nil {
		t.Error("expected nil store")
	}

	if err := cache.Clear(); err != nil {
		t.Errorf("Clear() should not error for disabled cache: %v", err)
	}

	if _, err := cache.Size(); err != nil {
		t.Errorf("Size() should not error for disabled cache: %v", err)
	}
}

func TestLoadCacheConfigNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")

	cfg, err := LoadCacheConfig(configPath)
	if err != nil {
		t.Fatalf("LoadCacheConfig() should not error for non-existent file: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected default config to be enabled")
	}
}

func TestSaveAndLoadCacheConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cache.json")

	cfg := &CacheConfig{
		Enabled:  true,
		CacheDir: tmpDir,
		MaxSize:  100 * 1024 * 1024 * 1024,
	}

	if err := SaveCacheConfig(cfg, configPath); err != nil {
		t.Fatalf("SaveCacheConfig() failed: %v", err)
	}

	loadedCfg, err := LoadCacheConfig(configPath)
	if err != nil {
		t.Fatalf("LoadCacheConfig() failed: %v", err)
	}

	if loadedCfg.Enabled != cfg.Enabled {
		t.Errorf("Enabled mismatch: got %v, want %v", loadedCfg.Enabled, cfg.Enabled)
	}
	if loadedCfg.CacheDir != cfg.CacheDir {
		t.Errorf("CacheDir mismatch: got %v, want %v", loadedCfg.CacheDir, cfg.CacheDir)
	}
	if loadedCfg.MaxSize != cfg.MaxSize {
		t.Errorf("MaxSize mismatch: got %v, want %v", loadedCfg.MaxSize, cfg.MaxSize)
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	layoutPath1 := filepath.Join(tmpDir, "layout1")
	layoutPath2 := filepath.Join(tmpDir, "layout2")

	for _, p := range []string{layoutPath1, layoutPath2} {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create layout dir: %v", err)
		}
	}

	entry1 := CacheEntry{
		Reference:    "test.example.com/image1:latest",
		Architecture: "amd64",
		Digest:       "sha256:abc123",
		CachedAt:     time.Now(),
		Size:         1024,
		LayoutPath:   layoutPath1,
	}
	entry2 := CacheEntry{
		Reference:    "test.example.com/image2:latest",
		Architecture: "amd64",
		Digest:       "sha256:def456",
		CachedAt:     time.Now(),
		Size:         2048,
		LayoutPath:   layoutPath2,
	}

	cache.addEntry(entry1)
	cache.addEntry(entry2)

	if err := cache.Remove("test.example.com/image1:latest"); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	if cache.Exists("test.example.com/image1:latest", "amd64") {
		t.Error("expected image1 to be removed")
	}
	if !cache.Exists("test.example.com/image2:latest", "amd64") {
		t.Error("expected image2 to still exist")
	}
}

func TestGetImageLayoutPathConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewImageCache(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create image cache: %v", err)
	}
	defer cache.Close()

	path1 := cache.getImageLayoutPath("docker.io/library/nginx:latest", "amd64")
	path2 := cache.getImageLayoutPath("docker.io/library/nginx:latest", "amd64")

	if path1 != path2 {
		t.Error("same input should produce same path")
	}

	path3 := cache.getImageLayoutPath("docker.io/library/nginx:latest", "arm64")
	if path1 == path3 {
		t.Error("different arch should produce different path")
	}
}

func TestCacheEntryDigest(t *testing.T) {
	ref := "docker.io/library/nginx:latest"
	ar, err := registry.ParseReference(ref)
	if err != nil {
		t.Fatalf("failed to parse reference: %v", err)
	}

	if ar.Host() == "" {
		t.Error("expected non-empty host")
	}
}
