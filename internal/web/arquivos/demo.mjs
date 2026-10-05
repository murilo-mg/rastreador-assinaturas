// Demonstração: cada página tem seus próprios exemplos, somente em memória.
const comparar = new Intl.Collator('pt-BR', { sensitivity: 'base', numeric: true });
const json = (dados, status = 200) => new Response(JSON.stringify(dados), { status, headers: { 'Content-Type': 'application/json' } });
const erro = (texto, status = 400) => json({ erro: texto }, status);
const civil = data => new Date(Date.UTC(data.getFullYear(), data.getMonth(), data.getDate()));
const diaNoMes = (ano, mes, dia) => new Date(Date.UTC(ano, mes, Math.min(dia, new Date(Date.UTC(ano, mes + 1, 0)).getUTCDate())));

function validar(entrada, edicao) {
  if (!entrada || typeof entrada !== 'object' || Array.isArray(entrada) || Object.keys(entrada).some(k => !['nome', 'valor', 'categoria', 'dia_cobranca', 'ativa'].includes(k))) return 'Envie os campos da assinatura.';
  if (typeof entrada.nome !== 'string' || !entrada.nome.trim() || [...entrada.nome.trim()].length > 100) return 'O nome deve ter entre 1 e 100 caracteres.';
  if (entrada.categoria !== undefined && (typeof entrada.categoria !== 'string' || [...entrada.categoria.trim()].length > 50)) return 'A categoria deve ter até 50 caracteres.';
  if (typeof entrada.valor !== 'number' || !Number.isFinite(entrada.valor) || entrada.valor < 0 || entrada.valor > 99999999.99 || Math.abs(entrada.valor * 100 - Math.round(entrada.valor * 100)) > 0.000001) return 'Informe um valor não negativo com até duas casas decimais.';
  if (!Number.isInteger(entrada.dia_cobranca) || entrada.dia_cobranca < 1 || entrada.dia_cobranca > 31) return 'O dia da cobrança deve estar entre 1 e 31.';
  if ((edicao || entrada.ativa !== undefined) && typeof entrada.ativa !== 'boolean') return 'Informe se a assinatura está ativa.';
  return '';
}

function campoCSV(valor) {
  let texto = String(valor);
  const inicio = texto.replace(/^[\p{White_Space}\p{Cc}\p{Cf}]*/u, '');
  if (/^[=+\-@＝＋－＠]/u.test(inicio)) texto = '\t' + texto;
  if (/[;"\r\n]/u.test(texto) || /^\s/u.test(texto)) return '"' + texto.replaceAll('"', '""') + '"';
  return texto;
}

export function criarDemonstracao({ agora = () => new Date() } = {}) {
  let registros;
  let proximoID;
  function restaurar() {
    registros = [['Cinema em casa', 39.9, 'Streaming', 1], ['Academia', 89.9, 'Saúde', 3], ['Curso de idiomas', 59.9, 'Educação', 5], ['Ferramenta de trabalho', 24.9, 'Trabalho', 14]].map(([nome, valor, categoria, intervalo], i) => {
      const data = civil(agora()); data.setUTCDate(data.getUTCDate() + intervalo);
      return { id: i + 1, nome, valor, categoria, dia_cobranca: data.getUTCDate(), ativa: true };
    });
    proximoID = registros.length + 1;
  }
  restaurar();
  function responder(caminho, opcoes = {}) {
    const url = new URL(caminho, 'https://demonstracao.invalid');
    const metodo = opcoes.method || 'GET';
    const lista = [...registros].sort((a, b) => comparar.compare(a.nome, b.nome));
    const centavos = registros.filter(a => a.ativa).reduce((soma, a) => soma + Math.round(a.valor * 100), 0);
    if (metodo === 'GET') {
      if (url.pathname === '/assinaturas') return json(lista);
      if (url.pathname === '/relatorios/gasto-mensal') return json({ Total: centavos / 100 });
      if (url.pathname === '/relatorios/projecao-anual') return json({ GastoMensal: centavos / 100, GastoAnual: centavos * 12 / 100 });
      if (url.pathname === '/relatorios/proximos-vencimentos') {
        const parametro = url.searchParams.get('dias');
        if (parametro !== null && !/^\d+$/u.test(parametro)) return erro('Informe dias entre 1 e 365.');
        const dias = parametro === null ? 7 : Number(parametro);
        if (dias < 1 || dias > 365) return erro('Informe dias entre 1 e 365.');
        const hoje = civil(agora());
        const vencimentos = registros.filter(a => a.ativa).map(a => {
          let data = diaNoMes(hoje.getUTCFullYear(), hoje.getUTCMonth(), a.dia_cobranca);
          if (data < hoje) data = diaNoMes(hoje.getUTCFullYear(), hoje.getUTCMonth() + 1, a.dia_cobranca);
          return { ID: a.id, Nome: a.nome, Valor: a.valor, DiaCobranca: a.dia_cobranca, DataCobranca: data.toISOString().slice(0, 10), DiasAteCobranca: (data - hoje) / 86400000 };
        }).filter(v => v.DiasAteCobranca <= dias).sort((a, b) => a.DataCobranca.localeCompare(b.DataCobranca) || comparar.compare(a.Nome, b.Nome) || a.ID - b.ID);
        return json(vencimentos);
      }
      if (url.pathname === '/assinaturas/exportar.csv') {
        const linhas = [['ID', 'Nome', 'Valor mensal (R$)', 'Categoria', 'Dia da cobrança', 'Ativa'], ...lista.map(a => [a.id, a.nome, a.valor.toFixed(2).replace('.', ','), a.categoria, a.dia_cobranca, a.ativa ? 'Sim' : 'Não'])];
        return new Response('\uFEFF' + linhas.map(l => l.map(campoCSV).join(';')).join('\n') + '\n', { headers: { 'Content-Type': 'text/csv; charset=utf-8' } });
      }
    }
    const rota = /^\/assinaturas\/([1-9]\d*)$/u.exec(url.pathname);
    if (metodo === 'POST' && url.pathname === '/assinaturas' || rota && ['PUT', 'PATCH', 'DELETE'].includes(metodo)) {
      const id = rota ? Number(rota[1]) : null;
      const indice = registros.findIndex(a => a.id === id);
      if (rota && indice < 0) return erro('Assinatura não encontrada.', 404);
      if (metodo === 'DELETE') { registros.splice(indice, 1); return new Response(null, { status: 204 }); }
      if (typeof opcoes.body !== 'string' || new TextEncoder().encode(opcoes.body).length > 65536) return erro('Envie um JSON de até 64 KiB.', 413);
      let entrada;
      try { entrada = JSON.parse(opcoes.body); } catch { return erro('Envie um objeto JSON válido.'); }
      if (metodo === 'PATCH') {
        if (!entrada || typeof entrada.ativa !== 'boolean' || Object.keys(entrada).length !== 1) return erro('Informe somente o campo ativa como true ou false.');
        registros[indice].ativa = entrada.ativa;
      } else {
        const problema = validar(entrada, metodo === 'PUT');
        if (problema) return erro(problema);
        const a = { id: id ?? proximoID++, nome: entrada.nome.trim(), valor: entrada.valor, categoria: entrada.categoria?.trim() || '', dia_cobranca: entrada.dia_cobranca, ativa: entrada.ativa ?? true };
        if (metodo === 'POST') { registros.push(a); return json({ id: a.id }, 201); }
        registros[indice] = a;
      }
      return new Response(null, { status: 204 });
    }
    return erro('Operação não disponível na demonstração.', 404);
  }
  return { responder, restaurar };
}
