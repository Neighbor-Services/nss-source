export interface DispatchIncidentSummary {
  id: string;
  seeker_id: string;
  seeker_name: string;
  seeker_phone: string;
  seeker_avatar?: string;
  title: string;
  description: string;
  status: string; // BROADCASTING, ACCEPTED, IN_PROGRESS, COMPLETED, CANCELLED, EXPIRED
  latitude: number;
  longitude: number;
  address: string;
  broadcast_radius_km: number;
  max_budget?: number;
  candidate_count?: number;
  accepted_provider_id?: string;
  provider_name?: string;
  provider_phone?: string;
  provider_latitude?: number;
  provider_longitude?: number;
  estimated_eta?: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DispatchOverrideRequest {
  action: 'REASSIGN' | 'CANCEL' | 'EXTEND_RADIUS' | string;
  target_provider_id?: string;
  extend_radius_km?: number;
  reason?: string;
}

export interface ProviderFleetTelemetry {
  id: string;
  user_id: string;
  name: string;
  email: string;
  phone: string;
  avatar?: string;
  is_online: boolean;
  latitude: number;
  longitude: number;
  service: string;
  rating: number;
  reviews_count: number;
}
