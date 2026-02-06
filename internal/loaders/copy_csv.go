package loaders

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type CopyResult struct {
	Table string
	File  string
	Rows  int64
}

func EnsureTable(ctx context.Context, db *sql.DB, spec TableSpec, drop bool) error {
	if drop {
		if _, err := db.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS "%s";`, spec.Name)); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, CreateTableSQL(spec)); err != nil {
		return err
	}
	return nil
}

// CopyCSV streams a ';' separated (latin-1) file into Postgres via pgx CopyFrom.
// All columns are treated as TEXT.
// This replaces pandas to_sql chunking with faster streaming.
func CopyCSV(ctx context.Context, db *sql.DB, spec TableSpec, csvPath string) (CopyResult, error) {
	sqlConn, err := db.Conn(ctx)
	if err != nil {
		return CopyResult{}, err
	}
	defer sqlConn.Close()

	f, err := os.Open(csvPath)
	if err != nil {
		return CopyResult{}, err
	}
	defer f.Close()

	reader := csv.NewReader(transform.NewReader(f, charmap.ISO8859_1.NewDecoder()))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	src := &csvCopySource{
		r:       reader,
		cols:    len(spec.Columns),
	}

	var rows int64
	err = sqlConn.Raw(func(driverConn any) error {
		stdConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type %T", driverConn)
		}
		var copyErr error
		rows, copyErr = stdConn.Conn().CopyFrom(ctx, pgx.Identifier{spec.Name}, spec.Columns, src)
		return copyErr
	})
	if err != nil {
		return CopyResult{}, fmt.Errorf("copy %s (%s): %w", spec.Name, csvPath, err)
	}

	return CopyResult{Table: spec.Name, File: csvPath, Rows: rows}, nil
}

type csvCopySource struct {
	r    *csv.Reader
	cols int
	row  []string
	err  error
}

func (s *csvCopySource) Next() bool {
	rec, err := s.r.Read()
	if err == io.EOF {
		return false
	}
	if err != nil {
		s.err = err
		return false
	}

	// Ajusta número de colunas: se vier menos, completa com "".
	// Se vier mais, trunca.
	if len(rec) < s.cols {
		padded := make([]string, s.cols)
		copy(padded, rec)
		rec = padded
	} else if len(rec) > s.cols {
		rec = rec[:s.cols]
	}

	s.row = rec
	return true
}

func (s *csvCopySource) Values() ([]any, error) {
	out := make([]any, s.cols)
	for i := 0; i < s.cols; i++ {
		v := strings.TrimSpace(s.row[i])
		// mantém como texto; se quiser NULL real, troque para nil quando v == ""
		out[i] = v
	}
	return out, nil
}

func (s *csvCopySource) Err() error { return s.err }
