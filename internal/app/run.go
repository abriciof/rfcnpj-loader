package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/abriciof/rfcnpj-loader/internal/config"
	"github.com/abriciof/rfcnpj-loader/internal/dav"
	"github.com/abriciof/rfcnpj-loader/internal/db"
	"github.com/abriciof/rfcnpj-loader/internal/downloader"
	"github.com/abriciof/rfcnpj-loader/internal/email"
	"github.com/abriciof/rfcnpj-loader/internal/extract"
	"github.com/abriciof/rfcnpj-loader/internal/loaders"
	"github.com/abriciof/rfcnpj-loader/internal/scan"
	"github.com/abriciof/rfcnpj-loader/internal/state"
	"github.com/abriciof/rfcnpj-loader/internal/timeutil"
)

type report struct {
	Month      timeutil.YearMonth
	MonthURL   string
	StartedAt  time.Time
	FinishedAt time.Time
	Downloaded int
	Extracted  int
	LoadedRows map[string]int64
	Errors     []string
}

func Run(ctx context.Context, cfg config.Config) error {
	start := time.Now()
	slog.Info("pipeline started",
		"start_month", cfg.StartMonth,
		"force_month", cfg.ForceMonth,
		"enable_download", cfg.EnableDownload,
		"enable_extract", cfg.EnableExtract,
		"create_indexes", cfg.CreateIndexes,
		"output_path", cfg.OutputFilesPath,
		"extracted_path", cfg.ExtractedFilesPath,
	)

	// ensure dirs
	_ = os.MkdirAll(cfg.OutputFilesPath, 0o755)
	_ = os.MkdirAll(cfg.ExtractedFilesPath, 0o755)

	sqlDB, err := db.OpenSQL(ctx, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	slog.Info("database connected", "host", cfg.DBHost, "port", cfg.DBPort, "db_name", cfg.DBName)

	meta := state.NewMetaStore(sqlDB)
	if err := meta.Ensure(ctx); err != nil {
		return err
	}

	res, items, err := resolveTargetMonth(ctx, cfg, meta)
	if err != nil {
		return err
	}

	rep := report{
		Month:      res,
		MonthURL:   fmt.Sprintf(cfg.DavListURLTemplate, res.String()),
		StartedAt:  start,
		LoadedRows: map[string]int64{},
	}

	if items == nil {
		// up-to-date
		msg := fmt.Sprintf("✅ Já atualizado. Próximo mês (%s) ainda não disponível.", res.HumanPTBR())
		slog.Info("up-to-date", "month", res.String(), "message", msg)

		if cfg.MailNotifyUpToDate && email.Enabled(email.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, To: cfg.MailTo}) {
			_ = email.Send(email.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, To: cfg.MailTo},
				"RFCNPJ Loader - Atualizado ("+res.String()+")",
				msg,
			)
		}
		return nil
	}
	slog.Info("remote files listed", "month", res.String(), "count", len(items))

	// Filter wanted zips based on enabled tables
	want := downloader.Wanted{
		Empresas:         cfg.LoadEmpresa,
		Estabelecimentos: cfg.LoadEstabelecimento,
		Socios:           cfg.LoadSocios,
		Simples:          cfg.LoadSimples,
		Motivos:          cfg.LoadMoti,
		Qualificacoes:    cfg.LoadQuals,
		Cnaes:            cfg.LoadCnae,
		Municipios:       cfg.LoadMunic,
		Naturezas:        cfg.LoadNatju,
		Paises:           cfg.LoadPais,
	}
	wantedItems := downloader.FilterWanted(items, want)
	rep.Downloaded = len(wantedItems)
	slog.Info("filtered wanted zip files", "count", len(wantedItems))

	// Download (equivalente ao bloco comentado do Python, controlado por ENABLE_DOWNLOAD)
	down := downloader.NewDAVDownloader(cfg.DavBaseDomain, cfg.OutputFilesPath, cfg.DownloadWorkers, cfg.EnableDownload)
	if err := down.DownloadAll(ctx, wantedItems); err != nil {
		return err
	}
	slog.Info("download stage finished", "planned_files", len(wantedItems), "enabled", cfg.EnableDownload)

	// Extract (equivalente ao bloco comentado do Python, controlado por ENABLE_EXTRACT)
	zipPaths := make([]string, 0, len(wantedItems))
	for _, it := range wantedItems {
		zipPaths = append(zipPaths, filepath.Join(cfg.OutputFilesPath, filepath.Base(it.Href)))
	}
	extractedMonthDir := filepath.Join(cfg.ExtractedFilesPath, res.String())
	ext := extract.NewExtractor(cfg.ExtractWorkers, cfg.EnableExtract)
	if err := ext.ExtractAll(ctx, zipPaths, extractedMonthDir); err != nil {
		return err
	}
	rep.Extracted = len(zipPaths)
	slog.Info("extract stage finished", "planned_files", len(zipPaths), "enabled", cfg.EnableExtract, "dest_dir", extractedMonthDir)

	// Scan extracted directory for CSV/TXT files
	filesByType, err := scan.ScanExtracted(extractedMonthDir)
	if err != nil {
		return err
	}
	slog.Info("scan stage finished",
		"empresa_files", len(filesByType.Empresa),
		"estabelecimento_files", len(filesByType.Estabelecimento),
		"socios_files", len(filesByType.Socios),
		"simples_files", len(filesByType.Simples),
		"cnae_files", len(filesByType.Cnae),
		"moti_files", len(filesByType.Moti),
		"munic_files", len(filesByType.Munic),
		"natju_files", len(filesByType.Natju),
		"pais_files", len(filesByType.Pais),
		"quals_files", len(filesByType.Quals),
	)

	// Load enabled tables in parallel (TABLE_WORKERS)
	tasks := buildLoadTasks(cfg, filesByType)
	if err := runLoadTasks(ctx, sqlDB, cfg, tasks, &rep); err != nil {
		return err
	}
	slog.Info("load stage finished", "tables", len(tasks))

	// Optional indexes (equivalente ao bloco comentado do Python)
	if cfg.CreateIndexes {
		if err := createIndexes(ctx, sqlDB, cfg); err != nil {
			return err
		}
		slog.Info("index stage finished")
	}

	// Save meta month + url
	_ = meta.Set(ctx, "loaded_month", res.String())
	_ = meta.Set(ctx, "loaded_url", rep.MonthURL)

	rep.FinishedAt = time.Now()

	// Email notify
	if email.Enabled(email.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, To: cfg.MailTo}) {
		subject := fmt.Sprintf("RFCNPJ Loader finalizado - %s", res.String())
		body := formatReport(rep)
		_ = email.Send(email.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, To: cfg.MailTo}, subject, body)
	}

	slog.Info("pipeline finished", "month", res.String(), "duration", time.Since(start).String())
	return nil
}

func resolveTargetMonth(ctx context.Context, cfg config.Config, meta *state.MetaStore) (timeutil.YearMonth, []dav.Item, error) {
	client := dav.NewClient()

	// FORCE_MONTH
	if strings.TrimSpace(cfg.ForceMonth) != "" {
		ym, err := timeutil.ParseYearMonth(cfg.ForceMonth)
		if err != nil {
			return timeutil.YearMonth{}, nil, fmt.Errorf("FORCE_MONTH inválido: %w", err)
		}
		url := fmt.Sprintf(cfg.DavListURLTemplate, ym.String())
		items, err := client.ListZips(ctx, url)
		if err != nil {
			return ym, nil, fmt.Errorf("FORCE_MONTH não disponível: %w", err)
		}
		return ym, items, nil
	}

	lastStr, ok, err := meta.Get(ctx, "loaded_month")
	if err != nil {
		return timeutil.YearMonth{}, nil, err
	}

	var target timeutil.YearMonth
	if ok {
		last, err := timeutil.ParseYearMonth(lastStr)
		if err != nil {
			return timeutil.YearMonth{}, nil, fmt.Errorf("loaded_month inválido no BD: %w", err)
		}
		target = last.Next()
	} else {
		if strings.TrimSpace(cfg.StartMonth) == "" {
			return timeutil.YearMonth{}, nil, fmt.Errorf("primeira execução: START_MONTH é obrigatório (não existe loaded_month no BD)")
		}
		first, err := timeutil.ParseYearMonth(cfg.StartMonth)
		if err != nil {
			return timeutil.YearMonth{}, nil, fmt.Errorf("START_MONTH inválido: %w", err)
		}
		target = first
	}

	url := fmt.Sprintf(cfg.DavListURLTemplate, target.String())
	items, err := client.ListZips(ctx, url)
	if err != nil {
		// mês ainda não publicado -> up-to-date
		return target, nil, nil
	}
	return target, items, nil
}

type loadTask struct {
	spec  loaders.TableSpec
	files []string
}

func buildLoadTasks(cfg config.Config, fb scan.FilesByType) []loadTask {
	var tasks []loadTask
	if cfg.LoadEmpresa {
		tasks = append(tasks, loadTask{spec: loaders.Empresa, files: fb.Empresa})
	}
	if cfg.LoadEstabelecimento {
		tasks = append(tasks, loadTask{spec: loaders.Estabelecimento, files: fb.Estabelecimento})
	}
	if cfg.LoadSocios {
		tasks = append(tasks, loadTask{spec: loaders.Socios, files: fb.Socios})
	}
	if cfg.LoadSimples {
		tasks = append(tasks, loadTask{spec: loaders.Simples, files: fb.Simples})
	}
	if cfg.LoadCnae {
		tasks = append(tasks, loadTask{spec: loaders.Cnae, files: fb.Cnae})
	}
	if cfg.LoadMoti {
		tasks = append(tasks, loadTask{spec: loaders.Moti, files: fb.Moti})
	}
	if cfg.LoadMunic {
		tasks = append(tasks, loadTask{spec: loaders.Munic, files: fb.Munic})
	}
	if cfg.LoadNatju {
		tasks = append(tasks, loadTask{spec: loaders.Natju, files: fb.Natju})
	}
	if cfg.LoadPais {
		tasks = append(tasks, loadTask{spec: loaders.Pais, files: fb.Pais})
	}
	if cfg.LoadQuals {
		tasks = append(tasks, loadTask{spec: loaders.Quals, files: fb.Quals})
	}

	// keep stable order
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].spec.Name < tasks[j].spec.Name })
	return tasks
}

func runLoadTasks(ctx context.Context, sqlDB *sql.DB, cfg config.Config, tasks []loadTask, rep *report) error {
	sem := make(chan struct{}, cfg.TableWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(tasks))

	for _, t := range tasks {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if len(t.files) == 0 {
				slog.Warn("no files found for table", "table", t.spec.Name)
				return
			}

			// drop+create table once
			if err := loaders.EnsureTable(ctx, sqlDB, t.spec, true); err != nil {
				errCh <- err
				return
			}

			// file-level parallelism
			fileSem := make(chan struct{}, cfg.FileWorkers)
			var wgf sync.WaitGroup
			localErr := make(chan error, len(t.files))

			for _, fp := range t.files {
				fp := fp
				wgf.Add(1)
				go func() {
					defer wgf.Done()
					fileSem <- struct{}{}
					defer func() { <-fileSem }()

					r, err := loaders.CopyCSV(ctx, sqlDB, t.spec, fp)
					if err != nil {
						localErr <- err
						return
					}
					mu.Lock()
					rep.LoadedRows[t.spec.Name] += r.Rows
					mu.Unlock()
				}()
			}

			wgf.Wait()
			close(localErr)
			for e := range localErr {
				if e != nil {
					errCh <- e
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func createIndexes(ctx context.Context, sqlDB *sql.DB, cfg config.Config) error {
	// equivalente ao bloco comentado do Python
	stmts := []string{}
	if cfg.LoadEmpresa {
		stmts = append(stmts, `CREATE INDEX IF NOT EXISTS empresa_cnpj ON empresa(cnpj_basico);`)
	}
	if cfg.LoadEstabelecimento {
		stmts = append(stmts, `CREATE INDEX IF NOT EXISTS estabelecimento_cnpj ON estabelecimento(cnpj_basico);`)
	}
	if cfg.LoadSocios {
		stmts = append(stmts, `CREATE INDEX IF NOT EXISTS socios_cnpj ON socios(cnpj_basico);`)
	}
	if cfg.LoadSimples {
		stmts = append(stmts, `CREATE INDEX IF NOT EXISTS simples_cnpj ON simples(cnpj_basico);`)
	}
	for _, s := range stmts {
		if _, err := sqlDB.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func formatReport(rep report) string {
	dur := rep.FinishedAt.Sub(rep.StartedAt)
	sb := strings.Builder{}
	sb.WriteString("RFCNPJ Loader - Finalizado\n")
	sb.WriteString("Mês: " + rep.Month.HumanPTBR() + " (" + rep.Month.String() + ")\n")
	sb.WriteString("URL: " + rep.MonthURL + "\n")
	sb.WriteString("Início: " + rep.StartedAt.Format(time.RFC3339) + "\n")
	sb.WriteString("Fim: " + rep.FinishedAt.Format(time.RFC3339) + "\n")
	sb.WriteString(fmt.Sprintf("Duração: %s\n", dur))
	sb.WriteString(fmt.Sprintf("Downloads planejados: %d\n", rep.Downloaded))
	sb.WriteString(fmt.Sprintf("Arquivos extraídos: %d\n", rep.Extracted))
	sb.WriteString("\nLinhas carregadas por tabela:\n")
	keys := make([]string, 0, len(rep.LoadedRows))
	for k := range rep.LoadedRows {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", k, rep.LoadedRows[k]))
	}
	return sb.String()
}
