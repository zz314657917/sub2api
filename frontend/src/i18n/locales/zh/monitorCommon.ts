export default {
    status: {
      operational: '正常',
      degraded: '部分可用',
      failed: '失败',
      error: '错误',
      unknown: '-'
    },
    providers: {
      openai: 'OpenAI',
      anthropic: 'Anthropic',
      gemini: 'Gemini',
      grok: 'Grok',
      antigravity: 'Antigravity',
      kimi: 'Kimi',
      zhipu: '智谱',
      deepseek: 'DeepSeek',
      minimax: 'MiniMax',
    },
    // 检查模式（监控条目的工作方式）
    checkMode: {
      probe: '探活',
      quota: '配额',
      quota_probe: '探活 + 配额'
    },
    // 配额快照展示（MonitorQuotaView，管理端与用户端共用）
    quota: {
      unavailable: '配额信息不可用',
      windows: {
        '5h': '5 小时',
        '7d': '7 天',
        '7dSonnet': '7 天 Sonnet',
        '7dFable': '7 天 Fable',
        weekly: '周',
        monthly: '月',
        daily: '日',
        '30d': '30 天',
        total: '总量'
      },
      labels: {
        requests: '请求',
        tokens: 'Token',
        shared: '共享',
        pro: 'Pro',
        flash: 'Flash'
      }
    },
    extraModelsHeader: '附加模型',
    extraModelsEmpty: '无附加模型',
    latencyEmpty: '-',
    availabilityPrefix: '可用性',
    dialogLatency: '对话延迟',
    endpointPing: '端点 PING',
    history60pts: '近 {n} 次记录',
    nextUpdateIn: '{n}s 后刷新',
    past: 'PAST',
    now: 'NOW',
    maintenancePaused: '维护中 · 已暂停时间线采集',
    extraModelsCount: '+ {n} 模型',
    pollEvery: '{n}s 轮询',
    updatedAt: '更新于 {time}',
    relativeSecondsAgo: '{n} 秒前',
    relativeMinutesAgo: '{n} 分钟前',
    relativeHoursAgo: '{n} 小时前',
    relativeDaysAgo: '{n} 天前'
  }
