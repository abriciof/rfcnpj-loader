package app

import (
	"strings"
	"testing"
	"time"

	"github.com/abriciof/rfcnpj-loader/internal/config"
	"github.com/abriciof/rfcnpj-loader/internal/scan"
	"github.com/abriciof/rfcnpj-loader/internal/timeutil"
)

func TestBuildLoadTasks_IncludesEnabledAndSorted(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		LoadSocios:  true,
		LoadEmpresa: true,
		LoadMoti:    true,
	}
	fb := scan.FilesByType{
		Empresa: []string{"empresa.csv"},
		Socios:  []string{"socios.csv"},
		Moti:    []string{"moti.csv"},
	}

	tasks := buildLoadTasks(cfg, fb)
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	gotOrder := []string{tasks[0].spec.Name, tasks[1].spec.Name, tasks[2].spec.Name}
	wantOrder := []string{"empresa", "moti", "socios"}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Fatalf("unexpected order at %d: got %s want %s", i, gotOrder[i], wantOrder[i])
		}
	}
}

func TestFormatReport(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	finish := start.Add(2 * time.Minute)
	rep := report{
		Month:      timeutil.YearMonth{Year: 2026, Month: 1},
		MonthURL:   "https://example.test/2026-01/",
		StartedAt:  start,
		FinishedAt: finish,
		UTCOffset:  "-04:00",
		Downloaded: 3,
		Extracted:  2,
		LoadedRows: map[string]int64{
			"socios":  20,
			"empresa": 10,
		},
	}

	out := formatReport(rep)
	required := []string{
		"RFCNPJ Loader - Finalizado",
		"Mês: Janeiro de 2026 (2026-01)",
		"URL: https://example.test/2026-01/",
		"Início: 2026-01-01T06:00:00-04:00",
		"Fim: 2026-01-01T06:02:00-04:00",
		"Downloads planejados: 3",
		"Arquivos extraídos: 2",
		"- empresa: 10",
		"- socios: 20",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Fatalf("report missing %q\nreport:\n%s", s, out)
		}
	}
}
