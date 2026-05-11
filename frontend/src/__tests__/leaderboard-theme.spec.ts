import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const leaderboardView = readFileSync(resolve(process.cwd(), 'src/views/user/LeaderboardView.vue'), 'utf8')

describe('leaderboard visual identity', () => {
  it('renders each user identity on one line with a masked-account fallback', () => {
    expect(leaderboardView).toContain('function getLeaderboardDisplayName(item: UserLeaderboardItem): string')
    expect(leaderboardView).toContain("item.display_name?.trim() || item.email_masked?.trim()")
    expect(leaderboardView).toContain('{{ getLeaderboardDisplayName(item) }}')
    expect(leaderboardView).toContain('getInitial(getLeaderboardDisplayName(item))')
    expect(leaderboardView).not.toContain('shouldShowAccountHint')
    expect(leaderboardView).not.toContain('{{ item.email_masked }}')
  })
})
