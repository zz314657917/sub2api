import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const leaderboardView = readFileSync(resolve(process.cwd(), 'src/views/user/LeaderboardView.vue'), 'utf8')

describe('leaderboard visual identity', () => {
  it('renders each user identity on one line with a masked-account fallback', () => {
    expect(leaderboardView).toContain('function getLeaderboardDisplayName(item: UserLeaderboardItem): string')
    expect(leaderboardView).toContain("item.display_name?.trim() || item.email_masked?.trim()")
    expect(leaderboardView).toContain('{{ getLeaderboardDisplayName(item) }}')
    expect(leaderboardView).not.toContain('shouldShowAccountHint')
    expect(leaderboardView).not.toContain('{{ item.email_masked }}')
  })

  it('uses a lightweight rolling odometer for total token volume', () => {
    expect(leaderboardView).toContain('leaderboard-token-odometer')
    expect(leaderboardView).toContain('leaderboard-token-summary-inner')
    expect(leaderboardView).toContain('leaderboard-token-summary-main')
    expect(leaderboardView).toContain('data-testid="leaderboard-recent-token-trend"')
    expect(leaderboardView).toContain('recentTokenTrendChartData')
    expect(leaderboardView).toContain('leaderboard.recentTokenTrend.title')
    expect(leaderboardView).toContain('rollingTokenParts')
    expect(leaderboardView).toContain('function digitReelStyle(value: string')
    expect(leaderboardView).toContain('transform: translateY(var(--target-offset))')
    expect(leaderboardView).not.toContain('formatDateRange(')
    expect(leaderboardView).toContain('@media (prefers-reduced-motion: reduce)')
  })

  it('focuses the main ranking on token usage bars', () => {
    expect(leaderboardView).toContain('data-testid="leaderboard-token-ranking"')
    expect(leaderboardView).toContain('leaderboard-token-bar-fill')
    expect(leaderboardView).toContain('function tokenBarWidth(item: UserLeaderboardItem): string')
    expect(leaderboardView).toContain('function tokenBarStyle(item: UserLeaderboardItem): Record<string, string>')
    expect(leaderboardView).toContain('--token-bar-width')
    expect(leaderboardView).toContain('data-testid="leaderboard-rank-title"')
    expect(leaderboardView).toContain('function visibleLeaderboardTitleBadges')
    expect(leaderboardView).toContain('leaderboard.tokenRankingTitle')
    expect(leaderboardView).toContain(':title="leaderboardTokenMetricsLabel(item)"')
    expect(leaderboardView).toContain(':aria-label="leaderboardTokenMetricsLabel(item)"')
    expect(leaderboardView).toContain('leaderboard.inputTokensShort')
    expect(leaderboardView).toContain('leaderboard.outputTokensShort')
    expect(leaderboardView).toContain('leaderboard.costPerMillionShort')
    expect(leaderboardView).toContain('function formatLeaderboardCostPerMillion(item: UserLeaderboardItem): string')
    expect(leaderboardView).not.toContain('data-testid="leaderboard-cost-efficiency-summary"')
  })

  it('uses provider icons for model ranking avatars', () => {
    expect(leaderboardView).toContain("import ModelIcon from '@/components/common/ModelIcon.vue'")
    expect(leaderboardView).toContain('data-testid="leaderboard-model-rank-icon"')
    expect(leaderboardView).toContain('<ModelIcon :model="item.model" size="16px" />')
    expect(leaderboardView).not.toContain('function modelAvatarInitial')
  })

  it('shows growth and rank movement on model ranking rows', () => {
    expect(leaderboardView).toContain('data-testid="leaderboard-model-growth"')
    expect(leaderboardView).toContain('data-testid="leaderboard-model-rank-change"')
    expect(leaderboardView).toContain('leaderboard-model-rank-insights')
    expect(leaderboardView).toContain('function modelGrowthLabel(item: UserLeaderboardModelItem): string')
    expect(leaderboardView).toContain('function modelRankChangeLabel(item: UserLeaderboardModelItem): string')
    expect(leaderboardView).toContain("t('leaderboard.growth')")
    expect(leaderboardView).toContain("t('leaderboard.rankChange')")
  })

  it('keeps the ranking panel bound to a scoped-css-safe dark descendant selector', () => {
    expect(leaderboardView).toContain(':global(.dark .leaderboard-token-ranking-card)')
    expect(leaderboardView).toContain(':global(.dark .leaderboard-token-bar-track)')
    expect(leaderboardView).not.toContain(':global(html.dark) .leaderboard-token-ranking-card')
    expect(leaderboardView).not.toContain(':global(.dark) .leaderboard-token-ranking-card')
  })
})
