// Respostas compartilhadas pela API. Detalhes internos ficam no log do servidor.
package httpjson

import (
	"encoding/json"
	"log"
	"net/http"
)

func Responder(w http.ResponseWriter, status int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(valor); err != nil {
		log.Printf("erro ao escrever resposta: %v", err)
	}
}

func Erro(w http.ResponseWriter, status int, mensagem string) {
	Responder(w, status, map[string]string{"erro": mensagem})
}

func Interno(w http.ResponseWriter, err error) {
	log.Printf("erro na API: %v", err)
	Erro(w, http.StatusInternalServerError, "Não foi possível concluir a operação. Tente novamente.")
}
