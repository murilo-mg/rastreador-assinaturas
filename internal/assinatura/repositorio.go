// Responsável pelas queries SQL relacionadas a assinaturas.
package assinatura

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNaoEncontrada = errors.New("assinatura não encontrada")

type Repositorio struct {
	banco *sql.DB
}

func NovoRepositorio(banco *sql.DB) *Repositorio {
	return &Repositorio{banco: banco}
}

func (r *Repositorio) Criar(ctx context.Context, a Assinatura) (int, error) {
	var id int
	query := `
		INSERT INTO assinaturas (nome, valor, categoria, dia_cobranca, ativa)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.banco.QueryRowContext(ctx, query, a.Nome, a.Valor, a.Categoria, a.DiaCobranca, a.Ativa).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar assinatura: %w", err)
	}
	return id, nil
}

func (r *Repositorio) Listar(ctx context.Context) ([]Assinatura, error) {
	query := `
		SELECT id, nome, valor, COALESCE(categoria, ''), dia_cobranca, ativa
		FROM assinaturas
		ORDER BY nome
	`
	linhas, err := r.banco.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar assinaturas: %w", err)
	}
	defer linhas.Close()

	assinaturas := []Assinatura{}
	for linhas.Next() {
		var a Assinatura
		if err := linhas.Scan(&a.ID, &a.Nome, &a.Valor, &a.Categoria, &a.DiaCobranca, &a.Ativa); err != nil {
			return nil, fmt.Errorf("erro ao ler assinatura: %w", err)
		}
		assinaturas = append(assinaturas, a)
	}

	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	return assinaturas, nil
}

func (r *Repositorio) Remover(ctx context.Context, id int) error {
	query := `DELETE FROM assinaturas WHERE id = $1`
	resultado, err := r.banco.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao remover assinatura: %w", err)
	}

	linhasAfetadas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar remoção: %w", err)
	}
	if linhasAfetadas == 0 {
		return ErrNaoEncontrada
	}

	return nil
}
