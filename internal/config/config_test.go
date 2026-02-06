package config

import "testing"

func TestLoad_DefaultsAndRequiredTemplate(t *testing.T) {
	t.Setenv("DAV_LIST_URL_TEMPLATE", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DAV_LIST_URL_TEMPLATE is empty")
	}

	t.Setenv("DAV_LIST_URL_TEMPLATE", "https://example.test/%s/")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.DavBaseDomain != "https://arquivos.receitafederal.gov.br" {
		t.Fatalf("unexpected DavBaseDomain default: %q", cfg.DavBaseDomain)
	}
	if !cfg.EnableDownload || !cfg.EnableExtract {
		t.Fatal("expected default EnableDownload/EnableExtract=true")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DAV_LIST_URL_TEMPLATE", "https://example.test/%s/")
	t.Setenv("ENABLE_DOWNLOAD", "false")
	t.Setenv("ENABLE_EXTRACT", "0")
	t.Setenv("CREATE_INDEXES", "yes")
	t.Setenv("DOWNLOAD_WORKERS", "8")
	t.Setenv("EXTRACT_WORKERS", "x") // invalid -> default 2
	t.Setenv("MAIL_NOTIFY_UPTODATE", "on")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.EnableDownload {
		t.Fatal("expected EnableDownload=false")
	}
	if cfg.EnableExtract {
		t.Fatal("expected EnableExtract=false")
	}
	if !cfg.CreateIndexes {
		t.Fatal("expected CreateIndexes=true")
	}
	if cfg.DownloadWorkers != 8 {
		t.Fatalf("unexpected DownloadWorkers: %d", cfg.DownloadWorkers)
	}
	if cfg.ExtractWorkers != 2 {
		t.Fatalf("expected fallback ExtractWorkers=2, got %d", cfg.ExtractWorkers)
	}
	if !cfg.MailNotifyUpToDate {
		t.Fatal("expected MailNotifyUpToDate=true")
	}
}
