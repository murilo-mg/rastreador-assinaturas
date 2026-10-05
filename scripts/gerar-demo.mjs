import { cp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const raiz = resolve(dirname(fileURLToPath(import.meta.url)), '..');
export const cabecalhos = {
  'Content-Security-Policy': "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'",
  'X-Content-Type-Options': 'nosniff', 'X-Frame-Options': 'DENY',
  'Referrer-Policy': 'no-referrer', 'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
  'Cache-Control': 'no-store',
};
export const recursos = ['app.css', 'app.js', 'demo.mjs', 'favicon.svg'];

export async function gerarDemonstracao(saida = resolve(raiz, 'public')) {
  const origem = resolve(raiz, 'internal/web/arquivos');
  let html = await readFile(resolve(origem, 'index.html'), 'utf8');
  if (!html.includes('<body>') || !html.includes('<main>')) throw new Error('O modelo do painel mudou: revise o gerador da demonstração.');
  html = html.replace('<body>', '<body data-modo="demonstracao">')
    .replace('<title>Rastreador de Assinaturas</title>', '<title>Rastreador de Assinaturas · Demonstração</title>')
    .replace('>Controle pessoal</span>', '>Demonstração</span>')
    .replace('<main>', `<main>
    <section class="aviso-demo" aria-labelledby="titulo-demo"><div><h2 id="titulo-demo">Experimente com dados fictícios</h2><p>Você pode cadastrar, editar e remover exemplos. As alterações ficam somente nesta aba e são apagadas ao recarregar. Use apenas dados de teste.</p></div><button id="restaurar-exemplos" class="botao discreto" type="button">Restaurar exemplos</button></section>`)
    .replace('O registro será excluído do banco.', 'O exemplo será removido desta demonstração.');
  await rm(saida, { recursive: true, force: true });
  await mkdir(resolve(saida, 'assets'), { recursive: true });
  for (const nome of recursos) await cp(resolve(origem, nome), resolve(saida, 'assets', nome));
  await writeFile(resolve(saida, 'index.html'), html);
  await writeFile(resolve(saida, '_headers'), '/*\n' + Object.entries(cabecalhos).map(([nome, valor]) => `  ${nome}: ${valor}`).join('\n') + '\n');
  // Um 404 explícito evita que URLs de API recebam o index como fallback de SPA.
  await writeFile(resolve(saida, '404.html'), '<!doctype html><html lang="pt-BR"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Página não encontrada</title><link rel="stylesheet" href="/assets/app.css"><main><h1>Página não encontrada</h1><p><a href="/">Voltar à demonstração</a></p></main></html>');
  return saida;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  console.log('Demonstração gerada em ' + await gerarDemonstracao());
}
