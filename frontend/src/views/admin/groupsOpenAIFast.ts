export function supportsGroupOpenAIFast(platform: string): boolean {
  return platform === 'openai' || platform === 'composite'
}

export function normalizeGroupOpenAIFast(
  platform: string,
  enabled: boolean | undefined,
): boolean {
  return supportsGroupOpenAIFast(platform) && enabled === true
}
