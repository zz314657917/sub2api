export default {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    navHome: '首页',
    navTutorial: '教程',
    navModels: '模型广场',
    navContact: '联系客服',
    getStarted: '立即开始',
    goToDashboard: '前往仪表盘',
    heroEyebrow: '开发者 API 工作台',
    heroTitleTop: 'AI Coding',
    heroTitleBottom: '一站式大模型服务',
    heroSubtitle: 'AI Coding 一站式大模型服务',
    heroDescription: '一个兼容 OpenAI 的入口网址，把模型调用、密钥接入和用量记录收进同一个轻量工作流。',
    heroDescriptionAltModels: '完成教程接入 API，领取试用积分',
    heroDescriptionAltSupport: '一把密钥连接你的 AI 工作流',
    heroProductTitle: '把 AI 接入变成一条清晰路径',
    heroProductMetricsLabel: '首页产品能力指标',
    heroGatewayConsole: 'API 工作台',
    apiEntryLabel: 'API 入口',
    apiEntryAriaLabel: '兼容 OpenAI 的 API 入口网址',
    heroStats: {
      endpoint: '统一端点',
      billing: '账单可追踪',
      routing: '多模型路由'
    },
    heroFlow: {
      request: {
        label: '请求进入统一网关',
        text: 'OpenAI compatible'
      },
      route: {
        label: '按分组路由到模型池',
        text: 'Chat / Image / Video'
      },
      usage: {
        label: '记录用量与倍率',
        text: 'Credit ledger'
      }
    },
    heroProofGateway: '统一 API 网关',
    heroProofBilling: '倍率账单',
    heroProofVisibility: '用量透明',
    claimButton: '注册领取试用',
    quickStartButton: '快速开始',
    contactSupport: '添加客服',
    accountWorkbench: {
      ariaLabel: '账号工作台概览',
      kicker: 'Account',
      title: '账号工作台',
      description: '登录状态下快速查看余额、调用量和累计消耗。',
      currentBalance: '当前余额',
      totalTokens: '总 Token',
      totalRequests: '总请求数',
      totalCost: '总花费',
      rechargeBalance: '充值余额',
      loading: '正在同步账号数据...',
      loadFailed: '账号数据暂时无法同步'
    },
    gatewayExplain: {
      ariaLabel: '统一 API 网关说明',
      kicker: '什么是统一模型网关',
      title: '一个入口，连接多种模型与账号池。',
      description: '你只需要接入一个兼容 OpenAI 的 API 地址；模型选择、分组倍率、账号路由和用量记录都交给网关处理，业务代码不用为每个上游重新改一遍。',
      mapLabel: '统一 API 网关连接模型池、账号池和用量记录的关系图',
      gatewayLabel: '统一入口',
      gatewayName: 'API Gateway',
      accountPoolLabel: '账号与会话',
      accountPoolName: 'Account Pool',
      poolLabel: '多模型路由',
      poolName: 'Model Pool'
    },
    modelCarousel: {
      ariaLabel: '可接入模型轮播展示',
      kicker: '模型接入范围',
      title: '提供众多顶级优质大模型',
      highlight: '',
      description: '从编码、推理到图像和多模态能力，统一放进同一个入口和账号池策略里。'
    },
    faq: {
      ariaLabel: '常见问题解答',
      kicker: 'FAQ 常见问题',
      title: '常见问题解答',
      description: '整理了关于接入、计费、模型和账号池的常见疑问，先把关键问题说清楚。',
      tabs: {
        service: '关于服务',
        billing: '价格与计费',
        usage: '接入与使用'
      },
      items: [
        {
          question: '这是官方平台吗？',
          answer: '不是官方平台，而是兼容 OpenAI 的统一 API 网关，负责接入、路由、账号池策略和用量记录。'
        },
        {
          question: 'API 地址需要怎么改？',
          answer: '多数兼容 OpenAI 的工具只需要替换 Base URL 和 API Key；业务侧继续按 OpenAI-compatible 接口调用。'
        },
        {
          question: '模型和分组倍率在哪里看？',
          answer: '模型广场会展示公开模型与倍率信息，最终价格和可用分组以控制台、下单页和实际账户配置为准。'
        },
        {
          question: '账号池会影响我的业务代码吗？',
          answer: '不会。你的应用继续调用统一入口，模型路由、账号会话和失败切换由网关侧处理。'
        }
      ]
    },
    finalCta: {
      ariaLabel: '开始使用',
      kicker: 'Ready',
      title: '开始使用',
      description: '替换 API 地址和密钥，就可以把模型调用接入统一入口，先从教程和试用额度跑通第一条请求。',
      button: '注册领取试用'
    },
    pricing: {
      ctaRecharge: '购买积分',
      ctaKey: '创建 API Key',
      formula: {
        label: '价格计算公式',
        stepsLabel: '账单换算步骤',
        kicker: '费用公式',
        title: '先看官方用量，再套本站分组倍率',
        sampleGroupLabel: '示例分组',
        sampleGroupValue: 'claude-稳定',
        sampleRateLabel: '倍率',
        sampleRateValue: '0.15x',
        lead: '这里的 1 人民币可以按官方 1 美元 API 额度理解；真正扣多少积分，取决于你调用时命中的分组倍率。',
        note: '也就是说，当某个分组的倍率是 0.15x 时，你花 0.150 元人民币，就可以获得大致等同于官方 1 美元 API 的调用用量。更换分组时，只需要替换中间的倍率数字。',
        references: '官方定价参考',
        links: {
          claude: 'Claude 官方价目表',
          openai: 'OpenAI / Codex 官方价目表',
          gemini: 'Gemini 官方价目表'
        },
        steps: {
          base: {
            badge: '基础换算',
            value: '1 人民币 = 1 美元',
            desc: '先把官方美元消耗映射到本站积分单位。'
          },
          rate: {
            badge: '乘以分组倍率',
            value: '0.15x',
            desc: '倍率以当前模型分组和下单页显示为准。'
          },
          cost: {
            badge: '实际扣费',
            value: '0.150 元人民币',
            desc: '这是本次等价官方 1 美元用量的积分成本。'
          },
          usage: {
            badge: '对应官方用量',
            value: '官方 $1 API 用量',
            desc: '你获得的是按官方价目表折算后的调用量。'
          }
        }
      },
      signals: {
        label: '价格计算辅助说明',
        recharge: {
          title: '轻量试跑',
          desc: '少量积分也能先跑通教程、脚本和日常开发任务。'
        },
        groups: {
          title: '环境隔离',
          desc: '把个人、团队、测试和生产用途拆成不同策略。'
        },
        usage: {
          title: '请求可回溯',
          desc: '从模型、分组到扣费记录，排查成本问题更直接。'
        }
      }
    },
    tags: {
      subscriptionToApi: '订阅转 API',
      stickySession: '会话保持',
      realtimeBilling: '按量计费',
      directConnect: '无需额外配置',
      teamReady: '团队统一接入',
      clearUsage: '稳定低延迟',
      routing: '即开即用'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '统一接入',
      unifiedGatewayDesc: '一把密钥接入编码模型与上游能力。',
      multiAccount: '稳定接入体验',
      multiAccountDesc: '号池、代理与会话链路协同，减少单点异常对编码流程的影响。',
      balanceQuota: '用量状态清晰',
      balanceQuotaDesc: '请求、积分、账单和链路状态集中查看，方便团队控制成本。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '已支持的 AI 模型',
      description: '一个 API，多种选择',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: '更多'
    },
    // CTA 区块
    cta: {
      title: '准备好开始了吗？',
      description: '注册即可获得免费试用积分，体验一站式 AI 服务',
      button: '免费注册'
    },
    footer: {
      allRightsReserved: '保留所有权利。',
      linksLabel: '页脚链接',
      terms: '服务条款',
      privacy: '隐私政策',
      usagePolicy: '使用政策'
    }
  }
