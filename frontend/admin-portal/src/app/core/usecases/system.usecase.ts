import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { SystemRepository } from '../repositories/system.repository';
import {
  SystemHealthItem,
  BackupSnapshot,
  TOTPSetupResponse,
  AuditLogItem
} from '../domain/entities/system.model';

@Injectable({
  providedIn: 'root'
})
export class SystemUseCase {
  constructor(private systemRepo: SystemRepository) {}

  listAuditLogs(limit?: number): Observable<AuditLogItem[]> {
    return this.systemRepo.listAuditLogs(limit);
  }

  getSystemHealth(): Observable<SystemHealthItem> {
    return this.systemRepo.getSystemHealth();
  }

  getWorkersStatus(): Observable<any> {
    return this.systemRepo.getWorkersStatus();
  }

  triggerBackup(): Observable<{ success: boolean; message: string }> {
    return this.systemRepo.triggerBackup();
  }

  listBackups(): Observable<BackupSnapshot[]> {
    return this.systemRepo.listBackups();
  }

  downloadBackup(id: string): Observable<Blob> {
    return this.systemRepo.downloadBackup(id);
  }

  setup2FA(): Observable<TOTPSetupResponse> {
    return this.systemRepo.setup2FA();
  }

  verify2FA(code: string): Observable<{ success: boolean }> {
    return this.systemRepo.verify2FA(code);
  }

  disable2FA(code: string): Observable<{ success: boolean }> {
    return this.systemRepo.disable2FA(code);
  }

  getMaintenanceMode(): Observable<{ maintenance_mode: boolean; checked_at: string }> {
    return this.systemRepo.getMaintenanceMode();
  }

  setMaintenanceMode(enabled: boolean, reason?: string): Observable<{ maintenance_mode: boolean; status: string }> {
    return this.systemRepo.setMaintenanceMode(enabled, reason);
  }

  listWebhookEvents(): Observable<{ events: any[]; count: number }> {
    return this.systemRepo.listWebhookEvents();
  }
}
