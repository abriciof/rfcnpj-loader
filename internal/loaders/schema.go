package loaders

import "strings"

type TableSpec struct {
	Name    string
	Columns []string
}

// All columns as TEXT for resilience (parsing can be added later).
func CreateTableSQL(t TableSpec) string {
	var sb strings.Builder
	sb.WriteString(`CREATE TABLE IF NOT EXISTS "`)
	sb.WriteString(t.Name)
	sb.WriteString(`" (`)
	for i, c := range t.Columns {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`"`)
		sb.WriteString(c)
		sb.WriteString(`" TEXT`)
	}
	sb.WriteString(");")
	return sb.String()
}
