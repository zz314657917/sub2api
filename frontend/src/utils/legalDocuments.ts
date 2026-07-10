import type { LoginAgreementDocument } from '@/types'

export type LegalDocumentLink = {
  documentId: string
  title: string
}

export type LegalDocumentTitleFallbacks = Record<string, string>

const invalidTitlePattern = /^[?\uFFFD\s]+$/

export function toLegalDocumentLink(
  doc: LoginAgreementDocument,
  titleFallbacks: LegalDocumentTitleFallbacks = {}
): LegalDocumentLink | null {
  const rawTitle = doc.title?.trim() || ''
  const documentId = (doc.id || rawTitle).trim()

  if (!documentId) {
    return null
  }

  const title = invalidTitlePattern.test(rawTitle)
    ? titleFallbacks[documentId]?.trim() || documentId
    : rawTitle || titleFallbacks[documentId]?.trim() || documentId

  return { documentId, title }
}
