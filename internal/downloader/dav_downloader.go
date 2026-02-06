package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/abriciof/rfcnpj-loader/internal/dav"
)

type Wanted struct {
	Empresas         bool
	Estabelecimentos bool
	Socios           bool
	Simples          bool
	Motivos          bool
	Qualificacoes    bool
	Cnaes            bool
	Municipios       bool
	Naturezas        bool
	Paises           bool
}

var (
	reEmpresas         = regexp.MustCompile(`(?i)/Empresas\d+\.zip$`)
	reEstabelecimentos = regexp.MustCompile(`(?i)/Estabelecimentos\d+\.zip$`)
	reSocios           = regexp.MustCompile(`(?i)/Socios\d+\.zip$`)
)

func FilterWanted(items []dav.Item, want Wanted) []dav.Item {
	out := make([]dav.Item, 0, len(items))
	for _, it := range items {
		h := it.Href
		lh := strings.ToLower(h)
		switch {
		case want.Simples && strings.HasSuffix(lh, "/simples.zip"):
			out = append(out, it)
		case want.Motivos && strings.HasSuffix(lh, "/motivos.zip"):
			out = append(out, it)
		case want.Qualificacoes && strings.HasSuffix(lh, "/qualificacoes.zip"):
			out = append(out, it)
		case want.Cnaes && strings.HasSuffix(lh, "/cnaes.zip"):
			out = append(out, it)
		case want.Municipios && strings.HasSuffix(lh, "/municipios.zip"):
			out = append(out, it)
		case want.Naturezas && strings.HasSuffix(lh, "/naturezas.zip"):
			out = append(out, it)
		case want.Paises && strings.HasSuffix(lh, "/paises.zip"):
			out = append(out, it)
		case want.Empresas && reEmpresas.MatchString(h):
			out = append(out, it)
		case want.Estabelecimentos && reEstabelecimentos.MatchString(h):
			out = append(out, it)
		case want.Socios && reSocios.MatchString(h):
			out = append(out, it)
		}
	}
	return out
}

type DAVDownloader struct {
	BaseDomain     string
	OutputDir      string
	Workers        int
	EnableDownload bool // equivalente ao bloco comentado do Python
	http           *http.Client
}

func NewDAVDownloader(baseDomain, outputDir string, workers int, enable bool) *DAVDownloader {
	if workers <= 0 {
		workers = 4
	}
	return &DAVDownloader{
		BaseDomain:     strings.TrimRight(baseDomain, "/"),
		OutputDir:      outputDir,
		Workers:        workers,
		EnableDownload: enable,
		http: &http.Client{
			Timeout: 0, // downloads grandes -> sem timeout global
		},
	}
}

func (d *DAVDownloader) DownloadAll(ctx context.Context, items []dav.Item) error {
	if !d.EnableDownload {
		// equivalente ao seu bloco comentado: não baixar caso dê erro / reprocessamento
		return nil
	}
	if err := os.MkdirAll(d.OutputDir, 0o755); err != nil {
		return err
	}

	jobs := make(chan dav.Item)
	errs := make(chan error, d.Workers)
	var wg sync.WaitGroup

	for i := 0; i < d.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range jobs {
				if err := d.downloadOne(ctx, it); err != nil {
					errs <- err
					return
				}
			}
		}()
	}

	go func() {
		for _, it := range items {
			jobs <- it
		}
		close(jobs)
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DAVDownloader) downloadOne(ctx context.Context, it dav.Item) error {
	url := d.BaseDomain + it.Href
	fileName := path.Base(it.Href)
	dst := filepath.Join(d.OutputDir, fileName)

	// check_diff por tamanho (equivalente ao Python)
	if st, err := os.Stat(dst); err == nil {
		if it.ContentLength > 0 && st.Size() == it.ContentLength {
			return nil // já baixado e igual
		}
		_ = os.Remove(dst)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := d.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("download falhou %s (%d): %s", fileName, resp.StatusCode, strings.TrimSpace(string(b)))
	}

	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	start := time.Now()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	_ = f.Close()

	if err := os.Rename(tmp, dst); err != nil {
		return err
	}

	_ = start // se quiser logar tempo por arquivo
	return nil
}
