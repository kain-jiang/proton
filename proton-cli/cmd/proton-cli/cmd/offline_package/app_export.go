package offline_package

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

var timeNow = time.Now

type appExportOptions struct {
	manifest            string
	output              string
	platform            string
	overrideRegistry    string
	ignoreMissingImages bool
	disableDependencies bool
}

func newAppExportCommand() *cobra.Command {
	opts := &appExportOptions{
		output:   "offline-app-package.tar",
		platform: "linux/amd64",
	}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export an application offline package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAppExport(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.manifest, "manifest", "f", "", "Path or URL to a VersionSet manifest")
	cmd.Flags().StringVarP(&opts.output, "output", "o", opts.output, "Output tar file")
	cmd.Flags().StringVar(&opts.platform, "platform", opts.platform, "Target platform")
	cmd.Flags().StringVar(&opts.overrideRegistry, "override-registry", "", "Override chart values .image.registry when pulling images")
	cmd.Flags().BoolVar(&opts.ignoreMissingImages, "ignore-missing-images", false, "Continue exporting when some images cannot be pulled")
	cmd.Flags().BoolVar(&opts.disableDependencies, "disable-dependencies", false, "Only process the root manifest and ignore dependencies")
	_ = cmd.MarkFlagRequired("manifest")

	return cmd
}

func runAppExport(ctx context.Context, opts *appExportOptions) error {
	log.Printf("reading manifest %q", opts.manifest)
	manifestBytes, _, manifestDocuments, err := loadAppManifestTree(ctx, opts.manifest, !opts.disableDependencies)
	if err != nil {
		return err
	}

	platforms, err := normalizeAppPlatforms(opts.platform)
	if err != nil {
		return err
	}
	platformLabel := strings.Join(platforms, ",")

	workdir, err := os.MkdirTemp("", "proton-cli-offline-app-export-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	chartsDir := filepath.Join(workdir, "charts")
	imagesDir := filepath.Join(workdir, "images")
	if err := os.MkdirAll(chartsDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		return err
	}

	log.Printf("resolving releases from %d manifest(s)", len(manifestDocuments))
	charts, err := downloadAppCharts(ctx, collectAppChartRequests(manifestDocuments), chartsDir)
	if err != nil {
		return err
	}

	log.Printf("extracting images from charts")
	images, warnings, err := extractAppImagesFromCharts(charts)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		log.Printf("warning: %s", warning)
	}

	log.Printf("pulling %d images for platforms %s", len(images), strings.Join(platforms, ", "))
	imageMetadata, imageErrors, err := exportAppImages(ctx, images, platforms, imagesDir, opts.overrideRegistry, opts.ignoreMissingImages)
	if err != nil {
		return err
	}
	for _, imageErr := range imageErrors {
		log.Printf("warning: %s", imageErr)
	}

	exportedManifest, err := buildExportedManifest(manifestBytes, platforms, opts.overrideRegistry, imageMetadata, imageErrors)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workdir, "manifest.yaml"), exportedManifest, 0o644); err != nil {
		return err
	}

	log.Printf("packaging %q", opts.output)
	if err := tarAppPackage(workdir, opts.output); err != nil {
		return err
	}

	exportedImages := countExportedAppImages(imageMetadata)
	fmt.Printf("export completed\n- platform: %s\n- charts: %d\n- images: %d/%d\n- output: %s\n", platformLabel, len(charts), exportedImages, len(imageMetadata), opts.output)
	return nil
}

func downloadAppCharts(ctx context.Context, charts []appChartArtifact, chartsDir string) ([]appChartArtifact, error) {
	indexFiles := map[string]*repo.IndexFile{}
	out := make([]appChartArtifact, 0, len(charts))
	for _, chart := range charts {
		indexFile, err := loadChartRepositoryIndex(chart.RepoURL, indexFiles)
		if err != nil {
			return nil, err
		}

		cv, err := indexFile.Get(chart.Name, chart.Version)
		if err != nil {
			return nil, fmt.Errorf("find chart %s-%s in repo %s index: %w", chart.Name, chart.Version, chart.RepoURL, err)
		}
		if len(cv.URLs) == 0 {
			return nil, fmt.Errorf("chart %s-%s has no downloadable URL", chart.Name, chart.Version)
		}

		resolved, err := resolveChartURL(chart.RepoURL, cv.URLs[0])
		if err != nil {
			return nil, fmt.Errorf("resolve chart %s-%s URL: %w", chart.Name, chart.Version, err)
		}

		log.Printf("downloading chart %s-%s", chart.Name, chart.Version)
		targetPath := filepath.Join(chartsDir, chart.Path)
		if err := downloadFile(ctx, resolved, targetPath); err != nil {
			return nil, fmt.Errorf("download chart %s-%s: %w", chart.Name, chart.Version, err)
		}

		chart.URL = resolved
		chart.Path = targetPath
		out = append(out, chart)
	}

	return out, nil
}

func loadChartRepositoryIndex(repoURL string, cache map[string]*repo.IndexFile) (*repo.IndexFile, error) {
	if indexFile, ok := cache[repoURL]; ok {
		return indexFile, nil
	}

	repoEntry := &repo.Entry{
		Name: "offline-package-app",
		URL:  repoURL,
	}
	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(cli.New()))
	if err != nil {
		return nil, fmt.Errorf("create chart repository client for %s: %w", repoURL, err)
	}

	indexPath, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("download chart repository index for %s: %w", repoURL, err)
	}
	defer os.Remove(indexPath)

	indexFile, err := repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("load chart repository index for %s: %w", repoURL, err)
	}

	cache[repoURL] = indexFile
	return indexFile, nil
}

func resolveChartURL(repoURL, raw string) (string, error) {
	base, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(ref).String(), nil
}

func downloadFile(ctx context.Context, rawURL, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractAppImagesFromCharts(charts []appChartArtifact) ([]appImageRef, []string, error) {
	seen := map[string]appImageRef{}
	var warnings []string

	for _, chart := range charts {
		f, err := os.Open(chart.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("open chart %s: %w", chart.Path, err)
		}

		loaded, err := loader.LoadArchive(f)
		_ = f.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("load chart %s: %w", chart.Path, err)
		}

		refs, ok := extractAppImagesFromValues(loaded.Values)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("chart %s did not expose recognizable image fields", filepath.Base(chart.Path)))
			continue
		}

		for _, ref := range refs {
			seen[ref.Source] = ref
		}
	}

	out := make([]appImageRef, 0, len(seen))
	for _, ref := range seen {
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Source < out[j].Source
	})
	return out, warnings, nil
}

func extractAppImagesFromValues(values map[string]any) ([]appImageRef, bool) {
	rawImage, ok := values["image"]
	if !ok {
		return nil, false
	}

	imageMap, ok := rawImage.(map[string]any)
	if !ok {
		return nil, false
	}

	registry, _ := imageMap["registry"].(string)
	var refs []appImageRef

	if repository, okRepo := imageMap["repository"].(string); okRepo {
		tag, okTag := imageMap["tag"].(string)
		if okTag && registry != "" && repository != "" && tag != "" {
			ref, err := newAppImageRef(registry, repository, tag)
			if err == nil {
				refs = append(refs, ref)
			}
		}
	}

	for key, raw := range imageMap {
		if key == "registry" || key == "repository" || key == "tag" {
			continue
		}

		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		repository, okRepo := child["repository"].(string)
		tag, okTag := child["tag"].(string)
		if !okRepo || !okTag || registry == "" || repository == "" || tag == "" {
			continue
		}

		ref, err := newAppImageRef(registry, repository, tag)
		if err == nil {
			refs = append(refs, ref)
		}
	}

	return refs, len(refs) > 0
}

func newAppImageRef(registryHost, repositoryName, tag string) (appImageRef, error) {
	source := fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(registryHost, "/"), strings.TrimPrefix(repositoryName, "/"), tag)
	named, err := reference.ParseNormalizedNamed(source)
	if err != nil {
		return appImageRef{}, err
	}

	tagged, ok := reference.TagNameOnly(named).(reference.NamedTagged)
	if !ok {
		return appImageRef{}, fmt.Errorf("image %q is missing a tag", source)
	}

	return appImageRef{
		Source: tagged.String(),
		// The values field `image.registry` is treated as the full registry address.
		// The offline package name must drop that whole prefix and keep only
		// `image.repository`.
		Repository: strings.TrimPrefix(repositoryName, "/"),
		Tag:        tagged.Tag(),
	}, nil
}

func exportAppImages(ctx context.Context, images []appImageRef, platforms []string, imagesDir, overrideRegistry string, ignoreMissing bool) ([]appPackageImage, []string, error) {
	dst, err := oci.New(imagesDir)
	if err != nil {
		return nil, nil, err
	}

	metadata := make([]appPackageImage, 0, len(images))
	var imageErrors []string
	for _, image := range images {
		entry := appPackageImage{
			Source:             image.Source,
			PullSource:         image.Source,
			Repository:         image.Repository,
			Tag:                image.Tag,
			LocalRef:           image.LocalRef(),
			RequestedPlatforms: append([]string(nil), platforms...),
		}

		pullSource, err := overrideAppImageSource(image, overrideRegistry)
		if err != nil {
			entry.Error = fmt.Sprintf("override registry: %v", err)
			metadata = append(metadata, entry)
			if !ignoreMissing {
				return metadata, imageErrors, fmt.Errorf("override registry for %s: %w", image.Source, err)
			}
			imageErrors = append(imageErrors, fmt.Sprintf("image %s skipped: %s", image.Source, entry.Error))
			continue
		}
		entry.PullSource = pullSource

		log.Printf("pull image %s", pullSource)
		srcRepo, srcRef, err := newRemoteRepositoryForReference(pullSource, "", "", false)
		if err != nil {
			entry.Error = fmt.Sprintf("prepare remote repository: %v", err)
			metadata = append(metadata, entry)
			if !ignoreMissing {
				return metadata, imageErrors, fmt.Errorf("prepare remote repository for %s: %w", image.Source, err)
			}
			imageErrors = append(imageErrors, fmt.Sprintf("image %s skipped: %s", image.Source, entry.Error))
			continue
		}

		selectedPlatforms, err := copyAppImageForPlatforms(ctx, srcRepo, srcRef, dst, image.LocalRef(), platforms)
		if err != nil {
			entry.Error = err.Error()
			metadata = append(metadata, entry)
			if !ignoreMissing {
				return metadata, imageErrors, fmt.Errorf("pull image %s: %w", image.Source, err)
			}
			imageErrors = append(imageErrors, fmt.Sprintf("image %s skipped: %s", image.Source, entry.Error))
			continue
		}

		entry.ExportedPlatforms = selectedPlatforms
		entry.Exported = true
		metadata = append(metadata, entry)
	}
	return metadata, imageErrors, nil
}

func overrideAppImageSource(image appImageRef, overrideRegistry string) (string, error) {
	overrideRegistry = strings.TrimSpace(overrideRegistry)
	if overrideRegistry == "" {
		return image.Source, nil
	}

	ref := fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(overrideRegistry, "/"), image.Repository, image.Tag)
	named, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return "", err
	}

	tagged, ok := reference.TagNameOnly(named).(reference.NamedTagged)
	if !ok {
		return "", fmt.Errorf("image %q is missing a tag", ref)
	}
	return tagged.String(), nil
}

func copyAppImageForPlatforms(ctx context.Context, srcRepo *remote.Repository, srcRef string, dst *oci.Store, localRef string, platforms []string) ([]string, error) {
	descriptors := make([]ocispec.Descriptor, 0, len(platforms))
	selected := make([]string, 0, len(platforms))

	for _, platformName := range platforms {
		targetPlatform, err := appPlatformSpec(platformName)
		if err != nil {
			return nil, err
		}

		desc, err := oras.Resolve(ctx, srcRepo, srcRef, oras.ResolveOptions{TargetPlatform: targetPlatform})
		if err != nil {
			return nil, fmt.Errorf("select platform %s: %w", platformName, err)
		}
		desc.Platform = targetPlatform

		if err := oras.CopyGraph(ctx, srcRepo, dst, desc, oras.DefaultCopyGraphOptions); err != nil {
			return nil, fmt.Errorf("copy platform %s: %w", platformName, err)
		}

		descriptors = append(descriptors, desc)
		selected = append(selected, platformName)
	}

	if len(descriptors) == 1 {
		if err := dst.Tag(ctx, descriptors[0], localRef); err != nil {
			return nil, fmt.Errorf("tag single-platform image: %w", err)
		}
		return selected, nil
	}

	indexDesc, err := pushAppImageIndex(ctx, dst, descriptors)
	if err != nil {
		return nil, err
	}
	if err := dst.Tag(ctx, indexDesc, localRef); err != nil {
		return nil, fmt.Errorf("tag multi-platform image: %w", err)
	}
	return selected, nil
}

func appPlatformSpec(platformName string) (*ocispec.Platform, error) {
	switch platformName {
	case "linux/amd64":
		return &ocispec.Platform{OS: "linux", Architecture: "amd64"}, nil
	case "linux/arm64":
		return &ocispec.Platform{OS: "linux", Architecture: "arm64"}, nil
	default:
		return nil, fmt.Errorf("unsupported platform %q", platformName)
	}
}

func pushAppImageIndex(ctx context.Context, dst *oci.Store, manifests []ocispec.Descriptor) (ocispec.Descriptor, error) {
	index := ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: manifests,
	}

	payload, err := json.Marshal(index)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("marshal image index: %w", err)
	}

	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageIndex,
		Digest:    digest.FromBytes(payload),
		Size:      int64(len(payload)),
	}
	if err := dst.Push(ctx, desc, bytes.NewReader(payload)); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("push image index: %w", err)
	}
	return desc, nil
}

func tarAppPackage(srcDir, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	return tw.AddFS(os.DirFS(srcDir))
}

func loadOCIImageTags(imagesDir string) ([]string, error) {
	indexBytes, err := os.ReadFile(filepath.Join(imagesDir, "index.json"))
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Manifests []struct {
			Annotations map[string]string `json:"annotations,omitempty"`
		} `json:"manifests,omitempty"`
	}
	if err := json.Unmarshal(indexBytes, &parsed); err != nil {
		return nil, err
	}

	var tags []string
	for _, manifest := range parsed.Manifests {
		if tag := manifest.Annotations["org.opencontainers.image.ref.name"]; tag != "" {
			tags = append(tags, tag)
		}
	}
	sort.Strings(tags)
	return tags, nil
}

func newRemoteRepositoryForReference(rawRef, username, password string, plainHTTP bool) (*remote.Repository, string, error) {
	parsed, err := registry.ParseReference(rawRef)
	if err != nil {
		return nil, "", err
	}

	repo, err := remote.NewRepository(parsed.Registry + "/" + parsed.Repository)
	if err != nil {
		return nil, "", err
	}
	repo.PlainHTTP = plainHTTP
	if username != "" || password != "" {
		repo.Client = authClient(username, password)
	}

	return repo, parsed.Reference, nil
}

func countExportedAppImages(images []appPackageImage) int {
	var count int
	for _, image := range images {
		if image.Exported {
			count++
		}
	}
	return count
}
