package relatorio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type consultasTeste struct{ dias int }

func (c *consultasTeste) GastoMensalTotal(context.Context) (GastoMensal, error) {
	return GastoMensal{Total: 15.9}, nil
}
func (c *consultasTeste) ProjecaoAnual(context.Context) (ProjecaoAnual, error) {
	return ProjecaoAnual{GastoMensal: 15.9, GastoAnual: 190.8}, nil
}
func (c *consultasTeste) ProximosVencimentos(_ context.Context, dias int) ([]ProximoVencimento, error) {
	c.dias = dias
	return nil, nil
}

func TestParametroDias(t *testing.T) {
	for _, tt := range []struct {
		consulta     string
		status, dias int
	}{
		{"", 200, 7}, {"?dias=1", 200, 1}, {"?dias=365", 200, 365},
		{"?dias=0", 400, 0}, {"?dias=-1", 400, 0}, {"?dias=366", 400, 0}, {"?dias=abc", 400, 0}, {"?dias=", 400, 0},
	} {
		t.Run(tt.consulta, func(t *testing.T) {
			c := &consultasTeste{}
			mux := http.NewServeMux()
			NovoHandler(c).RegistrarRotas(mux)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("GET", "/relatorios/proximos-vencimentos"+tt.consulta, nil))
			if w.Code != tt.status || c.dias != tt.dias {
				t.Fatalf("status %d, dias %d", w.Code, c.dias)
			}
			if tt.status == 200 && w.Body.String() != "[]\n" {
				t.Fatal("lista vazia deve retornar []")
			}
		})
	}
}

func TestRelatoriosRecusamEscrita(t *testing.T) {
	for _, rota := range []string{"gasto-mensal", "projecao-anual", "proximos-vencimentos"} {
		mux := http.NewServeMux()
		NovoHandler(&consultasTeste{}).RegistrarRotas(mux)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/relatorios/"+rota, nil))
		if w.Code != 405 || w.Header().Get("Allow") != "GET" {
			t.Fatal("relatório aceitou método incorreto")
		}
	}
}
