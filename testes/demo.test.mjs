import test from 'node:test';
import assert from 'node:assert/strict';
import { criarDemonstracao } from '../internal/web/arquivos/demo.mjs';

const criar = data => criarDemonstracao({ agora: () => new Date(data || '2026-04-30T12:00:00Z') });
const enviar = (demo, caminho, metodo = 'GET', corpo) => demo.responder(caminho, { method: metodo, body: corpo === undefined ? undefined : JSON.stringify(corpo) });
const cadastro = { nome: 'Clube', valor: 29.9, categoria: 'Educação', dia_cobranca: 31, ativa: true };

test('exemplos, totais e isolamento entre instâncias', async () => {
  const a = criar(), b = criar();
  assert.equal((await enviar(a, '/assinaturas').json()).length, 4);
  assert.deepEqual(await enviar(a, '/relatorios/gasto-mensal').json(), { Total: 214.6 });
  assert.equal((await enviar(a, '/relatorios/projecao-anual').json()).GastoAnual, 2575.2);
  assert.equal(enviar(a, '/assinaturas/1', 'DELETE').status, 204);
  assert.equal((await enviar(a, '/assinaturas').json()).length, 3);
  assert.equal((await enviar(b, '/assinaturas').json()).length, 4);
  a.restaurar();
  assert.equal((await enviar(a, '/assinaturas').json()).length, 4);
  assert.equal((await enviar(a, '/relatorios/gasto-mensal').json()).Total, 214.6);
});

test('cadastro, edição, estado, relatórios e remoção', async () => {
  const d = criar();
  const id = (await enviar(d, '/assinaturas', 'POST', cadastro).json()).id;
  assert.equal((await enviar(d, '/relatorios/gasto-mensal').json()).Total, 244.5);
  assert.equal(enviar(d, '/assinaturas/' + id, 'PUT', { ...cadastro, nome: 'Clube novo', valor: 10, ativa: false }).status, 204);
  assert.equal((await enviar(d, '/relatorios/gasto-mensal').json()).Total, 214.6);
  for (const ativa of [true, true, false, true]) {
    assert.equal(enviar(d, '/assinaturas/' + id, 'PATCH', { ativa }).status, 204);
    const a = (await enviar(d, '/assinaturas').json()).find(a => a.id === id);
    assert.equal(a.nome, 'Clube novo'); assert.equal(a.valor, 10); assert.equal(a.ativa, ativa);
    assert.equal((await enviar(d, '/relatorios/gasto-mensal').json()).Total, ativa ? 224.6 : 214.6);
    const vencimentos = await enviar(d, '/relatorios/proximos-vencimentos?dias=365').json();
    assert.equal(vencimentos.some(v => v.ID === id), ativa);
  }
  assert.equal(enviar(d, '/assinaturas/' + id, 'DELETE').status, 204);
  assert.equal(enviar(d, '/assinaturas/' + id, 'DELETE').status, 404);
  assert.equal(enviar(d, '/assinaturas/999', 'PUT', cadastro).status, 404);
});

test('entradas inválidas não alteram os dados', async () => {
  const d = criar();
  for (const entrada of [null, [], {}, { ...cadastro, nome: ' ' }, { ...cadastro, valor: -1 }, { ...cadastro, valor: 1.001 }, { ...cadastro, valor: null }, { ...cadastro, dia_cobranca: 32 }, { ...cadastro, dia_cobranca: 1.5 }, { ...cadastro, ativa: 'false' }, { ...cadastro, categoria: 'a'.repeat(51) }, { ...cadastro, id: 8 }]) {
    assert.equal(enviar(d, '/assinaturas', 'POST', entrada).status, 400);
  }
  assert.equal(enviar(d, '/assinaturas/1', 'PUT', { nome: 'Curso', valor: 1, dia_cobranca: 1 }).status, 400);
  for (const entrada of [{}, { ativa: null }, { ativa: true, nome: 'Outro' }]) assert.equal(enviar(d, '/assinaturas/1', 'PATCH', entrada).status, 400);
  assert.equal(d.responder('/assinaturas', { method: 'POST', body: '{' }).status, 400);
  assert.equal(d.responder('/assinaturas', { method: 'POST', body: ' '.repeat(65537) }).status, 413);
  assert.equal((await enviar(d, '/assinaturas').json()).length, 4);
  for (const dias of ['', '0', '-1', '366', 'abc', '1.5']) assert.equal(enviar(d, '/relatorios/proximos-vencimentos?dias=' + dias).status, 400);
});

test('datas civis, meses curtos, bissexto e virada de ano', async () => {
  for (const [hoje, dia, esperado, distancia] of [
    ['2026-04-30', 31, '2026-04-30', 0], ['2026-02-28', 31, '2026-02-28', 0],
    ['2028-02-29', 31, '2028-02-29', 0], ['2026-04-30', 1, '2026-05-01', 1],
    ['2026-12-31', 1, '2027-01-01', 1], ['2026-03-01', 31, '2026-03-31', 30],
  ]) {
    const d = criar(hoje + 'T12:00:00Z');
    const id = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, dia_cobranca: dia }).json()).id;
    const v = (await enviar(d, '/relatorios/proximos-vencimentos?dias=365').json()).find(v => v.ID === id);
    assert.equal(v.DataCobranca, esperado); assert.equal(v.DiasAteCobranca, distancia);
  }
});

test('CSV inclui inativas, acentos, valores, aspas e fórmulas tratadas', async () => {
  const d = criar();
  const nome = '=1+1";@SUM(1)\nEducação';
  const id = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, nome, categoria: '\u200B=1+1', ativa: false }).json()).id;
  const csv = await enviar(d, '/assinaturas/exportar.csv').text();
  // Response.text remove BOM; os bytes mantêm a marca UTF-8.
  assert.deepEqual([...new Uint8Array(await enviar(d, '/assinaturas/exportar.csv').arrayBuffer()).slice(0, 3)], [239, 187, 191]);
  assert.ok(csv.includes('"\t=1+1"";@SUM(1)\nEducação";29,90;"\t\u200B=1+1";31;Não'));
  assert.equal((await enviar(d, '/assinaturas').json()).find(a => a.id === id).nome, nome);
  for (const a of await enviar(d, '/assinaturas').json()) enviar(d, '/assinaturas/' + a.id, 'DELETE');
  assert.equal((await enviar(d, '/assinaturas/exportar.csv').text()).trim().split('\n').length, 1);
});

test('vencimentos incluem hoje e o limite do período, ordenados e sem inativas', async () => {
  const d = criar();
  const hoje = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, nome: 'Hoje', dia_cobranca: 30 }).json()).id;
  const limite = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, nome: 'Limite', dia_cobranca: 7 }).json()).id;
  const fora = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, nome: 'Fora', dia_cobranca: 8 }).json()).id;
  const inativa = (await enviar(d, '/assinaturas', 'POST', { ...cadastro, nome: 'Inativa', dia_cobranca: 30, ativa: false }).json()).id;
  const v = await enviar(d, '/relatorios/proximos-vencimentos').json();
  assert.equal(v[0].ID, hoje); assert.equal(v[0].DiasAteCobranca, 0);
  assert.equal(v.at(-1).ID, limite); assert.equal(v.at(-1).DiasAteCobranca, 7);
  assert.equal(v.some(a => a.ID === fora || a.ID === inativa), false);
});
