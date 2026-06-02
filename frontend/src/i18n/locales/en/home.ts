export default {
    viewOnGithub: 'View on GitHub',
    viewDocs: 'View Documentation',
    docs: 'Docs',
    switchToLight: 'Switch to Light Mode',
    switchToDark: 'Switch to Dark Mode',
    dashboard: 'Dashboard',
    login: 'Login',
    navHome: 'Home',
    navTutorial: 'Tutorial',
    navModels: 'Model Plaza',
    navContact: 'Contact',
    getStarted: 'Get Started',
    goToDashboard: 'Open Dashboard',
    heroEyebrow: 'AI access hub',
    heroTitleTop: 'mapleAI',
    heroTitleBottom: 'Intelligent Solutions',
    heroSubtitle: 'mapleAI intelligent solutions',
    heroDescription: 'Connect AI model capabilities through one gateway.',
    heroDescriptionAltModels: 'Follow the quick start and claim trial credits.',
    heroDescriptionAltSupport: 'One key for your AI workflow.',
    heroProofGateway: 'Unified API gateway',
    heroProofBilling: 'Pay as you go',
    heroProofVisibility: 'Usage visibility',
    claimButton: 'Claim Trial',
    contactSupport: 'Contact Support',
    pricing: {
      ctaRecharge: 'Recharge Now',
      ctaKey: 'Create API Key',
      formula: {
        label: 'Price calculation formula',
        stepsLabel: 'Billing conversion steps',
        kicker: 'Cost equation',
        title: 'Start from official usage, then apply the selected group rate',
        sampleGroupLabel: 'Sample group',
        sampleGroupValue: 'claude-stable',
        sampleRateLabel: 'Rate',
        sampleRateValue: '0.15x',
        lead: 'Here, 1 CNY can be understood as 1 USD of official API quota; the actual balance cost depends on the group rate hit by the request.',
        note: 'For example, when a group rate is 0.15x, spending 0.150 CNY gives roughly the same callable usage as official $1 API usage. When switching groups, replace only the middle rate number.',
        references: 'Official pricing references',
        links: {
          claude: 'Claude official pricing',
          openai: 'OpenAI / Codex official pricing',
          gemini: 'Gemini official pricing'
        },
        steps: {
          base: {
            badge: 'Base mapping',
            value: '1 CNY = 1 USD',
            desc: 'Map official dollar usage into the site balance unit first.'
          },
          rate: {
            badge: 'Apply group rate',
            value: '0.15x',
            desc: 'Use the rate shown for the active model group and order page.'
          },
          cost: {
            badge: 'Actual cost',
            value: '0.150 CNY',
            desc: 'This is the balance cost for equivalent official $1 usage.'
          },
          usage: {
            badge: 'Official usage',
            value: 'Official $1 API usage',
            desc: 'You receive callable usage converted from the official price table.'
          }
        }
      },
      signals: {
        label: 'Pricing calculation support notes',
        recharge: {
          title: 'Lightweight trials',
          desc: 'Use a small balance to run tutorials, scripts, and daily development work.'
        },
        groups: {
          title: 'Environment isolation',
          desc: 'Split personal, team, test, and production usage into different policies.'
        },
        usage: {
          title: 'Request traceability',
          desc: 'Debug cost issues directly from model, group, and balance records.'
        }
      }
    },
    tags: {
      subscriptionToApi: 'Subscription to API',
      stickySession: 'Session Persistence',
      realtimeBilling: 'Pay As You Go',
      directConnect: 'No extra setup',
      teamReady: 'Unified team access',
      clearUsage: 'Stable low latency',
      routing: 'Ready in minutes'
    },
    // Pain points section
    painPoints: {
      title: 'Sound Familiar?',
      items: {
        expensive: {
          title: 'High Subscription Costs',
          desc: 'Paying for multiple AI subscriptions that add up every month'
        },
        complex: {
          title: 'Account Chaos',
          desc: 'Managing scattered accounts and API keys across different platforms'
        },
        unstable: {
          title: 'Service Interruptions',
          desc: 'Single accounts hitting rate limits and disrupting your workflow'
        },
        noControl: {
          title: 'No Usage Control',
          desc: "Can't track where your money goes or limit team member usage"
        }
      }
    },
    // Solutions section
    solutions: {
      title: 'We Solve These Problems',
      subtitle: 'Three simple steps to stress-free AI access'
    },
    features: {
      unifiedGateway: 'Unified Access',
      unifiedGatewayDesc: 'Use one key for coding models and upstream capabilities.',
      multiAccount: 'Stable Access',
      multiAccountDesc: 'Coordinate account pools, proxy routing, and session links to reduce single-account disruption.',
      balanceQuota: 'Clear Usage State',
      balanceQuotaDesc: 'Track requests, quota, billing, and route status in one place for team cost control.'
    },
    // Comparison section
    comparison: {
      title: 'Why Choose Us?',
      headers: {
        feature: 'Comparison',
        official: 'Official Subscriptions',
        us: 'Our Platform'
      },
      items: {
        pricing: {
          feature: 'Pricing',
          official: 'Fixed monthly fee, pay even if unused',
          us: 'Pay only for what you use'
        },
        models: {
          feature: 'Model Selection',
          official: 'Single provider only',
          us: 'Switch between models freely'
        },
        management: {
          feature: 'Account Management',
          official: 'Manage each service separately',
          us: 'Unified key, one dashboard'
        },
        stability: {
          feature: 'Stability',
          official: 'Single account rate limits',
          us: 'Multi-account pool, auto-failover'
        },
        control: {
          feature: 'Usage Control',
          official: 'Not available',
          us: 'Quotas & detailed analytics'
        }
      }
    },
    providers: {
      title: 'Supported AI Models',
      description: 'One API, Multiple Choices',
      supported: 'Supported',
      soon: 'Soon',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: 'More'
    },
    // CTA section
    cta: {
      title: 'Ready to Get Started?',
      description: 'Sign up now and get free trial credits to experience seamless AI access',
      button: 'Sign Up Free'
    },
    footer: {
      allRightsReserved: 'All rights reserved.',
      linksLabel: 'Footer links',
      terms: 'Terms',
      privacy: 'Privacy Policy',
      usagePolicy: 'Usage Policy'
    }
  }
