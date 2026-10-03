package assinatura

import (
	"encoding/csv"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func lerCSVTeste(t *testing.T, corpo string) [][]string {
	t.Helper()
	if !strings.HasPrefix(corpo, "\uFEFF") {
		t.Fatal("BOM UTF-8 ausente")
	}
	leitor := csv.NewReader(strings.NewReader(strings.TrimPrefix(corpo, "\uFEFF")))
	leitor.Comma = ';'
	linhas, err := leitor.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return linhas
}

func TestExportacaoCSV(t *testing.T) {
	m := &memoria{lista: []Assinatura{
		{ID: 2, Nome: "Educação; \"Mensal\"\nCurso", Valor: 39.9, Categoria: "Saúde", DiaCobranca: 31, Ativa: true},
		{ID: 3, Nome: "Grátis", Valor: 0, DiaCobranca: 1, Ativa: false},
		{ID: 4, Nome: "Limite", Valor: 99999999.99, Categoria: "=1+1\";@SUM(1)", DiaCobranca: 10, Ativa: true},
		{ID: 5, Nome: "=1+1", Valor: 0.01, DiaCobranca: 20, Ativa: false},
	}}
	w := enviar(m, "GET", "/assinaturas/exportar.csv", "", "")
	if w.Code != 200 || w.Header().Get("Content-Type") != "text/csv; charset=utf-8" ||
		w.Header().Get("Content-Disposition") != `attachment; filename="assinaturas.csv"` || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("resposta inválida: %d %v", w.Code, w.Header())
	}
	linhas := lerCSVTeste(t, w.Body.String())
	esperado := [][]string{
		{"ID", "Nome", "Valor mensal (R$)", "Categoria", "Dia da cobrança", "Ativa"},
		{"2", "Educação; \"Mensal\"\nCurso", "39,90", "Saúde", "31", "Sim"},
		{"3", "Grátis", "0,00", "", "1", "Não"},
		{"4", "Limite", "99999999,99", "\t=1+1\";@SUM(1)", "10", "Sim"},
		{"5", "\t=1+1", "0,01", "", "20", "Não"},
	}
	if !reflect.DeepEqual(linhas, esperado) {
		t.Fatalf("CSV diferente: %#v", linhas)
	}
	if !strings.Contains(w.Body.String(), "\"\t=1+1\"") || !strings.Contains(w.Body.String(), "\"\t=1+1\"\";@SUM(1)\"") {
		t.Fatal("prefixo de fórmula deve ficar dentro do campo entre aspas")
	}
	if m.lista[3].Nome != "=1+1" || m.lista[2].Categoria != "=1+1\";@SUM(1)" || m.criacoes != 0 || m.alteracoes != 0 || m.estados != 0 {
		t.Fatal("exportação alterou os registros")
	}
}

func TestTextoCSV(t *testing.T) {
	for _, texto := range []string{
		"=1+1", "+1", "-1", "@SUM(1)", "＝1", "＋1", "－1", "＠SUM(1)",
		" \t\r\n=1+1", "\uFEFF=1+1", "\u200B=1+1", "\x00=1+1",
	} {
		t.Run(texto, func(t *testing.T) {
			if obtido := textoCSV(texto); obtido != "\t"+texto {
				t.Fatalf("prefixo não neutralizado: %q", obtido)
			}
		})
	}
	for _, texto := range []string{"", "Educação", "Plano + família", "Curso;=1+1", "\";=1+1", "Curso\n=1+1"} {
		if textoCSV(texto) != texto {
			t.Fatalf("texto comum mudou: %q", texto)
		}
	}
}

func TestExportacaoVaziaEMetodos(t *testing.T) {
	m := &memoria{}
	w := enviar(m, "GET", "/assinaturas/exportar.csv", "", "")
	if w.Code != 200 || len(lerCSVTeste(t, w.Body.String())) != 1 {
		t.Fatal("lista vazia deve gerar somente o cabeçalho")
	}
	for _, metodo := range []string{"POST", "PUT", "PATCH", "DELETE", "HEAD"} {
		m := &memoria{}
		w := enviar(m, metodo, "/assinaturas/exportar.csv", "", "")
		if w.Code != 405 || w.Header().Get("Allow") != "GET" || m.listagens != 0 {
			t.Fatalf("método %s aceito ou consultou o banco", metodo)
		}
	}
	w = enviar(&memoria{erro: errors.New("password=privado")}, "GET", "/assinaturas/exportar.csv", "", "")
	if w.Code != 500 || w.Header().Get("Content-Disposition") != "" || strings.Contains(w.Body.String(), "privado") || strings.HasPrefix(w.Body.String(), "\uFEFF") {
		t.Fatal("falha no banco gerou download ou expôs erro interno")
	}
}
