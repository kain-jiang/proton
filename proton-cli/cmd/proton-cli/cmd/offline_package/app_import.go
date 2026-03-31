package offline_package

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chartmuseum/helm-push/pkg/chartmuseum"
	"github.com/distribution/reference"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
)

type appImportOptions struct {
	input               string
	registry            string
	registryUsername    string
	registryPassword    string
	registryPlainHTTP   bool
	force               bool
	chartmuseumURL      string
	chartmuseumUsername string
	chartmuseumPassword string
}

func newAppImportCommand() *cobra.Command {
	opts := &appImportOptions{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an application offline package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAppImport(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.input, "input", "i", "", "Input offline package tar file")
	cmd.Flags().StringVar(&opts.registry, "registry", "", "Target registry host")
	cmd.Flags().StringVar(&opts.registryUsername, "registry-username", "", "Target registry username")
	cmd.Flags().StringVar(&opts.registryPassword, "registry-password", "", "Target registry password")
	cmd.Flags().BoolVar(&opts.registryPlainHTTP, "registry-plain-http", false, "Allow plain HTTP registry pushes")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing charts in ChartMuseum")
	cmd.Flags().StringVar(&opts.chartmuseumURL, "chartmuseum-url", "", "Target ChartMuseum URL")
	cmd.Flags().StringVar(&opts.chartmuseumUsername, "chartmuseum-username", "", "Target ChartMuseum username")
	cmd.Flags().StringVar(&opts.chartmuseumPassword, "chartmuseum-password", "", "Target ChartMuseum password")
	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("registry")
	_ = cmd.MarkFlagRequired("chartmuseum-url")

	return cmd
}

func runAppImport(ctx context.Context, opts *appImportOptions) error {
	if fi, err := os.Stat(opts.input); err != nil {
		return fmt.Errorf("stat input package: %w", err)
	} else if fi.IsDir() {
		return fmt.Errorf("input package must be a tar file")
	}

	workdir, err := os.MkdirTemp("", "proton-cli-offline-app-import-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	log.Printf("extracting package %q", opts.input)
	if err := extractAppPackage(opts.input, workdir); err != nil {
		return err
	}

	manifestPath, chartsDir, imagesDir, err := validateAppPackageLayout(workdir)
	if err != nil {
		return err
	}
	_ = manifestPath

	log.Printf("pushing images")
	imageCount, err := importAppImages(ctx, imagesDir, opts.registry, opts.registryUsername, opts.registryPassword, opts.registryPlainHTTP)
	if err != nil {
		return err
	}

	log.Printf("uploading charts")
	chartCount, err := importAppCharts(chartsDir, opts.chartmuseumURL, opts.chartmuseumUsername, opts.chartmuseumPassword, opts.force)
	if err != nil {
		return err
	}

	fmt.Printf("import completed\n- registry: %s\n- chartmuseum: %s\n- charts imported: %d\n- images imported: %d\n", opts.registry, opts.chartmuseumURL, chartCount, imageCount)
	return nil
}

func extractAppPackage(tarPath, dst string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if h.FileInfo().IsDir() {
			if err := os.MkdirAll(filepath.Join(dst, h.Name), h.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}
		if err := extractFileFromTar(tr, h, dst); err != nil {
			return err
		}
	}
}

func validateAppPackageLayout(root string) (string, string, string, error) {
	manifestPath := filepath.Join(root, "manifest.yaml")
	chartsDir := filepath.Join(root, "charts")
	imagesDir := filepath.Join(root, "images")

	for _, required := range []string{manifestPath, chartsDir, imagesDir} {
		if _, err := os.Stat(required); err != nil {
			return "", "", "", fmt.Errorf("invalid app package, missing %s: %w", filepath.Base(required), err)
		}
	}

	return manifestPath, chartsDir, imagesDir, nil
}

func importAppImages(ctx context.Context, imagesDir, registryHost, username, password string, plainHTTP bool) (int, error) {
	tags, err := loadOCIImageTags(imagesDir)
	if err != nil {
		return 0, fmt.Errorf("read image tags from package: %w", err)
	}

	src, err := oci.New(imagesDir)
	if err != nil {
		return 0, fmt.Errorf("open oci layout: %w", err)
	}

	for _, tag := range tags {
		repositoryName, tagName, err := splitAppLocalRef(tag)
		if err != nil {
			return 0, fmt.Errorf("parse image tag %s: %w", tag, err)
		}

		destination := fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(registryHost, "/"), repositoryName, tagName)
		log.Printf("push image %s", destination)

		dstRepo, _, err := newRemoteRepositoryForReference(destination, username, password, plainHTTP)
		if err != nil {
			return 0, fmt.Errorf("prepare registry repository for %s: %w", destination, err)
		}

		if _, err := oras.Copy(ctx, src, tag, dstRepo, tagName, oras.DefaultCopyOptions); err != nil {
			return 0, fmt.Errorf("push image %s: %w", destination, err)
		}
	}

	return len(tags), nil
}

func importAppCharts(chartsDir, host, username, password string, force bool) (int, error) {
	client, err := chartmuseum.NewClient(
		chartmuseum.URL(host),
		chartmuseum.Username(username),
		chartmuseum.Password(password),
	)
	if err != nil {
		return 0, fmt.Errorf("create chartmuseum client: %w", err)
	}

	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return 0, err
	}

	var count int
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".tgz" {
			continue
		}

		chartPath := filepath.Join(chartsDir, entry.Name())
		resp, err := client.UploadChartPackage(chartPath, force)
		if err != nil {
			return 0, fmt.Errorf("upload chart %s: %w", entry.Name(), err)
		}

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != 201 && resp.StatusCode != 202 {
			return 0, fmt.Errorf("upload chart %s failed: http %d: %s", entry.Name(), resp.StatusCode, strings.TrimSpace(string(body)))
		}

		count++
	}

	return count, nil
}

func splitAppLocalRef(localRef string) (string, string, error) {
	named, err := reference.ParseNormalizedNamed("example.invalid/" + localRef)
	if err != nil {
		return "", "", err
	}

	tagged, ok := reference.TagNameOnly(named).(reference.NamedTagged)
	if !ok {
		return "", "", fmt.Errorf("missing tag in %q", localRef)
	}

	return strings.TrimPrefix(reference.Path(tagged), "example.invalid/"), tagged.Tag(), nil
}
