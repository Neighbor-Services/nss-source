import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { VerificationUseCase } from '../../../core/usecases/verification.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { VerificationItem } from '../../../core/domain/entities/verification.model';

export type VerificationTabFilter = 'ALL' | 'PENDING' | 'APPROVED' | 'REJECTED';
export type VerificationSectionTab = 'QUEUE' | 'OCR_SECURITY' | 'STANDARDS';

@Component({
  selector: 'app-verifications',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './verifications.component.html',
  styleUrl: './verifications.component.css'
})
export class VerificationsComponent implements OnInit {
  verifications = signal<VerificationItem[]>([]);
  isLoading = signal(false);
  isBatchProcessing = signal(false);
  actionNotice = signal<string | null>(null);
  
  // Section & Filter
  mainSection = signal<VerificationSectionTab>('QUEUE');
  selectedStatus = '';
  activeTab = signal<VerificationTabFilter>('ALL');
  searchQuery = '';
  viewMode = signal<'grid' | 'table'>('grid');
  
  selectedIds = signal<string[]>([]);
  previewImage = signal<string | null>(null);

  // Computed Queue Stats
  totalItems = computed(() => this.verifications().length);
  pendingCount = computed(() => this.verifications().filter(v => v.status === 'pending').length);
  approvedCount = computed(() => this.verifications().filter(v => v.status === 'approved').length);
  rejectedCount = computed(() => this.verifications().filter(v => v.status === 'rejected').length);

  filteredVerifications = computed(() => {
    let list = this.verifications();
    const tab = this.activeTab();
    if (tab === 'PENDING') list = list.filter(v => v.status === 'pending');
    else if (tab === 'APPROVED') list = list.filter(v => v.status === 'approved');
    else if (tab === 'REJECTED') list = list.filter(v => v.status === 'rejected');

    if (this.searchQuery.trim()) {
      const q = this.searchQuery.toLowerCase();
      list = list.filter(v => 
        v.userName.toLowerCase().includes(q) || 
        v.userEmail.toLowerCase().includes(q) || 
        (v.documentType && v.documentType.toLowerCase().includes(q))
      );
    }
    return list;
  });

  constructor(
    private verificationUC: VerificationUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit() {
    this.loadVerifications();
  }

  setTab(tab: VerificationTabFilter) {
    this.activeTab.set(tab);
    if (tab === 'PENDING') this.selectedStatus = 'pending';
    else if (tab === 'APPROVED') this.selectedStatus = 'approved';
    else if (tab === 'REJECTED') this.selectedStatus = 'rejected';
    else this.selectedStatus = '';
  }

  loadVerifications() {
    this.isLoading.set(true);
    this.selectedIds.set([]);
    this.verificationUC.listVerifications().subscribe({
      next: (items) => {
        this.verifications.set(items);
        this.isLoading.set(false);
      },
      error: () => {
        this.isLoading.set(false);
      }
    });
  }


  async approve(id: string) {
    const confirmed = await this.dialog.confirm({
      title: 'Approve Identity Verification',
      message: 'Are you sure you want to approve this provider identity document? The provider will immediately gain verified badge credentials.',
      confirmText: 'Approve Verification',
      cancelText: 'Cancel'
    });
    if (!confirmed) return;

    this.verificationUC.approveVerification(id).subscribe({
      next: () => {
        this.actionNotice.set('Provider identity verified and approved.');
        this.loadVerifications();
        setTimeout(() => this.actionNotice.set(null), 4000);
      }
    });
  }

  async reject(id: string) {
    const reason = await this.dialog.prompt({
      title: 'Reject Identity Verification',
      message: 'Please specify the rejection reason to notify the provider (e.g. Blurry photo, Expired document, Name mismatch):',
      placeholder: 'e.g. Expired government ID or illegible scan',
      confirmText: 'Confirm Rejection',
      cancelText: 'Cancel'
    });

    if (reason && reason.trim()) {
      this.verificationUC.rejectVerification(id, reason.trim()).subscribe({
        next: () => {
          this.actionNotice.set('Provider verification rejected with reason.');
          this.loadVerifications();
          setTimeout(() => this.actionNotice.set(null), 4000);
        }
      });
    }
  }

  toggleSelect(id: string) {
    const ids = [...this.selectedIds()];
    const index = ids.indexOf(id);
    if (index > -1) {
      ids.splice(index, 1);
    } else {
      ids.push(id);
    }
    this.selectedIds.set(ids);
  }

  selectAll() {
    if (this.selectedIds().length === this.verifications().length) {
      this.selectedIds.set([]);
    } else {
      this.selectedIds.set(this.verifications().map(v => v.id));
    }
  }

  async batchProcess(action: 'APPROVE' | 'REJECT') {
    const ids = this.selectedIds();
    if (ids.length === 0) return;

    let notes = '';
    if (action === 'REJECT') {
      const reason = await this.dialog.prompt({
        title: 'Batch Reject Verifications',
        message: `Please specify the rejection reason for all ${ids.length} selected identity submissions:`,
        placeholder: 'Documentation does not meet verification requirements',
        defaultValue: 'Documentation does not meet verification requirements',
        confirmText: 'Reject Selected',
        cancelText: 'Cancel'
      });
      if (!reason) return;
      notes = reason;
    } else {
      const confirmed = await this.dialog.confirm({
        title: 'Batch Approve Verifications',
        message: `Are you sure you want to approve all ${ids.length} selected provider identity verifications?`,
        confirmText: 'Approve All Selected',
        cancelText: 'Cancel'
      });
      if (!confirmed) return;
    }

    this.isBatchProcessing.set(true);
    this.verificationUC.batchVerifications(ids, action, notes).subscribe({
      next: (res) => {
        this.isBatchProcessing.set(false);
        this.actionNotice.set(`Batch ${action.toLowerCase()} processed for ${res.count} provider identity submissions!`);
        this.loadVerifications();
        setTimeout(() => this.actionNotice.set(null), 4000);
      },
      error: () => {
        this.isBatchProcessing.set(false);
      }
    });
  }

  openPreview(url: string) {
    this.previewImage.set(url);
  }

  closePreview() {
    this.previewImage.set(null);
  }

  exportCSV(): void {
    const data = this.verifications();
    if (!data.length) return;

    const headers = ['ID', 'Provider Name', 'Provider Email', 'Document Type', 'Status', 'Rejection Reason', 'Submitted At'];
    const rows = data.map(v => [
      `"${v.id}"`,
      `"${v.userName || ''}"`,
      `"${v.userEmail || ''}"`,
      `"${v.documentType || 'ID Document'}"`,
      `"${v.status}"`,
      `"${v.rejectionReason || ''}"`,
      `"${v.submittedAt || ''}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `verifications_export_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }
}

