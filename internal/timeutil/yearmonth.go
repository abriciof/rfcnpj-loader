package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type YearMonth struct {
	Year  int
	Month time.Month
}

func ParseYearMonth(s string) (YearMonth, error) {
	parts := strings.Split(strings.TrimSpace(s), "-")
	if len(parts) != 2 {
		return YearMonth{}, fmt.Errorf("formato inválido (esperado YYYY-MM): %q", s)
	}
	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return YearMonth{}, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return YearMonth{}, err
	}
	if m < 1 || m > 12 {
		return YearMonth{}, fmt.Errorf("mês inválido: %d", m)
	}
	return YearMonth{Year: y, Month: time.Month(m)}, nil
}

func (ym YearMonth) String() string {
	return fmt.Sprintf("%04d-%02d", ym.Year, int(ym.Month))
}

func (ym YearMonth) Next() YearMonth {
	y := ym.Year
	m := ym.Month + 1
	if m > 12 {
		m = 1
		y++
	}
	return YearMonth{Year: y, Month: m}
}

func (ym YearMonth) HumanPTBR() string {
	nomes := []string{
		"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
	}
	return fmt.Sprintf("%s de %d", nomes[int(ym.Month)], ym.Year)
}
