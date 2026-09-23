import { Component, OnInit, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ReportUseCase } from '../../../core/usecases/report.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { FinancialReport } from '../../../core/domain/entities/report.model';

@Component({
  selector: 'app-financial',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './financial.component.html',
  styleUrl: './financial.component.css'
})
export class FinancialComponent implements OnInit, OnDestroy {
  report = signal<FinancialReport | null>(null);
  loading = signal<boolean>(false);
  exporting = signal<string | null>(null);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);
  isReconciling = signal<boolean>(false);
  activePreset = signal<string>('30d');
  selectedTaxYear = signal<number>(2026);
  selectedTaxQuarter = signal<string>('Q3');

  // Stripe Live Balance
  stripeBalance = signal<{ available: number; pending: number; reserved: number; currency: string; lastRefreshed: string } | null>(null);
  isLoadingBalance = signal(false);
  private balanceRefreshTimer: any = null;

  startDate = '';
  endDate = '';

  constructor(
    private reportUseCase: ReportUseCase,
    private dialog: DialogService
  ) {
    this.setPreset('30d');
  }

  ngOnInit(): void {
    this.fetchFinancialReport();
    this.loadStripeBalance();
    // Auto-refresh Stripe balance every 30s
    this.balanceRefreshTimer = setInterval(() => this.loadStripeBalance(), 30000);
  }

  ngOnDestroy(): void {
    if (this.balanceRefreshTimer) {
      clearInterval(this.balanceRefreshTimer);
    }
  }

  setPreset(preset: '7d' | '30d' | '90d' | 'ytd' | 'all') {
    this.activePreset.set(preset);
    const end = new Date();
    const start = new Date();

    switch (preset) {
      case '7d':
        start.setDate(end.getDate() - 7);
        break;
      case '30d':
        start.setDate(end.getDate() - 30);
        break;
      case '90d':
        start.setDate(end.getDate() - 90);
        break;
      case 'ytd':
        start.setMonth(0, 1);
        break;
      case 'all':
        start.setFullYear(2023, 0, 1);
        break;
    }

    this.startDate = start.toISOString().split('T')[0];
    this.endDate = end.toISOString().split('T')[0];
    this.fetchFinancialReport();
  }

  fetchFinancialReport() {
    this.loading.set(true);
    this.errorMessage.set(null);

    this.reportUseCase.getFinancialReport(this.startDate, this.endDate).subscribe({
      next: (data) => {
        this.report.set(data);
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load GAAP financial report metrics');
        this.loading.set(false);
      }
    });
  }

  runReconciliationAudit() {
    this.isReconciling.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.reportUseCase.auditLedger().subscribe({
      next: (res) => {
        this.isReconciling.set(false);
        this.successMessage.set(
          `Audit Verified: Stripe Inflows ($${res.stripeInflow.toFixed(2)}) match Platform Escrows ($${res.platformEscrow.toFixed(2)}) with ${res.reconciledCount} transactions perfectly balanced.`
        );
        this.fetchFinancialReport();
        setTimeout(() => this.successMessage.set(null), 7000);
      },
      error: (err) => {
        this.isReconciling.set(false);
        this.errorMessage.set(err.error?.message || 'Stripe Ledger discrepancy detected. Review failed transfers.');
      }
    });
  }

  generateTaxPackage() {
    this.exporting.set('tax');
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.reportUseCase.export1099KPackage(this.selectedTaxYear(), this.selectedTaxQuarter()).subscribe({
      next: () => {
        this.exporting.set(null);
        this.successMessage.set(`IRS 1099-K Tax Package for ${this.selectedTaxYear()} (${this.selectedTaxQuarter()}) exported successfully!`);
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.exporting.set(null);
        this.errorMessage.set(err.error?.message || 'Failed to generate IRS 1099-K Tax Package.');
      }
    });
  }

  downloadExport(type: 'users' | 'payouts' | 'disputes') {
    this.exporting.set(type);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.reportUseCase.exportData(type, this.startDate, this.endDate).subscribe({
      next: () => {
        this.exporting.set(null);
        this.successMessage.set(`${type.toUpperCase()} data exported to CSV successfully!`);
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.exporting.set(null);
        this.errorMessage.set(err.error?.message || `Failed to export ${type} data.`);
      }
    });
  }

  formatCurrency(val: number | undefined): string {
    if (val === undefined || isNaN(val)) return '$0.00';
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
  }

  formatNumber(val: number | undefined): string {
    if (val === undefined || isNaN(val)) return '0';
    return new Intl.NumberFormat('en-US').format(val);
  }

  loadStripeBalance(): void {
    this.isLoadingBalance.set(true);
    this.reportUseCase.getStripeLiveBalance().subscribe({
      next: (bal) => {
        this.stripeBalance.set(bal);
        this.isLoadingBalance.set(false);
      },
      error: () => {
        this.isLoadingBalance.set(false);
      }
    });
  }
}
