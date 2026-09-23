import { Observable } from 'rxjs';
import { SystemHealthItem, BackupSnapshot, TOTPSetupResponse, AuditLogItem, WorkerTelemetryResponse } from '../domain/entities/system.model';

export abstract class SystemRepository {
  abstract listAuditLogs(limit?: number): Observable<AuditLogItem[]>;
  abstract getSystemHealth(): Observable<SystemHealthItem>;
  abstract getWorkersStatus(): Observable<WorkerTelemetryResponse>;
  abstract triggerBackup(): Observable<{ success: boolean; message: string }>;
  abstract listBackups(): Observable<BackupSnapshot[]>;
  abstract downloadBackup(id: string): Observable<Blob>;
  abstract setup2FA(): Observable<TOTPSetupResponse>;
  abstract verify2FA(code: string): Observable<{ success: boolean }>;
  abstract disable2FA(code: string): Observable<{ success: boolean }>;
  abstract getMaintenanceMode(): Observable<{ maintenance_mode: boolean; checked_at: string }>;
  abstract setMaintenanceMode(enabled: boolean, reason?: string): Observable<{ maintenance_mode: boolean; status: string }>;
  abstract listWebhookEvents(): Observable<{ events: any[]; count: number }>;
}
