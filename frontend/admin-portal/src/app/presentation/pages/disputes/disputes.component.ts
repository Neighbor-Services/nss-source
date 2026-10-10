import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DisputeUseCase } from '../../../core/usecases/dispute.usecase';
import { GovernanceUseCase } from '../../../core/usecases/governance.usecase';
import { DisputeItem } from '../../../core/domain/entities/dispute.model';
import { ChargebackEvidencePackage } from '../../../core/domain/entities/governance.model';

export type DisputeTabFilter = 'ALL' | 'OPEN' | 'RESOLVED' | 'REJECTED';
export type DisputeSectionTab = 'CASES' | 'ARBITRATION_RULES';

@Component({
  selector: 'app-disputes',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './disputes.component.html',
  styleUrl: './disputes.component.css'
})
export class DisputesComponent implements OnInit {
  disputes = signal<DisputeItem[]>([]);
  isLoading = signal(false);
  isResolving = signal(false);
  selectedStatus = '';
  activeTab = signal<DisputeTabFilter>('ALL');
  mainSection = signal<DisputeSectionTab>('CASES');

  selectedDispute = signal<DisputeItem | null>(null);
  selectedDisputeForDetail = signal<DisputeItem | null>(null);
  showResolveModal = signal(false);
  showRejectModal = signal(false);
  showDetailModal = signal(false);

  // Chargeback Evidence Kit
  chargebackEvidence = signal<ChargebackEvidencePackage | null>(null);
  loadingChargeback = signal(false);
  showChargebackModal = signal(false);

  resolutionNotes = '';
  refundType: 'FULL' | 'PARTIAL' | 'RELEASE' = 'FULL';
  refundAmount = 0;
  rejectReason = '';

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalCount = computed(() => this.disputes().length);
  openCount = computed(() => this.disputes().filter(d => this.isOpen(d)).length);
  resolvedCount = computed(() => this.disputes().filter(d => this.isResolved(d)).length);
  rejectedCount = computed(() => this.disputes().filter(d => this.isRejected(d)).length);
  totalDisputedAmount = computed(() => this.disputes().reduce((sum, d) => sum + (d.amount || 0), 0));

  filteredDisputes = computed(() => {
    const tab = this.activeTab();
    if (tab === 'OPEN') return this.disputes().filter(d => this.isOpen(d));
    if (tab === 'RESOLVED') return this.disputes().filter(d => this.isResolved(d));
    if (tab === 'REJECTED') return this.disputes().filter(d => this.isRejected(d));
    return this.disputes();
  });

  isOpen(d: DisputeItem): boolean {
    const s = (d.status || '').toUpperCase();
    return s === 'OPEN' || s === 'UNDER_REVIEW' || s === 'ESCALATED' || s === 'PENDING';
  }

  isResolved(d: DisputeItem): boolean {
    return (d.status || '').toUpperCase() === 'RESOLVED';
  }

  isRejected(d: DisputeItem): boolean {
    const s = (d.status || '').toUpperCase();
    return s === 'REJECTED' || s === 'DISMISSED';
  }

  getEvidenceTimeRemaining(createdAt: string): { text: string; isExpired: boolean } {
    if (!createdAt) return { text: 'Awaiting submission', isExpired: false };
    const created = new Date(createdAt).getTime();
    const deadline = created + (48 * 60 * 60 * 1000);
    const now = Date.now();
    const diffMs = deadline - now;
    if (diffMs <= 0) {
      return { text: '48h Evidence Window Closed', isExpired: true };
    }
    const hours = Math.floor(diffMs / (1000 * 60 * 60));
    return { text: `${hours}h remaining for evidence`, isExpired: false };
  }

  constructor(
    private disputeUC: DisputeUseCase,
    private governanceUC: GovernanceUseCase
  ) {}

  ngOnInit() {
    this.loadDisputes();
  }

  openChargebackEvidence(d: DisputeItem) {
    this.loadingChargeback.set(true);
    this.showChargebackModal.set(true);
    this.chargebackEvidence.set(null);

    this.governanceUC.compileChargebackEvidence(d.id).subscribe({
      next: (pkg) => {
        this.chargebackEvidence.set(pkg);
        this.loadingChargeback.set(false);
      },
      error: () => {
        this.loadingChargeback.set(false);
      }
    });
  }

  closeChargebackModal() {
    this.showChargebackModal.set(false);
    this.chargebackEvidence.set(null);
  }

  loadDisputes() {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.disputeUC.listDisputes(this.selectedStatus || undefined).subscribe({
      next: (items) => {
        this.disputes.set(items);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to load disputes');
      }
    });
  }

  openDetail(d: DisputeItem) {
    this.selectedDisputeForDetail.set(d);
    this.showDetailModal.set(true);
  }

  closeDetail() {
    this.showDetailModal.set(false);
    this.selectedDisputeForDetail.set(null);
  }

  openResolve(d: DisputeItem) {
    this.selectedDispute.set(d);
    this.refundType = 'FULL';
    this.refundAmount = d.amount;
    this.resolutionNotes = `Escrow dispute #${d.appointmentId} resolved by platform arbitration. Refund granted to seeker.`;
    this.showResolveModal.set(true);
  }

  onRefundTypeChange() {
    const d = this.selectedDispute();
    if (!d) return;
    if (this.refundType === 'FULL') {
      this.refundAmount = d.amount;
    } else if (this.refundType === 'RELEASE') {
      this.refundAmount = 0;
    }
  }

  closeResolve() {
    this.showResolveModal.set(false);
    this.selectedDispute.set(null);
  }

  submitResolve() {
    const d = this.selectedDispute();
    if (!d) return;

    this.isResolving.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const notes = this.resolutionNotes || (this.refundAmount > 0 
      ? `Refund of $${this.refundAmount.toFixed(2)} granted to customer.` 
      : 'Dispute resolved in favor of provider. Escrow released.');

    this.disputeUC.resolveDispute(d.id, notes, this.refundAmount).subscribe({
      next: () => {
        this.isResolving.set(false);
        this.showResolveModal.set(false);
        if (this.showDetailModal()) this.closeDetail();
        this.successMessage.set(`Dispute #${d.appointmentId} resolved successfully.`);
        this.loadDisputes();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isResolving.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to resolve dispute');
      }
    });
  }

  openReject(d: DisputeItem) {
    this.selectedDispute.set(d);
    this.rejectReason = 'Evidence provided is insufficient to justify claim. Dispute dismissed and funds settled.';
    this.showRejectModal.set(true);
  }

  closeReject() {
    this.showRejectModal.set(false);
    this.selectedDispute.set(null);
  }

  submitReject() {
    const d = this.selectedDispute();
    if (!d) return;

    this.isResolving.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.disputeUC.rejectDispute(d.id, this.rejectReason).subscribe({
      next: () => {
        this.isResolving.set(false);
        this.showRejectModal.set(false);
        if (this.showDetailModal()) this.closeDetail();
        this.successMessage.set(`Dispute #${d.appointmentId} dismissed.`);
        this.loadDisputes();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isResolving.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to dismiss dispute');
      }
    });
  }

  exportCSV(): void {
    const data = this.disputes();
    if (!data.length) return;

    const headers = ['ID', 'Appointment Reference', 'Seeker Name', 'Provider Name', 'Amount', 'Reason', 'Status', 'Resolution Notes', 'Created At'];
    const rows = data.map(d => [
      `"${d.id}"`,
      `"${d.appointmentId || ''}"`,
      `"${d.seekerName || ''}"`,
      `"${d.providerName || ''}"`,
      d.amount || 0,
      `"${d.reason || ''}"`,
      `"${d.status}"`,
      `"${d.resolutionNotes || ''}"`,
      `"${d.createdAt || ''}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `disputes_export_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }

  // ─── MEDIATION STUDIO ───────────────────────────────────────────────────────

  showMediationStudio = signal(false);
  mediationDispute = signal<DisputeItem | null>(null);
  mediationResolutionType = signal<'FULL_REFUND' | 'PARTIAL_SPLIT' | 'RELEASE_TO_PROVIDER' | 'ISSUE_CREDIT'>('PARTIAL_SPLIT');
  mediationRefundPct = signal(50);
  mediationCustomVoucher = signal(0);
  mediationNotes = signal('');

  seekerRefundAmount = computed(() => {
    const d = this.mediationDispute();
    if (!d) return 0;
    const total = d.amount || 0;
    const type = this.mediationResolutionType();
    if (type === 'FULL_REFUND') return total;
    if (type === 'RELEASE_TO_PROVIDER' || type === 'ISSUE_CREDIT') return 0;
    return Math.round((total * (this.mediationRefundPct() / 100)) * 100) / 100;
  });

  providerPayoutAmount = computed(() => {
    const d = this.mediationDispute();
    if (!d) return 0;
    const total = d.amount || 0;
    const type = this.mediationResolutionType();
    if (type === 'RELEASE_TO_PROVIDER') return total;
    if (type === 'FULL_REFUND') return 0;
    if (type === 'ISSUE_CREDIT') return total; // Provider gets full, platform issues credit
    return Math.round((total - this.seekerRefundAmount()) * 100) / 100;
  });

  platformVoucherCredit = computed(() => {
    const type = this.mediationResolutionType();
    if (type === 'ISSUE_CREDIT') {
      const d = this.mediationDispute();
      return (d?.amount || 0) * 0.5; // 50% courtesy platform voucher
    }
    return this.mediationCustomVoucher();
  });

  openMediationStudio(d: DisputeItem) {
    this.mediationDispute.set(d);
    this.mediationResolutionType.set('PARTIAL_SPLIT');
    this.mediationRefundPct.set(50);
    this.mediationCustomVoucher.set(0);
    this.mediationNotes.set(`Mediation agreed between ${d.seekerName} and ${d.providerName}`);
    this.showMediationStudio.set(true);
  }

  closeMediationStudio() {
    this.showMediationStudio.set(false);
    this.mediationDispute.set(null);
  }

  submitMediationResolution() {
    const d = this.mediationDispute();
    if (!d) return;

    this.isResolving.set(true);
    this.errorMessage.set(null);

    const payload = {
      resolution_type: this.mediationResolutionType(),
      refund_amount: this.seekerRefundAmount(),
      provider_payout: this.providerPayoutAmount(),
      credit_voucher: this.platformVoucherCredit(),
      notes: this.mediationNotes()
    };

    this.disputeUC.mediateDispute(d.id, payload).subscribe({
      next: () => {
        this.isResolving.set(false);
        this.showMediationStudio.set(false);
        if (this.showDetailModal()) this.closeDetail();
        this.successMessage.set(`Mediation Studio agreement executed for Dispute #${d.appointmentId}!`);
        this.loadDisputes();
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.isResolving.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to execute mediation resolution.');
      }
    });
  }
}


