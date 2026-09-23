export interface DashboardAuditLog {
  id: string;
  action: string;
  resource_type?: string;
  resource_id?: string;
  created_at?: string;
  user?: {
    email?: string;
  };
}

export interface DashboardStats {
  totalUsers: number;
  activeProviders: number;
  activeSeekers: number;
  openDisputes: number;
  pendingVerifications: number;
  pendingBackgroundChecks: number;
  totalGMV: number;
  platformRevenue: number;
  activeJobsCount: number;
  completedJobsCount: number;
  activeSubscriptions: number;
  pendingPayoutsCount: number;
  pendingPayoutsAmount: number;
  totalWalletBalance: number;
  recentAuditLogs: DashboardAuditLog[];
  signupsLast30Days?: Record<string, number>;
}
