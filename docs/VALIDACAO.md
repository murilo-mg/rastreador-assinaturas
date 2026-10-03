# Validação do painel e da API

Revisão de 3 de outubro de 2026, preparada sobre a `main` em `1999cca`, incluindo exportação em CSV.

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

As capturas de `imagens/` usam dados fictícios. Incluem o formulário de edição
e as ações na lista. Login não faz parte desta versão.

## Limite da revisão

Não foi possível executar Docker neste ambiente. O workflow prepara PostgreSQL
16 nativo e verifica o build da imagem no GitHub Actions. A versão anterior
(`1999cca`) passou nessas etapas no [run 37157419204](https://github.com/murilo-mg/rastreador-assinaturas/actions/runs/37157419204).
O resultado das alterações desta revisão deve ser conferido no novo PR antes do merge.

O conteúdo e o download do CSV foram conferidos; Excel e LibreOffice não
foram executados neste ambiente. O formato UTF-8/BOM segue a documentação
da Microsoft e o tratamento de fórmulas segue a OWASP, com as referências
e o escopo descritos no README.

O detector de condições de corrida e os testes verificam comportamentos
específicos. Esta revisão não equivale a uma auditoria completa de segurança.
O aplicativo continua destinado a uso local, sem autenticação.
