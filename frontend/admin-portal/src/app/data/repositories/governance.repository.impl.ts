import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { GovernanceRepository } from '../../core/repositories/governance.repository';
import {
  ApprovalRequest,
  ChargebackEvidencePackage,
  ImpersonationSession,
  SLADispatchAlert,
  ProviderQualityHealth,
  SurgePricingRule
} from '../../core/domain/entities/governance.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class GovernanceRepositoryImpl implements GovernanceRepository {
  private base = ADMIN_API_CONFIG.baseUrl;

  constructor(private http: HttpClient) {}

  listApprovalRequests(status?: string): Observable<{ results: ApprovalRequest[]; count: number }> {
    const query = status && status !== 'ALL' ? `?status=${status}` : '';
    return this.http.get<{ results: ApprovalRequest[]; count: number }>(`${this.base}/admin/approvals${query}`);
  }

  createApprovalRequest(req: Partial<ApprovalRequest>): Observable<ApprovalRequest> {
    return this.http.post<ApprovalRequest>(`${this.base}/admin/approvals/`, req);
  }

  resolveApprovalRequest(id: string, status: 'APPROVED' | 'REJECTED', reason?: string): Observable<{ status: string }> {
    return this.http.post<{ status: string }>(`${this.base}/admin/approvals/${id}/resolve/`, {
      status,
      reason
    });
  }

  compileChargebackEvidence(disputeId: string): Observable<ChargebackEvidencePackage> {
    return this.http.get<ChargebackEvidencePackage>(`${this.base}/admin/disputes/${disputeId}/chargeback-evidence/`);
  }

  generateImpersonationToken(userId: string, reason: string): Observable<ImpersonationSession> {
    return this.http.post<ImpersonationSession>(`${this.base}/admin/users/impersonate-token/`, {
      user_id: userId,
      reason
    });
  }

  listSLADispatchAlerts(): Observable<{ results: SLADispatchAlert[]; count: number }> {
    return this.http.get<{ results: SLADispatchAlert[]; count: number }>(`${this.base}/admin/dispatch/sla-alerts/`);
  }

  escalateSLADispatch(appointmentId: string): Observable<{ status: string }> {
    return this.http.post<{ status: string }>(`${this.base}/admin/dispatch/sla-alerts/${appointmentId}/escalate/`, {});
  }

  listProviderQualityHealth(): Observable<{ results: ProviderQualityHealth[]; count: number }> {
    return this.http.get<{ results: ProviderQualityHealth[]; count: number }>(`${this.base}/admin/providers/quality-health/`);
  }

  listSurgePricingRules(): Observable<{ results: SurgePricingRule[]; count: number }> {
    return this.http.get<{ results: SurgePricingRule[]; count: number }>(`${this.base}/admin/pricing/surge-rules/`);
  }

  saveSurgePricingRule(rule: Partial<SurgePricingRule>): Observable<SurgePricingRule> {
    return this.http.post<SurgePricingRule>(`${this.base}/admin/pricing/surge-rules/`, rule);
  }

  deleteSurgePricingRule(id: string): Observable<{ status: string }> {
    return this.http.delete<{ status: string }>(`${this.base}/admin/pricing/surge-rules/${id}/`);
  }
}
