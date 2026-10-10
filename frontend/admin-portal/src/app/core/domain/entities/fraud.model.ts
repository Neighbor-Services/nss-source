export interface FraudRiskAlert {
  id: string;
  userId: string;
  userName?: string;
  userEmail?: string;
  riskScore: number;
  riskLevel: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  flags: string[];
  status: 'OPEN' | 'REVIEWED' | 'DISMISSED' | 'ACTIONED';
  details?: any;
  createdAt?: string;
  created_at?: string;
}

export interface LeakageAlert {
  id: string;
  sender_id: string;
  sender_name: string;
  receiver_id: string;
  receiver_name: string;
  message_snippet: string;
  matched_keyword: string;
  risk_score: number;
  detected_at: string;
  status: 'PENDING' | 'REVIEWED' | 'SHADOWBANNED' | 'DISMISSED';
}

export interface ExpiringCredential {
  id: string;
  provider_id: string;
  provider_name: string;
  provider_email: string;
  credential_type: string;
  document_number: string;
  expires_at: string;
  days_remaining: number;
  status: 'EXPIRING_SOON' | 'EXPIRED';
}

