# Validação do painel e da API

Revisão de 3 de outubro de 2026, preparada sobre a `main` em `6701e43`.

## Código e testes

- Go 1.27.1: `gofmt`, `go vet`, suíte completa e execução com `-race`.
- Integração local pelo protocolo PostgreSQL usando PGlite, que executa o
  motor PostgreSQL em WebAssembly. Não é uma simulação das consultas.
- Cadastro, persistência, remoção, soma decimal no banco, exclusão de inativas
  dos relatórios e projeção anual verificados com dados de teste.
- Vencimentos: virada de mês e ano, fevereiro, ano bissexto, dias 29 a 31,
  inclusão de hoje e do último dia do período e ordenação cronológica.
- JSON inválido, campos obrigatórios, campos desconhecidos, corpo acima de
  64 KiB, tipos de conteúdo, métodos e registros inexistentes cobertos.
- Build estático com `CGO_ENABLED=0`, usado na verificação do navegador.

## Interface

Verificada em Chromium 154, com a aplicação Go e o banco de revisão:

- Estado inicial vazio e adição explícita de quatro exemplos fictícios.
- Totais mensal e anual, contagem de ativas e distribuição por categoria.
- Busca, filtro, ordenação e mudança do período de vencimentos.
- Cadastro, mensagem de validação e recuperação após corrigir o campo.
- Exclusão com confirmação e cancelamento da exclusão.
- Assinaturas inativas fora dos totais.
- Nome com aparência de HTML exibido como texto, sem criar elemento.
- Mensagem de falha quando a conexão da listagem é interrompida.
- Largura de 390 px, sem transbordamento horizontal, e fechamento por Escape.
- Recursos do aplicativo na mesma origem, sem erros de JavaScript ou
  violações de CSP nos fluxos normais.

As capturas de `imagens/` usam dados fictícios. Não demonstram edição de
assinaturas nem login, funcionalidades que ainda não foram implementadas.

## Limite da revisão

Não foi possível executar Docker neste ambiente. O workflow incluído prepara
um PostgreSQL 16 nativo e verifica o build da imagem no GitHub Actions.
Esse resultado deve ser conferido no PR antes do merge.

O detector de condições de corrida e os testes verificam comportamentos
específicos. Esta revisão não equivale a uma auditoria completa de segurança.
O aplicativo continua destinado a uso local, sem autenticação.
