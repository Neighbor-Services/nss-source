export interface VerificationItem {
  id: string;
  userId: string;
  userName: string;
  userEmail: string;
  documentType: string;
  documentUrl: string;
  documentBackUrl?: string;
  selfieUrl?: string;
  tradeLicenseUrl?: string;
  licenseNumber?: string;
  status: 'pending' | 'approved' | 'rejected';
  submittedAt: string;
  rejectionReason?: string;
}
