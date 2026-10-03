package assinatura

import (
	"math"
	"strings"
	"testing"
)

func TestValidacao(t *testing.T) {
	testes := []struct {
		nome     string
		mudar    func(*Assinatura)
		invalida bool
	}{
		{"valida", func(a *Assinatura) {}, false},
		{"gratuita", func(a *Assinatura) { a.Valor = 0 }, false},
		{"nome vazio", func(a *Assinatura) { a.Nome = "   " }, true},
		{"nome longo", func(a *Assinatura) { a.Nome = strings.Repeat("á", 101) }, true},
		{"unicode no limite", func(a *Assinatura) { a.Nome = strings.Repeat("á", 100) }, false},
		{"categoria longa", func(a *Assinatura) { a.Categoria = strings.Repeat("a", 51) }, true},
		{"valor negativo", func(a *Assinatura) { a.Valor = -1 }, true},
		{"valor fracionado", func(a *Assinatura) { a.Valor = 1.001 }, true},
		{"centavos comuns", func(a *Assinatura) { a.Valor = 19.99 }, false},
		{"limite decimal", func(a *Assinatura) { a.Valor = 99999999.99 }, false},
		{"valor acima do banco", func(a *Assinatura) { a.Valor = 100000000 }, true},
		{"NaN", func(a *Assinatura) { a.Valor = math.NaN() }, true},
		{"infinito", func(a *Assinatura) { a.Valor = math.Inf(1) }, true},
		{"dia zero", func(a *Assinatura) { a.DiaCobranca = 0 }, true},
		{"dia 32", func(a *Assinatura) { a.DiaCobranca = 32 }, true},
		{"dia 31", func(a *Assinatura) { a.DiaCobranca = 31 }, false},
	}
	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			a := Assinatura{Nome: "Curso", Valor: 39.90, DiaCobranca: 10}
			teste.mudar(&a)
			if err := a.Validar(); (err != nil) != teste.invalida {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
	a := Assinatura{Nome: "  Academia  ", Categoria: "  Saúde  ", Valor: 1, DiaCobranca: 1}
	if err := a.Validar(); err != nil {
		t.Fatal(err)
	}
	if a.Nome != "Academia" || a.Categoria != "Saúde" {
		t.Fatal("espaços não foram removidos")
	}
}
