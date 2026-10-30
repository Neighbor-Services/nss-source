import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RefundUseCase } from '../../../core/usecases/refund.usecase';
import { RefundRequestItem } from '../../../core/domain/entities/refund.model';

@Component({
  selector: 'app-refunds',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './refunds.component.html',
  styleUrl: './refunds.component.css'
})
export class RefundsComponent implements OnInit {
  refunds = signal<RefundRequestItem[]>([]);
  isLoading = signal(false);
  isProcessing = signal(false);

  selectedStatus = signal<string>('ALL');
  searchTerm = signal<string>('');

  // Modals
  showProcessModal = signal(false);
  showRejectModal = signal(false);
  showDirectRefundModal = signal(false);
  showDetailModal = signal(false);

  selectedRefund = signal<RefundRequestItem | null>(null);

  // Form values
  processAmount = 0;
  processNotes = '';
  rejectReason = '';

  directPaymentIntentId = '';
  directAppointmentId = '';
  directAmount = 0;
  directReason = '';
  directNotes = '';

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // KPIs
  totalCount = computed(() => this.refunds().length);
  pendingCount = computed(() => this.refunds().filter(r => r.status === 'PENDING').length);
  processedCount = computed(() => this.refunds().filter(r => r.status === 'PROCESSED').length);
  rejectedCount = computed(() => this.refunds().filter(r => r.status === 'REJECTED').length);

  totalPendingAmount = computed(() =>
    this.refunds().filter(r => r.status === 'PENDING').reduce((sum, r) => sum + (r.amount || 0), 0)
  );
  totalProcessedAmount = computed(() =>
    this.refunds().filter(r => r.status === 'PROCESSED').reduce((sum, r) => sum + (r.amount || 0), 0)
  );

  filteredRefunds = computed(() => {
    let list = this.refunds();
    const status = this.selectedStatus();
    if (status !== 'ALL') {
      list = list.filter(r => r.status === status);
    }
    const term = this.searchTerm().toLowerCase().trim();
    if (term) {
      list = list.filter(r =>
        (r.payment_intent_id && r.payment_intent_id.toLowerCase().includes(term)) ||
        (r.reason && r.reason.toLowerCase().includes(term)) ||
        (r.user?.email && r.user.email.toLowerCase().includes(term)) ||
        (r.appointment?.title && r.appointment.title.toLowerCase().includes(term)) ||
        (r.stripe_refund_id && r.stripe_refund_id.toLowerCase().includes(term))
      );
    }
    return list;
  });

  constructor(private refundUseCase: RefundUseCase) {}

  ngOnInit(): void {
    this.loadRefunds();
  }

  loadRefunds(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);
    this.refundUseCase.listRefundRequests(this.selectedStatus()).subscribe({
      next: res => {
        this.refunds.set(res.results || []);
        this.isLoading.set(false);
      },
      error: err => {
        this.errorMessage.set('Failed to load refund requests: ' + (err.error?.detail || err.message));
        this.isLoading.set(false);
      }
    });
  }

  setStatusFilter(status: string): void {
    this.selectedStatus.set(status);
    this.loadRefunds();
  }

  exportCSV(): void {
    this.isLoading.set(true);
    this.refundUseCase.exportRefundsCSV(this.selectedStatus()).subscribe({
      next: (blob) => {
        this.isLoading.set(false);
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `refunds_export_${this.selectedStatus().toLowerCase()}_${new Date().toISOString().slice(0, 10)}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
        this.successMessage.set('Refund report CSV exported successfully.');
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to export CSV: ' + (err.error?.detail || err.message));
      }
    });
  }

  getStripeDashboardUrl(paymentIntentId: string): string {
    if (!paymentIntentId) return 'https://dashboard.stripe.com/payments';
    return `https://dashboard.stripe.com/payments/${paymentIntentId}`;
  }

  setProcessAmountFraction(fraction: number): void {
    const refund = this.selectedRefund();
    if (refund) {
      this.processAmount = parseFloat((refund.amount * fraction).toFixed(2));
    }
  }

  openProcessModal(refund: RefundRequestItem): void {
    this.selectedRefund.set(refund);
    this.processAmount = refund.amount;
    this.processNotes = '';
    this.showProcessModal.set(true);
  }

  submitProcessRefund(): void {
    const refund = this.selectedRefund();
    if (!refund) return;

    if (this.processAmount <= 0) {
      this.errorMessage.set('Refund amount must be greater than zero');
      return;
    }

    this.isProcessing.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.refundUseCase.processRefund(refund.id, this.processAmount, this.processNotes).subscribe({
      next: () => {
        this.isProcessing.set(false);
        this.showProcessModal.set(false);
        this.successMessage.set(`Stripe refund of $${this.processAmount.toFixed(2)} processed successfully back to the customer's card!`);
        this.loadRefunds();
        setTimeout(() => this.successMessage.set(null), 6000);
      },
      error: err => {
        this.isProcessing.set(false);
        this.errorMessage.set('Stripe Refund Error: ' + (err.error?.detail || err.message));
      }
    });
  }

  openRejectModal(refund: RefundRequestItem): void {
    this.selectedRefund.set(refund);
    this.rejectReason = '';
    this.showRejectModal.set(true);
  }

  submitRejectRefund(): void {
    const refund = this.selectedRefund();
    if (!refund) return;

    if (!this.rejectReason.trim()) {
      this.errorMessage.set('Please provide a reason for declining the refund');
      return;
    }

    this.isProcessing.set(true);
    this.errorMessage.set(null);

    this.refundUseCase.rejectRefund(refund.id, this.rejectReason.trim()).subscribe({
      next: () => {
        this.isProcessing.set(false);
        this.showRejectModal.set(false);
        this.successMessage.set('Refund request declined and customer notified.');
        this.loadRefunds();
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: err => {
        this.isProcessing.set(false);
        this.errorMessage.set('Failed to decline refund: ' + (err.error?.detail || err.message));
      }
    });
  }

  openDirectRefundModal(): void {
    this.directPaymentIntentId = '';
    this.directAppointmentId = '';
    this.directAmount = 0;
    this.directReason = 'Customer service refund';
    this.directNotes = '';
    this.showDirectRefundModal.set(true);
  }

  submitDirectRefund(): void {
    if (!this.directPaymentIntentId.trim()) {
      this.errorMessage.set('Stripe Payment Intent ID is required (e.g., pi_...)');
      return;
    }
    if (this.directAmount <= 0) {
      this.errorMessage.set('Refund amount must be greater than zero');
      return;
    }

    this.isProcessing.set(true);
    this.errorMessage.set(null);

    this.refundUseCase.processDirectRefund(
      this.directPaymentIntentId.trim(),
      this.directAmount,
      this.directReason.trim(),
      this.directNotes.trim(),
      this.directAppointmentId.trim() || undefined
    ).subscribe({
      next: () => {
        this.isProcessing.set(false);
        this.showDirectRefundModal.set(false);
        this.successMessage.set(`Direct Stripe refund of $${this.directAmount.toFixed(2)} processed successfully!`);
        this.loadRefunds();
        setTimeout(() => this.successMessage.set(null), 6000);
      },
      error: err => {
        this.isProcessing.set(false);
        this.errorMessage.set('Stripe Refund Execution Error: ' + (err.error?.detail || err.message));
      }
    });
  }

  viewDetail(refund: RefundRequestItem): void {
    this.selectedRefund.set(refund);
    this.showDetailModal.set(true);
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'PROCESSED':
      case 'APPROVED':
        return 'badge-success';
      case 'PENDING':
        return 'badge-warning';
      case 'REJECTED':
      case 'FAILED':
        return 'badge-danger';
      default:
        return 'badge-secondary';
    }
  }

  getRequesterName(r: RefundRequestItem): string {
    if (r.user?.profile?.first_name || r.user?.profile?.last_name) {
      return `${r.user.profile.first_name || ''} ${r.user.profile.last_name || ''}`.trim();
    }
    if (r.user?.first_name || r.user?.last_name) {
      return `${r.user.first_name || ''} ${r.user.last_name || ''}`.trim();
    }
    if (r.user?.email) {
      return r.user.email;
    }
    return `User #${r.user_id.slice(0, 8)}`;
  }
}

