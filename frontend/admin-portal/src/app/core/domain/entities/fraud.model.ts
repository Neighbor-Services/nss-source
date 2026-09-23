export interface FraudRiskAlert {
  id: string;
  userId: string;
  userName?: string;
  userEmail?: string;
  riskScore: number;
  riskLevel: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  flags: string[];
  status: 'OPEN' | 'REVIEWED' | 'DISMISSED' | 'ACTIONED';
  createdAt: string;
  details?: any;
}
