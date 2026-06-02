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
      gemini: 'Gemini'
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
