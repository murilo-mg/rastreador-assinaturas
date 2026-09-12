# Rastreador de Assinaturas

API em Go pra acompanhar assinaturas e gastos recorrentes — streaming, academia, qualquer serviço cobrado periodicamente. A ideia é saber quanto está sendo gasto por mês, o que vence em breve, e ter uma projeção de gasto anual.

## Status

Em construção. Por enquanto só a conexão com o banco e uma rota de saúde estão prontas.

## Stack

- Go
- PostgreSQL

## Como rodar

Precisa de um PostgreSQL rodando e do banco criado com o schema em `db/schema.sql`.

Configure as variáveis de ambiente (ou use os valores padrão, feitos pra desenvolvimento local):

\`\`\`bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=rastreador_assinaturas
\`\`\`

Depois:

\`\`\`bash
go run main.go
\`\`\`

O servidor sobe na porta 8080.