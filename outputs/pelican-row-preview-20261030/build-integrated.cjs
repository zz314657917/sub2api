const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const root = __dirname;
const repo = path.resolve(root, '../..');
const requireFrontend = createRequire(path.join(repo, 'frontend/package.json'));
const postcss = requireFrontend('postcss');
const original = fs.readFileSync(path.join(root, 'index.html'), 'utf8');
let shell = original.slice(original.indexOf('    <section class="browse-bar"'), original.indexOf('  <script src="./preview.js">'));
shell = shell.replace('  </main>', '').replace(/    <footer class="page-footer">.*?<\/footer>/, '');
fs.writeFileSync(path.join(root, 'integrated-gallery.html'), shell);

const css = postcss.parse(fs.readFileSync(path.join(root, 'preview.css'), 'utf8'));
css.walkRules(rule => {
  if (rule.parent.type === 'atrule' && /keyframes$/.test(rule.parent.name)) return;
  rule.selectors = rule.selectors.map(selector => {
    if (selector === ':root' || selector === 'body') return '.integrated-gallery';
    if (selector === 'body.dark') return '.dark .integrated-gallery';
    return `.integrated-gallery ${selector}`;
  });
});
css.walkAtRules(/keyframes$/, rule => { rule.params = `pelican-preview-${rule.params}`; });
css.walkDecls(/^animation/, decl => { decl.value = decl.value.replace(/\b(spin|pulse)\b/g, 'pelican-preview-$1'); });
const overrides = `
.console-shell.console-shell-viewport { height: 100vh; height: 100dvh; min-height: 0; }
.console-shell-viewport .console-main-viewport { min-height: 0; height: 100%; }
.console-shell-viewport .console-content-viewport { min-height: 0; }
.integrated-gallery {
  --bg: var(--console-bg); --surface: var(--console-input); --ink: var(--console-text);
  --muted: var(--console-muted); --line: var(--console-border);
  --accent: var(--console-accent); --accent-soft: var(--console-accent-soft);
  --card-border: var(--console-border); --shadow: var(--console-shadow);
  display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; min-width: 0;
  width: 100%; overflow: hidden; background: transparent; font-size: 12px;
}
.dark .integrated-gallery {
  --bg: var(--console-bg); --surface: var(--console-input); --ink: var(--console-text);
  --muted: var(--console-muted); --line: var(--console-border);
  --accent: var(--console-accent); --accent-soft: var(--console-accent-soft);
  --card-border: var(--console-border);
}
.integrated-gallery .browse-bar { flex-shrink: 0; padding-bottom: 12px; flex-wrap: wrap; }
.integrated-gallery .browse-left { flex-wrap: wrap; }
.integrated-gallery .browse-right { margin-left: auto; gap: 8px; }
.integrated-gallery .search-field { width: 215px; }
.integrated-gallery .select-field select { min-width: 0; padding: 0 32px 0 12px; max-width: 180px; }
.integrated-gallery .list-caption { flex-shrink: 0; margin: 13px 0 0; }
.integrated-gallery #groups { flex: 1 1 auto; min-height: 0; overflow-y: auto; padding-right: 3px; scrollbar-width: thin; }
.integrated-gallery .group-row { grid-template-columns: 158px minmax(0, 1fr); gap: 20px; padding: 20px 0 14px; }
.integrated-gallery .group-title { margin: 0 0 4px; font-size: 13px; }
.integrated-gallery .rail { --card-width: 242px; }
.integrated-gallery #empty-state { overflow-y: auto; }
.pelican-page > .pelican-controls > .admin-toggle { margin-right: 8px; }
@media (max-width: 700px) {
  .integrated-gallery .browse-bar { flex-wrap: nowrap; align-items: flex-start; }
  .integrated-gallery .browse-left { min-width: 0; }
  .integrated-gallery .search-field { width: 100%; }
  .integrated-gallery .select-field select { max-width: 180px; }
  .integrated-gallery .browse-right { margin-left: 0; flex-shrink: 0; }
  .integrated-gallery .group-row { display: block; padding: 18px 0 12px; }
  .integrated-gallery .group-title span { display: none; }
}
`;
fs.writeFileSync(path.join(root, 'integrated-gallery.css'), css.toString() + overrides);
let js = fs.readFileSync(path.join(root, 'preview.js'), 'utf8');
// The integrated page uses AppSidebar's existing theme control.
js = js.replace(/\$\('\.theme-toggle'\)\.addEventListener\('click', \(\) => \{[\s\S]*?\n\}\);\n/, '');
js = js.replace("$('.theme-toggle').innerHTML = icon('moon');\n", '');
fs.writeFileSync(path.join(root, 'integrated-gallery.js'), js);
console.log('Integrated gallery generated; production source unchanged.');
