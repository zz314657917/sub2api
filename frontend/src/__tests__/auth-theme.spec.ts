import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const registerView = readFileSync(resolve(process.cwd(), 'src/views/auth/RegisterView.vue'), 'utf8')
const loginView = readFileSync(resolve(process.cwd(), 'src/views/auth/LoginView.vue'), 'utf8')

describe('auth page visual direction', () => {
  it('keeps the register card title readable on the dark glass panel', () => {
    expect(registerView).toContain('text-2xl font-bold text-white')
    expect(registerView).toContain('text-sm font-medium text-violet-100/74')
    expect(registerView).not.toContain('text-2xl font-bold text-gray-900 dark:text-white')
  })

  it('defaults login agreement checkboxes to accepted on login and register pages', () => {
    const expectedDefault = "loginAgreementMode.value === 'checkbox'"

    expect(loginView).toContain(expectedDefault)
    expect(registerView).toContain(expectedDefault)
  })
})
