function getApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    if ((window as any).__API_BASE_URL__) {
      return (window as any).__API_BASE_URL__;
    }
    const hostname = window.location.hostname;
    if (hostname.includes('staging')) {
      return 'https://staging-api.neighborservice.com/api/v1';
    }
    if (hostname.includes('neighborservice.com')) {
      return 'https://api.neighborservice.com/api/v1';
    }
  }
  return 'http://localhost:8000/api/v1';
}

export const ADMIN_API_CONFIG = {
  get baseUrl(): string {
    return getApiBaseUrl();
  },
  endpoints: {
    login: '/accounts/login/',
    dashboardStats: '/admin/dashboard/stats',
    users: '/admin/users',
    verifications: '/admin/verifications',
    verificationsBatch: '/admin/verifications/batch',
    backgroundChecks: '/admin/background-checks',
    reports: '/admin/reports',
    disputes: '/admin/disputes',
    payouts: '/admin/payouts',
    wallets: '/admin/wallets',
    subscriptions: '/admin/subscriptions',
    subscriptionPlans: '/admin/subscription-plans',
    categories: '/admin/categories',
    catalogServices: '/admin/catalog-services',
    settings: '/admin/settings',
    featureFlags: '/admin/feature-flags',
    auditLogs: '/admin/audit-logs',
    systemHealth: '/admin/system/health',
    systemBackup: '/admin/system/backup',
    systemBackups: '/admin/system/backups',
    cacheClear: '/admin/cache/clear',
    financialReport: '/admin/reports/financial',
    exportUsers: '/admin/export/users',
    exportPayouts: '/admin/export/payouts',
    exportDisputes: '/admin/export/disputes',
    broadcastNotifications: '/admin/notifications/broadcast',
    roles: '/admin/roles',
    fraudRiskAlerts: '/admin/fraud/risk-alerts',
    fraudEvaluate: '/admin/fraud/evaluate',
    templatesEmails: '/admin/templates/emails',
    templatesTestSend: '/admin/templates/emails/test-send',
    twoFASetup: '/admin/2fa/setup',
    twoFAVerify: '/admin/2fa/verify',
    twoFADisable: '/admin/2fa/disable',
    notificationFeed: '/admin/notifications/feed',
    payoutsBatchApprove: '/admin/payouts/batch-approve',
    payoutsBatchReject: '/admin/payouts/batch-reject',
    providerFunnel: '/admin/funnel/providers',
    appointments: '/admin/appointments',
    exportAppointments: '/admin/export/appointments',
    reviews: '/admin/reviews',
    promoCodes: '/admin/promo-codes',
    exportVerifications: '/admin/export/verifications',
    exportBackgroundChecks: '/admin/export/background-checks',
    legalDocuments: '/admin/legal',
    supportMessages: '/admin/support/messages',
    supportResolutions: '/admin/support/resolutions',
    cmsFaqs: '/admin/cms/faqs',
    cmsTestimonials: '/admin/cms/testimonials',
    cmsHero: '/admin/cms/hero',
    cmsStats: '/admin/cms/stats',
    cmsAbout: '/admin/cms/about',
    stripeBalance: '/admin/financial/stripe-balance',
    maintenance: '/admin/maintenance',
    webhookEvents: '/admin/webhooks/events'
  }
};

