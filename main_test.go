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
	"time"

	_ "github.com/lib/pq"

	"rastreador-assinaturas/internal/assinatura"
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
	// Editar preserva identidade/data de criação e atualiza todos os campos editáveis.
	var criada time.Time
	if err := banco.QueryRow("SELECT criado_em FROM assinaturas WHERE id=1").Scan(&criada); err != nil {
		t.Fatal(err)
	}
	if w := enviar("PUT", "/assinaturas/1", `{"nome":"Editada","valor":4.40,"categoria":"Teste","dia_cobranca":31,"ativa":false}`); w.Code != 204 {
		t.Fatalf("edição: %d %s", w.Code, w.Body.String())
	}
	var depois time.Time
	if err := banco.QueryRow("SELECT criado_em FROM assinaturas WHERE id=1").Scan(&depois); err != nil || !depois.Equal(criada) {
		t.Fatal("edição mudou a data de criação")
	}
	listar := func() []assinatura.Assinatura {
		t.Helper()
		w := enviar("GET", "/assinaturas", "")
		var lista []assinatura.Assinatura
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &lista) != nil {
			t.Fatal("listagem inválida")
		}
		return lista
	}
	lista := listar()
	if len(lista) != 3 || lista[0].ID != 1 || lista[0].Nome != "Editada" || lista[0].Valor != 4.4 || lista[0].Categoria != "Teste" || lista[0].DiaCobranca != 31 || lista[0].Ativa {
		t.Fatalf("edição não persistida: %+v", lista)
	}
	for _, tt := range []struct {
		ativa        string
		total, anual float64
	}{
		{"false", 2.2, 26.4}, {"true", 6.6, 79.2}, {"true", 6.6, 79.2}, {"false", 2.2, 26.4},
	} {
		if w := enviar("PATCH", "/assinaturas/1", `{"ativa":`+tt.ativa+`}`); w.Code != 204 {
			t.Fatalf("estado: %d %s", w.Code, w.Body.String())
		}
		w := enviar("GET", "/relatorios/gasto-mensal", "")
		if json.Unmarshal(w.Body.Bytes(), &mensal) != nil || mensal.Total != tt.total {
			t.Fatalf("total após estado: %+v", mensal)
		}
		w = enviar("GET", "/relatorios/projecao-anual", "")
		if json.Unmarshal(w.Body.Bytes(), &anual) != nil || anual.GastoAnual != tt.anual {
			t.Fatalf("anual após estado: %+v", anual)
		}
		w = enviar("GET", "/relatorios/proximos-vencimentos?dias=365", "")
		var vencimentos []struct {
			ID          int
			Nome        string
			Valor       float64
			DiaCobranca int
		}
		if json.Unmarshal(w.Body.Bytes(), &vencimentos) != nil {
			t.Fatal("vencimentos inválidos")
		}
		achou := false
		for _, v := range vencimentos {
			if v.ID == 1 {
				achou = true
				if v.Nome != "Editada" || v.Valor != 4.4 || v.DiaCobranca != 31 {
					t.Fatal("vencimento não atualizado")
				}
			}
			if v.ID == 3 {
				t.Fatal("inativa nos vencimentos")
			}
		}
		if achou != (tt.ativa == "true") {
			t.Fatal("estado não refletido nos vencimentos")
		}
		a := listar()[0]
		if a.Nome != "Editada" || a.Valor != 4.4 || a.Categoria != "Teste" || a.DiaCobranca != 31 || a.Ativa != (tt.ativa == "true") {
			t.Fatal("PATCH alterou outros campos")
		}
	}
	for _, metodo := range []string{"PUT", "PATCH"} {
		corpo := `{"ativa":true}`
		if metodo == "PUT" {
			corpo = `{"nome":"Ausente","valor":1,"dia_cobranca":1,"ativa":true}`
		}
		if w := enviar(metodo, "/assinaturas/999", corpo); w.Code != 404 {
			t.Fatal("alteração de inexistente deveria retornar 404")
		}
	}
	if w := enviar("PUT", "/assinaturas/1", `{"nome":"Inválida","valor":-1,"dia_cobranca":1,"ativa":true}`); w.Code != 400 {
		t.Fatal("edição inválida aceita")
	}
	if listar()[0].Nome != "Editada" {
		t.Fatal("edição inválida mudou o banco")
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
