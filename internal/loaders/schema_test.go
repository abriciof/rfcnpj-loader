package loaders

import (
	"strings"
	"testing"
)

func TestCreateTableSQL(t *testing.T) {
	t.Parallel()

	sql := CreateTableSQL(TableSpec{
		Name:    "empresa",
		Columns: []string{"cnpj_basico", "razao_social"},
	})

	if !strings.Contains(sql, `CREATE TABLE IF NOT EXISTS "empresa"`) {
		t.Fatalf("unexpected SQL prefix: %s", sql)
	}
	if !strings.Contains(sql, `"cnpj_basico" TEXT`) || !strings.Contains(sql, `"razao_social" TEXT`) {
		t.Fatalf("expected TEXT columns in SQL: %s", sql)
	}
	if !strings.HasSuffix(sql, ");") {
		t.Fatalf("expected SQL to end with ); got %s", sql)
	}
}

