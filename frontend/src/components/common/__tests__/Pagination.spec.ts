import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import Pagination from '../Pagination.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (key === 'pagination.pageOf') return `${params?.page}/${params?.total}`
      return key
    },
  }),
}))

function mountPagination(props: Record<string, unknown>) {
  return mount(Pagination, {
    props: {
      total: 2000,
      page: 50,
      pageSize: 20,
      showPageSizeSelector: false,
      ...props,
    },
    global: {
      stubs: {
        Icon: { template: '<span />' },
        Select: { template: '<div />' },
      },
    },
  })
}

describe('Pagination', () => {
  it('keeps compact pagination short on middle pages', () => {
    const wrapper = mountPagination({ compact: true })

    const pageButtons = wrapper.findAll('nav button').map((button) => button.text().trim())

    expect(pageButtons).toEqual(['', '1', '...', '50', '...', '100', ''])
  })

  it('keeps the default pagination range for wide layouts', () => {
    const wrapper = mountPagination({ compact: false })

    const pageButtons = wrapper.findAll('nav button').map((button) => button.text().trim())

    expect(pageButtons).toEqual(['', '1', '...', '48', '49', '50', '51', '52', '...', '100', ''])
  })
})
