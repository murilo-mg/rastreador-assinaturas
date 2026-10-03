# Rastreador de Assinaturas

API em Go com um painel para acompanhar assinaturas mensais: cadastrar serviços, consultar os gastos e ver as próximas cobranças. Fiz esse projeto para praticar Go, PostgreSQL e a integração entre uma interface e o backend.

![Painel do rastreador com assinaturas fictícias, resumo mensal e próximas cobranças](docs/imagens/painel.png)

As capturas usam quatro serviços fictícios adicionados pelo botão **Experimentar com exemplos**. O painel começa vazio; nenhum exemplo é inserido automaticamente.

## O que dá para fazer

- Cadastrar assinaturas ativas ou inativas, com nome, valor, categoria e dia de cobrança.
- Buscar pelo nome, filtrar por categoria e ordenar por nome, valor ou dia.
- Ver o gasto mensal, a projeção anual e a quantidade de assinaturas ativas.
- Consultar as próximas cobranças em períodos de 7, 15 ou 30 dias.
- Comparar os gastos mensais por categoria.
- Remover uma assinatura com confirmação.

Os filtros mudam a lista. O resumo e a distribuição por categoria continuam considerando **todas as assinaturas ativas**.

<details>
<summary>Cadastro e versão para celular</summary>

![Formulário para cadastrar uma assinatura](docs/imagens/cadastro.png)

<img src="docs/imagens/mobile.png" width="390" alt="Painel do rastreador em uma tela de celular">

</details>

## Como rodar

Requisito: Docker com Docker Compose.

Na raiz do repositório:

~~~bash
docker compose up --build
~~~

Abra **[http://localhost:8080](http://localhost:8080)**.

O Compose sobe o PostgreSQL e espera o banco ficar pronto antes de iniciar a aplicação. O painel, o JavaScript e o CSS são servidos pelo mesmo processo Go, sem instalar dependências de frontend.

Para acompanhar os logs:

~~~bash
docker compose logs -f api
~~~

Para parar:

~~~bash
docker compose down
~~~

As assinaturas ficam no volume `dados_postgres` e permanecem disponíveis quando você sobe os serviços novamente. O schema é aplicado na primeira criação do banco.

## Regras dos cálculos

Todos os valores representam **cobranças mensais em reais**.

- O total mensal soma somente assinaturas ativas.
- A projeção anual multiplica o total mensal atual por 12. Não considera reajustes, descontos futuros nem histórico de pagamentos.
- O dia de cobrança pode ser de 1 a 31. Se o mês não tiver aquele dia, a próxima cobrança fica no último dia do mês.
- Exemplo: dia 31 corresponde a 30 de abril e a 28 de fevereiro, ou 29 em ano bissexto. Em março, volta a corresponder ao dia 31.
- Os vencimentos incluem hoje e a data final do período, em ordem cronológica. Cada assinatura aparece com sua próxima cobrança.
- O fuso usado no Compose é `America/Manaus`; pode ser alterado na variável `TZ`.

Remover uma assinatura apaga o registro deste aplicativo. **Não cancela o serviço contratado nem uma cobrança real.**

## API

| Método | Rota | Resultado |
| --- | --- | --- |
| GET | `/saude` | Verifica a conexão com o banco |
| POST | `/assinaturas` | Cadastra uma assinatura |
| GET | `/assinaturas` | Lista as assinaturas |
| DELETE | `/assinaturas/{id}` | Remove um registro |
| GET | `/relatorios/gasto-mensal` | Soma os valores ativos |
| GET | `/relatorios/projecao-anual` | Calcula a projeção anual |
| GET | `/relatorios/proximos-vencimentos?dias=7` | Lista as próximas cobranças |

### Exemplo de cadastro

~~~bash
curl -X POST http://localhost:8080/assinaturas \
  -H "Content-Type: application/json" \
  -d '{"nome":"Academia","valor":89.90,"categoria":"Saúde","dia_cobranca":10,"ativa":true}'
~~~

Resposta, com status `201`:

~~~json
{"id":1}
~~~

Nome, valor e dia de cobrança são obrigatórios. A categoria é opcional. Se `ativa` não for enviada, o cadastro começa ativo; `false` explícito é respeitado.

O nome aceita até 100 caracteres, a categoria até 50 e o valor até duas casas decimais, entre zero e R$ 99.999.999,99. O corpo do cadastro é limitado a 64 KiB; campos desconhecidos são rejeitados.

### Exemplo de relatório

~~~bash
curl http://localhost:8080/relatorios/gasto-mensal
~~~

~~~json
{"Total":89.9}
~~~

As chaves existentes dos relatórios foram preservadas: `Total`, `GastoMensal`, `GastoAnual`, `ID`, `Nome`, `Valor` e `DiaCobranca`. Os vencimentos também retornam `DataCobranca` no formato `AAAA-MM-DD` e `DiasAteCobranca`.

O parâmetro `dias` aceita inteiros de 1 a 365. Sem o parâmetro, o padrão é 7; valores inválidos retornam `400`.

Erros da API retornam um objeto `{"erro":"mensagem"}`. Entradas inválidas recebem `400`, corpo grande recebe `413`, formato diferente de JSON recebe `415` e uma assinatura inexistente recebe `404`. Detalhes internos do banco ficam no log do servidor.

## Testes

Com Go 1.27:

~~~bash
go test ./...
go vet ./...
~~~

A suíte cobre validação, limites do JSON, métodos HTTP, respostas de erro, recursos do painel e cálculo de datas, incluindo meses curtos, anos bissextos e viradas de mês e ano.

Para executar também a integração, use um banco PostgreSQL acessível:

~~~bash
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/rastreador_assinaturas?sslmode=disable' go test -race ./...
~~~

Os testes de integração usam tabelas temporárias e não alteram as assinaturas existentes. Sem `TEST_DATABASE_URL`, esses testes são indicados como ignorados; os demais continuam sendo executados.

O GitHub Actions executa formatação, `go vet`, testes com detector de condições de corrida, integração com PostgreSQL 16 e build do Docker. O escopo da conferência local está em [docs/VALIDACAO.md](docs/VALIDACAO.md).

## Organização

| Área | Arquivos |
| --- | --- |
| Inicialização e rotas | `main.go` |
| Cadastros, validação e persistência | `internal/assinatura/` |
| Relatórios e calendário | `internal/relatorio/` |
| Conexão com o banco | `internal/database/` |
| Respostas JSON | `internal/httpjson/` |
| Painel incorporado ao binário | `internal/web/` |
| Schema do PostgreSQL | `db/schema.sql` |
| Capturas e validação | `docs/` |

## Escopo atual

Esta versão é para **uso local e pessoal**. Não possui login nem separação por usuário: todos os acessos à mesma instância usam o mesmo conjunto de assinaturas.

No Compose, API e banco ficam acessíveis somente em `127.0.0.1`. A senha de desenvolvimento é `postgres`. Uma publicação na internet exige um ciclo próprio de autenticação, isolamento de dados e configuração da hospedagem.

A interface usa recursos locais, exibe os dados como texto e não inclui analytics ou serviços externos. Os registros são persistidos no PostgreSQL.

Os próximos passos são edição e ativação/desativação de registros existentes, exportação e uma demonstração pública com dados fictícios isolados.
