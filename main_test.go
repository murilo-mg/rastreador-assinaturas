package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func TestPainelERecursos(t *testing.T) {
	h := criarHandler(nil)
	for _, tt := range []struct {
		caminho, trecho string
		status          int
	}{
		{"/", "Rastreador de Assinaturas", 200},
		{"/assets/app.js", "textContent", 200},
		{"/assets/app.css", ":root", 200},
		{"/assets/favicon.svg", "<svg", 200},
		{"/README.md", "", 404}, {"/main.go", "", 404}, {"/assets/", "", 404},
	} {
		t.Run(tt.caminho, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", tt.caminho, nil))
			if w.Code != tt.status || !strings.Contains(w.Body.String(), tt.trecho) {
				t.Fatalf("recurso %s: %d", tt.caminho, w.Code)
			}
			if w.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
				t.Fatal("cabeçalhos ausentes")
			}
		})
	}
}

func TestOrigemDaEscrita(t *testing.T) {
	for _, tt := range []struct {
		origem    string
		permitida bool
	}{
		{"", true}, {"http://localhost:8080", true}, {"http://outro.site", false}, {"null", false},
	} {
		t.Run(tt.origem, func(t *testing.T) {
			chamado := false
			h := proteger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { chamado = true; w.WriteHeader(204) }))
			r := httptest.NewRequest("POST", "http://localhost:8080/assinaturas", nil)
			r.Header.Set("Origin", tt.origem)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if chamado != tt.permitida {
				t.Fatal("origem não verificada")
			}
		})
	}
}

func TestAPIComBanco(t *testing.T) {
	conexao := os.Getenv("TEST_DATABASE_URL")
	if conexao == "" {
		t.Skip("defina TEST_DATABASE_URL para executar a integração com PostgreSQL")
	}
	banco, err := sql.Open("postgres", conexao)
	if err != nil {
		t.Fatal(err)
	}
	defer banco.Close()
	banco.SetMaxOpenConns(1)
	ctx := context.Background()
	schema, err := os.ReadFile("db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	// Tabela temporária: os dados existentes do banco não são alterados.
	comando := strings.Replace(string(schema), "CREATE TABLE IF NOT EXISTS", "CREATE TEMP TABLE", 1)
	if _, err := banco.ExecContext(ctx, comando); err != nil {
		t.Fatal(err)
	}
	defer banco.ExecContext(ctx, "DROP TABLE IF EXISTS pg_temp.assinaturas")
	h := criarHandler(banco)
	enviar := func(metodo, rota, corpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(metodo, rota, strings.NewReader(corpo))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, corpo := range []string{
		`{"nome":"Primeira","valor":1.10,"dia_cobranca":10}`,
		`{"nome":"Segunda","valor":2.20,"dia_cobranca":20,"ativa":true}`,
		`{"nome":"Inativa","valor":50,"dia_cobranca":15,"ativa":false}`,
	} {
		if w := enviar("POST", "/assinaturas", corpo); w.Code != 201 {
			t.Fatalf("cadastro: %d %s", w.Code, w.Body.String())
		}
	}
	w := enviar("GET", "/relatorios/gasto-mensal", "")
	var mensal struct{ Total float64 }
	if err := json.Unmarshal(w.Body.Bytes(), &mensal); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || mensal.Total != 3.3 {
		t.Fatalf("total mensal incorreto: %+v", mensal)
	}
	w = enviar("GET", "/relatorios/projecao-anual", "")
	var anual struct{ GastoAnual float64 }
	if err := json.Unmarshal(w.Body.Bytes(), &anual); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || anual.GastoAnual != 39.6 {
		t.Fatalf("projeção incorreta: %+v", anual)
	}
	if w := enviar("DELETE", "/assinaturas/1", ""); w.Code != 204 {
		t.Fatalf("remoção: %s", w.Body.String())
	}
	if w := enviar("DELETE", "/assinaturas/1", ""); w.Code != 404 {
		t.Fatal("remoção repetida deveria retornar 404")
	}
	if w := enviar("GET", "/saude", ""); w.Code != 200 {
		t.Fatal("saúde não disponível")
	}
}
