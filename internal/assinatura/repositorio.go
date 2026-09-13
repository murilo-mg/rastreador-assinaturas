// Responsável pelas queries SQL relacionadas a assinaturas.
package assinatura

import (
	"database/sql"
	"fmt"
)

type Repositorio struct {
	banco *sql.DB
}

func NovoRepositorio(banco *sql.DB) *Repositorio {
	return &Repositorio{banco: banco}
}

func (r *Repositorio) Criar(a Assinatura) (int, error) {
	var id int
	query := `
		INSERT INTO assinaturas (nome, valor, categoria, dia_cobranca, ativa)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.banco.QueryRow(query, a.Nome, a.Valor, a.Categoria, a.DiaCobranca, a.Ativa).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar assinatura: %w", err)
	}
	return id, nil
}

func (r *Repositorio) Listar() ([]Assinatura, error) {
	query := `
		SELECT id, nome, valor, categoria, dia_cobranca, ativa
		FROM assinaturas
		ORDER BY nome
	`
	linhas, err := r.banco.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar assinaturas: %w", err)
	}
	defer linhas.Close()

	var assinaturas []Assinatura
	for linhas.Next() {
		var a Assinatura
		if err := linhas.Scan(&a.ID, &a.Nome, &a.Valor, &a.Categoria, &a.DiaCobranca, &a.Ativa); err != nil {
			return nil, fmt.Errorf("erro ao ler assinatura: %w", err)
		}
		assinaturas = append(assinaturas, a)
	}

	return assinaturas, nil
}

func (r *Repositorio) Remover(id int) error {
	query := `DELETE FROM assinaturas WHERE id = $1`
	resultado, err := r.banco.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erro ao remover assinatura: %w", err)
	}

	linhasAfetadas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar remoção: %w", err)
	}
	if linhasAfetadas == 0 {
		return fmt.Errorf("nenhuma assinatura encontrada com id %d", id)
	}

	return nil
}