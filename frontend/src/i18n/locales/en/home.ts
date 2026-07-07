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
    goToDashboard: 'Go to Dashboard',
    heroEyebrow: 'Developer API workspace',
    heroTitleTop: 'AI Coding',
    heroTitleBottom: 'One-stop model service',
    heroSubtitle: 'AI Coding one-stop model service',
    heroDescription: 'Use one OpenAI-compatible entry URL to keep model calls, API keys, and usage records in a lightweight workflow.',
    heroDescriptionAltModels: 'Follow the quick start and claim trial credits.',
    heroDescriptionAltSupport: 'One key for your AI workflow.',
    heroProductTitle: 'Make AI access a clear path',
    heroProductMetricsLabel: 'Homepage product capability metrics',
    heroGatewayConsole: 'API workspace',
    apiEntryLabel: 'API entry',
    apiEntryAriaLabel: 'OpenAI-compatible API entry URL',
    heroStats: {
      endpoint: 'Unified endpoint',
      billing: 'Traceable billing',
      routing: 'Multi-model routing'
    },
    heroFlow: {
      request: {
        label: 'Requests enter one gateway',
        text: 'OpenAI compatible'
      },
      route: {
        label: 'Route by group policy',
        text: 'Chat / Image / Video'
      },
      usage: {
        label: 'Record usage and rates',
        text: 'Credit ledger'
      }
    },
    heroProofGateway: 'Unified API gateway',
    heroProofBilling: 'Rate billing',
    heroProofVisibility: 'Usage visibility',
    claimButton: 'Claim Trial',
    quickStartButton: 'Quick Start',
    contactSupport: 'Contact Support',
    accountWorkbench: {
      ariaLabel: 'Account workbench overview',
      kicker: 'Account',
      title: 'Account workbench',
      description: 'Review balance, usage volume, and lifetime spend while signed in.',
      currentBalance: 'Current balance',
      totalTokens: 'Total tokens',
      totalRequests: 'Total requests',
      totalCost: 'Total spend',
      loading: 'Syncing account data...',
      loadFailed: 'Account data is temporarily unavailable'
    },
    gatewayExplain: {
      ariaLabel: 'Unified API gateway explanation',
      kicker: 'What a unified model gateway does',
      title: 'One entry point for models, accounts, and usage.',
      description: 'Connect to one OpenAI-compatible API URL. The gateway handles model selection, group rates, account routing, and usage records, so your product code does not need a new integration for every upstream provider.',
      mapLabel: 'Diagram showing one API gateway connected to a model pool, account pool, and usage records',
      gatewayLabel: 'Unified entry',
      gatewayName: 'API Gateway',
      accountPoolLabel: 'Accounts and sessions',
      accountPoolName: 'Account Pool',
      poolLabel: 'Multi-model routing',
      poolName: 'Model Pool'
    },
    modelCarousel: {
      ariaLabel: 'Supported model carousel',
      kicker: 'Model coverage',
      title: 'Access a broad set of top-tier models',
      highlight: '',
      description: 'Coding, reasoning, image, and multimodal capabilities can live behind the same entry point and account-pool policy.'
    },
    faq: {
      ariaLabel: 'Frequently asked questions',
      kicker: 'FAQ',
      title: 'Frequently asked questions',
      description: 'A short guide to the questions teams usually ask about setup, billing, model access, and account-pool routing.',
      tabs: {
        service: 'About the service',
        billing: 'Pricing and billing',
        usage: 'Setup and usage'
      },
      items: [
        {
          question: 'Is this an official provider platform?',
          answer: 'No. It is an OpenAI-compatible API gateway that manages access, routing, account-pool policy, and usage records.'
        },
        {
          question: 'What do I need to change in my tools?',
          answer: 'Most OpenAI-compatible tools only need a new Base URL and API key. Your application can keep using the same compatible API shape.'
        },
        {
          question: 'Where can I view models and group rates?',
          answer: 'The model plaza shows public model and rate information. Final availability and pricing follow the dashboard, order page, and account configuration.'
        },
        {
          question: 'Will the account pool affect my application code?',
          answer: 'No. Your application calls the unified entry point while routing, sessions, and failover are handled by the gateway.'
        }
      ]
    },
    finalCta: {
      ariaLabel: 'Start using the gateway',
      kicker: 'Ready',
      title: 'Start using',
      description: 'Replace the API URL and key to route model calls through one entry point, then use the tutorial and trial credits to send the first request.',
      button: 'Claim Trial'
    },
    pricing: {
      ctaRecharge: 'Buy Credits',
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
        lead: 'Here, 1 CNY can be understood as 1 USD of official API quota; the actual credit cost depends on the group rate hit by the request.',
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
            desc: 'Map official dollar usage into the site credit unit first.'
          },
          rate: {
            badge: 'Apply group rate',
            value: '0.15x',
            desc: 'Use the rate shown for the active model group and order page.'
          },
          cost: {
            badge: 'Actual cost',
            value: '0.150 CNY',
            desc: 'This is the credit cost for equivalent official $1 usage.'
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
          desc: 'Use a small credit balance to run tutorials, scripts, and daily development work.'
        },
        groups: {
          title: 'Environment isolation',
          desc: 'Split personal, team, test, and production usage into different policies.'
        },
        usage: {
          title: 'Request traceability',
          desc: 'Debug cost issues directly from model, group, and credit records.'
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
