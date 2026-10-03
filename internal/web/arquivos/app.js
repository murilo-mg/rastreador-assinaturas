'use strict';

const el = seletor => document.querySelector(seletor);
const reais = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });
const comparador = new Intl.Collator('pt-BR', { sensitivity: 'base', numeric: true });
let dados = null;
let tentativa = 0;
let controlador = null;
let cadastroOcupado = false;
let removendo = false;
let assinaturaRemover = null;

function elemento(tag, texto, classe) {
  const no = document.createElement(tag);
  if (texto !== undefined) no.textContent = texto;
  if (classe) no.className = classe;
  return no;
}

function mensagem(texto, erro = false) {
  const no = el('#mensagem');
  no.textContent = texto;
  no.classList.toggle('erro', erro);
  no.hidden = !texto;
}

async function requisicao(caminho, opcoes = {}, sinal) {
  let resposta;
  try {
    resposta = await fetch(caminho, { ...opcoes, signal: sinal ?? AbortSignal.timeout(10000) });
  } catch (erro) {
    if (sinal?.aborted) throw erro;
    throw new Error('Não consegui acessar o servidor. Confira se a aplicação está rodando e tente novamente.');
  }
  if (!resposta.ok) {
    let texto = 'Não foi possível concluir a operação.';
    try { texto = (await resposta.json()).erro || texto; } catch { /* Resposta sem JSON. */ }
    throw new Error(texto);
  }
  return resposta.status === 204 ? null : resposta.json();
}

async function atualizar() {
  const atual = ++tentativa;
  controlador?.abort();
  controlador = new AbortController();
  const sinal = controlador.signal;
  const prazo = setTimeout(() => controlador?.signal === sinal && controlador.abort(), 10000);
  el('#atualizar').disabled = true;
  el('#atualizar').textContent = 'Atualizando…';
  el('#lista').setAttribute('aria-busy', 'true');
  el('.resumo').setAttribute('aria-busy', 'true');
  try {
    const [assinaturas, mensal, anual, vencimentos] = await Promise.all([
      requisicao('/assinaturas', {}, sinal), requisicao('/relatorios/gasto-mensal', {}, sinal),
      requisicao('/relatorios/projecao-anual', {}, sinal),
      requisicao('/relatorios/proximos-vencimentos?dias=' + el('#dias').value, {}, sinal),
    ]);
    if (atual !== tentativa) return false;
    dados = { assinaturas, mensal, anual, vencimentos };
    renderizar();
    return true;
  } catch (erro) {
    if (atual !== tentativa) return false;
    mensagem(sinal.aborted ? 'A atualização demorou demais. Tente novamente.' : erro.message, true);
    if (!dados) {
      for (const seletor of ['#lista', '#vencimentos', '#categorias']) {
        const tag = seletor === '#vencimentos' ? 'li' : 'p';
        el(seletor).replaceChildren(elemento(tag, 'Dados indisponíveis. Use Atualizar para tentar novamente.', 'estado-vazio'));
      }
      el('#contagem').textContent = 'Aguardando conexão';
      el('#nota-quantidade').textContent = 'Resumo indisponível';
    }
    return false;
  } finally {
    clearTimeout(prazo);
    if (atual === tentativa) {
      el('#atualizar').disabled = false;
      el('#atualizar').textContent = '↻ Atualizar';
      el('#lista').setAttribute('aria-busy', 'false');
      el('.resumo').setAttribute('aria-busy', 'false');
    }
  }
}

function categoria(a) { return a.categoria || 'Sem categoria'; }

function renderizar() {
  const ativas = dados.assinaturas.filter(a => a.ativa).length;
  el('#gasto-mensal').textContent = reais.format(dados.mensal.Total);
  el('#gasto-anual').textContent = reais.format(dados.anual.GastoAnual);
  el('#quantidade').textContent = ativas;
  const inativas = dados.assinaturas.length - ativas;
  el('#nota-quantidade').textContent = inativas ? `${inativas} inativa${inativas > 1 ? 's' : ''} fora do total` : 'Todas incluídas no resumo';
  const filtro = el('#categoria-filtro');
  const anterior = filtro.value;
  const opcoes = [...new Set(dados.assinaturas.map(categoria))].sort(comparador.compare);
  filtro.replaceChildren(new Option('Todas', ''), ...opcoes.map(nome => new Option(nome, nome)));
  filtro.value = opcoes.includes(anterior) ? anterior : '';
  renderizarLista(); renderizarVencimentos(); renderizarCategorias();
}

function renderizarLista() {
  if (!dados) return;
  const texto = el('#busca').value.trim().toLocaleLowerCase('pt-BR');
  const filtro = el('#categoria-filtro').value;
  const lista = dados.assinaturas.filter(a =>
    a.nome.toLocaleLowerCase('pt-BR').includes(texto) && (!filtro || categoria(a) === filtro));
  const ordem = el('#ordem').value;
  lista.sort((a, b) => (ordem === 'valor' ? b.valor - a.valor : ordem === 'dia' ? a.dia_cobranca - b.dia_cobranca : 0) || comparador.compare(a.nome, b.nome));
  el('#contagem').textContent = `${lista.length} de ${dados.assinaturas.length} assinaturas · resumo considera todas as ativas`;
  if (!lista.length) {
    const vazio = elemento('div', undefined, 'vazio-lista');
    if (dados.assinaturas.length) {
      vazio.append(elemento('h3', 'Nenhuma assinatura encontrada'), elemento('p', 'Tente outro nome ou selecione todas as categorias.'));
    } else {
      vazio.append(elemento('h3', 'Comece pelo que você paga todo mês'), elemento('p', 'Cadastre sua primeira assinatura ou adicione quatro exemplos fictícios para conhecer o painel.'));
      const acoes = elemento('div', undefined, 'vazio-acoes');
      const criar = elemento('button', 'Nova assinatura', 'botao primario'); criar.type = 'button'; criar.addEventListener('click', abrirCadastro);
      const exemplo = elemento('button', 'Experimentar com exemplos', 'botao discreto'); exemplo.type = 'button'; exemplo.addEventListener('click', () => carregarExemplos(exemplo));
      acoes.append(criar, exemplo); vazio.append(acoes);
    }
    el('#lista').replaceChildren(vazio); return;
  }
  const tabela = elemento('table', undefined, 'tabela');
  const cabecalho = elemento('thead'); const linha = elemento('tr');
  for (const titulo of ['Assinatura', 'Valor / mês', 'Categoria', 'Cobrança', 'Ação']) {
    const th = elemento('th', titulo); th.scope = 'col'; linha.append(th);
  }
  cabecalho.append(linha); const corpo = elemento('tbody');
  for (const a of lista) {
    const tr = elemento('tr'); if (!a.ativa) tr.className = 'inativa';
    const nome = elemento('td'); const servico = elemento('div', undefined, 'nome-servico');
    const avatar = elemento('span', [...a.nome][0]?.toLocaleUpperCase('pt-BR') || '•', 'avatar'); avatar.setAttribute('aria-hidden', 'true');
    const descricao = elemento('div'); descricao.append(elemento('strong', a.nome), elemento('small', a.ativa ? 'Ativa' : 'Inativa'));
    servico.append(avatar, descricao); nome.append(servico);
    const acao = elemento('td'); const remover = elemento('button', 'Remover', 'remover'); remover.type = 'button';
    remover.setAttribute('aria-label', 'Remover ' + a.nome); remover.addEventListener('click', () => abrirRemocao(a)); acao.append(remover);
    tr.append(nome, elemento('td', reais.format(a.valor), 'valor'), elemento('td', categoria(a), 'categoria'), elemento('td', 'Dia ' + a.dia_cobranca, 'dia'), acao); corpo.append(tr);
  }
  tabela.append(cabecalho, corpo); el('#lista').replaceChildren(tabela);
}

function renderizarVencimentos() {
  const lista = el('#vencimentos');
  if (!dados.vencimentos.length) { lista.replaceChildren(elemento('li', 'Nenhuma cobrança ativa nesse período.', 'estado-vazio')); return; }
  lista.replaceChildren(...dados.vencimentos.map(v => {
    const li = elemento('li', undefined, 'vencimento');
    const data = new Date(v.DataCobranca + 'T12:00:00');
    const caixa = elemento('div', undefined, 'data-box');
    caixa.append(elemento('strong', data.getDate()), elemento('small', data.toLocaleDateString('pt-BR', { month: 'short' }).replace('.', '')));
    caixa.setAttribute('aria-label', data.toLocaleDateString('pt-BR', { day: 'numeric', month: 'long', year: 'numeric' }));
    const info = elemento('div', undefined, 'vencimento-info');
    info.append(elemento('strong', v.Nome), elemento('small', v.DiasAteCobranca === 0 ? 'Hoje' : v.DiasAteCobranca === 1 ? 'Amanhã' : `Em ${v.DiasAteCobranca} dias`));
    li.append(caixa, info, elemento('span', reais.format(v.Valor), 'valor')); return li;
  }));
}

function renderizarCategorias() {
  const totais = new Map();
  for (const a of dados.assinaturas.filter(a => a.ativa)) {
    const nome = categoria(a); totais.set(nome, (totais.get(nome) || 0) + Math.round(a.valor * 100));
  }
  const soma = [...totais.values()].reduce((a, b) => a + b, 0);
  if (!soma) { el('#categorias').replaceChildren(elemento('p', 'Cadastre uma assinatura ativa com valor para ver a distribuição.', 'estado-vazio')); return; }
  el('#categorias').replaceChildren(...[...totais].sort((a, b) => b[1] - a[1]).map(([nome, valor]) => {
    const linha = elemento('div', undefined, 'linha-categoria'); const rotulo = elemento('div', undefined, 'rotulo-categoria');
    rotulo.append(elemento('span', nome), elemento('strong', reais.format(valor / 100)));
    const barra = elemento('meter', undefined, 'barra'); barra.min = 0; barra.max = soma; barra.value = valor;
    barra.setAttribute('aria-label', nome + ': ' + Math.round(valor / soma * 100) + '% do gasto mensal'); linha.append(rotulo, barra); return linha;
  }));
}

function abrirCadastro() {
  el('#erro-cadastro').hidden = true; el('#dialogo-cadastro').showModal();
}

function ocuparCadastro(ocupado) {
  cadastroOcupado = ocupado;
  for (const id of ['#salvar', '#fechar-cadastro', '#cancelar-cadastro']) el(id).disabled = ocupado;
  el('#salvar').textContent = ocupado ? 'Salvando…' : 'Salvar assinatura';
  el('#formulario').setAttribute('aria-busy', String(ocupado));
}

el('#formulario').addEventListener('submit', async evento => {
  evento.preventDefault(); if (cadastroOcupado) return;
  ocuparCadastro(true); el('#erro-cadastro').hidden = true;
  try {
    await requisicao('/assinaturas', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({
      nome: el('#nome').value.trim(), valor: Number(el('#valor').value), categoria: el('#categoria').value.trim(),
      dia_cobranca: Number(el('#dia-cobranca').value), ativa: el('#ativa').checked,
    }) });
    el('#formulario').reset(); el('#dialogo-cadastro').close(); mensagem('Assinatura salva.');
    await atualizar();
  } catch (erro) { el('#erro-cadastro').textContent = erro.message; el('#erro-cadastro').hidden = false;
  } finally { ocuparCadastro(false); }
});

function abrirRemocao(a) {
  assinaturaRemover = a; el('#texto-remover').textContent = `Você vai remover “${a.nome}”, de ${reais.format(a.valor)} por mês.`;
  el('#erro-remover').hidden = true; el('#dialogo-remover').showModal(); el('#cancelar-remover').focus();
}

el('#formulario-remover').addEventListener('submit', async evento => {
  evento.preventDefault(); if (removendo || !assinaturaRemover) return;
  removendo = true; el('#erro-remover').hidden = true;
  el('#confirmar-remover').disabled = el('#cancelar-remover').disabled = true;
  try {
    await requisicao('/assinaturas/' + assinaturaRemover.id, { method: 'DELETE' });
    el('#dialogo-remover').close(); mensagem('Assinatura removida do rastreador.');
    await atualizar(); el('#assinaturas').focus({ preventScroll: true });
  } catch (erro) { el('#erro-remover').textContent = erro.message; el('#erro-remover').hidden = false;
  } finally { removendo = false; el('#confirmar-remover').disabled = el('#cancelar-remover').disabled = false; }
});

async function carregarExemplos(botao) {
  botao.disabled = true; botao.textContent = 'Adicionando…';
  const exemplo = [['Cinema em casa', 39.90, 'Streaming', 1], ['Academia', 89.90, 'Saúde', 3], ['Curso de idiomas', 59.90, 'Educação', 5], ['Ferramenta de trabalho', 24.90, 'Trabalho', 14]];
  let criadas = 0;
  try {
    for (const [nome, valor, categoria, intervalo] of exemplo) {
      const data = new Date(); data.setDate(data.getDate() + intervalo);
      await requisicao('/assinaturas', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ nome, valor, categoria, dia_cobranca: data.getDate(), ativa: true }) }); criadas++;
    }
    mensagem('Quatro assinaturas fictícias adicionadas. Você pode removê-las quando quiser.');
  } catch (erro) { mensagem(`${criadas} exemplos adicionados. ${erro.message}`, true);
  } finally { await atualizar(); botao.disabled = false; botao.textContent = 'Experimentar com exemplos'; }
}

el('#abrir-cadastro').addEventListener('click', abrirCadastro);
el('#formulario').addEventListener('input', () => { el('#erro-cadastro').hidden = true; });
for (const id of ['#fechar-cadastro', '#cancelar-cadastro']) el(id).addEventListener('click', () => el('#dialogo-cadastro').close());
el('#dialogo-cadastro').addEventListener('cancel', evento => { if (cadastroOcupado) evento.preventDefault(); });
el('#cancelar-remover').addEventListener('click', () => el('#dialogo-remover').close());
el('#dialogo-remover').addEventListener('cancel', evento => { if (removendo) evento.preventDefault(); });
el('#atualizar').addEventListener('click', () => { mensagem(''); atualizar(); });
el('#dias').addEventListener('change', () => { mensagem(''); atualizar(); });
el('#busca').addEventListener('input', renderizarLista);
for (const id of ['#categoria-filtro', '#ordem']) el(id).addEventListener('change', renderizarLista);
atualizar();
