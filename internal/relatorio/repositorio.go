// Consultas de totais e cálculo de vencimentos mensais.
package relatorio

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

type Repositorio struct {
	banco *sql.DB
	agora func() time.Time
}

func NovoRepositorio(banco *sql.DB) *Repositorio {
	return &Repositorio{banco: banco, agora: time.Now}
}

// As chaves JSON existentes são preservadas para os clientes da API.
type GastoMensal struct{ Total float64 }
type ProjecaoAnual struct {
	GastoMensal float64
	GastoAnual  float64
}
type ProximoVencimento struct {
	ID              int
	Nome            string
	Valor           float64
	DiaCobranca     int
	DataCobranca    string
	DiasAteCobranca int
}

func (r *Repositorio) GastoMensalTotal(ctx context.Context) (GastoMensal, error) {
	query := "SELECT COALESCE(SUM(valor), 0) FROM assinaturas WHERE ativa = TRUE"
	var total float64
	if err := r.banco.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return GastoMensal{}, fmt.Errorf("erro ao calcular gasto mensal: %w", err)
	}
	return GastoMensal{Total: total}, nil
}

func (r *Repositorio) ProjecaoAnual(ctx context.Context) (ProjecaoAnual, error) {
	gasto, err := r.GastoMensalTotal(ctx)
	if err != nil {
		return ProjecaoAnual{}, err
	}
	// Arredonda em centavos antes da multiplicação para evitar resíduos de float no JSON.
	centavos := int64(gasto.Total*100 + 0.5)
	return ProjecaoAnual{GastoMensal: gasto.Total, GastoAnual: float64(centavos*12) / 100}, nil
}

// Datas de calendário em UTC evitam que horário de verão mude a contagem de dias.
// O dia atual é obtido no fuso configurado para a aplicação.
func dataCivil(data time.Time) time.Time {
	return time.Date(data.Year(), data.Month(), data.Day(), 0, 0, 0, 0, time.UTC)
}

func cobrancaNoMes(dia int, mes time.Time) time.Time {
	ultimo := time.Date(mes.Year(), mes.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if dia > ultimo {
		dia = ultimo
	}
	return time.Date(mes.Year(), mes.Month(), dia, 0, 0, 0, 0, time.UTC)
}

func proximaCobranca(dia int, hoje time.Time) time.Time {
	hoje = dataCivil(hoje)
	vencimento := cobrancaNoMes(dia, hoje)
	if vencimento.Before(hoje) {
		proximoMes := time.Date(hoje.Year(), hoje.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		vencimento = cobrancaNoMes(dia, proximoMes)
	}
	return vencimento
}

func (r *Repositorio) ProximosVencimentos(ctx context.Context, diasAFrente int) ([]ProximoVencimento, error) {
	if diasAFrente < 1 || diasAFrente > 365 {
		return nil, fmt.Errorf("antecedência deve estar entre 1 e 365 dias")
	}
	linhas, err := r.banco.QueryContext(ctx,
		"SELECT id, nome, valor, dia_cobranca FROM assinaturas WHERE ativa = TRUE")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar vencimentos: %w", err)
	}
	defer linhas.Close()

	hoje := dataCivil(r.agora())
	limite := hoje.AddDate(0, 0, diasAFrente)
	vencimentos := []ProximoVencimento{}
	for linhas.Next() {
		var v ProximoVencimento
		if err := linhas.Scan(&v.ID, &v.Nome, &v.Valor, &v.DiaCobranca); err != nil {
			return nil, fmt.Errorf("erro ao ler vencimento: %w", err)
		}
		data := proximaCobranca(v.DiaCobranca, hoje)
		if data.After(limite) {
			continue
		}
		v.DataCobranca = data.Format("2006-01-02")
		v.DiasAteCobranca = int(data.Sub(hoje).Hours() / 24)
		vencimentos = append(vencimentos, v)
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar vencimentos: %w", err)
	}
	sort.Slice(vencimentos, func(i, j int) bool {
		if vencimentos[i].DataCobranca != vencimentos[j].DataCobranca {
			return vencimentos[i].DataCobranca < vencimentos[j].DataCobranca
		}
		if vencimentos[i].Nome != vencimentos[j].Nome {
			return vencimentos[i].Nome < vencimentos[j].Nome
		}
		return vencimentos[i].ID < vencimentos[j].ID
	})
	return vencimentos, nil
}
