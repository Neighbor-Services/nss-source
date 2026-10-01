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
  evidence?: string;
  resolutionNotes?: string;
}
