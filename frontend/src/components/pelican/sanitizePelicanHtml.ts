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

export function sanitizePelicanHtml(html: string, fitSvgPreview = false): string {
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
  if (fitSvgPreview) fitSingleSvgPreview(clean)
  // Serialize the sanitized DOM without extracting/reinserting untrusted style text.
  return `${clean.querySelector('head')?.innerHTML ?? ''}${clean.querySelector('body')?.innerHTML ?? ''}`
}

function fitSingleSvgPreview(clean: HTMLElement): void {
  const body = clean.querySelector('body')
  if (!body) return
  const wrappers: HTMLElement[] = [body]
  let container: Element = body
  // Only normalize a lone SVG and its otherwise empty wrapper chain.
  // Multi-section HTML and SVGs used as decorative icons retain their layout.
  while (true) {
    const children = Array.from(container.children).filter(child => child.localName !== 'style')
    if (children.length !== 1 || Array.from(container.childNodes).some(node => node.nodeType === 3 && node.textContent?.trim())) return
    const child = children[0]!
    if (child.namespaceURI === 'http://www.w3.org/2000/svg' && child.localName === 'svg') {
      const viewBox = (child.getAttribute('viewBox') ?? '').trim().split(/[\s,]+/).map(Number)
      if (viewBox.length !== 4 || viewBox.some(value => !Number.isFinite(value)) || viewBox[2]! <= 0 || viewBox[3]! <= 0) return
      const svg = child as SVGSVGElement
      for (const wrapper of wrappers) {
        for (const [property, value] of Object.entries({ display: 'block', width: '100%', 'min-width': '0', 'max-width': 'none', height: 'auto', 'min-height': '0', 'max-height': 'none', margin: '0', padding: '0', 'box-sizing': 'border-box', position: 'static' })) {
          wrapper.style.setProperty(property, value, 'important')
        }
      }
      for (const [property, value] of Object.entries({ display: 'block', width: '100%', 'min-width': '0', 'max-width': `${580 * viewBox[2]! / viewBox[3]!}px`, height: 'auto', 'min-height': '0', 'max-height': '580px', 'aspect-ratio': `${viewBox[2]} / ${viewBox[3]}`, margin: '0 auto', position: 'static', 'box-sizing': 'border-box' })) {
        svg.style.setProperty(property, value, 'important')
      }
      svg.setAttribute('preserveAspectRatio', 'xMidYMid meet')
      const pageStyle = clean.ownerDocument.createElement('style')
      pageStyle.textContent = 'html,body{display:block!important;width:100%!important;min-width:0!important;max-width:none!important;margin:0!important;padding:0!important;box-sizing:border-box!important}'
      clean.querySelector('head')?.append(pageStyle)
      return
    }
    if (!(child instanceof HTMLElement) || !['div', 'main', 'section', 'article', 'figure'].includes(child.localName)) return
    wrappers.push(child)
    container = child
  }
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
  return `<!doctype html><html><head><meta http-equiv="Content-Security-Policy" content="default-src 'none'; ${scriptPolicy} style-src 'unsafe-inline'; img-src data:; font-src 'none'; connect-src 'none'; frame-src 'none'; form-action 'none'; base-uri 'none'"></head><body>${sanitizePelicanHtml(html, preview)}${previewStyle}${bridge}</body></html>`
}
