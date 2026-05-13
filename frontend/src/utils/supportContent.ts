import type { PublicSettings, SupportPopupItem } from '@/types'

type SupportContentSettings = Pick<PublicSettings, 'contact_info'> & {
  support_popup_items?: SupportPopupItem[]
}

export function hasSupportPopupImages(settings: Partial<SupportContentSettings> | null | undefined): boolean {
  const items = settings?.support_popup_items
  return Array.isArray(items) && items.some((item) => item.title?.trim() && item.image_url?.trim())
}

export function hasSupportContactText(settings: Partial<SupportContentSettings> | null | undefined, fallback = ''): boolean {
  return Boolean((settings?.contact_info || fallback).trim())
}

export function hasSupportContent(settings: Partial<SupportContentSettings> | null | undefined, fallback = ''): boolean {
  return hasSupportPopupImages(settings) || hasSupportContactText(settings, fallback)
}
