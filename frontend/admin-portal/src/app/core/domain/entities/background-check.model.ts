export interface BackgroundCheckItem {
  id: string;
  userId: string;
  userName: string;
  userEmail: string;
  checkrCandidateId: string;
  checkrReportId: string;
  status: 'clear' | 'consider' | 'pending' | 'suspended';
  package: string;
  turnaroundTime?: string;
  completedAt?: string;
  manualOverride?: boolean;
}
