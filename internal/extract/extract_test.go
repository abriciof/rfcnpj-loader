package extract

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewExtractor_DefaultWorkers(t *testing.T) {
	t.Parallel()

	e := NewExtractor(0, true)
	if e.Workers != 2 {
		t.Fatalf("expected default workers=2, got %d", e.Workers)
	}
}

func TestExtractAll_Disabled(t *testing.T) {
	t.Parallel()

	e := NewExtractor(1, false)
	dest := filepath.Join(t.TempDir(), "out")
	if err := e.ExtractAll(context.Background(), []string{"missing.zip"}, dest); err != nil {
		t.Fatalf("ExtractAll disabled should not fail, got: %v", err)
	}
	if _, err := os.Stat(dest); err == nil {
		t.Fatal("dest dir should not be created when extraction is disabled")
	}
}

func TestExtractAll_ExtractsFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "sample.zip")
	createTestZip(t, zipPath, map[string]string{
		"a.txt":       "alpha",
		"nested/b.txt": "beta",
	})

	dest := filepath.Join(dir, "out")
	e := NewExtractor(2, true)
	if err := e.ExtractAll(context.Background(), []string{zipPath}, dest); err != nil {
		t.Fatalf("ExtractAll returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "a.txt"), "alpha")
	assertFileContent(t, filepath.Join(dest, "nested", "b.txt"), "beta")
}

func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	if string(b) != expected {
		t.Fatalf("unexpected content for %s: got %q want %q", path, string(b), expected)
	}
}

