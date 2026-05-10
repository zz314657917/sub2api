import { describe, expect, it } from 'vitest'

import { parseOAuthCallbackInput } from '../oauthCallback'

describe('parseOAuthCallbackInput', () => {
  it('extracts code and state from a full callback URL', () => {
    const parsed = parseOAuthCallbackInput('https://app.example.com/auth/callback?code=abc123&state=state456')

    expect(parsed).toEqual({
      code: 'abc123',
      state: 'state456',
    })
  })

  it('keeps a raw authorization code when the input is not a URL', () => {
    const parsed = parseOAuthCallbackInput('plain-code')

    expect(parsed).toEqual({
      code: 'plain-code',
      state: '',
    })
  })
})
