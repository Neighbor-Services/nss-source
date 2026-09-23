import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FraudUseCase } from '../../../core/usecases/fraud.usecase';
import { FraudRiskAlert } from '../../../core/domain/entities/fraud.model';

export type FraudMainTab = 'ALERTS' | 'SCANNER' | 'RADAR_RULES';

@Component({
  selector: 'app-fraud',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './fraud.component.html',
  styleUrl: './fraud.component.css'
})
export class FraudComponent implements OnInit {
  alerts = signal<FraudRiskAlert[]>([]);
  totalCount = signal<number>(0);
  loading = signal<boolean>(false);
  evaluating = signal<boolean>(false);
  resolvingAlertId = signal<string | null>(null);

  activeTab = signal<FraudMainTab>('ALERTS');
  selectedStatus = '';
  manualUserId = '';

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  constructor(private fraudUseCase: FraudUseCase) {}

  ngOnInit(): void {
    this.fetchRiskAlerts();
  }

  fetchRiskAlerts(): void {
    this.loading.set(true);
    this.errorMessage.set(null);
    this.fraudUseCase.listFraudRiskAlerts(this.selectedStatus || undefined).subscribe({
      next: (res) => {
        this.alerts.set(res.results);
        this.totalCount.set(res.count);
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load fraud risk alerts');
        this.loading.set(false);
      }
    });
  }

  runManualEvaluation(): void {
    if (!this.manualUserId.trim()) {
      this.errorMessage.set('Please provide a valid User ID to evaluate');
      return;
    }

    this.evaluating.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.fraudUseCase.evaluateUserRisk(this.manualUserId.trim()).subscribe({
      next: (alert) => {
        this.evaluating.set(false);
        this.successMessage.set(`Evaluated risk for user ${this.manualUserId}: Score ${alert.riskScore}/100 (${alert.riskLevel})`);
        this.manualUserId = '';
        this.fetchRiskAlerts();
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to evaluate user risk');
        this.evaluating.set(false);
      }
    });
  }

  resolveAlert(alertId: string, action: 'ACTIONED' | 'DISMISSED'): void {
    this.resolvingAlertId.set(alertId);
    this.errorMessage.set(null);

    this.fraudUseCase.resolveRiskAlert(alertId, action).subscribe({
      next: () => {
        this.resolvingAlertId.set(null);
        this.successMessage.set(`Risk alert marked as ${action}`);
        this.fetchRiskAlerts();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to resolve risk alert');
        this.resolvingAlertId.set(null);
      }
    });
  }

  getRiskBadgeClass(level: string): string {
    switch (level) {
      case 'CRITICAL': return 'badge-critical';
      case 'HIGH': return 'badge-danger';
      case 'MEDIUM': return 'badge-warning';
      default: return 'badge-low';
    }
  }
}
