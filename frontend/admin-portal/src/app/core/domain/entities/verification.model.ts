export interface VerificationItem {
  id: string;
  userId: string;
  userName: string;
  userEmail: string;
  documentType: string;
  documentUrl: string;
  selfieUrl?: string;
  status: 'pending' | 'approved' | 'rejected';
  submittedAt: string;
  rejectionReason?: string;
}
