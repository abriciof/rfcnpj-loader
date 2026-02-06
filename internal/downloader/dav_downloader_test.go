package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/abriciof/rfcnpj-loader/internal/dav"
)

func TestFilterWanted(t *testing.T) {
	t.Parallel()

	items := []dav.Item{
		{Href: "/x/Simples.zip"},
		{Href: "/x/Motivos.zip"},
		{Href: "/x/Empresas0.zip"},
		{Href: "/x/Estabelecimentos2.zip"},
		{Href: "/x/Socios1.zip"},
		{Href: "/x/README.txt"},
	}

	got := FilterWanted(items, Wanted{
		Simples:          true,
		Motivos:          true,
		Empresas:         true,
		Estabelecimentos: true,
		Socios:           true,
	})
	if len(got) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(got))
	}
}

func TestNewDAVDownloader_Defaults(t *testing.T) {
	t.Parallel()

	d := NewDAVDownloader("https://example.test/", "out", 0, true)
	if d.Workers != 4 {
		t.Fatalf("expected default workers=4, got %d", d.Workers)
	}
	if d.BaseDomain != "https://example.test" {
		t.Fatalf("unexpected BaseDomain: %q", d.BaseDomain)
	}
}

func TestDownloadAll_Disabled(t *testing.T) {
	t.Parallel()

	d := NewDAVDownloader("https://example.test", filepath.Join(t.TempDir(), "out"), 1, false)
	if err := d.DownloadAll(context.Background(), []dav.Item{{Href: "/a.zip"}}); err != nil {
		t.Fatalf("DownloadAll disabled should not fail: %v", err)
	}
}

func TestDownloadOne_DownloadsFile(t *testing.T) {
	t.Parallel()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path != "/file.zip" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("zip-content"))
	}))
	defer srv.Close()

	out := t.TempDir()
	d := NewDAVDownloader(srv.URL, out, 1, true)
	d.http = srv.Client()

	item := dav.Item{Href: "/file.zip", ContentLength: int64(len("zip-content"))}
	if err := d.downloadOne(context.Background(), item); err != nil {
		t.Fatalf("downloadOne returned error: %v", err)
	}

	gotPath := filepath.Join(out, "file.zip")
	b, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(b) != "zip-content" {
		t.Fatalf("unexpected file content: %q", string(b))
	}

	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 request, got %d", hits)
	}
}

func TestDownloadOne_SkipsWhenSameSize(t *testing.T) {
	t.Parallel()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("new-data"))
	}))
	defer srv.Close()

	out := t.TempDir()
	existingPath := filepath.Join(out, "file.zip")
	if err := os.WriteFile(existingPath, []byte("old-data"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	d := NewDAVDownloader(srv.URL, out, 1, true)
	d.http = srv.Client()
	item := dav.Item{Href: "/file.zip", ContentLength: int64(len("old-data"))}
	if err := d.downloadOne(context.Background(), item); err != nil {
		t.Fatalf("downloadOne returned error: %v", err)
	}

	if atomic.LoadInt32(&hits) != 0 {
		t.Fatalf("expected no request when same size, got %d", hits)
	}
}

