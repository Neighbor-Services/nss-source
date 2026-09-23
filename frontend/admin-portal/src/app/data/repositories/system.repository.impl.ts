import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { SystemRepository } from '../../core/repositories/system.repository';
import {
  SystemHealthItem,
  BackupSnapshot,
  TOTPSetupResponse,
  AuditLogItem
} from '../../core/domain/entities/system.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class SystemRepositoryImpl implements SystemRepository {
  constructor(private http: HttpClient) {}

  listAuditLogs(limit?: number): Observable<AuditLogItem[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.auditLogs}?limit=${limit || 50}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((a: any) => ({
          id: a.id?.toString() || '',
          actorEmail: a.actor_email || (a.user ? a.user.email : 'system@neighborservice.com'),
          action: a.action || 'ACTION',
          resource: a.resource_type || a.resource || 'System',
          ipAddress: a.ip_address || '127.0.0.1',
          timestamp: a.created_at || a.timestamp || new Date().toISOString(),
          details: a.details
        }));
      })
    );
  }

  getSystemHealth(): Observable<SystemHealthItem> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.systemHealth}`).pipe(
      map(res => ({
        status: res.status || 'healthy',
        uptimeSeconds: res.uptime_seconds || 0,
        databaseStatus: res.database || res.db_status || 'connected',
        redisStatus: res.redis || res.redis_status || 'connected',
        memoryUsageMB: res.memory_mb || res.memory_usage_mb || 0,
        cpuPercent: res.cpu_percent || 0,
        activeGoroutines: res.goroutines || res.active_goroutines || 0,
        version: res.version || '1.0.0'
      }))
    );
  }

  getWorkersStatus(): Observable<any> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}/admin/system/workers`);
  }

  triggerBackup(): Observable<{ success: boolean; message: string }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.systemBackup}`, {}).pipe(
      map(res => ({ success: true, message: res.message || 'Database snapshot created successfully.' }))
    );
  }

  listBackups(): Observable<BackupSnapshot[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.systemBackups}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((b: any) => ({
          id: b.id?.toString() || '',
          filename: b.filename || 'backup.sql.gz',
          fileSize: b.file_size || 0,
          status: b.status || 'COMPLETED',
          storagePath: b.storage_path || '',
          checksum: b.checksum || '',
          createdAt: b.created_at || new Date().toISOString(),
          completedAt: b.completed_at
        }));
      })
    );
  }

  downloadBackup(id: string): Observable<Blob> {
    return this.http.get(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.systemBackups}/${id}/download`, {
      responseType: 'blob'
    });
  }

  setup2FA(): Observable<TOTPSetupResponse> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.twoFASetup}`, {}).pipe(
      map(res => ({
        secret: res.secret || '',
        otpAuthUrl: res.otp_auth_url || '',
        qrCodeUrl: res.qr_code_url || ''
      }))
    );
  }

  verify2FA(code: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.twoFAVerify}`, { code }).pipe(
      map(() => ({ success: true }))
    );
  }

  disable2FA(code: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.twoFADisable}`, { code }).pipe(
      map(() => ({ success: true }))
    );
  }

  getMaintenanceMode(): Observable<{ maintenance_mode: boolean; checked_at: string }> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.maintenance}`);
  }

  setMaintenanceMode(enabled: boolean, reason: string = ''): Observable<{ maintenance_mode: boolean; status: string }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.maintenance}`, { enabled, reason });
  }

  listWebhookEvents(): Observable<{ events: any[]; count: number }> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.webhookEvents}`);
  }
}
