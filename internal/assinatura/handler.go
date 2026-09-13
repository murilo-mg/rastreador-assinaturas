// Expõe os endpoints HTTP relacionados a assinaturas.
package assinatura

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	repositorio *Repositorio
}

func NovoHandler(repositorio *Repositorio) *Handler {
	return &Handler{repositorio: repositorio}
}

func (h *Handler) RegistrarRotas(mux *http.ServeMux) {
	mux.HandleFunc("/assinaturas", h.roteirarPorMetodo)
	mux.HandleFunc("/assinaturas/", h.remover)
}

func (h *Handler) roteirarPorMetodo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listar(w, r)
	case http.MethodPost:
		h.criar(w, r)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	assinaturas, err := h.repositorio.Listar()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assinaturas)
}

func (h *Handler) criar(w http.ResponseWriter, r *http.Request) {
	var a Assinatura
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	id, err := h.repositorio.Criar(a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *Handler) remover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	idTexto := strings.TrimPrefix(r.URL.Path, "/assinaturas/")
	id, err := strconv.Atoi(idTexto)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := h.repositorio.Remover(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}