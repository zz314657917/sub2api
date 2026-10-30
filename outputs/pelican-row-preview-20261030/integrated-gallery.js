const icons = {
  search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4 4"/>',
  'chevron-down': '<path d="m6 9 6 6 6-6"/>',
  left: '<path d="m14 6-6 6 6 6"/>', right: '<path d="m10 6 6 6-6 6"/>',
  clock: '<circle cx="12" cy="12" r="8"/><path d="M12 7v5l3 2"/>',
  refresh: '<path d="M20 7v5h-5M4 17v-5h5"/><path d="M6.1 7a7 7 0 0 1 11.6-1L20 9M4 15l2.3 3A7 7 0 0 0 18 17"/>',
  check: '<path d="m5 12 4 4L19 6"/>',
  move: '<path d="M3 12h18M7 8l-4 4 4 4m10-8 4 4-4 4"/>',
  expand: '<path d="M9 4H4v5m11-5h5v5M4 15v5h5m11-5v5h-5"/>',
  close: '<path d="m6 6 12 12M6 18 18 6"/>',
  moon: '<path d="M20 14A8 8 0 0 1 10 4a8 8 0 1 0 10 10Z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.4 1.4m11.2 11.2L19 19M5 19l1.4-1.4M17.6 6.4 19 5"/>',
  error: '<path d="M12 7v6m0 4h.01"/>',
};
const icon = (name) => `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true">${icons[name] || ''}</svg>`;
function hydrateIcons(root = document) { root.querySelectorAll('[data-icon]').forEach(element => { element.innerHTML = icon(element.dataset.icon); }); }
const groups = [
  { id: 'gpt', name: '官方 GPT · 稳定组', subtitle: '稳定输出，持续观察', vendor: 'OPENAI', color: '#557c70', models: ['gpt-6-astra', 'gpt-6.1-sol'], plan: 12, records: [
    ['gpt-6-astra', 'success', '12:42', 118.48, 0],
    ['gpt-6.1-sol', 'running', '12:54', null, 1],
    ['gpt-6.1-sol', 'success', '12:38', 109.39, 2],
    ['gpt-6-astra', 'success', '12:26', 150.00, 3],
    ['gpt-6.1-sol', 'success', '12:14', 96.72, 4],
    ['gpt-6-astra', 'success', '12:02', 124.86, 5],
  ] },
  { id: 'claude', name: 'Claude Pro · 优选组', subtitle: '不同模型，一起对照', vendor: 'ANTHROPIC', color: '#c28d73', models: ['claude-sonnet-4', 'claude-opus-4'], plan: 8, records: [
    ['claude-sonnet-4', 'success', '12:42', 83.26, 4],
    ['claude-opus-4', 'success', '12:36', 162.58, 1],
    ['claude-sonnet-4', 'success', '12:24', 78.65, 5],
    ['claude-opus-4', 'failed', '12:12', 180.00, 0],
    ['claude-sonnet-4', 'success', '12:00', 88.13, 3],
    ['claude-opus-4', 'success', '11:48', 155.46, 2],
  ] },
  { id: 'gemini', name: 'Gemini · 日常组', subtitle: '轻快创作，更多灵感', vendor: 'GOOGLE', color: '#668bbd', models: ['gemini-2.5-pro', 'gemini-2.5-flash'], plan: 6, records: [
    ['gemini-2.5-pro', 'success', '12:38', 72.43, 2],
    ['gemini-2.5-flash', 'success', '12:42', 36.82, 3],
    ['gemini-2.5-pro', 'success', '12:23', 84.35, 0],
    ['gemini-2.5-flash', 'success', '12:17', 32.19, 4],
    ['gemini-2.5-pro', 'success', '12:05', 91.26, 1],
    ['gemini-2.5-flash', 'success', '11:59', 38.24, 5],
  ] },
];
const artworkRetention = Object.freeze({ defaultCount: 5, maxCount: 20 });
for (const group of groups) {
  for (const [modelIndex, model] of group.models.entries()) {
    const existingCount = group.records.filter(record => record[0] === model).length;
    for (let index = existingCount; index < artworkRetention.defaultCount; index++) {
      group.records.push([
        model, 'success', `11:${String(48 - (index - existingCount) * 12 - modelIndex * 6).padStart(2, '0')}`,
        62.18 + index * 9.23 + modelIndex * 12, (index + modelIndex) % 6
      ]);
    }
  }
  group.records.sort((a, b) => b[2].localeCompare(a[2]));
}
function retainedRecords(group) {
  const count = Number(group.maxResults ?? artworkRetention.defaultCount);
  const limit = Number.isFinite(count) ? Math.min(artworkRetention.maxCount, Math.max(1, Math.trunc(count))) : artworkRetention.defaultCount;
  const perModel = new Map();
  return group.records.map((record, index) => ({ record, index })).filter(({ record }) => {
    const used = perModel.get(record[0]) ?? 0;
    perModel.set(record[0], used + 1);
    return used < limit;
  });
}
const statusNames = { success: '生成成功', running: '生成中', failed: '生成失败' };
const selectedModels = Object.fromEntries(groups.map(group => [group.id, 'all']));
const $ = (selector) => document.querySelector(selector);
const esc = (value) => String(value).replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[char]));
let logos = {};
function logo(group) { return `<span class="model-logo" aria-hidden="true">${logos[group.id] || '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8" fill="none" stroke="currentColor"/><circle cx="12" cy="12" r="3"/></svg>'}</span>`; }
function card(group, record, index) {
  const [model, status, time, duration, art] = record;
  const visual = status === 'success'
    ? `<button class="art-button" data-group="${group.id}" data-index="${index}" aria-label="放大 ${esc(model)} ${time} 的作品"><img src="assets/scene-${art}.svg" alt="鹈鹕在海边骑自行车的演示作品" loading="lazy"><span class="round-label">第 ${18 - index} 轮</span><span class="expand-icon">${icon('expand')}</span></button>`
    : `<div class="art-placeholder ${status}">${status === 'running' ? '<span class="running-orbit" aria-hidden="true"></span><strong>灵感正在路上</strong><small>正在生成本轮作品</small>' : `<span class="failed-symbol">${icon('error')}</span><strong>本轮未完成</strong><small>生成超时 · 可查看其他轮次</small>`}</div>`;
  return `<article class="result-card" data-model-id="${esc(model)}">${visual}<div class="card-copy"><h3 class="model-title">${logo(group)}<span class="model-name" title="${esc(model)}">${esc(model)}</span><span class="status ${status}"><i></i>${statusNames[status]}</span></h3><div class="card-footer"><time datetime="2026-10-30T${time}:00+08:00" title="2026-10-30 ${time}:00 (UTC+08:00)">10 月 30 日 ${time}</time><span class="duration">${icon('clock')}${duration == null ? '—' : `${duration.toFixed(2)} 秒`}</span></div></div></article>`;
}
function filteredRecords(group) {
  const query = $('#search').value.trim().toLowerCase();
  return retainedRecords(group).filter(({ record }) =>
    (selectedModels[group.id] === 'all' || record[0] === selectedModels[group.id]) &&
    ($('#status-filter').value === 'all' || record[1] === $('#status-filter').value) &&
    (!query || group.name.toLowerCase().includes(query) || record[0].includes(query))
  );
}
function updateNavigation(row) {
  const rail = row.querySelector('.rail'), cards = rail.querySelectorAll('.result-card');
  const max = rail.scrollWidth - rail.clientWidth;
  row.querySelector('[data-direction="left"]').disabled = rail.scrollLeft < 2 || max < 2;
  row.querySelector('[data-direction="right"]').disabled = rail.scrollLeft > max - 2 || max < 2;
  const first = cards.length ? Math.min(cards.length, Math.round(rail.scrollLeft / (cards[0].offsetWidth + 14)) + 1) : 0;
  row.querySelector('.position-label').textContent = `${first} / ${cards.length}`;
}
function render() {
  const query = $('#search').value.trim().toLowerCase(), groupId = $('#group-filter').value;
  let visibleGroups = 0, visibleRecords = 0;
  $('#groups').innerHTML = groups.filter(group => (groupId === 'all' || group.id === groupId) && (!query || group.name.toLowerCase().includes(query) || group.models.some(model => model.includes(query)))).map(group => {
    const entries = filteredRecords(group);
    visibleGroups++; visibleRecords += entries.length;
    return `<section class="group-row" data-group="${group.id}" aria-label="${esc(group.name)}"><aside class="group-meta"><div class="platform" style="--brand-color:${group.color}">${group.vendor}</div><h2 class="group-title">${esc(group.name)}<span>${group.subtitle}</span></h2><p class="group-count"><strong>${retainedRecords(group).length}</strong> 条演示记录</p><nav class="model-filters" aria-label="${esc(group.name)}模型筛选">${['all', ...group.models].map(model => `<button type="button" class="model-filter ${selectedModels[group.id] === model ? 'active' : ''}" data-model="${model}" aria-pressed="${selectedModels[group.id] === model}">${model === 'all' ? '全部模型' : esc(model)}${icon('check')}</button>`).join('')}</nav><div class="row-navigation"><button class="icon-button" data-direction="left" aria-label="向前浏览${esc(group.name)}">${icon('left')}</button><button class="icon-button" data-direction="right" aria-label="向后浏览${esc(group.name)}">${icon('right')}</button><span class="position-label"></span></div></aside><div class="row-content"><div class="rail" tabindex="0" role="region" aria-label="${esc(group.name)}作品列表">${entries.length ? entries.map(({ record, index }) => card(group, record, index)).join('') : '<div class="rail-empty"><span>该模型暂无符合条件的结果</span><button class="text-button" data-clear-row>查看本组全部结果</button></div>'}</div></div></section>`;
  }).join('');
  $('#group-count').textContent = visibleGroups;
  $('#result-count').textContent = visibleRecords;
  $('#empty-state').hidden = visibleGroups > 0;
  document.querySelectorAll('.group-row').forEach(row => {
    row.querySelectorAll('[data-model]').forEach(button => button.addEventListener('click', () => { selectedModels[row.dataset.group] = button.dataset.model; render(); document.querySelector(`.group-row[data-group="${row.dataset.group}"] [data-model="${button.dataset.model}"]`)?.focus({ preventScroll: true }); }));
    row.querySelectorAll('[data-direction]').forEach(button => button.addEventListener('click', () => { const rail = row.querySelector('.rail'); rail.scrollBy({ left: (button.dataset.direction === 'right' ? 1 : -1) * (rail.clientWidth * .85), behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' }); }));
    row.querySelector('[data-clear-row]')?.addEventListener('click', () => { selectedModels[row.dataset.group] = 'all'; $('#status-filter').value = 'all'; $('#search').value = ''; render(); });
    row.querySelector('.rail').addEventListener('scroll', () => updateNavigation(row), { passive: true });
    updateNavigation(row);
  });
  document.querySelectorAll('.art-button').forEach(button => button.addEventListener('click', () => {
    const group = groups.find(item => item.id === button.dataset.group), record = group.records[Number(button.dataset.index)];
    $('#dialog-group').textContent = `${group.name} · 计划 #${group.plan}`;
    $('#dialog-title').textContent = record[0];
    $('#dialog-art').innerHTML = `<img src="assets/scene-${record[4]}.svg" alt="放大的鹈鹕骑行演示作品">`;
    $('#dialog-meta').textContent = `2026-10-30 ${record[2]} · 耗时 ${record[3].toFixed(2)} 秒`;
    $('#art-dialog').showModal();
  }));
}
for (const group of groups) $('#group-filter').insertAdjacentHTML('beforeend', `<option value="${group.id}">${esc(group.name)}</option>`);
$('#search').addEventListener('input', render);
$('#group-filter').addEventListener('change', render);
$('#status-filter').addEventListener('change', render);
$('#reset').addEventListener('click', () => { $('#search').value = ''; $('#group-filter').value = 'all'; $('#status-filter').value = 'all'; render(); });
$('#refresh').addEventListener('click', () => { $('#updated-at').textContent = `${new Date().toLocaleTimeString('zh-CN', { hour12: false })} 更新`; render(); });
$('#dialog-close').addEventListener('click', () => $('#art-dialog').close());
$('#art-dialog').addEventListener('click', event => { if (event.target === $('#art-dialog')) { const box = event.target.getBoundingClientRect(); if (event.clientX < box.left || event.clientX > box.right || event.clientY < box.top || event.clientY > box.bottom) event.target.close(); } });
addEventListener('keydown', event => { if (event.key === '/' && !['INPUT', 'SELECT', 'TEXTAREA'].includes(document.activeElement.tagName) && !$('#art-dialog').open) { event.preventDefault(); $('#search').focus(); } });
addEventListener('resize', () => document.querySelectorAll('.group-row').forEach(updateNavigation));
hydrateIcons();
render();
fetch('./assets/logos.json').then(response => response.json()).then(result => { logos = result; render(); }).catch(() => {});
