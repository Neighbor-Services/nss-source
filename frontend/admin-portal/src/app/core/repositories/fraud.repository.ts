import { Observable } from 'rxjs';
import { FraudRiskAlert } from '../domain/entities/fraud.model';

export abstract class FraudRepository {
  abstract listFraudRiskAlerts(status?: string): Observable<{ results: FraudRiskAlert[]; count: number }>;
  abstract evaluateUserRisk(userId: string): Observable<FraudRiskAlert>;
  abstract resolveRiskAlert(alertId: string, action: string): Observable<{ success: boolean }>;
}
