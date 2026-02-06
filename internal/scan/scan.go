package scan

import (
	"os"
	"path/filepath"
	"strings"
)

type FilesByType struct {
	Empresa         []string
	Estabelecimento []string
	Socios          []string
	Simples         []string
	Cnae            []string
	Moti            []string
	Munic           []string
	Natju           []string
	Pais            []string
	Quals           []string
}

func ScanExtracted(root string) (FilesByType, error) {
	var out FilesByType
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToUpper(filepath.Base(path))

		switch {
		case strings.Contains(name, "EMPRE"):
			out.Empresa = append(out.Empresa, path)
		case strings.Contains(name, "ESTABELE"):
			out.Estabelecimento = append(out.Estabelecimento, path)
		case strings.Contains(name, "SOCIO"):
			out.Socios = append(out.Socios, path)
		case strings.Contains(name, "SIMPLES"):
			out.Simples = append(out.Simples, path)
		case strings.Contains(name, "CNAE"):
			out.Cnae = append(out.Cnae, path)
		case strings.Contains(name, "MOTI") || strings.Contains(name, "MOTIVO"):
			out.Moti = append(out.Moti, path)
		case strings.Contains(name, "MUNIC"):
			out.Munic = append(out.Munic, path)
		case strings.Contains(name, "NATJU") || strings.Contains(name, "NATURE"):
			out.Natju = append(out.Natju, path)
		case strings.Contains(name, "PAIS"):
			out.Pais = append(out.Pais, path)
		case strings.Contains(name, "QUAL"):
			out.Quals = append(out.Quals, path)
		}
		return nil
	})
	return out, err
}
