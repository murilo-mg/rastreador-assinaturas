package assinatura

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type memoria struct {
	criada     Assinatura
	criacoes   int
	erro       error
	atualizada Assinatura
	alteracoes int
	estados    int
	ativa      bool
	lista      []Assinatura
	listagens  int
}

func (m *memoria) Listar(context.Context) ([]Assinatura, error) {
	m.listagens++
	return m.lista, m.erro
}
func (m *memoria) Criar(_ context.Context, a Assinatura) (int, error) {
	m.criada = a
	m.criacoes++
	return 8, m.erro
}
func (m *memoria) Remover(_ context.Context, id int) error {
	if id == 999 {
		return ErrNaoEncontrada
	}
	return m.erro
}

func (m *memoria) Atualizar(_ context.Context, a Assinatura) error {
	m.atualizada = a
	m.alteracoes++
	if a.ID == 999 {
		return ErrNaoEncontrada
	}
	return m.erro
}

func (m *memoria) DefinirAtiva(_ context.Context, id int, ativa bool) error {
	m.estados++
	m.ativa = ativa
	if id == 999 {
		return ErrNaoEncontrada
	}
	return m.erro
}

func enviar(m *memoria, metodo, caminho, corpo, tipo string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NovoHandler(m).RegistrarRotas(mux)
	r := httptest.NewRequest(metodo, caminho, strings.NewReader(corpo))
	r.Header.Set("Content-Type", tipo)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

func TestCadastroHTTP(t *testing.T) {
	testes := []struct {
		nome, corpo, tipo string
		status            int
	}{
		{"valido", `{"nome":"Curso","valor":39.9,"dia_cobranca":10}`, "application/json", 201},
		{"inativo", `{"nome":"Curso","valor":39.9,"dia_cobranca":10,"ativa":false}`, "application/json", 201},
		{"mime ausente", `{}`, "", 415},
		{"mime incorreto", `{}`, "text/plain", 415},
		{"json quebrado", `{`, "application/json", 400},
		{"null", `null`, "application/json", 400},
		{"objeto vazio", `{}`, "application/json", 400},
		{"campo desconhecido", `{"nome":"Curso","valor":1,"dia_cobranca":10,"extra":1}`, "application/json", 400},
		{"id enviado", `{"nome":"Curso","valor":1,"dia_cobranca":10,"id":12}`, "application/json", 400},
		{"dois objetos", `{"nome":"Curso","valor":1,"dia_cobranca":10} {}`, "application/json", 400},
		{"nome vazio", `{"nome":" ","valor":1,"dia_cobranca":10}`, "application/json", 400},
		{"valor ausente", `{"nome":"Curso","dia_cobranca":10}`, "application/json", 400},
		{"valor null", `{"nome":"Curso","valor":null,"dia_cobranca":10}`, "application/json", 400},
		{"negativo", `{"nome":"Curso","valor":-1,"dia_cobranca":10}`, "application/json", 400},
		{"dia 32", `{"nome":"Curso","valor":1,"dia_cobranca":32}`, "application/json", 400},
		{"corpo grande", `{"nome":"` + strings.Repeat("a", limiteCorpo) + `","valor":1,"dia_cobranca":10}`, "application/json", 413},
		{"espacos grandes", `{"nome":"Curso","valor":1,"dia_cobranca":10}` + strings.Repeat(" ", limiteCorpo), "application/json", 413},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			m := &memoria{}
			w := enviar(m, "POST", "/assinaturas", tt.corpo, tt.tipo)
			if w.Code != tt.status {
				t.Fatalf("status %d, queria %d: %s", w.Code, tt.status, w.Body.String())
			}
			if tt.status != 201 && m.criacoes != 0 {
				t.Fatal("entrada inválida chegou ao banco")
			}
			if tt.nome == "valido" && !m.criada.Ativa {
				t.Fatal("cadastro sem ativa deve começar ativo")
			}
			if tt.nome == "inativo" && m.criada.Ativa {
				t.Fatal("false explícito foi ignorado")
			}
		})
	}
}

func TestListagemERemocaoHTTP(t *testing.T) {
	m := &memoria{}
	if w := enviar(m, "GET", "/assinaturas", "", ""); w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatal("lista vazia deve retornar []")
	}
	for _, tt := range []struct {
		metodo, caminho string
		status          int
	}{
		{"DELETE", "/assinaturas/8", 204}, {"DELETE", "/assinaturas/999", 404},
		{"DELETE", "/assinaturas/-1", 400}, {"DELETE", "/assinaturas/0", 400},
		{"DELETE", "/assinaturas/8/extra", 400}, {"GET", "/assinaturas/8", 405}, {"PATCH", "/assinaturas", 405},
	} {
		t.Run(tt.metodo+tt.caminho, func(t *testing.T) {
			w := enviar(m, tt.metodo, tt.caminho, "", "")
			if w.Code != tt.status {
				t.Fatalf("status %d", w.Code)
			}
		})
	}
}

func TestErroInternoNaoVaza(t *testing.T) {
	m := &memoria{erro: errors.New("password=privado SELECT detalhes internos")}
	w := enviar(m, "GET", "/assinaturas", "", "")
	if w.Code != 500 || strings.Contains(w.Body.String(), "privado") {
		t.Fatal("erro interno exposto")
	}
}

func TestEdicaoHTTP(t *testing.T) {
	valido := `{"nome":" Curso atualizado ","valor":0,"categoria":" Educação ","dia_cobranca":31,"ativa":false}`
	for _, tt := range []struct {
		nome, caminho, corpo, tipo string
		status                     int
	}{
		{"valida", "/assinaturas/8", valido, "application/json", 204},
		{"inexistente", "/assinaturas/999", valido, "application/json", 404},
		{"id invalido", "/assinaturas/0", valido, "application/json", 400},
		{"rota extra", "/assinaturas/8/extra", valido, "application/json", 400},
		{"estado ausente", "/assinaturas/8", `{"nome":"Curso","valor":1,"dia_cobranca":10}`, "application/json", 400},
		{"estado null", "/assinaturas/8", `{"nome":"Curso","valor":1,"dia_cobranca":10,"ativa":null}`, "application/json", 400},
		{"valor ausente", "/assinaturas/8", `{"nome":"Curso","dia_cobranca":10,"ativa":true}`, "application/json", 400},
		{"nome vazio", "/assinaturas/8", `{"nome":" ","valor":1,"dia_cobranca":10,"ativa":true}`, "application/json", 400},
		{"decimal invalido", "/assinaturas/8", `{"nome":"Curso","valor":1.001,"dia_cobranca":10,"ativa":true}`, "application/json", 400},
		{"dia invalido", "/assinaturas/8", `{"nome":"Curso","valor":1,"dia_cobranca":32,"ativa":true}`, "application/json", 400},
		{"campo desconhecido", "/assinaturas/8", `{"nome":"Curso","valor":1,"dia_cobranca":10,"ativa":true,"id":3}`, "application/json", 400},
		{"mime", "/assinaturas/8", valido, "text/plain", 415},
		{"dois objetos", "/assinaturas/8", valido + ` {}`, "application/json", 400},
		{"limite", "/assinaturas/8", valido + strings.Repeat(" ", limiteCorpo), "application/json", 413},
	} {
		t.Run(tt.nome, func(t *testing.T) {
			m := &memoria{}
			w := enviar(m, "PUT", tt.caminho, tt.corpo, tt.tipo)
			if w.Code != tt.status {
				t.Fatalf("status %d, queria %d: %s", w.Code, tt.status, w.Body.String())
			}
			if tt.status != 204 && tt.status != 404 && m.alteracoes != 0 {
				t.Fatal("entrada inválida chegou ao banco")
			}
			if tt.status == 204 && (m.atualizada.ID != 8 || m.atualizada.Nome != "Curso atualizado" || m.atualizada.Categoria != "Educação" || m.atualizada.Ativa || m.criacoes != 0) {
				t.Fatalf("edição incorreta: %+v", m)
			}
		})
	}
}

func TestEstadoHTTP(t *testing.T) {
	for _, tt := range []struct {
		nome, corpo string
		status      int
	}{
		{"ativar", `{"ativa":true}`, 204}, {"desativar", `{"ativa":false}`, 204},
		{"ausente", `{}`, 400}, {"null", `{"ativa":null}`, 400}, {"objeto null", `null`, 400},
		{"string", `{"ativa":"false"}`, 400}, {"outro campo", `{"ativa":true,"nome":"Outro"}`, 400},
		{"dois objetos", `{"ativa":true} {}`, 400}, {"corpo grande", `{"ativa":true}` + strings.Repeat(" ", limiteCorpo), 413},
	} {
		t.Run(tt.nome, func(t *testing.T) {
			m := &memoria{}
			w := enviar(m, "PATCH", "/assinaturas/8", tt.corpo, "application/json")
			if w.Code != tt.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if m.alteracoes != 0 || m.criacoes != 0 {
				t.Fatal("PATCH alterou outros campos")
			}
			if tt.status == 204 && (m.estados != 1 || m.ativa != (tt.nome == "ativar")) {
				t.Fatal("estado incorreto")
			}
			if tt.status != 204 && m.estados != 0 {
				t.Fatal("estado inválido chegou ao banco")
			}
		})
	}
	if w := enviar(&memoria{}, "PATCH", "/assinaturas/999", `{"ativa":true}`, "application/json"); w.Code != 404 {
		t.Fatal("estado de registro inexistente deveria retornar 404")
	}
	for _, metodo := range []string{"PUT", "PATCH"} {
		corpo := `{"ativa":false}`
		if metodo == "PUT" {
			corpo = `{"nome":"Curso","valor":1,"dia_cobranca":1,"ativa":false}`
		}
		w := enviar(&memoria{erro: errors.New("password=privado")}, metodo, "/assinaturas/8", corpo, "application/json")
		if w.Code != 500 || strings.Contains(w.Body.String(), "privado") {
			t.Fatal("erro de alteração exposto")
		}
	}
}
