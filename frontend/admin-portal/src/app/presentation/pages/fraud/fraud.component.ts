import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FraudUseCase } from '../../../core/usecases/fraud.usecase';
import { FraudRiskAlert } from '../../../core/domain/entities/fraud.model';

export type FraudMainTab = 'ALERTS' | 'LEAKAGE' | 'CREDENTIALS' | 'SCANNER' | 'RADAR_RULES';

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

  // Leakage & Credentials State
  leakageAlerts = signal<any[]>([]);
  loadingLeakage = signal<boolean>(false);
  expiringCredentials = signal<any[]>([]);
  loadingCredentials = signal<boolean>(false);

  activeTab = signal<FraudMainTab>('ALERTS');
  selectedStatus = '';
  selectedLevel = '';
  searchQuery = signal<string>('');
  manualUserId = '';


  selectedAlert = signal<FraudRiskAlert | null>(null);
  evaluationResult = signal<FraudRiskAlert | null>(null);

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed metrics
  criticalCount = computed(() => this.alerts().filter(a => a.riskScore >= 70 && a.status === 'OPEN').length);
  openAlertsCount = computed(() => this.alerts().filter(a => a.status === 'OPEN').length);

  filteredAlerts = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const lvl = this.selectedLevel;
    const st = this.selectedStatus;

    return this.alerts().filter(a => {
      const matchQuery = !q || 
        (a.userName && a.userName.toLowerCase().includes(q)) ||
        (a.userEmail && a.userEmail.toLowerCase().includes(q)) ||
        a.userId.toLowerCase().includes(q) ||
        (a.flags && a.flags.some(f => f.toLowerCase().includes(q)));
      
      const matchLevel = !lvl || a.riskLevel === lvl;
      const matchStatus = !st || a.status === st;

      return matchQuery && matchLevel && matchStatus;
    });
  });

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

  openDossier(alert: FraudRiskAlert): void {
    this.selectedAlert.set(alert);
  }

  closeDossier(): void {
    this.selectedAlert.set(null);
  }

  runManualEvaluation(): void {
    if (!this.manualUserId.trim()) {
      this.errorMessage.set('Please provide a valid User ID to evaluate');
      return;
    }

    this.evaluating.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);
    this.evaluationResult.set(null);

    this.fraudUseCase.evaluateUserRisk(this.manualUserId.trim()).subscribe({
      next: (alert) => {
        this.evaluating.set(false);
        this.evaluationResult.set(alert);
        this.successMessage.set(`Evaluated risk for user ${this.manualUserId}: Score ${alert.riskScore}/100 (${alert.riskLevel})`);
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
        if (this.selectedAlert()?.id === alertId) {
          this.selectedAlert.update(a => a ? { ...a, status: action } : null);
        }
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

  formatFlag(flag: string): string {
    return flag.replace(/_/g, ' ');
  }

  fetchLeakageAlerts(): void {
    this.loadingLeakage.set(true);
    this.fraudUseCase.getLeakageAlerts().subscribe({
      next: (res) => {
        this.leakageAlerts.set(res.results || []);
        this.loadingLeakage.set(false);
      },
      error: () => {
        this.loadingLeakage.set(false);
      }
    });
  }

  fetchExpiringCredentials(): void {
    this.loadingCredentials.set(true);
    this.fraudUseCase.getExpiringCredentials().subscribe({
      next: (res) => {
        this.expiringCredentials.set(res.results || []);
        this.loadingCredentials.set(false);
      },
      error: () => {
        this.loadingCredentials.set(false);
      }
    });
  }
}

