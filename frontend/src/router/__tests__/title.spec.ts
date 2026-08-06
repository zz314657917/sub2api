import { describe, expect, it } from 'vitest'
import { resolveDocumentTitle } from '@/router/title'

describe('resolveDocumentTitle', () => {
  it('路由存在标题时，使用“路由标题 - 站点名”格式', () => {
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', () => {
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，回退默认站点名', () => {
    expect(resolveDocumentTitle('Dashboard', '')).toBe('Dashboard - Sub2API')
    expect(resolveDocumentTitle(undefined, '   ')).toBe('Sub2API')
  })

  it('站点名变更时仅影响后续路由标题计算', () => {
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })

  it('Token 拼团标题使用公开设置里的自定义功能名称', () => {
    expect(resolveDocumentTitle('Group Buy', 'My Site', 'nav.groupBuy', {
      group_buy_product_name: '我的拼团',
    } as any)).toBe('我的拼团 - My Site')
  })

  it('像素网吧启用时使用公开设置里的自定义标题', () => {
    expect(resolveDocumentTitle('Group Buy', 'My Site', 'nav.groupBuy', {
      pixel_cafe_enabled: true,
      pixel_cafe_title: '  Token网咖  ',
    } as any)).toBe('Token网咖 - My Site')
  })
})
