export interface OAuthCallbackParts {
  code: string
  state: string
}

function parseParams(raw: string): OAuthCallbackParts | null {
  const params = new URLSearchParams(raw.replace(/^[?#]/, ''))
  const code = params.get('code')?.trim() ?? ''
  const state = params.get('state')?.trim() ?? ''
  if (!code && !state) return null
  return { code, state }
}

export function parseOAuthCallbackInput(input: string): OAuthCallbackParts {
  const trimmed = input.trim()
  if (!trimmed) return { code: '', state: '' }

  try {
    const url = new URL(trimmed)
    const queryParts = parseParams(url.search)
    const hashParts = parseParams(url.hash)
    return {
      code: queryParts?.code || hashParts?.code || '',
      state: queryParts?.state || hashParts?.state || '',
    }
  } catch {
    const params = parseParams(trimmed)
    if (params) return params
    return { code: trimmed, state: '' }
  }
}
