import { describe, expect, it } from 'vitest'
import { toLegalDocumentLink } from '../legalDocuments'

describe('toLegalDocumentLink', () => {
  it('keeps a valid configured title', () => {
    expect(toLegalDocumentLink({
      id: 'terms',
      title: '自定义服务条款',
      content_md: '# Terms'
    }, {
      terms: '服务条款'
    })).toEqual({
      documentId: 'terms',
      title: '自定义服务条款'
    })
  })

  it('uses the localized fallback for question-mark mojibake', () => {
    expect(toLegalDocumentLink({
      id: 'usage-policy',
      title: '????',
      content_md: '# Policy'
    }, {
      'usage-policy': '使用政策'
    })).toEqual({
      documentId: 'usage-policy',
      title: '使用政策'
    })
  })

  it('uses the localized fallback for replacement-character mojibake', () => {
    expect(toLegalDocumentLink({
      id: 'supported-regions',
      title: '\uFFFD\uFFFD\uFFFD\uFFFD',
      content_md: '# Regions'
    }, {
      'supported-regions': '支持地区'
    })).toEqual({
      documentId: 'supported-regions',
      title: '支持地区'
    })
  })

  it('drops documents without an id or usable title', () => {
    expect(toLegalDocumentLink({
      id: '',
      title: '',
      content_md: ''
    })).toBeNull()
  })
})
