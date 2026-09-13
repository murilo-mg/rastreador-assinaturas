// Expõe os endpoints HTTP de relatórios sobre assinaturas.
package relatorio

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const diasPadraoAntecedencia = 7

type Handler struct {
	repositorio *Repositorio
}

func NovoHandler(repositorio *Repositorio) *Handler {
	return &Handler{repositorio: repositorio}
}

func (h *Handler) RegistrarRotas(mux *http.ServeMux) {
	mux.HandleFunc("/relatorios/gasto-mensal", h.gastoMensal)
	mux.HandleFunc("/relatorios/proximos-vencimentos", h.proximosVencimentos)
	mux.HandleFunc("/relatorios/projecao-anual", h.projecaoAnual)
}

func (h *Handler) gastoMensal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	gasto, err := h.repositorio.GastoMensalTotal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gasto)
}

func (h *Handler) projecaoAnual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	projecao, err := h.repositorio.ProjecaoAnual()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projecao)
}

func (h *Handler) proximosVencimentos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	dias := lerParametroDias(r)

	vencimentos, err := h.repositorio.ProximosVencimentos(dias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vencimentos)
}

func lerParametroDias(r *http.Request) int {
	valor := r.URL.Query().Get("dias")
	if valor == "" {
		return diasPadraoAntecedencia
	}

	dias, err := strconv.Atoi(valor)
	if err != nil || dias <= 0 {
		return diasPadraoAntecedencia
	}

	return dias
}