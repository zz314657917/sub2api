import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, patch, remove, credentialGet, credentialCreate } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  patch: vi.fn(),
  remove: vi.fn(),
  credentialGet: vi.fn(),
  credentialCreate: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    patch,
    delete: remove
  }
}))

import { passkeyAPI } from '@/api/passkey'

class FakePublicKeyCredential {
  id = 'credential-id'
  rawId = Uint8Array.from([1, 2, 3]).buffer
  type = 'public-key'
  authenticatorAttachment = 'platform'
  response = {
    authenticatorData: Uint8Array.from([4, 5]).buffer,
    clientDataJSON: Uint8Array.from([6, 7]).buffer,
    signature: Uint8Array.from([8, 9]).buffer,
    userHandle: Uint8Array.from([10, 11]).buffer
  }

  getClientExtensionResults(): AuthenticationExtensionsClientOutputs {
    return {}
  }
}

class FakeRegistrationCredential extends FakePublicKeyCredential {
  response = {
    attestationObject: Uint8Array.from([12, 13]).buffer,
    clientDataJSON: Uint8Array.from([6, 7]).buffer,
    getTransports: () => ['internal']
  }
}

describe('passkey api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    patch.mockReset()
    remove.mockReset()
    credentialGet.mockReset()
    credentialCreate.mockReset()

    vi.stubGlobal('PublicKeyCredential', FakePublicKeyCredential)
    Object.defineProperty(window, 'PublicKeyCredential', {
      configurable: true,
      value: FakePublicKeyCredential
    })
    Object.defineProperty(navigator, 'credentials', {
      configurable: true,
      value: { get: credentialGet, create: credentialCreate }
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('converts assertion options and response bytes to WebAuthn JSON', async () => {
    post
      .mockResolvedValueOnce({
        data: {
          session_token: 'one-time-session',
          options: {
            publicKey: {
              challenge: 'AQID',
              rpId: 'sub2api.example.com',
              userVerification: 'required'
            }
          }
        }
      })
      .mockResolvedValueOnce({
        data: {
          access_token: 'access',
          token_type: 'Bearer',
          user: { id: 1 }
        }
      })
    credentialGet.mockResolvedValue(new FakePublicKeyCredential())

    await passkeyAPI.login()

    const request = credentialGet.mock.calls[0][0] as CredentialRequestOptions
    expect(Array.from(new Uint8Array(request.publicKey!.challenge))).toEqual([1, 2, 3])
    expect(request.publicKey!.userVerification).toBe('required')

    expect(post).toHaveBeenNthCalledWith(2, '/auth/passkey/login/finish', {
      session_token: 'one-time-session',
      credential: {
        id: 'credential-id',
        rawId: 'AQID',
        type: 'public-key',
        authenticatorAttachment: 'platform',
        clientExtensionResults: {},
        response: {
          authenticatorData: 'BAU',
          clientDataJSON: 'Bgc',
          signature: 'CAk',
          userHandle: 'Cgs'
        }
      }
    })
  })

  it('sends the account password when beginning registration', async () => {
    post
      .mockResolvedValueOnce({
        data: {
          session_token: 'register-session',
          options: {
            publicKey: {
              challenge: 'AQID',
              user: { id: 'BAU', name: 'user@example.com', displayName: 'user' }
            }
          }
        }
      })
      .mockResolvedValueOnce({
        data: { id: 3, name: 'Laptop', created_at: '2026-07-28T00:00:00Z', backup: false }
      })
    credentialCreate.mockResolvedValue(new FakeRegistrationCredential())

    await passkeyAPI.register('Laptop', 'correct-password')

    expect(post).toHaveBeenNthCalledWith(1, '/user/passkeys/register/begin', {
      password: 'correct-password'
    })
  })

  it('sends the account password when revoking a credential', async () => {
    remove.mockResolvedValue({ data: null })

    await passkeyAPI.remove(12, 'correct-password')

    expect(remove).toHaveBeenCalledWith('/user/passkeys/12', {
      data: { password: 'correct-password' }
    })
  })
})
