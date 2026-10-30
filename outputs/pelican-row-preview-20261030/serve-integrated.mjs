import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { createRequire } from 'node:module';
const root = path.dirname(fileURLToPath(import.meta.url));
const repo = path.resolve(root, '../..');
const frontend = process.env.PELICAN_PREVIEW_FRONTEND || path.join(repo, 'frontend');
const port = Number(process.env.PELICAN_PREVIEW_PORT || 62080);
process.chdir(frontend);
const req = createRequire(path.join(frontend, 'package.json'));
const { createServer } = await import(pathToFileURL(path.join(path.dirname(req.resolve('vite/package.json')), 'dist/node/index.js')).href);
const { default: vue } = await import(pathToFileURL(req.resolve('@vitejs/plugin-vue')).href);
const gallery = '/@fs/' + path.join(root, 'IntegratedGallery.vue').replaceAll('\\', '/');
const server = await createServer({
  configFile: false,
  root: frontend,
  plugins: [
    {
      name: 'pelican-layout-preview-only',
      enforce: 'pre',
      transform(source, id) {
        const normalizedId = id.replaceAll('\\', '/');
        if (normalizedId === path.join(frontend, 'src/components/pelican/PelicanPlansPanel.vue').replaceAll('\\', '/')) {
          const defaultSetting = 'max_results: 20,';
          const retentionInput = 'min="1" max="50" aria-label="保留结果数"';
          if (!source.includes(defaultSetting) || !source.includes(retentionInput)) throw new Error('Original retention settings changed; refusing preview overlay.');
          return source.replace(defaultSetting, 'max_results: 5,')
            .replace(retentionInput, 'min="1" max="20" title="每个模型默认保留 5 张，最多 20 张" aria-label="保留结果数"')
            .replace('<label>保留结果数<input', '<label>每个模型保留作品数<input');
        }
        if (normalizedId !== path.join(frontend, 'src/views/user/PelicanTestsView.vue').replaceAll('\\', '/')) return;
        const start = source.indexOf('      <section class="toolbar"');
        const end = source.indexOf('      <footer class="pagination">');
        if (start < 0 || end < start) throw new Error('Original Pelican view structure changed; refusing preview overlay.');
        const next = source.slice(0, start) + '      </div>\n      <IntegratedGallery />\n\n' + source.slice(end);
        return next.replace('<script setup lang="ts">', `<script setup lang="ts">\nimport IntegratedGallery from '${gallery}'`);
      },
      configureServer(vite) {
        vite.middlewares.use(async (request, response, next) => {
          const pathname = new URL(request.url, 'http://127.0.0.1').pathname;
          if (pathname.startsWith('/api/')) {
            response.writeHead(404, { 'Content-Type': 'application/json' });
            response.end(JSON.stringify({ code: 404, message: 'Preview has no backend connection.' }));
            return;
          }
          if (pathname.startsWith('/assets/')) {
            const asset = path.resolve(root, '.' + pathname);
            if (!asset.startsWith(path.join(root, 'assets') + path.sep) || !fs.existsSync(asset)) return next();
            response.writeHead(200, { 'Content-Type': asset.endsWith('.svg') ? 'image/svg+xml' : 'application/json', 'Cache-Control': 'no-store' });
            fs.createReadStream(asset).pipe(response);
            return;
          }
          if (pathname === '/' || pathname === '/pelican-tests') {
            const entry = '/@fs/' + path.join(root, 'integrated-main.ts').replaceAll('\\', '/');
            const template = fs.readFileSync(path.join(root, 'integrated.html'), 'utf8')
              .replace(/src="\/@fs\/[^"]*\/integrated-main\.ts"/, `src="${entry}"`);
            const html = await vite.transformIndexHtml(request.url, template);
            response.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' });
            response.end(html);
            return;
          }
          next();
        });
      }
    },
    vue()
  ],
  resolve: {
    alias: {
      '@': path.join(frontend, 'src'),
      'vue-i18n': req.resolve('vue-i18n/dist/vue-i18n.runtime.esm-bundler.js')
    },
    dedupe: ['vue', 'pinia', 'vue-router']
  },
  define: { __INTLIFY_JIT_COMPILATION__: true },
  cacheDir: process.env.PELICAN_PREVIEW_CACHE || path.join(repo, '.cache', `pelican-integrated-preview-vite-${port}`),
  server: { host: '127.0.0.1', port, strictPort: true, fs: { allow: [repo, frontend] } },
  css: { postcss: frontend },
  optimizeDeps: { include: ['vue', 'pinia', 'vue-router', 'vue-i18n', 'axios', 'driver.js'] }
});
await server.listen();
console.log(`Original Pelican layout preview: http://127.0.0.1:${port}/pelican-tests`);
