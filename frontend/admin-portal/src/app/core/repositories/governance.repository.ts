import { Observable } from 'rxjs';
import {
  ApprovalRequest,
  ChargebackEvidencePackage,
  ImpersonationSession,
  SLADispatchAlert,
  ProviderQualityHealth,
  SurgePricingRule
} from '../domain/entities/governance.model';

export abstract class GovernanceRepository {
  // Maker-Checker Approvals
  abstract listApprovalRequests(status?: string): Observable<{ results: ApprovalRequest[]; count: number }>;
  abstract createApprovalRequest(req: Partial<ApprovalRequest>): Observable<ApprovalRequest>;
  abstract resolveApprovalRequest(id: string, status: 'APPROVED' | 'REJECTED', reason?: string): Observable<{ status: string }>;

  // Chargeback Evidence Kit
  abstract compileChargebackEvidence(disputeId: string): Observable<ChargebackEvidencePackage>;

  // Safe Impersonation Token
  abstract generateImpersonationToken(userId: string, reason: string): Observable<ImpersonationSession>;

  // Predictive SLA Dispatch
  abstract listSLADispatchAlerts(): Observable<{ results: SLADispatchAlert[]; count: number }>;
  abstract escalateSLADispatch(appointmentId: string): Observable<{ status: string }>;

  // Provider Quality Health Scores
  abstract listProviderQualityHealth(): Observable<{ results: ProviderQualityHealth[]; count: number }>;

  // Dynamic Surge Pricing
  abstract listSurgePricingRules(): Observable<{ results: SurgePricingRule[]; count: number }>;
  abstract saveSurgePricingRule(rule: Partial<SurgePricingRule>): Observable<SurgePricingRule>;
  abstract deleteSurgePricingRule(id: string): Observable<{ status: string }>;
}
