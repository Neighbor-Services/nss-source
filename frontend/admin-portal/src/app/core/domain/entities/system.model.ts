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
