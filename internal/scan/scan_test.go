package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanExtracted_GroupsFilesByType(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := []string{
		"EMPRECSV",
		"ESTABELE_001.txt",
		"SOCIO_A.csv",
		"SIMPLES.zip.txt",
		"CNAE.txt",
		"MOTIVO.txt",
		"MUNICIPIOS.txt",
		"NATUREZA.txt",
		"PAISES.txt",
		"QUALIFICACOES.txt",
		"IGNORAR.txt",
	}

	for _, name := range files {
		p := filepath.Join(root, name)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file %s: %v", p, err)
		}
	}

	got, err := ScanExtracted(root)
	if err != nil {
		t.Fatalf("ScanExtracted returned error: %v", err)
	}

	if len(got.Empresa) != 1 ||
		len(got.Estabelecimento) != 1 ||
		len(got.Socios) != 1 ||
		len(got.Simples) != 1 ||
		len(got.Cnae) != 1 ||
		len(got.Moti) != 1 ||
		len(got.Munic) != 1 ||
		len(got.Natju) != 1 ||
		len(got.Pais) != 1 ||
		len(got.Quals) != 1 {
		t.Fatalf("unexpected grouping counts: %+v", got)
	}
}

