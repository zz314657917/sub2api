import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const readSource = (relativePath: string) => readFileSync(resolve(process.cwd(), relativePath), 'utf8')

describe('Ops error detail list navigation', () => {
  it('tracks the invoking list and restores it without resetting state', () => {
    const dashboard = readSource('src/views/admin/ops/OpsDashboard.vue')
    const errorList = readSource('src/views/admin/ops/components/OpsErrorDetailsModal.vue')
    const requestList = readSource('src/views/admin/ops/components/OpsRequestDetailsModal.vue')
    const detail = readSource('src/views/admin/ops/components/OpsErrorDetailModal.vue')

    expect(dashboard).toContain("type DetailReturnTarget = 'errorList' | 'requestList' | null")
    expect(dashboard).toContain(':resume-state="resumeListState"')
    expect(dashboard).toContain('@back="handleBackToList"')
    expect(dashboard).toContain("detailReturnTarget.value = showRequestDetails.value ? 'requestList' : showErrorDetails.value ? 'errorList' : null")
    expect(dashboard).toContain('resumeListState.value = true')
    expect(errorList).toContain('resumeState?: boolean')
    expect(errorList).toContain('if (props.resumeState) return')
    expect(requestList).toContain('resumeState?: boolean')
    expect(requestList).toContain('if (props.resumeState) return')
    expect(detail).toContain('data-testid="error-detail-back-to-list"')
    expect(detail).toContain("(e: 'back'): void")
  })
})
