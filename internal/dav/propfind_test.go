package dav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListZips_ReturnsOnlyZipFiles(t *testing.T) {
	t.Parallel()

	xmlBody := `<?xml version="1.0" encoding="utf-8"?>
<d:multistatus xmlns:d="DAV:">
  <d:response>
    <d:href>/dados_abertos_cnpj/2025-10/</d:href>
    <d:propstat>
      <d:prop>
        <d:getcontentlength>0</d:getcontentlength>
        <d:getcontenttype>httpd/unix-directory</d:getcontenttype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dados_abertos_cnpj/2025-10/Empresas0.zip</d:href>
    <d:propstat>
      <d:prop>
        <d:getcontentlength>12345</d:getcontentlength>
        <d:getcontenttype>application/zip</d:getcontenttype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dados_abertos_cnpj/2025-10/Socios0.zip</d:href>
    <d:propstat>
      <d:prop>
        <d:getcontentlength>67890</d:getcontentlength>
        <d:getcontenttype>application/zip</d:getcontenttype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dados_abertos_cnpj/2025-10/README.txt</d:href>
    <d:propstat>
      <d:prop>
        <d:getcontentlength>10</d:getcontentlength>
        <d:getcontenttype>text/plain</d:getcontenttype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PROPFIND" {
			t.Fatalf("expected PROPFIND, got %s", r.Method)
		}
		if depth := r.Header.Get("Depth"); depth != "1" {
			t.Fatalf("expected Depth=1, got %q", depth)
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(xmlBody))
	}))
	defer srv.Close()

	client := NewClient()
	items, err := client.ListZips(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("ListZips returned error: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("expected at least one .zip file, got empty list")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 .zip files, got %d", len(items))
	}

	for _, it := range items {
		if !strings.HasSuffix(strings.ToLower(it.Href), ".zip") {
			t.Fatalf("found non-zip in result: %s", it.Href)
		}
	}
}

