export const SUPPORT_POPUP_EVENT = 'sub2api:open-support-popup'

export function openSupportPopup(): void {
  window.dispatchEvent(new Event(SUPPORT_POPUP_EVENT))
}
