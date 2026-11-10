import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { StripePaymentUseCase } from '../../../core/usecases/stripe-payment.usecase';
import { StripePaymentDetail, StripePayment } from '../../../core/domain/entities/stripe-payment.model';

@Component({
  selector: 'app-payment-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './payment-detail.component.html',
  styleUrl: './payment-detail.component.css'
})
export class PaymentDetailComponent implements OnInit {
  paymentId = signal<string>('');
  detail = signal<StripePaymentDetail | null>(null);

  isLoading = signal<boolean>(true);
  isProcessingRefund = signal<boolean>(false);

  // Refund Modal State
  showRefundModal = signal<boolean>(false);
  refundMode = signal<'full' | 'partial'>('full');
  refundAmount = 0;
  refundReason = 'Requested by customer';
  refundNotes = '';

  copiedKey = signal<string | null>(null);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  payment = computed<StripePayment | null>(() => this.detail()?.payment ?? null);

  netAmount = computed<number>(() => {
    const p = this.payment();
    if (!p) return 0;
    return Math.max(0, (p.amount || 0) - (p.amount_refunded || 0));
  });

  isRefundable = computed<boolean>(() => {
    return this.detail()?.is_refundable ?? false;
  });

  refundableAmount = computed<number>(() => {
    return this.detail()?.refundable_amount ?? 0;
  });

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private stripePaymentUseCase: StripePaymentUseCase
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.paymentId.set(id);
      this.loadDetail(id);
    } else {
      this.router.navigate(['/admin/payments']);
    }
  }

  loadDetail(id: string): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.stripePaymentUseCase.getPaymentDetail(id).subscribe({
      next: data => {
        this.detail.set(data);
        this.isLoading.set(false);
      },
      error: err => {
        this.errorMessage.set(
          err.error?.message || err.message || 'Failed to load payment detail'
        );
        this.isLoading.set(false);
      }
    });
  }

  openRefundModal(): void {
    const maxRefund = this.refundableAmount();
    this.refundMode.set('full');
    this.refundAmount = maxRefund;
    this.refundReason = 'Requested by customer';
    this.refundNotes = '';
    this.showRefundModal.set(true);
  }

  closeRefundModal(): void {
    this.showRefundModal.set(false);
  }

  setRefundMode(mode: 'full' | 'partial'): void {
    this.refundMode.set(mode);
    if (mode === 'full') {
      this.refundAmount = this.refundableAmount();
    } else {
      this.refundAmount = Number((this.refundableAmount() / 2).toFixed(2));
    }
  }

  submitRefund(): void {
    const max = this.refundableAmount();
    let amt = this.refundAmount;

    if (this.refundMode() === 'full') {
      amt = max;
    }

    if (amt <= 0 || amt > max + 0.001) {
      alert(`Refund amount must be between $0.01 and $${max.toFixed(2)}`);
      return;
    }

    this.isProcessingRefund.set(true);
    this.errorMessage.set(null);

    this.stripePaymentUseCase
      .refundPayment(this.paymentId(), {
        amount: amt,
        reason: this.refundReason,
        notes: this.refundNotes
      })
      .subscribe({
        next: res => {
          this.isProcessingRefund.set(false);
          this.showRefundModal.set(false);
          this.successMessage.set(
            res.message || `Successfully processed Stripe refund of $${amt.toFixed(2)}`
          );
          if (res.detail) {
            this.detail.set(res.detail);
          } else {
            this.loadDetail(this.paymentId());
          }
          setTimeout(() => this.successMessage.set(null), 6000);
        },
        error: err => {
          this.isProcessingRefund.set(false);
          alert('Refund error: ' + (err.error?.message || err.message || 'Failed to execute refund'));
        }
      });
  }

  copyText(text: string, key: string): void {
    navigator.clipboard.writeText(text).then(() => {
      this.copiedKey.set(key);
      setTimeout(() => this.copiedKey.set(null), 2000);
    });
  }

  formatPaymentType(type?: string): string {
    switch (type) {
      case 'APPOINTMENT_ESCROW':
        return 'Booking Escrow Funding';
      case 'BACKGROUND_CHECK':
        return 'Provider Background Screening';
      case 'TIP':
        return 'Provider Tip';
      case 'SUBSCRIPTION':
        return 'Provider Subscription';
      case 'GENERAL':
        return 'Direct In-App Payment';
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

