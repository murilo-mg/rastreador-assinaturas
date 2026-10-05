# Segurança e limites de uso

O rastreador tem dois modos: uma aplicação Go com PostgreSQL para uso local
e uma demonstração estática com dados fictícios. Nenhum deles recebe pagamentos,
credenciais de serviços de assinatura ou dados de cartão.

## Aplicação Go local

- O Compose publica API e banco somente em `127.0.0.1`. A execução Go direta
  também usa esse endereço por padrão.
- O servidor valida `Host` antes de servir o painel ou acessar a API. Aceita
  somente `localhost`, `127.0.0.1` e `[::1]`, com porta opcional de 1 a 65535.
  Cabeçalhos de proxy não ampliam essa lista.
- A validação não consulta DNS: um domínio externo que passe a apontar para
  o loopback continua bloqueado. Comparar somente `Origin` com `Host` não
  seria suficiente nesse cenário de DNS rebinding.
- Quando há `Origin`, ela precisa corresponder ao `Host`, usar HTTP/HTTPS
  e não conter usuário, caminho, consulta ou fragmento. A regra vale para
  leituras e alterações. Clientes locais como `curl` podem omitir `Origin`.
- Consultas SQL usam parâmetros; entradas JSON têm validação e limite de
  64 KiB nos endpoints de criação e alteração.
- A interface exibe nomes e categorias como texto. A exportação CSV trata
  prefixos de fórmula sem modificar os registros originais.
- CSP restringe recursos à própria aplicação; também são enviados cabeçalhos
  contra interpretação indevida de conteúdo e enquadramento em outra página.
- O servidor limita o tempo das requisições. Erros internos do banco ficam
  nos logs, sem detalhes técnicos na resposta da API.

Essas medidas não fornecem login nem isolamento por usuário. Qualquer cliente
local que consiga acessar a instância pode consultar e alterar as assinaturas.
Um programa malicioso executado na máquina não é impedido pela validação de
`Host` ou `Origin`. A senha `postgres` do Compose é de desenvolvimento.

Não exponha essa configuração por túnel, encaminhamento de portas ou proxy
público. Mudar `SERVER_ADDR` não transforma o aplicativo em um serviço com
autenticação. Contas de usuário e dados reais na internet exigem um projeto
próprio de controle de acesso, isolamento, transporte e operação do banco.

## Demonstração pública

- O build copia apenas HTML, CSS, JavaScript e favicon para `public/`.
  Código Go, banco, documentação e arquivos de configuração não são publicados.
- As assinaturas ficam somente na memória da aba. Recarregar restaura os
  exemplos; abas diferentes não compartilham os registros.
- O painel não envia assinaturas à API nem usa cookies, `localStorage` ou
  `sessionStorage` para guardá-las. Use apenas dados fictícios.
- A demonstração publicada em [rastreador-assinaturas.pages.dev](https://rastreador-assinaturas.pages.dev/) teve seus cabeçalhos conferidos em 5 de outubro de 2026. A resposta incluiu CSP com `connect-src 'none'`, proteção contra enquadramento, `nosniff`, política de referência e política de permissões, conforme [DEMONSTRACAO.md](DEMONSTRACAO.md).

Não existe autenticação nesse site de exemplos, porque ele não hospeda a API
nem compartilha um banco com registros dos visitantes.

## Manutenção

O job `seguranca` executa `govulncheck` em pushes, PRs, acionamento manual e
semanalmente na branch padrão. A ferramenta está fixada em `v1.8.0`, mas
consulta a base atual de vulnerabilidades Go. Uma descoberta aplicável ao
código faz o job falhar; confira o resultado antes do merge e acompanhe as
execuções semanais. O workflow não corrige nem atualiza dependências sozinho.

Testes e scanners verificam riscos específicos e vulnerabilidades já
conhecidas. Mantenha Go, PostgreSQL, Docker, navegador e sistema operacional
atualizados. O resultado desta revisão está em [VALIDACAO.md](VALIDACAO.md);
ele não é uma auditoria independente nem uma garantia de segurança total.

Referências: [GitHub Security Lab sobre localhost e DNS rebinding](https://github.blog/security/application-security/localhost-dangers-cors-and-dns-rebinding/)
e [ferramentas de segurança do Go](https://go.dev/doc/security/).
