// Responsável por abrir e validar a conexão com o banco de dados.
package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Conectar() (*sql.DB, error) {
	stringConexao := montarStringConexao()

	banco, err := sql.Open("postgres", stringConexao)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com o banco: %w", err)
	}

	if err := banco.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao testar conexão com o banco: %w", err)
	}

	return banco, nil
}

func montarStringConexao() string {
	host := obterVariavelOuPadrao("DB_HOST", "localhost")
	porta := obterVariavelOuPadrao("DB_PORT", "5432")
	usuario := obterVariavelOuPadrao("DB_USER", "postgres")
	senha := obterVariavelOuPadrao("DB_PASSWORD", "postgres")
	nomeBanco := obterVariavelOuPadrao("DB_NAME", "rastreador_assinaturas")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, porta, usuario, senha, nomeBanco,
	)
}

func obterVariavelOuPadrao(nomeVariavel string, valorPadrao string) string {
	valor := os.Getenv(nomeVariavel)
	if valor == "" {
		return valorPadrao
	}
	return valor
}