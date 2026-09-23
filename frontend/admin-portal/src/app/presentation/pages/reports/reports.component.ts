import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ReportUseCase } from '../../../core/usecases/report.usecase';
import { DialogService } from '../../../core/services/dialog.service';
import { ModerationReportItem } from '../../../core/domain/entities/report.model';

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './reports.component.html',
  styleUrl: './reports.component.css'
})
export class ReportsComponent implements OnInit {
  // Moderation Abuse Reports State
  moderationReports = signal<ModerationReportItem[]>([]);
  loading = signal<boolean>(false);
  modStatusFilter = signal<string>('');
  categoryFilter = signal<string>('');
  searchQuery = signal<string>('');
  resolvingReportId = signal<string | null>(null);
  selectedReportIds = signal<string[]>([]);
  isBulkActioning = signal<boolean>(false);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);

  // Detail Modal State
  selectedReport = signal<ModerationReportItem | null>(null);
  showDetailModal = signal<boolean>(false);
  detailActionNote = signal<string>('');
  detailActioning = signal<boolean>(false);
  copiedField = signal<string | null>(null);

  // Computed moderation lists & counts
  availableCategories = computed(() => {
    const set = new Set<string>();
    this.moderationReports().forEach(r => {
      if (r.category) set.add(r.category);
    });
    return Array.from(set);
  });

  filteredReports = computed(() => {
    let list = this.moderationReports();
    const query = this.searchQuery().toLowerCase().trim();
    const cat = this.categoryFilter();
    const status = this.modStatusFilter();

    if (status) {
      list = list.filter(r => r.status === status);
    }
    if (cat) {
      list = list.filter(r => r.category === cat);
    }
    if (query) {
      list = list.filter(r =>
        r.id.toLowerCase().includes(query) ||
        r.reporterName.toLowerCase().includes(query) ||
        r.reporterEmail.toLowerCase().includes(query) ||
        r.reportedUserName.toLowerCase().includes(query) ||
        r.reportedUserEmail.toLowerCase().includes(query) ||
        r.category.toLowerCase().includes(query) ||
        r.reason.toLowerCase().includes(query) ||
        r.description.toLowerCase().includes(query) ||
        r.resourceType.toLowerCase().includes(query)
      );
    }
    return list;
  });

  pendingReportsCount = computed(() => this.moderationReports().filter(r => r.status === 'PENDING').length);
  resolvedReportsCount = computed(() => this.moderationReports().filter(r => r.status === 'RESOLVED').length);
  dismissedReportsCount = computed(() => this.moderationReports().filter(r => r.status === 'DISMISSED').length);
  selectedReportsCount = computed(() => this.selectedReportIds().length);
  isAllReportsSelected = computed(() => {
    const list = this.filteredReports();
    return list.length > 0 && list.every(r => this.selectedReportIds().includes(r.id));
  });

  constructor(
    private reportUseCase: ReportUseCase,
    private dialog: DialogService
  ) {}

  ngOnInit(): void {
    this.fetchModerationReports();
  }

  fetchModerationReports() {
    this.loading.set(true);
    this.errorMessage.set(null);

    this.reportUseCase.listModerationReports().subscribe({
      next: (list) => {
        this.moderationReports.set(list);
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load moderation abuse reports');
        this.loading.set(false);
      }
    });
  }

  filterByStatus(status: string) {
    this.modStatusFilter.set(status);
  }

  filterByCategory(category: string) {
    this.categoryFilter.set(category);
  }

  // ─── DOSSIER MODAL ─────────────────────────────────────────────────────────

  openReportDetail(report: ModerationReportItem) {
    this.selectedReport.set(report);
    this.detailActionNote.set(report.adminNotes || '');
    this.showDetailModal.set(true);
  }

  closeDetailModal() {
    this.showDetailModal.set(false);
    this.selectedReport.set(null);
    this.detailActionNote.set('');
    this.copiedField.set(null);
  }

  resolveFromModal() {
    const r = this.selectedReport();
    if (!r) return;
    this.detailActioning.set(true);
    const note = this.detailActionNote().trim() || 'Resolved via Admin Incident Dossier';

    this.reportUseCase.resolveModerationReport(r.id, note).subscribe({
      next: () => {
        this.detailActioning.set(false);
        this.closeDetailModal();
        this.successMessage.set(`Report #${r.id.substring(0, 8)} successfully resolved.`);
        this.fetchModerationReports();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.detailActioning.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to resolve report.');
      }
    });
  }

  dismissFromModal() {
    const r = this.selectedReport();
    if (!r) return;
    this.detailActioning.set(true);
    const note = this.detailActionNote().trim() || 'Dismissed via Admin Incident Dossier';

    this.reportUseCase.dismissModerationReport(r.id, note).subscribe({
      next: () => {
        this.detailActioning.set(false);
        this.closeDetailModal();
        this.successMessage.set(`Report #${r.id.substring(0, 8)} dismissed as non-actionable.`);
        this.fetchModerationReports();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.detailActioning.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to dismiss report.');
      }
    });
  }

  copyToClipboard(text: string, fieldKey: string) {
    if (!text) return;
    navigator.clipboard.writeText(text);
    this.copiedField.set(fieldKey);
    setTimeout(() => {
      if (this.copiedField() === fieldKey) this.copiedField.set(null);
    }, 2000);
  }

  // ─── TRIAGE ACTIONS ─────────────────────────────────────────────────────────

  async resolveReport(report: ModerationReportItem) {
    const confirmed = await this.dialog.confirm(
      'Resolve Abuse Report',
      `Mark report against ${report.reportedUserName || 'reported party'} as RESOLVED? This records an administrative audit log.`,
      'Resolve Incident'
    );
    if (!confirmed) return;

    this.resolvingReportId.set(report.id);
    this.reportUseCase.resolveModerationReport(report.id, 'Resolved via Admin Portal quick triage').subscribe({
      next: () => {
        this.resolvingReportId.set(null);
        this.successMessage.set(`Report against ${report.reportedUserName} resolved.`);
        this.fetchModerationReports();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.resolvingReportId.set(null);
        this.errorMessage.set(err.error?.message || 'Failed to resolve abuse report');
      }
    });
  }

  async dismissReport(report: ModerationReportItem) {
    const confirmed = await this.dialog.dangerConfirm(
      'Dismiss Abuse Report',
      `Are you sure you want to DISMISS this report? It will be archived with no sanctions applied.`,
      'Dismiss Report'
    );
    if (!confirmed) return;

    this.resolvingReportId.set(report.id);
    this.reportUseCase.dismissModerationReport(report.id, 'Dismissed as false positive / unsubstantiated').subscribe({
      next: () => {
        this.resolvingReportId.set(null);
        this.successMessage.set(`Report dismissed.`);
        this.fetchModerationReports();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.resolvingReportId.set(null);
        this.errorMessage.set(err.error?.message || 'Failed to dismiss abuse report');
      }
    });
  }

  // ─── BATCH / SELECTION ACTIONS ──────────────────────────────────────────────

  toggleSelectAllReports() {
    if (this.isAllReportsSelected()) {
      this.selectedReportIds.set([]);
    } else {
      this.selectedReportIds.set(this.filteredReports().map(r => r.id));
    }
  }

  toggleReportSelection(id: string) {
    const curr = [...this.selectedReportIds()];
    const idx = curr.indexOf(id);
    if (idx >= 0) {
      curr.splice(idx, 1);
    } else {
      curr.push(id);
    }
    this.selectedReportIds.set(curr);
  }

  async bulkResolveReports() {
    const ids = this.selectedReportIds();
    if (!ids.length) return;

    const confirmed = await this.dialog.confirm(
      'Bulk Resolve Reports',
      `Resolve ${ids.length} selected abuse reports simultaneously?`,
      'Bulk Resolve'
    );
    if (!confirmed) return;

    this.isBulkActioning.set(true);
    let completed = 0;
    ids.forEach(id => {
      this.reportUseCase.resolveModerationReport(id, 'Bulk resolved via Admin Triage').subscribe({
        next: () => {
          completed++;
          if (completed >= ids.length) {
            this.isBulkActioning.set(false);
            this.selectedReportIds.set([]);
            this.successMessage.set(`${ids.length} reports successfully resolved.`);
            this.fetchModerationReports();
            setTimeout(() => this.successMessage.set(null), 4000);
          }
        },
        error: () => {
          completed++;
          if (completed >= ids.length) {
            this.isBulkActioning.set(false);
            this.fetchModerationReports();
          }
        }
      });
    });
  }

  async bulkDismissReports() {
    const ids = this.selectedReportIds();
    if (!ids.length) return;

    const confirmed = await this.dialog.dangerConfirm(
      'Bulk Dismiss Reports',
      `Dismiss ${ids.length} selected abuse reports as unsubstantiated?`,
      'Bulk Dismiss'
    );
    if (!confirmed) return;

    this.isBulkActioning.set(true);
    let completed = 0;
    ids.forEach(id => {
      this.reportUseCase.dismissModerationReport(id, 'Bulk dismissed via Admin Triage').subscribe({
        next: () => {
          completed++;
          if (completed >= ids.length) {
            this.isBulkActioning.set(false);
            this.selectedReportIds.set([]);
            this.successMessage.set(`${ids.length} reports dismissed.`);
            this.fetchModerationReports();
            setTimeout(() => this.successMessage.set(null), 4000);
          }
        },
        error: () => {
          completed++;
          if (completed >= ids.length) {
            this.isBulkActioning.set(false);
            this.fetchModerationReports();
          }
        }
      });
    });
  }

  exportReportsCSV() {
    const reports = this.filteredReports();
    if (!reports.length) return;

    const headers = ['Report ID', 'Reporter Name', 'Reporter Email', 'Reported Name', 'Reported Email', 'Category', 'Reason', 'Description', 'Status', 'Submitted At', 'Admin Notes'];
    const rows = reports.map(r => [
      `"${r.id}"`,
      `"${(r.reporterName || '').replace(/"/g, '""')}"`,
      `"${(r.reporterEmail || '').replace(/"/g, '""')}"`,
      `"${(r.reportedUserName || '').replace(/"/g, '""')}"`,
      `"${(r.reportedUserEmail || '').replace(/"/g, '""')}"`,
      `"${(r.category || '').replace(/"/g, '""')}"`,
      `"${(r.reason || '').replace(/"/g, '""')}"`,
      `"${(r.description || '').replace(/"/g, '""')}"`,
      `"${r.status}"`,
      `"${r.createdAt}"`,
      `"${(r.adminNotes || '').replace(/"/g, '""')}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `moderation-reports-${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    URL.revokeObjectURL(link.href);
  }

  formatDate(val: string | undefined): string {
    if (!val) return '—';
    try {
      return new Date(val).toLocaleDateString('en-US', {
        month: 'numeric',
        day: 'numeric',
        year: '2-digit'
      });
    } catch {
      return val;
    }
  }

  getCategoryBadgeClass(category: string): string {
    const c = (category || '').toLowerCase();
    if (c.includes('fraud') || c.includes('scam')) return 'badge-fraud';
    if (c.includes('safety') || c.includes('threat') || c.includes('danger')) return 'badge-safety';
    if (c.includes('abuse') || c.includes('harass')) return 'badge-abuse';
    if (c.includes('inappropriate') || c.includes('nudity') || c.includes('explicit')) return 'badge-inappropriate';
    if (c.includes('impersonat')) return 'badge-impersonate';
    if (c.includes('spam')) return 'badge-spam';
    return 'badge-category-default';
  }

  getAvatarInitial(name: string | undefined, email: string | undefined): string {
    if (name && name.trim()) return name.trim().charAt(0).toUpperCase();
    if (email && email.trim()) return email.trim().charAt(0).toUpperCase();
    return 'U';
  }

  getPartyDisplayName(name: string | undefined, email: string | undefined): string {
    if (name && name.trim()) return name.trim();
    if (email && email.trim()) return email.trim();
    return 'Anonymous User';
  }
}
