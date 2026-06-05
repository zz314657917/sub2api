export default {
    title: 'Channel Status',
    description: 'Inspect channel availability, latency and recent status',
    monitorTitle: 'Channel Monitor',
    monitorDescription: 'Review availability, latency, and recent probes by time window.',
    searchPlaceholder: 'Search channels...',
    allProviders: 'All Providers',
    loadError: 'Failed to load channel status',
    detailLoadError: 'Failed to load channel detail',
    detailTitle: 'Channel Detail',
    closeDetail: 'Close',
    availabilityPanel: {
      title: 'Service Availability',
      searchPlaceholder: 'Search service name',
      groupSuffix: 'Stability Monitor',
      availabilityLabel: 'Availability',
      noResultsTitle: 'No matching services',
      noResultsDescription: 'Try a different keyword.',
      legend: {
        abnormal: 'Abnormal',
        normal: 'Normal',
        highLatency: 'High latency',
        maintenance: 'Maintenance'
      }
    },
    windowTab: {
      '7d': '7 days',
      '15d': '15 days',
      '30d': '30 days'
    },
    overall: {
      operational: 'OPERATIONAL',
      degraded: 'DEGRADED',
      unavailable: 'UNAVAILABLE'
    },
    columns: {
      name: 'Name',
      provider: 'Provider',
      groupName: 'Group',
      primaryModel: 'Primary Model',
      availability7d: '7d Availability',
      latency: 'Latency (ms)'
    },
    detailColumns: {
      model: 'Model',
      latestStatus: 'Latest Status',
      latestLatency: 'Latest Latency (ms)',
      availability7d: '7d Availability',
      availability15d: '15d Availability',
      availability30d: '30d Availability',
      avgLatency7d: '7d Avg Latency (ms)'
    },
    empty: {
      title: 'No channels available',
      description: 'No monitored channels have been configured yet.'
    },
    capacityPools: {
      mine: 'MY POOL',
      shared: 'SHARED POOL',
      mineTitle: 'My Account Capacity Pool',
      sharedTitle: 'Platform Shared Capacity Pool',
      externalTitle: 'Public Shared Capacity Reference',
      mineDescription: 'Summarizes quotas and OpenAI OAuth window snapshots for accounts you own.',
      sharedDescription: 'Shows schedulable shared accounts on this platform, plus a public shared capacity reference.',
      externalDescription: 'Public shared account pool status reference.',
      total: 'Accounts',
      active: 'Active',
      configuredQuota: 'Quota',
      remainingQuota: 'Remaining',
      schedulable: 'schedulable',
      rateLimited: 'Rate limited',
      error: 'Errors',
      disabled: 'Disabled',
      other: 'Other',
      abnormal: 'Abnormal',
      accountStatus: 'Account status',
      groupCapacity: 'Group capacity',
      groupCapacityHint: 'Shared capacity is judged by group so a global total does not hide an unavailable group.',
      healthy: 'Healthy',
      degraded: 'Partially available',
      unavailable: 'Unavailable',
      window: '{window} window',
      quotaWindow: '{window} quota',
      schedulableSnapshot: 'Available accounts',
      schedulableRemaining: 'Remaining',
      percentOnly: 'Remaining',
      ownContributed: 'Mine',
      unavailableReason: 'Not schedulable',
      unavailableReasons: {
        daily_quota_exceeded: 'Daily quota exhausted',
        weekly_quota_exceeded: 'Weekly quota exhausted',
        monthly_quota_exceeded: 'Monthly quota exhausted',
        quota_exceeded: 'Quota exhausted',
        rate_limited: 'Rate limited',
        overloaded: 'Overload cooldown',
        temp_unschedulable: 'Temporarily unavailable',
        manual_unschedulable: 'Manually disabled',
        error: 'Error',
        disabled: 'Disabled',
        expired: 'Expired',
        unused: 'Unused',
        inactive: 'Inactive',
        unschedulable: 'Unschedulable'
      },
      empty: 'No accounts in this pool yet.',
      loadError: 'Failed to load account capacity pools'
    }
  }
