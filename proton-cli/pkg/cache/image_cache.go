package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	CacheDirName     = "proton-cli"
	ImageCacheSubDir = "image-cache"
	CacheMetaFile    = "cache.json"
	CacheVersion     = 1
)

var (
	ErrCacheMiss        = errors.New("cache miss")
	ErrCacheCorrupted   = errors.New("cache corrupted")
	ErrInvalidReference = errors.New("invalid image reference")
)

type ImageCache struct {
	cacheDir     string
	enableCache  bool
	indexFile    *os.File
	indexMu      bool
	meta         *CacheMeta
	manifestPath string
	mu           sync.RWMutex
}

type CacheMeta struct {
	Version    int          `json:"version"`
	LastUpdate time.Time    `json:"last_update"`
	Images     []CacheEntry `json:"images"`
}

type CacheEntry struct {
	Reference    string    `json:"reference"`
	Architecture string    `json:"architecture"`
	Digest       string    `json:"digest"`
	CachedAt     time.Time `json:"cached_at"`
	Size         int64     `json:"size"`
	LayoutPath   string    `json:"layout_path"`
}

type ImageCacheOption func(*ImageCache)

func WithCacheDir(cacheDir string) ImageCacheOption {
	return func(c *ImageCache) {
		c.cacheDir = cacheDir
	}
}

func WithCacheEnabled(enabled bool) ImageCacheOption {
	return func(c *ImageCache) {
		c.enableCache = enabled
	}
}

func NewImageCache(opts ...ImageCacheOption) (*ImageCache, error) {
	cacheDir := getDefaultCacheDir()

	c := &ImageCache{
		cacheDir:    cacheDir,
		enableCache: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return c, nil
}

func getDefaultCacheDir() string {
	xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
	if xdgCacheHome != "" {
		return filepath.Join(xdgCacheHome, CacheDirName, ImageCacheSubDir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), CacheDirName, ImageCacheSubDir)
	}
	return filepath.Join(home, ".cache", CacheDirName, ImageCacheSubDir)
}

func (c *ImageCache) initialize() error {
	if !c.enableCache {
		return nil
	}

	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	c.meta = &CacheMeta{
		Version:    CacheVersion,
		LastUpdate: time.Now(),
		Images:     make([]CacheEntry, 0),
	}

	c.manifestPath = filepath.Join(c.cacheDir, CacheMetaFile)

	if data, err := os.ReadFile(c.manifestPath); err == nil {
		if err := json.Unmarshal(data, c.meta); err != nil {
			return fmt.Errorf("failed to parse cache manifest: %w", err)
		}
	}

	return nil
}

func (c *ImageCache) CacheDir() string {
	return c.cacheDir
}

func (c *ImageCache) IsEnabled() bool {
	return c.enableCache
}

func (c *ImageCache) Close() error {
	if c.indexFile != nil {
		if err := c.indexFile.Close(); err != nil {
			return err
		}
	}
	return c.saveMeta()
}

func (c *ImageCache) saveMeta() error {
	if c.meta == nil {
		return nil
	}

	c.meta.LastUpdate = time.Now()
	data, err := json.MarshalIndent(c.meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache meta: %w", err)
	}

	if err := os.WriteFile(c.manifestPath, data, 0644); err != nil {
		log.Printf("error: failed to write cache meta to %s: %v", c.manifestPath, err)
		return fmt.Errorf("failed to write cache meta: %w", err)
	}

	return nil
}

func (c *ImageCache) getImageLayoutPath(ref, arch string) string {
	hash := hashReference(ref, arch)
	return filepath.Join(c.cacheDir, hash)
}

func hashReference(ref, arch string) string {
	data := fmt.Sprintf("%s:%s", ref, arch)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func ParseReference(ref string) (registry.Reference, error) {
	ref = strings.TrimPrefix(ref, "oci://")
	ref = strings.TrimPrefix(ref, "docker://")

	ar, err := registry.ParseReference(ref)
	if err != nil {
		return registry.Reference{}, fmt.Errorf("%w: %v", ErrInvalidReference, err)
	}

	return ar, nil
}

func (c *ImageCache) Get(ctx context.Context, ref string, arch string, credentialFunc func(string) (auth.Credential, error)) (*oci.Store, error) {
	if !c.enableCache {
		return nil, ErrCacheMiss
	}

	c.mu.RLock()
	entry := c.findEntry(ref, arch)
	c.mu.RUnlock()

	if entry == nil {
		return nil, ErrCacheMiss
	}

	layoutPath := entry.LayoutPath
	if _, err := os.Stat(layoutPath); err != nil {
		c.mu.Lock()
		if err := c.removeEntry(ref, arch); err != nil {
			log.Printf("warning: failed to remove corrupted cache entry: %v", err)
		}
		c.mu.Unlock()
		return nil, ErrCacheMiss
	}

	store, err := oci.New(layoutPath)
	if err != nil {
		c.mu.Lock()
		if err := c.removeEntry(ref, arch); err != nil {
			log.Printf("warning: failed to remove corrupted cache entry: %v", err)
		}
		c.mu.Unlock()
		log.Printf("debug: failed to create OCI store from %s: %v", layoutPath, err)
		return nil, fmt.Errorf("%w: %v", ErrCacheCorrupted, err)
	}

	log.Printf("debug: successfully created OCI store from %s", layoutPath)
	return store, nil
}

func (c *ImageCache) GetWithDigest(ctx context.Context, ref string, arch string, expectedDigest string, credentialFunc func(string) (auth.Credential, error)) (*oci.Store, error) {
	store, err := c.Get(ctx, ref, arch, credentialFunc)
	if err != nil {
		return nil, err
	}

	ar, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	desc, err := store.Resolve(ctx, ar.Reference)
	if err != nil {
		c.mu.Lock()
		if err := c.removeEntry(ref, arch); err != nil {
			log.Printf("warning: failed to remove corrupted cache entry: %v", err)
		}
		c.mu.Unlock()
		return nil, ErrCacheMiss
	}

	if desc.Digest.String() != expectedDigest {
		c.mu.Lock()
		if err := c.removeEntry(ref, arch); err != nil {
			log.Printf("warning: failed to remove corrupted cache entry: %v", err)
		}
		c.mu.Unlock()
		return nil, ErrCacheMiss
	}

	return store, nil
}

func (c *ImageCache) Set(ctx context.Context, ref string, arch string, src *remote.Repository, srcRef string) error {
	if !c.enableCache {
		return nil
	}

	layoutPath := c.getImageLayoutPath(ref, arch)
	if err := os.MkdirAll(layoutPath, 0755); err != nil {
		return fmt.Errorf("failed to create layout directory: %w", err)
	}

	dst, err := oci.New(layoutPath)
	if err != nil {
		return fmt.Errorf("failed to create OCI store: %w", err)
	}

	log.Printf("debug: copying image %s to cache at %s", srcRef, layoutPath)

	// Parse the reference to get the tag
	ar, err := registry.ParseReference(ref)
	if err != nil {
		log.Printf("debug: failed to parse reference %s: %v", ref, err)
		os.RemoveAll(layoutPath)
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	// Use the tag as the reference in the OCI store
	tagRef := ar.Reference
	if tagRef == "" {
		tagRef = "latest"
	}

	desc, err := oras.Copy(ctx, src, srcRef, dst, tagRef, oras.DefaultCopyOptions)
	if err != nil {
		log.Printf("debug: failed to copy image to cache: %v", err)
		os.RemoveAll(layoutPath)
		return fmt.Errorf("failed to copy image to cache: %w", err)
	}
	log.Printf("debug: successfully cached image %s as %s, size: %d", srcRef, tagRef, desc.Size)

	entry := CacheEntry{
		Reference:    ref,
		Architecture: arch,
		Digest:       desc.Digest.String(),
		CachedAt:     time.Now(),
		Size:         desc.Size,
		LayoutPath:   layoutPath,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addEntry(entry); err != nil {
		os.RemoveAll(layoutPath)
		return fmt.Errorf("failed to save cache meta: %w", err)
	}

	return nil
}

func (c *ImageCache) findEntry(ref, arch string) *CacheEntry {
	for i := range c.meta.Images {
		entry := &c.meta.Images[i]
		if entry.Reference == ref && entry.Architecture == arch {
			return entry
		}
	}
	return nil
}

func (c *ImageCache) addEntry(entry CacheEntry) error {
	if err := c.removeEntry(entry.Reference, entry.Architecture); err != nil {
		log.Printf("warning: failed to remove existing cache entry: %v", err)
	}
	c.meta.Images = append(c.meta.Images, entry)
	return c.saveMeta()
}

func (c *ImageCache) removeEntry(ref, arch string) error {
	images := make([]CacheEntry, 0, len(c.meta.Images))
	for _, entry := range c.meta.Images {
		if entry.Reference == ref && entry.Architecture == arch {
			os.RemoveAll(entry.LayoutPath)
			continue
		}
		images = append(images, entry)
	}
	c.meta.Images = images
	return c.saveMeta()
}

func (c *ImageCache) Clear() error {
	if !c.enableCache {
		return nil
	}

	for _, entry := range c.meta.Images {
		os.RemoveAll(entry.LayoutPath)
	}

	c.meta.Images = make([]CacheEntry, 0)
	return c.saveMeta()
}

func (c *ImageCache) Remove(ref string) error {
	if !c.enableCache {
		return nil
	}

	arches := []string{"amd64", "arm64", runtimeGOARCH()}
	for _, arch := range arches {
		if err := c.removeEntry(ref, arch); err != nil {
			log.Printf("warning: failed to remove cache entry: %v", err)
		}
	}

	return c.saveMeta()
}

func (c *ImageCache) Size() (int64, error) {
	if !c.enableCache {
		return 0, nil
	}

	var totalSize int64
	for _, entry := range c.meta.Images {
		totalSize += entry.Size
	}
	return totalSize, nil
}

func (c *ImageCache) Stats() CacheStats {
	stats := CacheStats{
		TotalImages: len(c.meta.Images),
		Enabled:     c.enableCache,
		CacheDir:    c.cacheDir,
	}

	for _, entry := range c.meta.Images {
		stats.TotalSize += entry.Size
		if entry.Architecture == runtimeGOARCH() {
			stats.HostArchImages++
		}
	}

	return stats
}

type CacheStats struct {
	TotalImages    int    `json:"total_images"`
	TotalSize      int64  `json:"total_size"`
	HostArchImages int    `json:"host_arch_images"`
	Enabled        bool   `json:"enabled"`
	CacheDir       string `json:"cache_dir"`
}

func (c *ImageCache) List() []CacheEntry {
	return c.meta.Images
}

func (c *ImageCache) Exists(ref string, arch string) bool {
	return c.findEntry(ref, arch) != nil
}

func (c *ImageCache) UpdateCredential(ref string, credential auth.Credential) error {
	return nil
}

type CacheConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CacheDir string `json:"cache_dir" yaml:"cache_dir"`
	MaxSize  int64  `json:"max_size" yaml:"max_size"`
}

func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:  true,
		CacheDir: getDefaultCacheDir(),
		MaxSize:  50 * 1024 * 1024 * 1024,
	}
}

func LoadCacheConfig(configPath string) (*CacheConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultCacheConfig(), nil
		}
		return nil, err
	}

	var cfg CacheConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse cache config: %w", err)
	}

	if cfg.CacheDir == "" {
		cfg.CacheDir = getDefaultCacheDir()
	}

	return &cfg, nil
}

func SaveCacheConfig(cfg *CacheConfig, configPath string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func NormalizeImageReference(ref string) string {
	ref = strings.TrimPrefix(ref, "oci://")
	ref = strings.TrimPrefix(ref, "docker://")
	ref = strings.TrimSpace(ref)

	if !strings.Contains(ref, ":") {
		ref = ref + ":latest"
	}

	if u, err := url.Parse("docker://" + ref); err == nil {
		if u.Host != "" && !strings.Contains(u.Host, ".") && !strings.HasPrefix(ref, u.Host+"/") {
			if _, ok := defaultPorts[u.Host]; ok {
				ref = "docker.io/" + ref
			}
		}
	}

	return ref
}

var defaultPorts = map[string]bool{
	"docker.io": true,
	"ghcr.io":   true,
}

func runtimeGOARCH() string {
	switch os.Getenv("GOARCH") {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return "amd64"
	}
}
