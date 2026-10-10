export interface AuditLogItem {
  id: string;
  actorEmail: string;
  action: string;
  resource: string;
  ipAddress: string;
  timestamp: string;
  details?: any;
}

export interface SystemHealthItem {
  status: 'healthy' | 'degraded' | 'unhealthy';
  uptimeSeconds: number;
  databaseStatus: 'connected' | 'error';
  redisStatus: 'connected' | 'error';
  memoryUsageMB: number;
  cpuPercent: number;
  activeGoroutines: number;
  version: string;
}

export interface BackupSnapshot {
  id: string;
  filename: string;
  fileSize: number;
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
  storagePath?: string;
  checksum?: string;
  createdAt: string;
  completedAt?: string;
}

export interface TOTPSetupResponse {
  secret: string;
  otpAuthUrl: string;
  qrCodeUrl: string;
}

export interface WorkerTelemetryItem {
  name: string;
  category: string;
  interval: string;
  status: string;
  last_run: string;
  processed_items: number;
  description: string;
}

export interface WorkerTelemetryResponse {
  total_workers: number;
  all_healthy: boolean;
  workers: WorkerTelemetryItem[];
}

export interface HeatmapCluster {
  latitude: number;
  longitude: number;
  request_count: number;
  provider_count: number;
  supply_deficit: number;
  zip_code: string;
  city: string;
}

export interface OperationsHeatmapData {
  clusters: HeatmapCluster[];
  total_requests: number;
  total_providers: number;
  underserved_pct: number;
}

export interface EscrowSummary {
  total_held_in_escrow: number;
  active_jobs_count: number;
  pending_clearance: number;
  disputed_funds: number;
  escrow_velocity_avg_hours: number;
}

export interface SubscriptionCohortStats {
  total_subscribers: number;
  silver_count: number;
  gold_count: number;
  platinum_count: number;
  mrr: number;
  arr: number;
  churn_rate_pct: number;
  tier_conversion_pct: Record<string, number>;
}

