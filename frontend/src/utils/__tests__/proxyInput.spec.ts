import { describe, expect, it } from 'vitest'
import { parseProxyInput } from '@/utils/proxyInput'

describe('parseProxyInput', () => {
  it('parses the socks5h host-port-user-password format', () => {
    expect(parseProxyInput('socks5h://203.0.113.10:9004:test-user:test-pass')).toEqual({
      protocol: 'socks5h',
      host: '203.0.113.10',
      port: 9004,
      username: 'test-user',
      password: 'test-pass'
    })
  })

  it('keeps standard URL credentials and decodes them once', () => {
    expect(parseProxyInput('socks5h://user%40name:p%3Ass@proxy.example.com:1080')).toEqual({
      protocol: 'socks5h',
      host: 'proxy.example.com',
      port: 1080,
      username: 'user@name',
      password: 'p:ss'
    })
  })

  it('supports no-auth, bare, and IPv6 inputs', () => {
    expect(parseProxyInput('http://proxy.example.com:80')).toMatchObject({
      protocol: 'http',
      host: 'proxy.example.com',
      port: 80
    })
    expect(parseProxyInput('socks5h://proxy.example.com:8080')).toMatchObject({
      protocol: 'socks5h',
      host: 'proxy.example.com',
      port: 8080
    })
    expect(parseProxyInput('proxy.example.com:8080:alice:secret', { defaultProtocol: 'socks5' })).toMatchObject({
      protocol: 'socks5',
      host: 'proxy.example.com',
      port: 8080,
      username: 'alice',
      password: 'secret'
    })
    expect(parseProxyInput('socks5h://[2001:db8::1]:1080')).toMatchObject({
      host: '2001:db8::1',
      port: 1080
    })
  })

  it('preserves additional colons in the compact password', () => {
    expect(parseProxyInput('socks5h://host.example:9004:user:pass:with:colon')).toMatchObject({
      username: 'user',
      password: 'pass:with:colon'
    })
  })

  it('rejects malformed protocols, ports, paths, and missing schemes when required', () => {
    expect(parseProxyInput('ftp://host:8080')).toBeNull()
    expect(parseProxyInput('socks5h://host:0')).toBeNull()
    expect(parseProxyInput('socks5h://host:70000')).toBeNull()
    expect(parseProxyInput('socks5h://host:8080/path')).toBeNull()
    expect(parseProxyInput('socks5h://host:8080:user:pass?query')).toBeNull()
    expect(parseProxyInput('host:8080', { requireProtocol: true })).toBeNull()
    expect(parseProxyInput('')).toBeNull()
  })
})
