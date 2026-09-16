import DOMPurify from 'dompurify'

const unsafeCss = /\\|@import\b|url\s*\(|expression\s*\(|-moz-binding|image-set\s*\(/i
const embeddedImage = /^data:image\/(?:png|gif|jpeg|webp|avif|svg\+xml);base64,[a-z\d+/=\s]+$/i
function hasUnsafeCss(css: string): boolean {
  return unsafeCss.test(css.replace(/url\(\s*(['"]?)#[\w-]+\1\s*\)/gi, ''))
}

export function sanitizePelicanHtml(html: string): string {
  const clean = DOMPurify.sanitize(html, {
    WHOLE_DOCUMENT: true, RETURN_DOM: true,
    FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'meta', 'base', 'link', 'a', 'animate', 'animateMotion', 'animateTransform', 'set', 'foreignObject', 'audio', 'video', 'source'],
    FORBID_ATTR: ['srcset', 'action', 'formaction', 'ping', 'target', 'download', 'autofocus'],
    ALLOW_DATA_ATTR: false,
  }) as HTMLElement
  for (const element of clean.querySelectorAll('*')) {
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
export function toPelicanSrcdoc(html: string): string {
  return `<!doctype html><html><head><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:; font-src 'none'; connect-src 'none'; frame-src 'none'; form-action 'none'; base-uri 'none'"></head><body>${sanitizePelicanHtml(html)}</body></html>`
}
