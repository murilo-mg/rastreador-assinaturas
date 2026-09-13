// Responsável pelas queries SQL de relatórios agregados sobre assinaturas.
package relatorio

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

type GastoMensal struct {
	Total float64
}

func (r *Repositorio) GastoMensalTotal() (GastoMensal, error) {
	query := `
		SELECT COALESCE(SUM(valor), 0)
		FROM assinaturas
		WHERE ativa = TRUE
	`
	var total float64
	if err := r.banco.QueryRow(query).Scan(&total); err != nil {
		return GastoMensal{}, fmt.Errorf("erro ao calcular gasto mensal: %w", err)
	}
	return GastoMensal{Total: total}, nil
}

type ProjecaoAnual struct {
	GastoMensal float64
	GastoAnual  float64
}

func (r *Repositorio) ProjecaoAnual() (ProjecaoAnual, error) {
	gastoMensal, err := r.GastoMensalTotal()
	if err != nil {
		return ProjecaoAnual{}, fmt.Errorf("erro ao calcular projeção anual: %w", err)
	}

	return ProjecaoAnual{
		GastoMensal: gastoMensal.Total,
		GastoAnual:  gastoMensal.Total * 12,
	}, nil
}

type ProximoVencimento struct {
	ID          int
	Nome        string
	Valor       float64
	DiaCobranca int
}

func (r *Repositorio) ProximosVencimentos(diasAFrente int) ([]ProximoVencimento, error) {
	query := `
		SELECT id, nome, valor, dia_cobranca
		FROM assinaturas
		WHERE ativa = TRUE
		AND (
			dia_cobranca BETWEEN EXTRACT(DAY FROM CURRENT_DATE) AND EXTRACT(DAY FROM CURRENT_DATE) + $1
			OR dia_cobranca < EXTRACT(DAY FROM CURRENT_DATE) + $1 - EXTRACT(DAY FROM (DATE_TRUNC('MONTH', CURRENT_DATE) + INTERVAL '1 MONTH - 1 DAY'))
		)
		ORDER BY dia_cobranca
	`
	linhas, err := r.banco.Query(query, diasAFrente)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar próximos vencimentos: %w", err)
	}
	defer linhas.Close()

	var vencimentos []ProximoVencimento
	for linhas.Next() {
		var v ProximoVencimento
		if err := linhas.Scan(&v.ID, &v.Nome, &v.Valor, &v.DiaCobranca); err != nil {
			return nil, fmt.Errorf("erro ao ler vencimento: %w", err)
		}
		vencimentos = append(vencimentos, v)
	}

	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	return vencimentos, nil
}