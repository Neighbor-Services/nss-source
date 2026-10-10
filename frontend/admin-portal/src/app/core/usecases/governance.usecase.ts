import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { GovernanceRepository } from '../repositories/governance.repository';
import {
  ApprovalRequest,
  ChargebackEvidencePackage,
  ImpersonationSession,
  SLADispatchAlert,
  ProviderQualityHealth,
  SurgePricingRule
} from '../domain/entities/governance.model';

@Injectable({
  providedIn: 'root'
})
export class GovernanceUseCase {
  constructor(private repo: GovernanceRepository) {}

  listApprovalRequests(status?: string): Observable<{ results: ApprovalRequest[]; count: number }> {
    return this.repo.listApprovalRequests(status);
  }

  createApprovalRequest(req: Partial<ApprovalRequest>): Observable<ApprovalRequest> {
    return this.repo.createApprovalRequest(req);
  }

  resolveApprovalRequest(id: string, status: 'APPROVED' | 'REJECTED', reason?: string): Observable<{ status: string }> {
    return this.repo.resolveApprovalRequest(id, status, reason);
  }

  compileChargebackEvidence(disputeId: string): Observable<ChargebackEvidencePackage> {
    return this.repo.compileChargebackEvidence(disputeId);
  }

  generateImpersonationToken(userId: string, reason: string): Observable<ImpersonationSession> {
    return this.repo.generateImpersonationToken(userId, reason);
  }

  listSLADispatchAlerts(): Observable<{ results: SLADispatchAlert[]; count: number }> {
    return this.repo.listSLADispatchAlerts();
  }

  escalateSLADispatch(appointmentId: string): Observable<{ status: string }> {
    return this.repo.escalateSLADispatch(appointmentId);
  }

  listProviderQualityHealth(): Observable<{ results: ProviderQualityHealth[]; count: number }> {
    return this.repo.listProviderQualityHealth();
  }

  listSurgePricingRules(): Observable<{ results: SurgePricingRule[]; count: number }> {
    return this.repo.listSurgePricingRules();
  }

  saveSurgePricingRule(rule: Partial<SurgePricingRule>): Observable<SurgePricingRule> {
    return this.repo.saveSurgePricingRule(rule);
  }

  deleteSurgePricingRule(id: string): Observable<{ status: string }> {
    return this.repo.deleteSurgePricingRule(id);
  }
}
