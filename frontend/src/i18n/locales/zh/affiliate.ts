export default {
    title: '邀请赚钱',
    description: '邀请好友注册并调用任意 API 后，即可领取固定奖励，后续充值还可按比例继续累积。',
    descriptionWithReward: '邀请好友注册并调用任意 API 后，即可领取 {amount}，后续充值还可按比例继续累积。',
    yourCode: '我的邀请码',
    inviteLink: '邀请链接',
    copyCode: '复制邀请码',
    copyLink: '复制链接',
    codeCopied: '邀请码已复制',
    linkCopied: '邀请链接已复制',
    loadFailed: '加载邀请赚钱数据失败',
    transferFailed: '转入余额失败',
    hero: {
      eyebrow: '推荐赚钱',
      title: '邀请好友，赚取额度'
    },
    summary: {
      earned: '已赚额度',
      earnedHint: '累计',
      availableHint: '可随时转入余额',
      invitedHint: '好友注册后统计',
      status: '状态',
      ready: '已开启',
      statusHint: '分享链接开始赚取额度'
    },
    share: {
      title: '你的邀请链接',
      description: '优先分享这个链接，我们会自动识别好友注册来源。',
      more: '也可以通过以下方式分享',
      email: '邮件',
      emailSubject: '邀请你一起使用 API 服务',
      emailBody: '通过我的邀请链接注册，调用任意 API 后即可开始赚取奖励：{link}'
    },
    stats: {
      rebateRate: '我的返利比例',
      rebateRateHint: '被邀请用户每次充值后你可获得的返利比例',
      invitedUsers: '邀请人数',
      availableQuota: '可转返利额度',
      frozenQuota: '冻结中',
      frozenQuotaHint: '新产生的返利正在冻结期中',
      totalQuota: '历史返利额度'
    },
    transfer: {
      title: '返利额度转余额',
      description: '将当前可用返利额度一键转入账户余额',
      button: '转入余额',
      transferring: '转入中...',
      empty: '当前没有可转入额度',
      success: '已转入余额：{amount}'
    },
    steps: {
      title: '如何运作',
      share: {
        title: '分享你的链接',
        description: '把邀请链接发给好友，或发布到社交媒体、社群和团队频道。'
      },
      verify: {
        title: '用户注册并调用',
        description: '好友通过你的链接创建账号后，调用任意 API 即可进入奖励流程。'
      },
      earn: {
        title: '领取 {amount}',
        description: '好友首次调用 API 后，你可以领取固定奖励；好友后续充值还会按 {rate} 计算返利。'
      }
    },
    invitees: {
      title: '已邀请用户',
      empty: '暂无邀请记录',
      columns: {
        email: '邮箱',
        username: '用户名',
        apiStatus: 'API 调用',
        rebate: '返利明细',
        joinedAt: '注册时间',
        action: '操作'
      },
      apiStatus: {
        used: '已调用',
        pending: '未调用'
      },
      actions: {
        claim: '领取返利',
        claiming: '领取中...',
        claimed: '已领取',
        waiting: '待调用',
        notConfigured: '未配置',
        claimSuccess: '已领取首次 API 调用返利：{amount}',
        claimFailed: '领取返利失败'
      }
    },
    recent: {
      title: '最近推荐',
      description: '查看通过你的链接注册的好友，以及每位好友的奖励状态。',
      empty: '分享上方邀请链接，开始赚取额度。'
    },
    notice: {
      title: '注意事项',
      line1: '只有好友通过你的邀请链接或邀请码注册，才会被记录为有效邀请。',
      line2: '好友通过你的链接注册并调用任意 API 后，固定奖励才可领取。',
      line3: '充值返利按当前生效比例计算，新产生的返利可能需要经过冻结期。',
      line4: '不要使用自己的邮箱、设备或支付方式互相邀请，异常邀请可能会被系统自动撤销。'
    },
    faq: {
      title: '常见问题',
      limit: {
        question: '邀请好友数量有上限吗？',
        answer: '没有固定上限；每一位通过你的链接注册并调用任意 API 的有效好友，都有机会带来 {amount} 固定奖励。'
      },
      when: {
        question: '额度何时到账？',
        answer: '好友通过你的链接注册并调用任意 API 后，可在最近推荐列表中领取固定奖励；充值返利会按站点配置进入可用或冻结额度。'
      },
      expire: {
        question: '额度会过期吗？',
        answer: '不会，转入余额后会一直保留在你的账户余额中，直到 API 用量消耗。'
      },
      selfInvite: {
        question: '我能用不同邮箱自我推荐吗？',
        answer: '不能。系统会结合 IP、设备和支付方式识别重复账号，自我推荐或异常邀请可能会被撤销。'
      }
    },
    tips: {
      title: '使用说明',
      line1: '将邀请码或邀请链接分享给新用户。',
      line2: '被邀请用户首次调用 API 后，可在列表中领取固定返利；充值返利比例为 {rate}。',
      line3: '返利额度可随时转入账户余额。',
      line4: '新产生的返利需要经过冻结期后才能提现。'
    }
  }
