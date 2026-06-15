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
    navContact: '联系管理',
    getStarted: '立即开始',
    goToDashboard: '登录控制台',
    heroEyebrow: 'AI 接入中枢',
    heroTitleTop: 'mapleAI',
    heroTitleBottom: '智能解决方案',
    heroSubtitle: 'mapleAI 智能解决方案',
    heroDescription: '统一接入模型能力，按需付费',
    heroDescriptionAltModels: '完成教程接入 API，领取试用积分',
    heroDescriptionAltSupport: '一把密钥连接你的 AI 工作流',
    heroProofGateway: '统一 API 网关',
    heroProofBilling: '按需付费',
    heroProofVisibility: '用量透明',
    claimButton: '注册领取试用',
    contactSupport: '添加客服',
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
