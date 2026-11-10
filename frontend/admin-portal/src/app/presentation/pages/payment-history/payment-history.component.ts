import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { StripePaymentUseCase } from '../../../core/usecases/stripe-payment.usecase';
import { StripePayment, StripePaymentStats } from '../../../core/domain/entities/stripe-payment.model';

@Component({
  selector: 'app-payment-history',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './payment-history.component.html',
  styleUrl: './payment-history.component.css'
})
export class PaymentHistoryComponent implements OnInit {
  payments = signal<StripePayment[]>([]);
  totalCount = signal<number>(0);
  stats = signal<StripePaymentStats | null>(null);

  isLoading = signal<boolean>(false);
  isSyncing = signal<boolean>(false);

  selectedStatus = signal<string>('ALL');
  selectedType = signal<string>('ALL');
  searchTerm = signal<string>('');

  currentPage = signal<number>(1);
  pageSize = signal<number>(15);

  copiedId = signal<string | null>(null);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  protected readonly Math = Math;

  totalPages = computed(() => Math.ceil(this.totalCount() / this.pageSize()) || 1);

  constructor(
    private stripePaymentUseCase: StripePaymentUseCase,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadPayments();
  }

  loadPayments(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.stripePaymentUseCase
      .listPayments(
        this.selectedStatus(),
        this.selectedType(),
        this.searchTerm(),
        this.currentPage(),
        this.pageSize()
      )
      .subscribe({
        next: res => {
          this.payments.set(res.results || []);
          this.totalCount.set(res.count ?? (res.results || []).length);
          if (res.stats) {
            this.stats.set(res.stats);
          }
          this.isLoading.set(false);
        },
        error: err => {
          this.errorMessage.set(
            err.error?.message || err.message || 'Failed to load payment history'
          );
          this.isLoading.set(false);
        }
      });
  }

  setStatus(status: string): void {
    this.selectedStatus.set(status);
    this.currentPage.set(1);
    this.loadPayments();
  }

  setType(type: string): void {
    this.selectedType.set(type);
    this.currentPage.set(1);
    this.loadPayments();
  }

  onSearchChange(): void {
    this.currentPage.set(1);
    this.loadPayments();
  }

  clearFilters(): void {
    this.selectedStatus.set('ALL');
    this.selectedType.set('ALL');
    this.searchTerm.set('');
    this.currentPage.set(1);
    this.loadPayments();
  }

  syncWithStripe(): void {
    this.isSyncing.set(true);
    this.successMessage.set(null);
    this.errorMessage.set(null);

    this.stripePaymentUseCase.syncPayments().subscribe({
      next: res => {
        this.successMessage.set(
          res.message || `Successfully synchronized Stripe payments (${res.synced} records)`
        );
        this.isSyncing.set(false);
        this.loadPayments();
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: err => {
        this.errorMessage.set(
          err.error?.message || err.message || 'Failed to synchronize with Stripe API'
        );
        this.isSyncing.set(false);
      }
    });
  }

  exportCSV(): void {
    this.isLoading.set(true);
    this.stripePaymentUseCase
      .exportPaymentsCSV(this.selectedStatus(), this.selectedType())
      .subscribe({
        next: blob => {
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = `stripe_payments_${new Date().toISOString().slice(0, 10)}.csv`;
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
          window.URL.revokeObjectURL(url);
          this.isLoading.set(false);
        },
        error: err => {
          this.errorMessage.set('Failed to export CSV: ' + (err.message || 'Unknown error'));
          this.isLoading.set(false);
        }
      });
  }

  copyId(id: string, event: Event): void {
    event.stopPropagation();
    navigator.clipboard.writeText(id).then(() => {
      this.copiedId.set(id);
      setTimeout(() => this.copiedId.set(null), 2000);
    });
  }

  goToPage(page: number): void {
    if (page >= 1 && page <= this.totalPages() && page !== this.currentPage()) {
      this.currentPage.set(page);
      this.loadPayments();
    }
  }

  viewDetail(payment: StripePayment): void {
    this.router.navigate(['/admin/payments', payment.id]);
  }

  getRequesterName(payment: StripePayment): string {
    if (payment.user_name && payment.user_name.trim()) {
      return payment.user_name;
    }
    if (payment.user?.first_name || payment.user?.last_name) {
      return `${payment.user.first_name || ''} ${payment.user.last_name || ''}`.trim();
    }
    if (payment.user?.profile?.first_name) {
      return `${payment.user.profile.first_name} ${payment.user.profile.last_name || ''}`.trim();
    }
    return payment.user_email || 'Guest Customer';
  }

  getRequesterEmail(payment: StripePayment): string {
    return payment.user_email || payment.user?.email || 'No email provided';
  }

  formatPaymentType(type?: string): string {
    switch (type) {
      case 'APPOINTMENT_ESCROW':
        return 'Booking Escrow';
      case 'BACKGROUND_CHECK':
        return 'Background Screening';
      case 'TIP':
        return 'Provider Gratuity';
      case 'SUBSCRIPTION':
        return 'Membership Plan';
      case 'GENERAL':
        return 'Direct In-App';
      default:
        return type || 'Payment';
    }
  }

  getStatusBadgeClass(status?: string): string {
    switch (status?.toLowerCase()) {
      case 'succeeded':
        return 'badge-succeeded';
      case 'refunded':
        return 'badge-refunded';
      case 'partially_refunded':
        return 'badge-partially-refunded';
      case 'processing':
      case 'requires_action':
      case 'requires_payment_method':
        return 'badge-processing';
      case 'failed':
      case 'canceled':
        return 'badge-failed';
      default:
        return 'badge-default';
    }
  }
}
