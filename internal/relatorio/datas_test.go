package relatorio

import (
	"testing"
	"time"
)

func TestProximaCobranca(t *testing.T) {
	for _, tt := range []struct {
		nome, hoje string
		dia        int
		esperada   string
	}{
		{"hoje", "2026-10-03", 3, "2026-10-03"},
		{"este mes", "2026-10-03", 10, "2026-10-10"},
		{"mes seguinte", "2026-10-03", 1, "2026-11-01"},
		{"virada do ano", "2026-12-31", 1, "2027-01-01"},
		{"dia31 em abril", "2026-04-29", 31, "2026-04-30"},
		{"dia31 no ultimo dia", "2026-04-30", 31, "2026-04-30"},
		{"dia31 seguinte", "2026-05-01", 31, "2026-05-31"},
		{"fevereiro comum", "2026-02-01", 31, "2026-02-28"},
		{"fevereiro dia29", "2026-02-01", 29, "2026-02-28"},
		{"bissexto", "2024-02-01", 31, "2024-02-29"},
		{"dia29 bissexto", "2024-02-28", 29, "2024-02-29"},
		{"janeiro para fevereiro", "2026-01-31", 30, "2026-02-28"},
	} {
		t.Run(tt.nome, func(t *testing.T) {
			hoje, _ := time.Parse("2006-01-02", tt.hoje)
			obtida := proximaCobranca(tt.dia, hoje).Format("2006-01-02")
			if obtida != tt.esperada {
				t.Fatalf("%s != %s", obtida, tt.esperada)
			}
		})
	}
}

func TestDataCivilPreservaFuso(t *testing.T) {
	local := time.FixedZone("Manaus", -4*3600)
	hoje := time.Date(2026, 10, 3, 23, 59, 0, 0, local)
	if dataCivil(hoje).Format("2006-01-02") != "2026-10-03" {
		t.Fatal("dia local convertido indevidamente para dia UTC")
	}
}
