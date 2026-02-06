package extract

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Extractor struct {
	Workers      int
	EnableExtract bool // equivalente ao bloco comentado do Python
}

func NewExtractor(workers int, enable bool) *Extractor {
	if workers <= 0 {
		workers = 2
	}
	return &Extractor{Workers: workers, EnableExtract: enable}
}

func (e *Extractor) ExtractAll(ctx context.Context, zipFiles []string, destDir string) error {
	if !e.EnableExtract {
		// equivalente ao bloco comentado: não extrair/reprocessar
		slog.Info("extract disabled by config")
		return nil
	}
	slog.Info("extract stage started", "files", len(zipFiles), "workers", e.Workers, "dest_dir", destDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	jobs := make(chan string)
	errs := make(chan error, e.Workers)
	var wg sync.WaitGroup

	for i := 0; i < e.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for zf := range jobs {
				if err := extractOne(zf, destDir); err != nil {
					errs <- err
					return
				}
			}
		}()
	}

	go func() {
		for _, zf := range zipFiles {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- zf:
			}
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
	slog.Info("extract stage completed", "files", len(zipFiles))
	return nil
}

func extractOne(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fp := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fp, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(fp)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}
		out.Close()
		rc.Close()
	}
	return nil
}
