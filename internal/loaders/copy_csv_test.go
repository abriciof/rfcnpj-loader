package loaders

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSVSource_PadsAndTruncates(t *testing.T) {
	t.Parallel()

	reader := csv.NewReader(strings.NewReader("a;b\nx;y;z\n"))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	src := &csvCopySource{r: reader, cols: 2}

	if !src.Next() {
		t.Fatal("expected first row")
	}
	v, err := src.Values()
	if err != nil {
		t.Fatalf("Values returned error: %v", err)
	}
	if len(v) != 2 || v[0] != "a" || v[1] != "b" {
		t.Fatalf("unexpected first row values: %#v", v)
	}

	if !src.Next() {
		t.Fatal("expected second row")
	}
	v, err = src.Values()
	if err != nil {
		t.Fatalf("Values returned error: %v", err)
	}
	if len(v) != 2 || v[0] != "x" || v[1] != "y" {
		t.Fatalf("unexpected second row values: %#v", v)
	}
}

func TestCSVSource_Err(t *testing.T) {
	t.Parallel()

	// invalid CSV to force reader error
	reader := csv.NewReader(strings.NewReader("\"a\n"))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	src := &csvCopySource{r: reader, cols: 1}
	if src.Next() {
		t.Fatal("expected Next=false on malformed CSV")
	}
	if src.Err() == nil {
		t.Fatal("expected Err() to return parse error")
	}
}

