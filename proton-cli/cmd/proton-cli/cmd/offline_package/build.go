package offline_package

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stianwa/createrepo"
	"golang.org/x/term"
	helmregistry "helm.sh/helm/v3/pkg/registry"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cache"
)

type artifactKind string

const (
	artifactKindBinary artifactKind = "binary"
	artifactKindChart  artifactKind = "chart"
	artifactKindImage  artifactKind = "image"
	artifactKindRPM    artifactKind = "rpm"
)

type buildOptions struct {
	manifest   string
	cacheDir   string
	noCache    bool
	cacheList  bool
	cacheClean bool
}

func defaultBuildOptions() *buildOptions {
	return &buildOptions{
		manifest: "manifest.yaml",
		cacheDir: "",
		noCache:  false,
	}
}

func (opts *buildOptions) AddFlag(s *pflag.FlagSet) {
	s.StringVar(&opts.manifest, "manifest", opts.manifest, "Path to the manifest file")
	s.StringVar(&opts.cacheDir, "cache-dir", opts.cacheDir, "Image cache directory (XDG_CACHE_HOME/proton-cli/image-cache by default)")
	s.BoolVar(&opts.noCache, "no-cache", opts.noCache, "Disable image cache")
	s.BoolVar(&opts.cacheList, "cache-list", opts.cacheList, "List cached images and exit")
	s.BoolVar(&opts.cacheClean, "cache-clean", opts.cacheClean, "Clean image cache and exit")
}

func newBuildCommand() *cobra.Command {
	opts := defaultBuildOptions()

	cmd := &cobra.Command{
		Use:   "build [flags]",
		Short: "Build a proton offline package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.cacheList {
				return listCache(opts.cacheDir)
			}
			if opts.cacheClean {
				return cleanCache(opts.cacheDir)
			}

			cacheEnabled := !opts.noCache
			cacheOpts := []cache.ImageCacheOption{
				cache.WithCacheEnabled(cacheEnabled),
			}
			if opts.cacheDir != "" {
				cacheOpts = append(cacheOpts, cache.WithCacheDir(opts.cacheDir))
			}

			imgCache, err := cache.NewImageCache(cacheOpts...)
			if err != nil {
				return fmt.Errorf("failed to initialize image cache: %w", err)
			}
			defer imgCache.Close()

			if cacheEnabled {
				log.Printf("Image cache enabled: %s", imgCache.CacheDir())
			} else {
				log.Println("Image cache disabled")
			}

			y, err := os.ReadFile(opts.manifest)
			if err != nil {
				return err
			}

			var m Manifest
			if err := yaml.Unmarshal(y, &m); err != nil {
				return err
			}

			return build(cmd.Context(), &m, imgCache)
		},
	}

	opts.AddFlag(cmd.Flags())

	return cmd
}

func listCache(cacheDir string) error {
	opts := []cache.ImageCacheOption{}
	if cacheDir != "" {
		opts = append(opts, cache.WithCacheDir(cacheDir))
	}

	imgCache, err := cache.NewImageCache(opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize image cache: %w", err)
	}
	defer imgCache.Close()

	stats := imgCache.Stats()
	fmt.Printf("Cache Directory: %s\n", stats.CacheDir)
	fmt.Printf("Total Images: %d\n", stats.TotalImages)
	fmt.Printf("Total Size: %s\n", formatSize(stats.TotalSize))
	fmt.Printf("Host Architecture Images: %d\n", stats.HostArchImages)
	fmt.Printf("Enabled: %v\n", stats.Enabled)
	fmt.Println()

	if len(stats.CacheDir) > 0 {
		entries := imgCache.List()
		if len(entries) > 0 {
			fmt.Println("Cached Images:")
			fmt.Printf("%-60s %-10s %-20s %s\n", "REFERENCE", "ARCH", "DIGEST", "SIZE")
			for _, e := range entries {
				fmt.Printf("%-60s %-10s %-20s %s\n", e.Reference, e.Architecture, e.Digest[:20], formatSize(e.Size))
			}
		}
	}

	return nil
}

func cleanCache(cacheDir string) error {
	opts := []cache.ImageCacheOption{}
	if cacheDir != "" {
		opts = append(opts, cache.WithCacheDir(cacheDir))
	}

	imgCache, err := cache.NewImageCache(opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize image cache: %w", err)
	}
	defer imgCache.Close()

	stats := imgCache.Stats()
	fmt.Printf("Cleaning cache directory: %s\n", stats.CacheDir)
	fmt.Printf("Removing %d cached images (%s)...\n", stats.TotalImages, formatSize(stats.TotalSize))

	if err := imgCache.Clear(); err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	fmt.Println("Cache cleaned successfully")
	return nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func build(ctx context.Context, m *Manifest, imgCache *cache.ImageCache) error {
	// create temporary directory as workspace
	w, err := os.MkdirTemp("", "proton-offline-package-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(w)
	log.Printf("working directory %q", w)

	var (
		binDir   = filepath.Join(w, "bin")
		chartDir = filepath.Join(w, "service-package", "charts")
		imageDir = filepath.Join(w, "service-package", "images")

		repoDir         = filepath.Join(w, "repos")
		repoPackagesDir = filepath.Join(repoDir, "Packages")
		repoRepodataDir = filepath.Join(repoDir, "repodata")
	)

	for _, p := range []string{
		binDir,
		chartDir,
		imageDir,
		repoPackagesDir,
		repoRepodataDir,
	} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}

	// create manifest file
	if err := createManifestFile(filepath.Join(w, "manifest.yaml"), m); err != nil {
		return err
	}

	// pull binaries (including proton-cli if defined in manifest)
	for _, a := range m.Spec.Binaries {
		if err := pullForAch(ctx, artifactKindBinary, &a, binDir, m.Spec.Architecture, imgCache); err != nil {
			return err
		}
	}

	// pull charts
	for _, a := range m.Spec.Charts {
		if err := pull(ctx, artifactKindChart, &a, chartDir, imgCache); err != nil {
			return err
		}
	}

	// pull images
	for _, a := range m.Spec.Images {
		if err := pullForAch(ctx, artifactKindImage, &a, imageDir, m.Spec.Architecture, imgCache); err != nil {
			return err
		}
	}

	// pull rpms
	for _, a := range m.Spec.RPMs {
		if err := pull(ctx, artifactKindRPM, &a, repoPackagesDir, imgCache); err != nil {
			return err
		}
	}

	// create install script
	if err := os.WriteFile(filepath.Join(w, "install.sh"), scriptInstallBytes, 0o644); err != nil {
		return err
	}

	// create rpm repository
	if err := createRPMRepository(repoDir); err != nil {
		return err
	}

	// package tarball
	f, err := os.Create("proton-offline-package.tar")
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	if err := tw.AddFS(os.DirFS(w)); err != nil {
		return err
	}

	return nil
}

func createManifestFile(p string, m *Manifest) error {
	y, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(p, y, 0o644)
}

func pull(ctx context.Context, kind artifactKind, a *Artifact, output string, imgCache *cache.ImageCache) error {
	return pullForAch(ctx, kind, a, output, runtime.GOARCH, imgCache)
}

func pullForAch(ctx context.Context, kind artifactKind, a *Artifact, output string, arch string, imgCache *cache.ImageCache) (err error) {
	switch {
	case a.HTTP != nil:
		err = pullHTTP(ctx, filepath.Join(output, a.Name), a.HTTP)
	case a.OCI != nil:
		if kind == artifactKindChart {
			err = pullChartOCI(ctx, filepath.Join(output, a.Name), a.OCI)
		} else {
			err = pullOCIForArch(ctx, output, a.Name, a.OCI, arch, imgCache)
		}
		if kind == artifactKindRPM {
			return pullRPMOCI(ctx, filepath.Join(output, a.Name), a.OCI)
		}
	case a.File != nil:
		err = pullFile(filepath.Join(output, a.Name), a.File)
	default:
		err = fmt.Errorf("failed to find artifact source of %q", a.Name)
	}
	if err != nil {
		return
	}

	// 检查可执行文件的格式
	if kind == artifactKindBinary {
		ok, err := isBinArch(filepath.Join(output, a.Name), arch)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%v is not %v", a.Name, arch)
		}
	}

	return
}

func pullHTTP(ctx context.Context, path string, s *HTTPSource) error {
	log.Println("pull http", s.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("pull http fail, status: %v, err: %w", resp.Status, err)
		}

		return fmt.Errorf("pull http fail, status: %v, body: %s", resp.Status, body)
	}

	var r io.Reader
	var mode os.FileMode
	switch s.Format {
	case "":
		r = resp.Body
		mode = 0o755
	case "tar+gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			h, err := tr.Next()
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("%s not found", s.Path)
			}
			if err != nil {
				return err
			}
			if h.Name != s.Path {
				continue
			}
			r = tr
			mode = h.FileInfo().Mode()
			break
		}
	default:
		return fmt.Errorf("invalid format %q", s.Format)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	if err := os.Chmod(path, mode); err != nil {
		return err
	}

	return nil
}

// container registry credentials cache
var credentials map[string]auth.Credential

func getCredential(hostPort string) (auth.Credential, error) {
	if cache, ok := credentials[hostPort]; ok {
		return cache, nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Login: %s\n", hostPort)
	fmt.Print("Username: ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return auth.EmptyCredential, fmt.Errorf("couldn't read from standard input: %w", err)
	}
	username := scanner.Text()

	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return auth.EmptyCredential, fmt.Errorf("couldn't read standard input: %w", err)
	}

	cache := auth.Credential{
		Username: username,
		Password: string(password),
	}

	if credentials == nil {
		credentials = make(map[string]auth.Credential)
	}

	credentials[hostPort] = cache

	return cache, nil
}

func pullOCIForArch(ctx context.Context, output, ref string, s *OCISource, arch string, imgCache *cache.ImageCache) error {
	log.Println("pull oci", s.Reference)
	ar, err := registry.ParseReference(s.Reference)
	if err != nil {
		return err
	}
	srcRef := ar.Reference

	if imgCache != nil && imgCache.IsEnabled() {
		log.Printf("checking cache for: %s (%s)", s.Reference, arch)
		cachedStore, err := imgCache.Get(ctx, s.Reference, arch, nil)
		if err == nil && cachedStore != nil {
			log.Printf("cache hit: %s (%s)", s.Reference, arch)
			dst, err := oci.New(output)
			if err != nil {
				return err
			}
			// Use the tag reference for cache copy
			tagRef := ar.Reference
			if tagRef == "" {
				tagRef = "latest"
			}

			if desc, err := oras.Copy(ctx, cachedStore, tagRef, dst, ref, oras.DefaultCopyOptions); err == nil {
				log.Printf("copied from cache: %s (size: %d)", s.Reference, desc.Size)
				return nil
			} else {
				log.Printf("cache copy failed, falling back to remote: %v", err)
			}
		} else if err != nil {
			log.Printf("cache miss: %v", err)
		} else {
			log.Printf("cache miss: not found")
		}
	}

	dst, err := oci.New(output)
	if err != nil {
		return err
	}

	var r *remote.Repository
	var desc ocispec.Descriptor
	var rc io.ReadCloser

	tryFetch := func(repo *remote.Repository) (ocispec.Descriptor, io.ReadCloser, error) {
		return repo.FetchReference(ctx, ar.Reference)
	}

	r = &remote.Repository{
		Client: &auth.Client{
			Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
			Cache:      auth.NewCache(),
		},
		Reference: ar,
	}

	desc, rc, err = tryFetch(r)
	if err != nil {
		if shouldRetryWithCredential(err) {
			r = &remote.Repository{
				Client: &auth.Client{
					Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
					Cache:      auth.NewCache(),
				},
				Reference: ar,
			}

			desc, rc, err = tryFetch(r)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer rc.Close()

	if isImageList(desc.MediaType) {
		if srcRef, err = choiceInstance(rc, "linux/"+arch); err != nil {
			return fmt.Errorf("choosing an image from list %s: %w", s.Reference, err)
		}
	}

	if _, err := oras.Copy(ctx, r, srcRef, dst, ref, oras.DefaultCopyOptions); err != nil {
		return err
	}

	if imgCache != nil && imgCache.IsEnabled() {
		log.Printf("caching: %s (%s)", s.Reference, arch)
		srcForCache := &remote.Repository{
			Client: &auth.Client{
				Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
				Cache:      auth.NewCache(),
			},
			Reference: ar,
		}
		if err := imgCache.Set(ctx, s.Reference, arch, srcForCache, srcRef); err != nil {
			log.Printf("warning: failed to cache image: %v", err)
		}
	}

	return nil
}

func pullChartOCI(ctx context.Context, path string, s *OCISource) error {
	log.Println("pull oci chart", s.Reference)

	ar, err := registry.ParseReference(s.Reference)
	if err != nil {
		return err
	}

	client, err := helmregistry.NewClient()
	if err != nil {
		return err
	}

	result, err := client.Pull(s.Reference)
	if err != nil {
		if !shouldRetryWithCredential(err) {
			return err
		}

		credential, err := getCredential(ar.Registry)
		if err != nil {
			return err
		}

		if err := client.Login(ar.Registry, helmregistry.LoginOptBasicAuth(credential.Username, credential.Password)); err != nil {
			return err
		}

		result, err = client.Pull(s.Reference)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(path, result.Chart.Data, 0o644)
}

// pullRPMOCI pulls an RPM file from an OCI registry using oras.
// The artifact is expected to be an OCI artifact bundle containing the RPM file.
// It copies the artifact to a temporary OCI store and extracts the RPM file.
func pullRPMOCI(ctx context.Context, path string, s *OCISource) error {
	log.Println("pull oci rpm", s.Reference)

	ar, err := registry.ParseReference(s.Reference)
	if err != nil {
		return err
	}

	// Create a temporary directory for the OCI store
	tmpDir, err := os.MkdirTemp("", "proton-cli-rpm-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Create source repository with authentication support
	r := &remote.Repository{
		Client: &auth.Client{
			Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
			Cache:      auth.NewCache(),
		},
		Reference: ar,
	}

	// Try to fetch first to check if authentication is needed
	_, rc, err := r.FetchReference(ctx, ar.Reference)
	if err != nil {
		if shouldRetryWithCredential(err) {
			r = &remote.Repository{
				Client: &auth.Client{
					Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
					Cache:      auth.NewCache(),
				},
				Reference: ar,
			}

			_, rc, err = r.FetchReference(ctx, ar.Reference)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if rc != nil {
		rc.Close()
	}

	// Create destination OCI store
	dst, err := oci.New(tmpDir)
	if err != nil {
		return err
	}

	// Copy the artifact from remote to local OCI store
	_, err = oras.Copy(ctx, r, ar.Reference, dst, ar.Reference, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to copy artifact: %w", err)
	}

	// Now extract the RPM file from the local OCI store
	// The artifact may contain multiple blobs, we need to find the RPM file
	src, err := oci.New(tmpDir)
	if err != nil {
		return err
	}

	// Get the manifest to find the blobs
	manifestDesc, err := src.Resolve(ctx, ar.Reference)
	if err != nil {
		return fmt.Errorf("failed to resolve artifact: %w", err)
	}

	// Fetch the manifest content
	manifestRC, err := src.Fetch(ctx, manifestDesc)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer manifestRC.Close()

	manifestBytes, err := io.ReadAll(manifestRC)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse the manifest to get blob digests
	var manifest struct {
		Layers []struct {
			MediaType   string `json:"mediaType"`
			Digest      string `json:"digest"`
			Size        int64  `json:"size"`
			Annotations struct {
				Title string `json:"org.opencontainers.image.title"`
			} `json:"annotations,omitempty"`
		} `json:"layers"`
		Blobs []struct {
			MediaType   string `json:"mediaType"`
			Digest      string `json:"digest"`
			Size        int64  `json:"size"`
			Annotations struct {
				Title string `json:"org.opencontainers.image.title"`
			} `json:"annotations,omitempty"`
		} `json:"blobs"`
		Config struct {
			Digest string `json:"digest"`
		} `json:"config"`
	}

	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		// If it's an OCI image manifest, try a different structure
		var imageManifest struct {
			Config struct {
				MediaType string `json:"mediaType"`
				Digest    string `json:"digest"`
				Size      int64  `json:"size"`
			} `json:"config"`
			Layers []struct {
				MediaType string `json:"mediaType"`
				Digest    string `json:"digest"`
				Size      int64  `json:"size"`
			} `json:"layers"`
		}
		if err2 := json.Unmarshal(manifestBytes, &imageManifest); err2 != nil {
			return fmt.Errorf("failed to parse manifest: %w, original error: %v", err, err2)
		}
		// For image manifests, treat layers as the content
		for _, layer := range imageManifest.Layers {
			d, err := digest.Parse(layer.Digest)
			if err != nil {
				continue
			}
			if err := extractBlobToFile(ctx, src, d, path); err != nil {
				continue
			}
			return nil
		}
		return fmt.Errorf("failed to extract RPM from image manifest")
	}

	// Try to find and extract the RPM file from layers/blobs
	// Look for a blob with .rpm extension in the title annotation
	for _, layer := range manifest.Layers {
		if strings.HasSuffix(layer.Annotations.Title, ".rpm") {
			d, err := digest.Parse(layer.Digest)
			if err != nil {
				continue
			}
			if err := extractBlobToFile(ctx, src, d, path); err != nil {
				return fmt.Errorf("failed to extract RPM blob: %w", err)
			}
			return nil
		}
	}

	// If no title annotation, try all blobs
	for _, blob := range manifest.Blobs {
		if strings.HasSuffix(blob.Annotations.Title, ".rpm") {
			d, err := digest.Parse(blob.Digest)
			if err != nil {
				continue
			}
			if err := extractBlobToFile(ctx, src, d, path); err != nil {
				return fmt.Errorf("failed to extract RPM blob: %w", err)
			}
			return nil
		}
	}

	// If no RPM file found by name, try to extract the first non-config blob
	for _, layer := range manifest.Layers {
		d, err := digest.Parse(layer.Digest)
		if err != nil {
			continue
		}
		if err := extractBlobToFile(ctx, src, d, path); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no RPM file found in artifact")
}

// extractBlobToFile extracts a blob from the OCI store and saves it to a file
func extractBlobToFile(ctx context.Context, store *oci.Store, d digest.Digest, path string) error {
	desc := ocispec.Descriptor{
		Digest: d,
	}

	rc, err := store.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return err
	}

	if err := os.Chmod(path, 0o644); err != nil {
		return err
	}

	return nil
}

func pullFile(path string, s *FileSource) error {
	log.Println("pull file", s.Path)

	src, err := os.Open(s.Path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func shouldRetryWithCredential(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	for _, token := range []string{
		"unauthorized",
		"authentication required",
		"denied",
		"forbidden",
	} {
		if strings.Contains(msg, token) {
			return true
		}
	}

	return false
}

func createRPMRepository(dir string) error {
	repo, err := createrepo.NewRepo(dir, &createrepo.Config{})
	if err != nil {
		return fmt.Errorf("init rpm repo fail: %w", err)
	}
	if _, err := repo.Create(); err != nil {
		return fmt.Errorf("create rpm repo fail: %w", err)
	}

	// create repository config template
	if err := os.WriteFile(filepath.Join(dir, "proton.repo.tmpl"), templateProtonRepoBytes, 0o644); err != nil {
		return err
	}

	return nil
}

func isImageList(mt string) bool {
	return mt == "application/vnd.docker.distribution.manifest.list.v2+json" || mt == "application/vnd.oci.image.index.v1+json"
}

// 对于 application/vnd.docker.distribution.manifest.list.v2+json 和 application/vnd.oci.image.index.v1+json 的简单定义
type imageList struct {
	Manifests []struct {
		MediaType string `json:"mediaType,omitzero"`
		Digest    string `json:"digest,omitzero"`
		Platform  struct {
			OS           string `json:"os,omitzero"`
			Architecture string `json:"architecture,omitzero"`
		} `json:"platform,omitzero"`
	} `json:"manifests,omitzero"`
}

func choiceInstance(r io.Reader, platform string) (digest string, err error) {
	var l imageList
	if err = json.NewDecoder(r).Decode(&l); err != nil {
		return
	}

	for _, m := range l.Manifests {
		if platform == m.Platform.OS+"/"+m.Platform.Architecture {
			digest = m.Digest
			return
		}
	}

	err = fmt.Errorf("no image found for %s", platform)
	return
}

func isBinArch(path, arch string) (ok bool, err error) {
	f, err := elf.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	switch arch {
	case "amd64":
		ok = f.Machine == elf.EM_X86_64
	case "arm64":
		ok = f.Machine == elf.EM_AARCH64
	default:
		err = fmt.Errorf("unsupported arch: %v", arch)
	}

	return
}
