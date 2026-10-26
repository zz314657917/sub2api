export default {
    status: {
      operational: 'Operational',
      degraded: 'Degraded',
      failed: 'Failed',
      error: 'Error',
      unknown: '-'
    },
    providers: {
      openai: 'OpenAI',
      anthropic: 'Anthropic',
      gemini: 'Gemini',
      grok: 'Grok',
      antigravity: 'Antigravity',
      kimi: 'Kimi',
      zhipu: 'Zhipu GLM',
      deepseek: 'DeepSeek',
      minimax: 'MiniMax',
    },
    // Check modes (how a monitor performs its checks)
    checkMode: {
      probe: 'Probe',
      quota: 'Quota',
      quota_probe: 'Probe + Quota'
    },
    // Quota snapshot rendering (MonitorQuotaView, shared by admin + user views)
    quota: {
      unavailable: 'Quota unavailable',
      windows: {
        '5h': '5h',
        '7d': '7d',
        '7dSonnet': '7d Sonnet',
        '7dFable': '7d Fable',
        weekly: 'Weekly',
        monthly: 'Monthly',
        daily: 'Daily',
        '30d': '30d',
        total: 'Total'
      },
      labels: {
        requests: 'Requests',
        tokens: 'Tokens',
        shared: 'Shared',
        pro: 'Pro',
        flash: 'Flash'
      }
    },
    extraModelsHeader: 'Extra Models',
    extraModelsEmpty: 'No extra models',
    latencyEmpty: '-',
    availabilityPrefix: 'Availability',
    dialogLatency: 'Dialog Latency',
    endpointPing: 'Endpoint PING',
    history60pts: 'HISTORY ({n} PTS)',
    nextUpdateIn: 'NEXT UPDATE IN {n}s',
    past: 'PAST',
    now: 'NOW',
    maintenancePaused: 'Maintenance · timeline paused',
    extraModelsCount: '+ {n} models',
    pollEvery: '{n}s polling',
    updatedAt: 'Updated {time}',
    relativeSecondsAgo: '{n}s ago',
    relativeMinutesAgo: '{n}m ago',
    relativeHoursAgo: '{n}h ago',
    relativeDaysAgo: '{n}d ago'
  }
