export interface DisputeItem {
  id: string;
  appointmentId: string;
  appointmentTitle?: string;
  seekerId?: string;
  seekerName: string;
  seekerEmail?: string;
  seekerPhone?: string;
  seekerAvatar?: string;
  providerId?: string;
  providerName: string;
  providerEmail?: string;
  providerPhone?: string;
  providerAvatar?: string;
  reason: string;
  description: string;
  amount: number;
  status: 'OPEN' | 'UNDER_REVIEW' | 'ESCALATED' | 'RESOLVED' | 'REJECTED' | 'DISMISSED' | string;
  createdAt: string;
  updatedAt?: string;
  evidenceUrls?: string[];
  resolutionNotes?: string;
}

export interface DisputeMediationInput {
  resolution_type: 'FULL_REFUND' | 'PARTIAL_SPLIT' | 'RELEASE_TO_PROVIDER' | 'ISSUE_CREDIT';
  refund_amount: number;
  provider_payout: number;
  credit_voucher: number;
  notes: string;
}

