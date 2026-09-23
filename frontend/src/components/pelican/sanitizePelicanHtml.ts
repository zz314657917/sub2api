import DOMPurify from 'dompurify'

const unsafeCss = /\\|@import\b|url\s*\(|expression\s*\(|-moz-binding|image-set\s*\(/i
const embeddedImage = /^data:image\/(?:png|gif|jpeg|webp|avif|svg\+xml);base64,[a-z\d+/=\s]+$/i
const svgAnimations = new Set(['animate', 'animatemotion', 'animatetransform', 'animatecolor', 'set'])
const animatedAttributes = new Set([
  'x', 'y', 'x1', 'x2', 'y1', 'y2', 'cx', 'cy', 'r', 'rx', 'ry', 'width', 'height',
  'd', 'points', 'transform', 'opacity', 'fill', 'fill-opacity', 'stroke',
  'stroke-width', 'stroke-opacity', 'stroke-dasharray', 'stroke-dashoffset',
  'color', 'stop-color', 'stop-opacity', 'offset', 'visibility',
])
function hasUnsafeCss(css: string): boolean {
  return unsafeCss.test(css.replace(/url\(\s*(['"]?)#[\w-]+\1\s*\)/gi, ''))
}

export function sanitizePelicanHtml(html: string): string {
  const clean = DOMPurify.sanitize(html, {
    WHOLE_DOCUMENT: true, RETURN_DOM: true,
    ADD_TAGS: ['animate', 'animatemotion', 'animatetransform', 'set'],
    ADD_ATTR: ['from', 'to', 'by', 'calcmode'],
    // Declarative SVG animation is allowed; executable JavaScript and embedded
    // documents remain blocked by DOMPurify and the empty iframe sandbox.
    FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'meta', 'base', 'link', 'a', 'foreignObject', 'audio', 'video', 'source'],
    FORBID_ATTR: ['srcset', 'action', 'formaction', 'ping', 'target', 'download', 'autofocus', 'nonce'],
    ALLOW_DATA_ATTR: false,
  }) as HTMLElement
  for (const element of clean.querySelectorAll('*')) {
    if (svgAnimations.has(element.localName.toLowerCase())) {
      const target = element.getAttribute('attributeName')?.toLowerCase() ?? ''
      const isMotion = element.localName.toLowerCase() === 'animatemotion'
      const unsafeValue = ['from', 'to', 'by', 'values'].some(name => {
        const value = element.getAttribute(name) ?? ''
        return hasUnsafeCss(value) || /(?:https?:|data:|javascript:|\/\/)/i.test(value)
      })
      if (element.namespaceURI !== 'http://www.w3.org/2000/svg' ||
          (!isMotion && !animatedAttributes.has(target)) || unsafeValue) {
        element.remove()
        continue
      }
    }
    for (const attribute of Array.from(element.attributes)) {
      const name = attribute.name.toLowerCase()
      if (['href', 'xlink:href', 'src', 'poster', 'background'].includes(name)) {
        const isImage = ['img', 'image'].includes(element.localName) && embeddedImage.test(attribute.value)
        const isLocalSvgReference = element.namespaceURI === 'http://www.w3.org/2000/svg' && /^#[\w-]+$/.test(attribute.value)
        if (!isImage && !isLocalSvgReference) element.removeAttributeNode(attribute)
      }
      if (name === 'style' && hasUnsafeCss(attribute.value)) element.removeAttributeNode(attribute)
    }
    if (element.localName === 'style' && hasUnsafeCss(element.textContent ?? '')) element.remove()
  }
  // Serialize the sanitized DOM without extracting/reinserting untrusted style text.
  return `${clean.querySelector('head')?.innerHTML ?? ''}${clean.querySelector('body')?.innerHTML ?? ''}`
}
// Constant application code: artwork never enters this executable string.
const measurementScript = `(() => {
  const token = document.currentScript.dataset.token;
  let pending = false, previous = -1, sequence = 0;
  function measure() {
    pending = false;
    const body = document.body;
    const style = getComputedStyle(body);
    let bottom = body.offsetTop + body.offsetHeight;
    // Layout offsets ignore transform-only animation. Unlike document scrollHeight,
    // they have no viewport floor, so shrinking content can shrink the iframe too.
    for (const element of Array.from(body.children).slice(0, 128)) {
      if (!(element instanceof HTMLElement) || !element.offsetHeight) continue;
      bottom = Math.max(bottom, element.offsetTop + element.offsetHeight);
    }
    const height = Math.max(1, Math.min(580, Math.ceil(bottom + (parseFloat(style.marginBottom) || 0))));
    if (height === previous) return;
    previous = height;
    parent.postMessage({ type: 'pelican-preview-height', token, height, sequence: ++sequence }, '*');
  }
  function schedule() {
    if (pending) return;
    pending = true;
    setTimeout(measure, 80);
  }
  const observer = new ResizeObserver(schedule);
  observer.observe(document.body);
  // Body catches normal-flow descendants, including SVG intrinsic-size changes.
  // Bounded direct-child observation also covers absolute-positioned canvases.
  for (const element of Array.from(document.body.children).slice(0, 128)) observer.observe(element);
  addEventListener('resize', schedule);
  addEventListener('load', schedule, true);
  schedule();
})();`

export function toPelicanSrcdoc(html: string, preview = false, identity?: { nonce: string; token: string }): string {
  // The public helper accepts only the app's hex secrets, never arbitrary script
  // attribute text. No identity means a script-free document, including details.
  const measured = preview && identity && /^[A-Za-z0-9+/_=-]{16,256}$/.test(identity.nonce) && /^[a-f0-9]{48}$/.test(identity.token)
  const scriptPolicy = measured ? `script-src 'nonce-${identity.nonce}'; script-src-attr 'none';` : "script-src 'none';"
  const previewStyle = preview ? '<style>html,body{height:auto!important;min-height:0!important;max-height:none!important;overflow:hidden!important}*{scrollbar-width:none!important}*::-webkit-scrollbar{display:none!important}</style>' : ''
  const bridge = measured ? `<script nonce="${identity.nonce}" data-token="${identity.token}">${measurementScript}</script>` : ''
  return `<!doctype html><html><head><meta http-equiv="Content-Security-Policy" content="default-src 'none'; ${scriptPolicy} style-src 'unsafe-inline'; img-src data:; font-src 'none'; connect-src 'none'; frame-src 'none'; form-action 'none'; base-uri 'none'"></head><body>${sanitizePelicanHtml(html)}${previewStyle}${bridge}</body></html>`
}
