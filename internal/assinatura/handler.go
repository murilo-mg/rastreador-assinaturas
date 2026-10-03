// Expõe os endpoints HTTP relacionados a assinaturas.
package assinatura

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"rastreador-assinaturas/internal/httpjson"
)

const limiteCorpo = 64 * 1024

type cadastro struct {
	Nome        string   `json:"nome"`
	Valor       *float64 `json:"valor"`
	Categoria   string   `json:"categoria"`
	DiaCobranca *int     `json:"dia_cobranca"`
	Ativa       *bool    `json:"ativa"`
}

type Armazenamento interface {
	Listar(context.Context) ([]Assinatura, error)
	Criar(context.Context, Assinatura) (int, error)
	Atualizar(context.Context, Assinatura) error
	DefinirAtiva(context.Context, int, bool) error
	Remover(context.Context, int) error
}

type Handler struct{ repositorio Armazenamento }

func NovoHandler(repositorio Armazenamento) *Handler { return &Handler{repositorio: repositorio} }

func (h *Handler) RegistrarRotas(mux *http.ServeMux) {
	mux.HandleFunc("/assinaturas", h.roteirarPorMetodo)
	mux.HandleFunc("/assinaturas/exportar.csv", h.exportarCSV)
	mux.HandleFunc("/assinaturas/", h.roteirarRegistro)
}

func (h *Handler) roteirarPorMetodo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		assinaturas, err := h.repositorio.Listar(r.Context())
		if err != nil {
			httpjson.Interno(w, err)
			return
		}
		if assinaturas == nil {
			assinaturas = []Assinatura{}
		}
		httpjson.Responder(w, http.StatusOK, assinaturas)
	case http.MethodPost:
		h.criar(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		httpjson.Erro(w, http.StatusMethodNotAllowed, "Método não permitido.")
	}
}

func lerJSON(w http.ResponseWriter, r *http.Request, entrada any) bool {
	tipo, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || tipo != "application/json" {
		httpjson.Erro(w, http.StatusUnsupportedMediaType, "Envie o cadastro como application/json.")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limiteCorpo)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(entrada); err != nil {
		var tamanho *http.MaxBytesError
		if errors.As(err, &tamanho) {
			httpjson.Erro(w, http.StatusRequestEntityTooLarge, "O cadastro excedeu 64 KiB.")
		} else {
			httpjson.Erro(w, http.StatusBadRequest, "Envie um objeto JSON válido com os campos do cadastro.")
		}
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		var tamanho *http.MaxBytesError
		if errors.As(err, &tamanho) {
			httpjson.Erro(w, http.StatusRequestEntityTooLarge, "O cadastro excedeu 64 KiB.")
		} else {
			httpjson.Erro(w, http.StatusBadRequest, "Envie somente um objeto JSON.")
		}
		return false
	}
	return true
}

func lerAssinatura(w http.ResponseWriter, r *http.Request, edicao bool) (Assinatura, bool) {
	var entrada cadastro
	if !lerJSON(w, r, &entrada) {
		return Assinatura{}, false
	}
	if entrada.Valor == nil || entrada.DiaCobranca == nil {
		httpjson.Erro(w, http.StatusBadRequest, "Informe o valor e o dia da cobrança.")
		return Assinatura{}, false
	}
	if edicao && entrada.Ativa == nil {
		httpjson.Erro(w, http.StatusBadRequest, "Informe se a assinatura está ativa.")
		return Assinatura{}, false
	}
	// Ausência de 'ativa' significa ativa; false explícito continua sendo respeitado.
	a := Assinatura{Nome: entrada.Nome, Valor: *entrada.Valor, Categoria: entrada.Categoria,
		DiaCobranca: *entrada.DiaCobranca, Ativa: true}
	if entrada.Ativa != nil {
		a.Ativa = *entrada.Ativa
	}
	if err := a.Validar(); err != nil {
		httpjson.Erro(w, http.StatusBadRequest, err.Error())
		return Assinatura{}, false
	}
	return a, true
}

func (h *Handler) criar(w http.ResponseWriter, r *http.Request) {
	a, valido := lerAssinatura(w, r, false)
	if !valido {
		return
	}
	id, err := h.repositorio.Criar(r.Context(), a)
	if err != nil {
		httpjson.Interno(w, err)
		return
	}
	w.Header().Set("Location", "/assinaturas/"+strconv.Itoa(id))
	httpjson.Responder(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) roteirarRegistro(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PUT, PATCH, DELETE")
		httpjson.Erro(w, http.StatusMethodNotAllowed, "Método não permitido.")
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/assinaturas/"))
	if err != nil || id <= 0 {
		httpjson.Erro(w, http.StatusBadRequest, "Id inválido.")
		return
	}
	switch r.Method {
	case http.MethodDelete:
		err = h.repositorio.Remover(r.Context(), id)
	case http.MethodPut:
		a, valido := lerAssinatura(w, r, true)
		if !valido {
			return
		}
		a.ID = id
		err = h.repositorio.Atualizar(r.Context(), a)
	case http.MethodPatch:
		var entrada struct {
			Ativa *bool `json:"ativa"`
		}
		if !lerJSON(w, r, &entrada) {
			return
		}
		if entrada.Ativa == nil {
			httpjson.Erro(w, http.StatusBadRequest, "Informe o campo ativa como true ou false.")
			return
		}
		err = h.repositorio.DefinirAtiva(r.Context(), id, *entrada.Ativa)
	}
	if errors.Is(err, ErrNaoEncontrada) {
		httpjson.Erro(w, http.StatusNotFound, "Assinatura não encontrada.")
		return
	}
	if err != nil {
		httpjson.Interno(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
