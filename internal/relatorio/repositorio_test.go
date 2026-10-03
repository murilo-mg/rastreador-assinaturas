package relatorio

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestVencimentosComBanco(t *testing.T) {
	conexao := os.Getenv("TEST_DATABASE_URL")
	if conexao == "" {
		t.Skip("defina TEST_DATABASE_URL para executar a integração com PostgreSQL")
	}
	banco, err := sql.Open("postgres", conexao)
	if err != nil {
		t.Fatal(err)
	}
	defer banco.Close()
	banco.SetMaxOpenConns(1)
	ctx := context.Background()
	_, err = banco.ExecContext(ctx, `CREATE TEMP TABLE assinaturas(id SERIAL PRIMARY KEY,nome TEXT,valor NUMERIC(10,2),dia_cobranca INTEGER,ativa BOOLEAN);
	INSERT INTO assinaturas(nome,valor,dia_cobranca,ativa) VALUES
	('Ultimo dia',10,31,true),('Primeiro dia',20,1,true),('Segundo dia',30,2,true),
	('Terceiro dia',40,3,true),('Inativa',50,1,false);`)
	if err != nil {
		t.Fatal(err)
	}
	defer banco.ExecContext(ctx, "DROP TABLE IF EXISTS pg_temp.assinaturas")
	r := NovoRepositorio(banco)
	r.agora = func() time.Time { return time.Date(2026, 4, 30, 15, 0, 0, 0, time.UTC) }
	vencimentos, err := r.ProximosVencimentos(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(vencimentos) != 3 {
		t.Fatalf("queria 3 vencimentos: %+v", vencimentos)
	}
	for i, data := range []string{"2026-04-30", "2026-05-01", "2026-05-02"} {
		if vencimentos[i].DataCobranca != data || vencimentos[i].DiasAteCobranca != i {
			t.Fatalf("data ou ordem incorreta: %+v", vencimentos[i])
		}
	}
}
