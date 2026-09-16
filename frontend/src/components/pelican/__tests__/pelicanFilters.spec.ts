import { describe, expect, it } from 'vitest'
import { buildPelicanListParams } from '@/api/pelicanTests'

describe('pelican list filters', () => {
  it('sends only a valid displayed account ID and selected authorized group ID', () => {
    expect(buildPelicanListParams(8, '#42', 2, 24)).toEqual({ group_id: 8, account_id: 42, page: 2, page_size: 24 })
    expect(buildPelicanListParams('', 'name-is-not-an-id', 1, 24)).toEqual({ group_id: undefined, account_id: undefined, page: 1, page_size: 24 })
  })
})
