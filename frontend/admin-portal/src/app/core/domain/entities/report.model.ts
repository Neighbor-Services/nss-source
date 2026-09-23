export interface FinancialBreakdownItem {
  date: string;
  gmv: number;
  revenue: number;
  payouts: number;
  transactionsCount: number;
}

export interface FinancialReport {
  totalRevenue?: number;
  grossMerchandiseValue?: number;
  totalPayouts?: number;
  heldInEscrow?: number;
  subscriptionRevenue?: number;
  backgroundCheckFees?: number;
  refundedAmount?: number;
  netMargin?: number;
  totalGMV: number;
  platformFeeRevenue: number;
  providerPayouts: number;
  netIncome: number;
  transactionsCount: number;
  avgTransactionValue: number;
  breakdown: FinancialBreakdownItem[];
}

export interface ModerationReportItem {
  id: string;
  reporterId: string;
  reporterName: string;
  reporterEmail: string;
  reporterPhone?: string;
  reporterType?: string;
  reporterAvatar?: string;
  
  reportedUserId?: string;
  reportedUserName: string;
  reportedUserEmail: string;
  reportedUserPhone?: string;
  reportedUserType?: string;
  reportedUserAvatar?: string;
  
  resourceType: string;
  resourceId: string;
  category: string;
  reason: string;
  description: string;
  fullContent: string;
  status: 'PENDING' | 'RESOLVED' | 'DISMISSED';
  adminNotes?: string;
  resolutionNotes?: string;
  createdAt: string;
  updatedAt?: string;
}
