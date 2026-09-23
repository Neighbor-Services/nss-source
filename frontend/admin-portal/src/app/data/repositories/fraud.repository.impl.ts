import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { FraudRepository } from '../../core/repositories/fraud.repository';
import { FraudRiskAlert } from '../../core/domain/entities/fraud.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class FraudRepositoryImpl implements FraudRepository {
  constructor(private http: HttpClient) {}

  listFraudRiskAlerts(status?: string): Observable<{ results: FraudRiskAlert[]; count: number }> {
    const url = status
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.fraudRiskAlerts}?status=${status}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.fraudRiskAlerts}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        const results: FraudRiskAlert[] = raw.map((a: any) => ({
          id: a.id?.toString() || '',
          userId: a.user_id || a.user?.id || '',
          userName: a.user ? `${a.user.first_name || ''} ${a.user.last_name || ''}`.trim() : 'User Account',
          userEmail: a.user?.email || '',
          riskScore: a.risk_score ?? 0,
          riskLevel: a.risk_level || 'LOW',
          flags: Array.isArray(a.flags) ? a.flags : (typeof a.flags === 'string' ? JSON.parse(a.flags) : []),
          status: a.status || 'OPEN',
          createdAt: a.created_at || new Date().toISOString(),
          details: a.details
        }));
        return {
          results,
          count: res.count ?? results.length
        };
      })
    );
  }

  evaluateUserRisk(userId: string): Observable<FraudRiskAlert> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.fraudEvaluate}/${userId}`, {}).pipe(
      map(a => ({
        id: a.id?.toString() || '',
        userId: a.user_id || userId,
        userName: a.user ? `${a.user.first_name || ''} ${a.user.last_name || ''}`.trim() : 'User Account',
        userEmail: a.user?.email || '',
        riskScore: a.risk_score ?? 0,
        riskLevel: a.risk_level || 'LOW',
        flags: Array.isArray(a.flags) ? a.flags : [],
        status: a.status || 'OPEN',
        createdAt: a.created_at || new Date().toISOString(),
        details: a.details
      }))
    );
  }

  resolveRiskAlert(alertId: string, action: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.fraudRiskAlerts}/${alertId}/resolve`, {
      action
    }).pipe(
      map(() => ({ success: true }))
    );
  }
}
