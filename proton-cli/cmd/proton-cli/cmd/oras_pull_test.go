package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewOrasCommand(t *testing.T) {
	cmd := newOrasCommand()
	require.NotNil(t, cmd)
	require.Equal(t, "oras", cmd.Use)

	pullCmd, _, err := cmd.Find([]string{"pull"})
	require.NoError(t, err)
	require.NotNil(t, pullCmd)
	require.Equal(t, "pull <oci-ref>", pullCmd.Use)
}

func TestOrasPullCommandDefaultsOutputToCurrentDir(t *testing.T) {
	cmd := newOrasPullCommand()
	outputFlag := cmd.Flags().Lookup("output")

	require.NotNil(t, outputFlag)
	require.Equal(t, ".", outputFlag.DefValue)
	_, required := outputFlag.Annotations[cobra.BashCompOneRequiredFlag]
	require.False(t, required)
}

func TestOrasPullCommandRequiresReference(t *testing.T) {
	cmd := newOrasPullCommand()
	cmd.SetArgs([]string{"-o", "/tmp/out"})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestArtifactTransferTotalSize(t *testing.T) {
	manifest := ocispec.Manifest{
		Layers: []ocispec.Descriptor{
			{Size: 128},
			{Size: 256},
		},
	}

	require.Equal(t, int64(384), artifactTransferTotalSize(manifest))
}

func TestPullProgressReporterWriteProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &pullProgressReporter{
		out:   buf,
		total: 1024,
	}

	reporter.add(256)

	require.Contains(t, buf.String(), "256 B/1.0 KiB")
}

func TestPullProgressReporterFromManifestDescriptors(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &pullProgressReporter{
		out:   buf,
		total: 1024,
	}

	reporter.add(512)
	require.Equal(t, int64(512), reporter.current)

	err := reporter.finish()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "512 B/1.0 KiB")
}

func TestPullProgressReporterNoopWithoutWriter(t *testing.T) {
	reporter := &pullProgressReporter{}

	reporter.add(10)
	require.NoError(t, reporter.finish())
}

func TestIsTerminalWriter(t *testing.T) {
	require.False(t, isTerminalWriter(io.Discard))
}

func TestProgressReaderTracksBytes(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &pullProgressReporter{
		out:   buf,
		total: 1024,
	}

	r := newProgressReader(bytes.NewReader([]byte("hello")), reporter)
	consumed, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "hello", string(consumed))
	require.Equal(t, int64(5), reporter.current)
	require.Contains(t, buf.String(), "5 B/1.0 KiB")
}

func TestSingleLayerArtifactWritesNamedFile(t *testing.T) {
	layer := ocispec.Descriptor{
		MediaType:   "application/vnd.oci.image.layer.v1.tar",
		Annotations: map[string]string{ocispec.AnnotationTitle: "proton-offline-package.tar"},
	}

	path, err := resolvePullOutputPath(".", []ocispec.Descriptor{layer})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(".", "proton-offline-package.tar"), path)
}

func TestMultiLayerArtifactKeepsDirectoryOutput(t *testing.T) {
	layers := []ocispec.Descriptor{
		{Annotations: map[string]string{ocispec.AnnotationTitle: "part1"}},
		{Annotations: map[string]string{ocispec.AnnotationTitle: "part2"}},
	}

	path, err := resolvePullOutputPath("/tmp/out", layers)
	require.NoError(t, err)
	require.Equal(t, "/tmp/out", path)
}

func TestSingleLayerArtifactRespectsExplicitFileOutput(t *testing.T) {
	layer := ocispec.Descriptor{
		Annotations: map[string]string{ocispec.AnnotationTitle: "pkg.tar"},
	}

	path, err := resolvePullOutputPath("/tmp/custom.tar", []ocispec.Descriptor{layer})
	require.NoError(t, err)
	require.Equal(t, "/tmp/custom.tar", path)
}

func TestLooksLikeFilePath(t *testing.T) {
	require.True(t, looksLikeFilePath("pkg.tar"))
	require.True(t, looksLikeFilePath(filepath.Join("tmp", "pkg.tgz")))
	require.False(t, looksLikeFilePath("."))
	require.False(t, looksLikeFilePath(string(os.PathSeparator)))
}
