package timeutil

import "testing"

func TestParseYearMonth(t *testing.T) {
	t.Parallel()

	got, err := ParseYearMonth("2026-01")
	if err != nil {
		t.Fatalf("ParseYearMonth returned error: %v", err)
	}
	if got.Year != 2026 || int(got.Month) != 1 {
		t.Fatalf("unexpected year/month: %+v", got)
	}
}

func TestParseYearMonth_Invalid(t *testing.T) {
	t.Parallel()

	cases := []string{"", "2026", "2026-13", "2026-00", "abc-01"}
	for _, in := range cases {
		_, err := ParseYearMonth(in)
		if err == nil {
			t.Fatalf("expected error for input %q", in)
		}
	}
}

func TestYearMonthNextAndString(t *testing.T) {
	t.Parallel()

	got := (YearMonth{Year: 2025, Month: 12}).Next()
	if got.String() != "2026-01" {
		t.Fatalf("unexpected next/string: %s", got.String())
	}
}

func TestYearMonthHumanPTBR(t *testing.T) {
	t.Parallel()

	got := (YearMonth{Year: 2026, Month: 3}).HumanPTBR()
	if got != "Março de 2026" {
		t.Fatalf("unexpected HumanPTBR: %q", got)
	}
}

