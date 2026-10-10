export interface ApprovalRequest {
  id: string;
  requester_id: string;
  requester_email?: string;
  approver_id?: string;
  approver_email?: string;
  action_type: 'HIGH_VALUE_REFUND' | 'USER_BAN' | 'ESCROW_RELEASE' | 'SURGE_OVERRIDE';
  target_resource: string;
  target_id: string;
  payload_json?: string;
  amount: number;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
  rejection_reason?: string;
  created_at: string;
  resolved_at?: string;
}

export interface ChargebackEvidencePackage {
  dispute_id: string;
  appointment_id: string;
  seeker_name: string;
  seeker_email: string;
  provider_name: string;
  provider_email: string;
  appointment_date: string;
  total_amount: number;
  security_code_verified: boolean;
  geofence_verified: boolean;
  check_in_timestamp: string;
  chat_transcript_snippet?: string;
  before_photos: string[];
  after_photos: string[];
  evidence_summary: string;
  stripe_dispute_id?: string;
}

export interface ImpersonationSession {
  admin_id: string;
  admin_email: string;
  target_user_id: string;
  target_user_role: string;
  target_email: string;
  token: string;
  expires_at: string;
  reason: string;
}

export interface SLADispatchAlert {
  appointment_id: string;
  service_name: string;
  seeker_name: string;
  address_city: string;
  wait_minutes: number;
  current_radius_km: number;
  escalation_stage: number;
  is_surge_boosted: boolean;
  surge_multiplier: number;
  available_pros_in_area: number;
  status: 'AT_RISK' | 'ESCALATED' | 'RESOLVED';
}

export interface ProviderQualityHealth {
  provider_id: string;
  provider_name: string;
  provider_email: string;
  overall_score: number;
  on_time_arrival_pct: number;
  dispute_frequency_pct: number;
  response_time_minutes: number;
  customer_sentiment_pct: number;
  completed_jobs_count: number;
  health_status: 'HEALTHY' | 'WARNING' | 'RETRAINING_REQUIRED';
}

export interface SurgePricingRule {
  id?: string;
  region_zip: string;
  service_category_id?: string;
  category_name?: string;
  multiplier: number;
  reason: string;
  is_active: boolean;
  created_at?: string;
  expires_at?: string;
}
