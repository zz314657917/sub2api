import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const leaderboardView = readFileSync(resolve(process.cwd(), 'src/views/user/LeaderboardView.vue'), 'utf8')

describe('leaderboard visual identity', () => {
  it('hides the secondary account line when it repeats the display name', () => {
    expect(leaderboardView).toContain('function shouldShowAccountHint(item: UserLeaderboardItem): boolean')
    expect(leaderboardView).toContain('displayName !== emailMasked')
    expect(leaderboardView).toContain('v-if="shouldShowAccountHint(item)"')
    expect(leaderboardView).toContain('v-if="myEntry && shouldShowAccountHint(myEntry)"')
  })
})
