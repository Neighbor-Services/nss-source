export interface DisputeItem {
  id: string;
  appointmentId: string;
  seekerName: string;
  providerName: string;
  reason: string;
  description: string;
  amount: number;
  status: 'open' | 'under_review' | 'resolved' | 'rejected';
  createdAt: string;
  evidenceUrls?: string[];
  resolutionNotes?: string;
}
