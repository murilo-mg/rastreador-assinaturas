# Validação do painel e da API

## Revisão de segurança de 5 de outubro de 2026

A revisão local começou sobre a `main` em `f464614`. As correções de segurança foram integradas à `main` em `36c2c74`.

- Um teste de regressão reproduziu a aceitação de um domínio externo em
  `Host`, mesmo com `Origin` igual, em todos os métodos testados. Ele falhou
  antes da correção e passou depois dela. A reprodução usa requisições HTTP
  de teste; não representa um ataque DNS completo executado em um navegador.
- A proteção agora rejeita hosts fora da lista local antes de chamar os
  handlers. Cobertura de IPv4/IPv6, portas, domínios parecidos com localhost,
  entradas malformadas e cabeçalhos de proxy que tentem alterar o host aceito.
- `Origin` é verificada também em GET/HEAD, além dos métodos de alteração;
  origens externas, porta diferente e componentes indevidos são rejeitados.
- Go 1.27.1: formatação, `go vet`, suíte com `-race` e integração pelo
  protocolo PostgreSQL/PGlite passaram após a mudança.
- O painel Go foi conferido em Chromium 154: cadastro, edição, estados,
  relatórios, filtros, remoção, download CSV, celular e recuperação de falha,
  sem erros de JavaScript ou CSP nos fluxos normais.
- Os sete grupos de testes Node passaram e o build estático foi gerado.
- `govulncheck` 1.8.0 não encontrou vulnerabilidades conhecidas no código
  analisado com Go 1.27.1 na data da revisão. Esse resultado pode mudar
  conforme novas vulnerabilidades forem publicadas.
- O workflow executa esse scanner em pushes, PRs, acionamento manual e
  semanalmente. A configuração foi integrada à `main`; os checks `testes` e
  `seguranca` passaram na revisão de segurança.

Os checks `testes` e `seguranca` da correção integrada em `36c2c74` passaram no [run 37261801997](https://github.com/murilo-mg/rastreador-assinaturas/actions/runs/37261801997). A demonstração pública em [rastreador-assinaturas.pages.dev](https://rastreador-assinaturas.pages.dev/) abriu com o aviso, os quatro exemplos e os indicadores. Em 5 de outubro de 2026, a resposta HTTP também confirmou status 200 e os cabeçalhos descritos em [DEMONSTRACAO.md](DEMONSTRACAO.md). A aplicação Go continua sem autenticação e é destinada a uso local. Veja [SEGURANCA.md](SEGURANCA.md).

## Revisão do painel e da demonstração

Revisão de 4 de outubro de 2026, preparada sobre a `main` em `076a794`, incluindo demonstração estática.

## Código e testes

- Go 1.27.1: `gofmt`, `go vet`, suíte completa e execução com `-race`.
- Integração local pelo protocolo PostgreSQL usando PGlite, que executa o
  motor PostgreSQL em WebAssembly. Não é uma simulação das consultas.
- Cadastro, persistência, remoção, soma decimal no banco, exclusão de inativas
  dos relatórios e projeção anual verificados com dados de teste.
- Edição de todos os campos, preservação do ID e da data de criação,
  rejeição de edição inválida sem alterar o banco e resposta `404` para IDs ausentes.
- Ativação/desativação sem alterar os outros campos, repetição do mesmo estado
  e atualização dos totais, da projeção e dos vencimentos.
- Edição exige `ativa` explícita; a mudança de estado aceita somente um booleano.
- Exportação CSV com BOM UTF-8, seis colunas, vírgula decimal, duas casas,
  acentos, aspas, ponto e vírgula, quebras de linha e registros inativos.
- Cabeçalho em lista vazia, método GET, nome de download, ausência de cache
  e falha do banco retornando JSON sem iniciar um arquivo parcial.
- Prefixos de fórmula em nomes e categorias, incluindo variantes Unicode
  e caracteres invisíveis iniciais, tratados na exportação sem alterar o banco.
- CSV gerado a partir dos registros reais da integração, após edição e
  mudança de estado, com os cabeçalhos de proteção do servidor.
- Vencimentos: virada de mês e ano, fevereiro, ano bissexto, dias 29 a 31,
  inclusão de hoje e do último dia do período e ordenação cronológica.
- JSON inválido, campos obrigatórios, campos desconhecidos, corpo acima de
  64 KiB, tipos de conteúdo, métodos e registros inexistentes cobertos.
- Build estático com `CGO_ENABLED=0`, usado na verificação do navegador.
- Node.js 24: exemplos, cálculos em centavos, CRUD, estado ativo, validação,
  calendário, CSV, isolamento entre instâncias e geração dos arquivos públicos.
- A geração copia somente os recursos permitidos e não altera o HTML da aplicação Go.

## Interface

Verificada em Chromium 154, com a aplicação Go e o banco de revisão:

- Estado inicial vazio e adição explícita de quatro exemplos fictícios.
- Totais mensal e anual, contagem de ativas e distribuição por categoria.
- Busca, filtro, ordenação e mudança do período de vencimentos.
- Cadastro, mensagem de validação e recuperação após corrigir o campo.
- Edição com campos preenchidos, cancelamento sem salvar, validação com
  preservação do formulário, gravação e atualização dos totais.
- Desativação e reativação pela lista, persistência após recarregar a página,
  categorias e vencimentos atualizados, busca/filtro/ordenação preservados
  durante a atualização e cadastro novo sem reutilizar os dados da edição.
- Exclusão com confirmação e cancelamento da exclusão.
- Assinaturas inativas fora dos totais.
- Nome com aparência de HTML exibido como texto, sem criar elemento.
- Mensagem de falha quando a conexão da listagem é interrompida.
- Largura de 390 px, sem transbordamento horizontal, e fechamento por Escape.
- Recursos do aplicativo na mesma origem, sem erros de JavaScript ou
  violações de CSP nos fluxos normais.
- Download de CSV vazio e preenchido: bytes idênticos à resposta da API,
  nome do arquivo, BOM, valores e inclusão de todas as assinaturas com filtros ativos.
- Exportação de inativas e nomes com prefixo de fórmula, download no celular,
  falha de conexão sem baixar um arquivo de erro e nova tentativa bem-sucedida.

Na demonstração estática, também foram conferidos:

- Quatro exemplos iniciais, aviso de dados fictícios e botão de restauração.
- Cadastro, validação e recuperação, edição, remoção/cancelamento, estado
  ativo, totais, categorias, vencimentos e CSV com o valor editado e inativas.
- Isolamento entre abas do mesmo navegador e entre contextos independentes.
- Restauração dos exemplos ao recarregar e pelo botão, com filtros reiniciados.
- Ausência de chamadas de rede à API, de recursos externos e de armazenamento
  das assinaturas em cookies, `localStorage` ou `sessionStorage`.
- Nome com aparência de HTML como texto, tela de 390 px sem transbordamento,
  fechamento por Escape e ausência de erros de JavaScript ou de CSP.
- Prévia responde 404 para código Go, README, arquivos privados e URLs de API.

As capturas de `imagens/` usam dados fictícios. Incluem o formulário de edição
e as ações na lista. Login não faz parte desta versão.

## Limite da revisão

O ambiente de revisão não tinha Docker instalado; o workflow do GitHub Actions executou a integração com PostgreSQL 16 e a compilação da imagem Docker. A revisão integrada em `36c2c74` passou nos checks remotos; veja o [run 37261801997](https://github.com/murilo-mg/rastreador-assinaturas/actions/runs/37261801997).

O conteúdo e o download do CSV foram conferidos; Excel e LibreOffice não
foram executados neste ambiente. O formato UTF-8/BOM segue a documentação
da Microsoft e o tratamento de fórmulas segue a OWASP, com as referências
e o escopo descritos no README.

A demonstração foi aberta no Cloudflare Pages em 5 de outubro de 2026. A página carregou o aviso e os quatro exemplos; a resposta HTTP retornou status 200 e os cabeçalhos de segurança esperados.

O detector de condições de corrida e os testes verificam comportamentos
específicos. Esta revisão não equivale a uma auditoria completa de segurança.
O aplicativo continua destinado a uso local, sem autenticação.
