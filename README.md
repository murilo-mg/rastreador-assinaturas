# Rastreador de Assinaturas

API em Go pra acompanhar assinaturas e gastos recorrentes — streaming, academia, qualquer serviço cobrado periodicamente. Cadastra o que você paga, e a API calcula quanto está sendo gasto por mês, o que vence em breve, e uma projeção de gasto anual.

## Stack

- Go
- PostgreSQL
- Docker

## Como rodar

Com Docker instalado, dentro da pasta do projeto:

\`\`\`bash
docker compose up --build
\`\`\`

Isso sobe o banco já com o schema aplicado e a API na porta 8080.

## Endpoints

- `POST /assinaturas` — cadastra uma assinatura
- `GET /assinaturas` — lista todas
- `DELETE /assinaturas/{id}` — remove uma assinatura
- `GET /relatorios/gasto-mensal` — soma o valor de todas as assinaturas ativas
- `GET /relatorios/proximos-vencimentos?dias=7` — lista o que vence nos próximos N dias (padrão: 7)
- `GET /relatorios/projecao-anual` — projeta o gasto anual com base no gasto mensal atual

## Exemplo de cadastro

\`\`\`bash
curl -X POST http://localhost:8080/assinaturas \\
  -H "Content-Type: application/json" \\
  -d '{"nome":"Netflix","valor":39.90,"categoria":"streaming","dia_cobranca":10,"ativa":true}'
\`\`\`