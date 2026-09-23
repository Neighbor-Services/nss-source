import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DisputeUseCase } from '../../../core/usecases/dispute.usecase';
import { DisputeItem } from '../../../core/domain/entities/dispute.model';

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
  showResolveModal = signal(false);
  showRejectModal = signal(false);

  resolutionNotes = '';
  refundAmount = 0;
  rejectReason = '';

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalCount = computed(() => this.disputes().length);
  openCount = computed(() => this.disputes().filter(d => d.status === 'open').length);
  resolvedCount = computed(() => this.disputes().filter(d => d.status === 'resolved').length);
  rejectedCount = computed(() => this.disputes().filter(d => d.status === 'rejected').length);
  totalDisputedAmount = computed(() => this.disputes().reduce((sum, d) => sum + (d.amount || 0), 0));

  filteredDisputes = computed(() => {
    const tab = this.activeTab();
    if (tab === 'OPEN') return this.disputes().filter(d => d.status === 'open');
    if (tab === 'RESOLVED') return this.disputes().filter(d => d.status === 'resolved');
    if (tab === 'REJECTED') return this.disputes().filter(d => d.status === 'rejected');
    return this.disputes();
  });


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

  constructor(private disputeUC: DisputeUseCase) {}

  ngOnInit() {
    this.loadDisputes();
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

  openResolve(d: DisputeItem) {
    this.selectedDispute.set(d);
    this.refundAmount = d.amount;
    this.resolutionNotes = 'Full escrow refund approved upon investigation';
    this.showResolveModal.set(true);
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

    this.disputeUC.resolveDispute(d.id, this.resolutionNotes, this.refundAmount).subscribe({
      next: () => {
        this.isResolving.set(false);
        this.showResolveModal.set(false);
        this.successMessage.set(`Dispute #${d.appointmentId} resolved with $${this.refundAmount.toFixed(2)} refund.`);
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
    this.rejectReason = 'Evidence provided is insufficient to justify escrow refund.';
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
        this.successMessage.set(`Dispute #${d.appointmentId} rejected/dismissed.`);
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
}

