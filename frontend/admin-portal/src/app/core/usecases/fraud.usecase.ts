import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { FraudRepository } from '../repositories/fraud.repository';
import { FraudRiskAlert } from '../domain/entities/fraud.model';

@Injectable({
  providedIn: 'root'
})
export class FraudUseCase {
  constructor(private fraudRepo: FraudRepository) {}

  listFraudRiskAlerts(status?: string): Observable<{ results: FraudRiskAlert[]; count: number }> {
    return this.fraudRepo.listFraudRiskAlerts(status);
  }

  evaluateUserRisk(userId: string): Observable<FraudRiskAlert> {
    return this.fraudRepo.evaluateUserRisk(userId);
  }

  resolveRiskAlert(alertId: string, action: string): Observable<{ success: boolean }> {
    return this.fraudRepo.resolveRiskAlert(alertId, action);
  }
}
