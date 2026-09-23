export interface PayoutRequestItem {
  id: string;
  providerId: string;
  providerName: string;
  amount: number;
  status: 'pending' | 'approved' | 'rejected' | 'processed';
  requestedAt: string;
  bankAccountLast4?: string;
}

export interface WalletItem {
  id: string;
  userId: string;
  userName: string;
  balance: number;
  pendingBalance: number;
  currency: string;
  updatedAt: string;
}

export interface BatchPayoutResult {
  processed_count: number;
  total_amount: number;
  failed_count: number;
  message: string;
}
