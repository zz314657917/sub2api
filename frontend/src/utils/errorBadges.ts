export function statusCodeBadgeClass(code: number): string {
  if (code >= 500) return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (code === 429) return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
  if (code >= 400) return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
  return 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-200'
}

export function mapUserErrorSortKey(key: string): 'created_at' | 'model' | 'status_code' {
  if (key === 'model') return 'model'
  if (key === 'status') return 'status_code'
  return 'created_at'
}

export const COMMON_ERROR_STATUS_CODES = [
  400, 401, 403, 404, 408, 413, 429, 499, 500, 502, 503, 504, 529,
]
