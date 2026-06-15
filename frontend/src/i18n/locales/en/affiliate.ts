export default {
    title: 'Invite Earnings',
    description: 'Invite friends to register and call any API. You can claim a fixed reward afterward, then keep earning from credit purchase rebates.',
    descriptionWithReward: 'Invite friends to register and call any API. You can claim {amount} afterward, then keep earning from credit purchase rebates.',
    yourCode: 'Your Affiliate Code',
    inviteLink: 'Invite Link',
    copyCode: 'Copy Code',
    copyLink: 'Copy Link',
    codeCopied: 'Affiliate code copied',
    linkCopied: 'Invite link copied',
    loadFailed: 'Failed to load affiliate data',
    transferFailed: 'Failed to transfer affiliate credits',
    hero: {
      eyebrow: 'Referral earnings',
      title: 'Invite friends and earn credit'
    },
    summary: {
      earned: 'Earned credit',
      earnedHint: 'All time',
      availableHint: 'Ready to transfer',
      invitedHint: 'Registered friends',
      status: 'Status',
      ready: 'Ready',
      statusHint: 'Share your link to start earning'
    },
    share: {
      title: 'Your invite link',
      description: 'Share this link first. We will automatically attribute registrations to you.',
      more: 'Or share another way',
      email: 'Email',
      emailSubject: 'Join this API service with my invite link',
      emailBody: 'Register through my invite link, then call any API to start the reward flow: {link}'
    },
    stats: {
      rebateRate: 'My Rebate Rate',
      rebateRateHint: 'What you earn each time an invitee buys credits',
      invitedUsers: 'Invited Users',
      availableQuota: 'Available Rebate Credits',
      frozenQuota: 'Frozen',
      frozenQuotaHint: 'Recently earned rebates pending release',
      totalQuota: 'Historical Rebate Credits'
    },
    transfer: {
      title: 'Transfer Rebate Credits',
      description: 'Move available rebate credits into your account credits',
      button: 'Transfer to Credits',
      transferring: 'Transferring...',
      empty: 'No available rebate credits',
      success: '{amount} has been transferred to your credits'
    },
    steps: {
      title: 'How it works',
      share: {
        title: 'Share your link',
        description: 'Send the invite link to friends, social channels, communities, or team chats.'
      },
      verify: {
        title: 'User registers and calls',
        description: 'Your friend creates an account through your link, then calls any API to enter the reward flow.'
      },
      earn: {
        title: 'Claim {amount}',
        description: 'After the first API call, claim the fixed reward here. Future credit purchases can keep earning at {rate}.'
      }
    },
    invitees: {
      title: 'Invited Users',
      empty: 'No invited users yet',
      columns: {
        email: 'Email',
        username: 'Username',
        apiStatus: 'API Call',
        rebate: 'Rebate',
        joinedAt: 'Joined At',
        action: 'Action'
      },
      apiStatus: {
        used: 'Called',
        pending: 'Pending'
      },
      actions: {
        claim: 'Claim',
        claiming: 'Claiming...',
        claimed: 'Claimed',
        waiting: 'Waiting',
        notConfigured: 'Not configured',
        claimSuccess: 'First API call rebate claimed: {amount}',
        claimFailed: 'Failed to claim rebate'
      }
    },
    recent: {
      title: 'Recent referrals',
      description: 'Track friends who registered through your link and each reward status.',
      empty: 'Share the invite link above to start earning credit.'
    },
    notice: {
      title: 'Important notes',
      line1: 'Only registrations through your invite link or code count as valid referrals.',
      line2: 'The fixed reward becomes claimable after your friend registers through your link and calls any API.',
      line3: 'Credit purchase rebates use your current effective rate. Newly earned rebates may enter a frozen period.',
      line4: 'Do not self-invite with your own email, device, or payment method. Abnormal referrals may be revoked.'
    },
    faq: {
      title: 'FAQ',
      limit: {
        question: 'Is there a referral limit?',
        answer: 'There is no fixed cap. Every valid friend who registers through your link and calls any API may bring a {amount} fixed reward.'
      },
      when: {
        question: 'When does credit arrive?',
        answer: 'After your friend registers through your link and calls any API, you can claim the fixed reward from Recent referrals. Credit purchase rebates follow the site configuration.'
      },
      expire: {
        question: 'Will the credit expire?',
        answer: 'No. Once transferred to credits, it stays in your account until consumed by API usage.'
      },
      selfInvite: {
        question: 'Can I refer myself with another email?',
        answer: 'No. The system may detect repeated accounts by IP, device, and payment method. Self-referrals or abnormal referrals may be revoked.'
      }
    },
    tips: {
      title: 'How It Works',
      line1: 'Share your affiliate code or invite link with new users.',
      line2: 'After an invitee makes the first API call, you can claim the fixed reward here. Credit purchase rebate rate: {rate}.',
      line3: 'Transfer rebate credits to account credits at any time.',
      line4: 'Newly earned rebates may have a waiting period before they can be transferred.'
    }
  }
