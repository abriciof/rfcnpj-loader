package loaders

import "testing"

func TestTableSpecs_AreDefined(t *testing.T) {
	t.Parallel()

	specs := []TableSpec{
		Empresa, Estabelecimento, Socios, Simples, Cnae,
		Moti, Munic, Natju, Pais, Quals,
	}

	for _, s := range specs {
		if s.Name == "" {
			t.Fatal("spec name cannot be empty")
		}
		if len(s.Columns) == 0 {
			t.Fatalf("spec %s must have columns", s.Name)
		}
	}
}

