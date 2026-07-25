import type { ProxyProtocol } from '@/types'

export interface ParsedProxyInput {
  protocol: ProxyProtocol
  host: string
  port: number
  username: string
  password: string
}

export interface ParseProxyInputOptions {
  defaultProtocol?: ProxyProtocol
  requireProtocol?: boolean
}

const supportedProtocols = new Set<ProxyProtocol>(['http', 'https', 'socks5', 'socks5h'])

function normalizeProtocol(value: string): ProxyProtocol | null {
  const protocol = value.toLowerCase() as ProxyProtocol
  return supportedProtocols.has(protocol) ? protocol : null
}

function decodeCredential(value: string): string | null {
  try {
    return decodeURIComponent(value)
  } catch {
    return null
  }
}

function parsePort(value: string): number | null {
  if (!/^\d+$/.test(value)) return null
  const port = Number(value)
  return Number.isInteger(port) && port >= 1 && port <= 65535 ? port : null
}

function normalizeHost(value: string): string | null {
  const host = value.trim()
  if (!host || /[\s/?#@]/.test(host)) return null
  return host
}

function parseColonSeparatedAuthority(
  authority: string,
  protocol: ProxyProtocol
): ParsedProxyInput | null {
  if (/[\s/?#]/.test(authority)) return null

  let host = ''
  let rest = ''

  if (authority.startsWith('[')) {
    const closingBracket = authority.indexOf(']')
    if (closingBracket <= 1) return null
    host = authority.slice(1, closingBracket)
    rest = authority.slice(closingBracket + 1)
  } else {
    const firstColon = authority.indexOf(':')
    if (firstColon <= 0) return null
    host = authority.slice(0, firstColon)
    rest = authority.slice(firstColon)
  }

  const match = rest.match(/^:(\d+)(?::([^:]*):([\s\S]*))?$/)
  if (!match) return null

  const port = parsePort(match[1])
  const normalizedHost = normalizeHost(host)
  if (port === null || normalizedHost === null) return null

  const username = match[2] === undefined ? '' : decodeCredential(match[2])
  const password = match[3] === undefined ? '' : decodeCredential(match[3])
  if (username === null || password === null) return null

  return { protocol, host: normalizedHost, port, username, password }
}

function parseUrlAuthority(authority: string, protocol: ProxyProtocol): ParsedProxyInput | null {
  if (!authority || /[\s/?#]/.test(authority)) return null

  let parsed: URL
  try {
    // Use a non-default scheme so URL keeps explicit ports such as http:80.
    parsed = new URL(`socks5h://${authority}`)
  } catch {
    return null
  }

  if (parsed.pathname || parsed.search || parsed.hash || !parsed.port) return null

  const host = normalizeHost(parsed.hostname.replace(/^\[|\]$/g, ''))
  const port = parsePort(parsed.port)
  const username = decodeCredential(parsed.username)
  const password = decodeCredential(parsed.password)
  if (host === null || port === null || username === null || password === null) return null

  return { protocol, host, port, username, password }
}

/**
 * Parse a proxy URL or the compact host:port:user:password form used by proxy lists.
 * Returns null for empty or malformed input so callers can show a safe generic error.
 */
export function parseProxyInput(
  raw: string,
  options: ParseProxyInputOptions = {}
): ParsedProxyInput | null {
  const trimmed = raw.trim()
  if (!trimmed) return null

  const defaultProtocol = options.defaultProtocol ?? 'http'
  if (!supportedProtocols.has(defaultProtocol)) return null

  const schemeMatch = trimmed.match(/^([a-z][a-z\d+.-]*):\/\/(.+)$/i)
  if (options.requireProtocol && !schemeMatch) return null

  const protocol = normalizeProtocol(schemeMatch?.[1] ?? defaultProtocol)
  if (protocol === null) return null

  const authority = schemeMatch?.[2] ?? trimmed
  // The compact form has no @ and cannot be parsed by URL because its port is
  // followed by credentials. Try it first, then fall back to normal URL syntax.
  return (
    (!authority.includes('@') && parseColonSeparatedAuthority(authority, protocol)) ||
    parseUrlAuthority(authority, protocol)
  )
}
