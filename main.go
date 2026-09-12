// Ponto de entrada da aplicação.
// Nesta etapa apenas conecta ao banco e sobe um servidor com rota de saúde.
package main

import (
	"log"
	"net/http"

	"rastreador-assinaturas/internal/database"
)

func main() {
	banco, err := database.Conectar()
	if err != nil {
		log.Fatalf("não foi possível conectar ao banco: %v", err)
	}
	defer banco.Close()

	log.Println("conexão com o banco estabelecida com sucesso")

	http.HandleFunc("/saude", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Println("servidor rodando na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}