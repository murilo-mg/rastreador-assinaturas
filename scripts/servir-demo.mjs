import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { cabecalhos, raiz, recursos } from './gerar-demo.mjs';

const arquivos = new Map([['/', ['index.html', 'text/html; charset=utf-8']], ...recursos.map(nome => ['/assets/' + nome, ['assets/' + nome, nome.endsWith('.css') ? 'text/css' : nome.endsWith('.svg') ? 'image/svg+xml' : 'text/javascript']])]);
const porta = Number(process.env.DEMO_PORT || 4173);
createServer(async (req, res) => {
  const recurso = arquivos.get(new URL(req.url, 'http://localhost').pathname);
  if (req.method !== 'GET' && req.method !== 'HEAD') { res.writeHead(405, { ...cabecalhos, Allow: 'GET, HEAD' }); res.end(); return; }
  try {
    const conteudo = await readFile(resolve(raiz, 'public', recurso?.[0] || '404.html'));
    res.writeHead(recurso ? 200 : 404, { ...cabecalhos, 'Content-Type': recurso?.[1] || 'text/html; charset=utf-8' });
    res.end(req.method === 'HEAD' ? undefined : conteudo);
  } catch { res.writeHead(503, cabecalhos); res.end('Execute node scripts/gerar-demo.mjs antes de abrir a demonstração.'); }
}).listen(porta, '127.0.0.1', () => console.log(`Demonstração: http://127.0.0.1:${porta}`));
