import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ADMIN_API_CONFIG } from '../../data/datasources/admin-api.config';
import { AuthService } from '../../data/datasources/auth.service';

export interface AuditLogPayload {
  actionType: string;
  targetEntity?: string;
  targetId?: string;
  details?: string;
  metadata?: Record<string, any>;
}

@Injectable({
  providedIn: 'root'
})
export class AuditService {
  private http = inject(HttpClient);
  private auth = inject(AuthService);

  logAction(payload: AuditLogPayload) {
    const adminUser = this.auth.currentUser();
    const body = {
      action: payload.actionType,
      target_entity: payload.targetEntity || 'SYSTEM',
      target_id: payload.targetId || '',
      details: payload.details || '',
      admin_email: adminUser?.email || 'admin@neighborservices.io',
      admin_id: adminUser?.id || '',
      timestamp: new Date().toISOString(),
      metadata: payload.metadata || {}
    };

    // Post to audit endpoint (fire and forget with fallback to console)
    this.http.post(`${ADMIN_API_CONFIG.baseUrl}/audit/`, body).subscribe({
      next: () => {},
      error: (err) => {
        // Fallback local logging
        console.info(`[AUDIT] ${body.action} on ${body.target_entity} (${body.target_id}): ${body.details}`);
      }
    });
  }
}
