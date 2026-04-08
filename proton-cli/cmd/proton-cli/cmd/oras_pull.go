package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	orasregistry "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type orasPullOptions struct {
	output string
}

func newOrasCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oras",
		Short: "Pull OCI artifacts",
	}

	cmd.AddCommand(newOrasPullCommand())

	return cmd
}

func newOrasPullCommand() *cobra.Command {
	opts := &orasPullOptions{}

	cmd := &cobra.Command{
		Use:   "pull <oci-ref>",
		Short: "Pull an OCI artifact and extract it to a local directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return orasPull(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", ".", "Output directory")

	return cmd
}

func orasPull(ctx context.Context, reference string, opts *orasPullOptions) error {
	ar, err := orasregistry.ParseReference(reference)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "proton-cli-oras-pull-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	dst, err := oci.New(tmpDir)
	if err != nil {
		return err
	}

	repo := &remote.Repository{
		Client: &auth.Client{
			Credential: auth.StaticCredential(ar.Host(), auth.EmptyCredential),
			Cache:      auth.NewCache(),
		},
		Reference: ar,
	}

	manifest, err := fetchArtifactManifest(ctx, repo, ar.Reference)
	if err != nil {
		return err
	}

	progress := newPullProgressReporter(os.Stderr, artifactTransferTotalSize(manifest))
	trackedDst := &progressTarget{
		Target:   dst,
		reporter: progress,
	}

	if _, err := oras.Copy(ctx, repo, ar.Reference, trackedDst, ar.Reference, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("pull artifact %q: %w", reference, err)
	}
	if err := progress.finish(); err != nil {
		return err
	}

	outputPath, err := resolvePullOutputPath(opts.output, manifest.Layers)
	if err != nil {
		return err
	}

	return extractOCIArtifact(ctx, dst, ar.Reference, outputPath)
}

func extractOCIArtifact(ctx context.Context, store *oci.Store, reference, outputPath string) error {
	manifest, err := fetchArtifactManifest(ctx, store, reference)
	if err != nil {
		return err
	}

	if len(manifest.Layers) == 0 {
		return fmt.Errorf("artifact %q has no layers to extract", reference)
	}

	if len(manifest.Layers) == 1 {
		return extractDescriptor(ctx, store, manifest.Layers[0], outputPath)
	}

	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return err
	}

	for _, layer := range manifest.Layers {
		layerPath, err := secureJoin(outputPath, artifactFileName(layer))
		if err != nil {
			return err
		}
		if err := extractDescriptor(ctx, store, layer, layerPath); err != nil {
			return err
		}
	}

	return nil
}

func extractDescriptor(ctx context.Context, store *oci.Store, desc ocispec.Descriptor, outputPath string) error {
	rc, err := store.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("fetch blob %s: %w", desc.Digest, err)
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return err
	}

	return f.Chmod(0o644)
}

func artifactFileName(desc ocispec.Descriptor) string {
	if title := desc.Annotations[ocispec.AnnotationTitle]; title != "" {
		return title
	}
	return desc.Digest.Encoded() + ".blob"
}

func isTarDescriptor(desc ocispec.Descriptor) bool {
	title := strings.ToLower(desc.Annotations[ocispec.AnnotationTitle])
	mt := strings.ToLower(desc.MediaType)
	return strings.HasSuffix(title, ".tar") ||
		mt == "application/vnd.oci.image.layer.v1.tar" ||
		mt == "application/vnd.cncf.oras.artifact.content.v1.tar"
}

func isTarGzipDescriptor(desc ocispec.Descriptor) bool {
	title := strings.ToLower(desc.Annotations[ocispec.AnnotationTitle])
	mt := strings.ToLower(desc.MediaType)
	return strings.HasSuffix(title, ".tar.gz") ||
		strings.HasSuffix(title, ".tgz") ||
		mt == "application/vnd.oci.image.layer.v1.tar+gzip" ||
		mt == "application/vnd.cncf.oras.artifact.content.v1.tar+gzip"
}

func secureJoin(root, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || cleanName == "" {
		return "", fmt.Errorf("invalid artifact path %q", name)
	}
	if filepath.IsAbs(cleanName) {
		return "", fmt.Errorf("absolute artifact path %q is not allowed", name)
	}
	if cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("artifact path %q escapes output directory", name)
	}
	return filepath.Join(root, cleanName), nil
}

func resolvePullOutputPath(output string, layers []ocispec.Descriptor) (string, error) {
	if len(layers) == 0 {
		return "", fmt.Errorf("artifact has no layers")
	}

	if len(layers) > 1 {
		return output, nil
	}

	if output == "." || output == "" {
		return filepath.Join(".", artifactFileName(layers[0])), nil
	}
	if looksLikeFilePath(output) {
		return output, nil
	}
	return filepath.Join(output, artifactFileName(layers[0])), nil
}

func looksLikeFilePath(path string) bool {
	clean := filepath.Clean(path)
	if clean == "." || clean == string(os.PathSeparator) {
		return false
	}
	return filepath.Ext(clean) != ""
}

type orasManifestFetcher interface {
	Resolve(ctx context.Context, reference string) (ocispec.Descriptor, error)
	Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error)
}

func fetchArtifactManifest(ctx context.Context, fetcher orasManifestFetcher, reference string) (ocispec.Manifest, error) {
	manifestDesc, err := fetcher.Resolve(ctx, reference)
	if err != nil {
		return ocispec.Manifest{}, fmt.Errorf("resolve artifact: %w", err)
	}

	manifestRC, err := fetcher.Fetch(ctx, manifestDesc)
	if err != nil {
		return ocispec.Manifest{}, fmt.Errorf("fetch manifest: %w", err)
	}
	defer manifestRC.Close()

	var manifest ocispec.Manifest
	if err := json.NewDecoder(manifestRC).Decode(&manifest); err != nil {
		return ocispec.Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}

	return manifest, nil
}

func artifactTransferTotalSize(manifest ocispec.Manifest) int64 {
	var total int64
	for _, layer := range manifest.Layers {
		if layer.Size > 0 {
			total += layer.Size
		}
	}
	return total
}

type pullProgressReporter struct {
	out        io.Writer
	total      int64
	current    int64
	isTerminal bool
	mu         sync.Mutex
}

func newPullProgressReporter(out io.Writer, total int64) *pullProgressReporter {
	return &pullProgressReporter{
		out:        out,
		total:      total,
		isTerminal: isTerminalWriter(out),
	}
}

func (p *pullProgressReporter) add(size int64) {
	if p.out == nil || size <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += size
	p.writeLocked(false)
}

func (p *pullProgressReporter) finish() error {
	if p.out == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.writeLocked(true)
	return nil
}

func (p *pullProgressReporter) writeLocked(done bool) {
	line := fmt.Sprintf("Downloading... %s/%s", humanize.IBytes(uint64(p.current)), humanize.IBytes(uint64(p.total)))
	if p.total == 0 {
		line = fmt.Sprintf("Downloading... %s", humanize.IBytes(uint64(p.current)))
	}

	if p.isTerminal {
		fmt.Fprintf(p.out, "\r%s", line)
		if done {
			fmt.Fprintln(p.out)
		}
		return
	}

	fmt.Fprintln(p.out, line)
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

type progressTarget struct {
	oras.Target
	reporter *pullProgressReporter
}

func (t *progressTarget) Push(ctx context.Context, expected ocispec.Descriptor, reader io.Reader) error {
	if t.reporter == nil {
		return t.Target.Push(ctx, expected, reader)
	}
	return t.Target.Push(ctx, expected, newProgressReader(reader, t.reporter))
}

type progressReader struct {
	reader   io.Reader
	reporter *pullProgressReporter
}

func newProgressReader(reader io.Reader, reporter *pullProgressReporter) io.Reader {
	return &progressReader{
		reader:   reader,
		reporter: reporter,
	}
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 && r.reporter != nil {
		r.reporter.add(int64(n))
	}
	return n, err
}

var _ content.Pusher = (*progressTarget)(nil)
