import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('subscription plan validity locale keys', () => {
  it('keeps the English keys with unit-neutral copy', () => {
    expect(en.payment.admin).toMatchObject({
      validityDays: 'Validity',
      validityDaysRequired: 'Validity must be greater than 0'
    })
  })

  it('keeps the Chinese keys with unit-neutral copy', () => {
    expect(zh.payment.admin).toMatchObject({
      validityDays: '有效期',
      validityDaysRequired: '有效期必须大于 0'
    })
  })
})
