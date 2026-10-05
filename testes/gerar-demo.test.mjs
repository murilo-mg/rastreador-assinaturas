import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, readFile, readdir, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { gerarDemonstracao, raiz } from '../scripts/gerar-demo.mjs';

test('site gerado contém só recursos públicos e não altera o painel Go', async () => {
  const saida = await mkdtemp(join(tmpdir(), 'rastreador-demo-'));
  const original = await readFile(join(raiz, 'internal/web/arquivos/index.html'), 'utf8');
  try {
    await gerarDemonstracao(saida);
    assert.deepEqual((await readdir(saida)).sort(), ['404.html', '_headers', 'assets', 'index.html']);
    assert.deepEqual((await readdir(join(saida, 'assets'))).sort(), ['app.css', 'app.js', 'demo.mjs', 'favicon.svg']);
    const html = await readFile(join(saida, 'index.html'), 'utf8');
    assert.match(html, /data-modo="demonstracao"/); assert.match(html, /Restaurar exemplos/);
    assert.match(html, /somente nesta aba/); assert.doesNotMatch(html, /excluído do banco/);
    const headers = await readFile(join(saida, '_headers'), 'utf8');
    assert.match(headers, /connect-src 'none'/); assert.match(headers, /frame-ancestors 'none'/);
    assert.equal(await readFile(join(raiz, 'internal/web/arquivos/index.html'), 'utf8'), original);
    assert.doesNotMatch(original, /data-modo="demonstracao"/);
    await gerarDemonstracao(saida); // A geração pode ser repetida sem acumular arquivos.
    assert.equal(await readFile(join(saida, 'index.html'), 'utf8'), html);
  } finally { await rm(saida, { recursive: true, force: true }); }
});
