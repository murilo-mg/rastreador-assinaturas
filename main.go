// Ponto de entrada da aplicação.
// Conecta ao banco e registra as rotas de assinaturas.
package main

import (
	"log"
	"net/http"

	"rastreador-assinaturas/internal/assinatura"
	"rastreador-assinaturas/internal/database"
)

func main() {
	banco, err := database.Conectar()
	if err != nil {
		log.Fatalf("não foi possível conectar ao banco: %v", err)
	}
	defer banco.Close()

	log.Println("conexão com o banco estabelecida com sucesso")

	mux := http.NewServeMux()
	mux.HandleFunc("/saude", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	repositorioAssinatura := assinatura.NovoRepositorio(banco)
	handlerAssinatura := assinatura.NovoHandler(repositorioAssinatura)
	handlerAssinatura.RegistrarRotas(mux)

	log.Println("servidor rodando na porta 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}