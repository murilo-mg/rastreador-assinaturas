// Expõe os endpoints HTTP de relatórios.
package relatorio

import (
	"context"
	"net/http"
	"rastreador-assinaturas/internal/httpjson"
	"strconv"
)

type Consultas interface {
	GastoMensalTotal(context.Context) (GastoMensal, error)
	ProjecaoAnual(context.Context) (ProjecaoAnual, error)
	ProximosVencimentos(context.Context, int) ([]ProximoVencimento, error)
}

type Handler struct{ repositorio Consultas }

func NovoHandler(repositorio Consultas) *Handler { return &Handler{repositorio: repositorio} }

func (h *Handler) RegistrarRotas(mux *http.ServeMux) {
	mux.HandleFunc("/relatorios/gasto-mensal", h.gastoMensal)
	mux.HandleFunc("/relatorios/proximos-vencimentos", h.proximosVencimentos)
	mux.HandleFunc("/relatorios/projecao-anual", h.projecaoAnual)
}

func somenteGet(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}
	w.Header().Set("Allow", "GET")
	httpjson.Erro(w, http.StatusMethodNotAllowed, "Método não permitido.")
	return false
}

func (h *Handler) gastoMensal(w http.ResponseWriter, r *http.Request) {
	if !somenteGet(w, r) {
		return
	}
	gasto, err := h.repositorio.GastoMensalTotal(r.Context())
	if err != nil {
		httpjson.Interno(w, err)
		return
	}
	httpjson.Responder(w, http.StatusOK, gasto)
}

func (h *Handler) projecaoAnual(w http.ResponseWriter, r *http.Request) {
	if !somenteGet(w, r) {
		return
	}
	projecao, err := h.repositorio.ProjecaoAnual(r.Context())
	if err != nil {
		httpjson.Interno(w, err)
		return
	}
	httpjson.Responder(w, http.StatusOK, projecao)
}

func (h *Handler) proximosVencimentos(w http.ResponseWriter, r *http.Request) {
	if !somenteGet(w, r) {
		return
	}
	dias := 7
	if r.URL.Query().Has("dias") {
		var err error
		dias, err = strconv.Atoi(r.URL.Query().Get("dias"))
		if err != nil || dias < 1 || dias > 365 {
			httpjson.Erro(w, http.StatusBadRequest, "Informe dias entre 1 e 365.")
			return
		}
	}
	vencimentos, err := h.repositorio.ProximosVencimentos(r.Context(), dias)
	if err != nil {
		httpjson.Interno(w, err)
		return
	}
	if vencimentos == nil {
		vencimentos = []ProximoVencimento{}
	}
	httpjson.Responder(w, http.StatusOK, vencimentos)
}
