import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BackgroundCheckUseCase } from '../../../core/usecases/background-check.usecase';
import { BackgroundCheckItem } from '../../../core/domain/entities/background-check.model';

export type CheckrTabFilter = 'ALL' | 'CLEAR' | 'PENDING' | 'CONSIDER' | 'OVERRIDDEN';
export type CheckrSectionTab = 'SCREENINGS' | 'WEBHOOKS' | 'FCRA_PROTOCOLS';

@Component({
  selector: 'app-background-checks',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './background-checks.component.html',
  styleUrl: './background-checks.component.css'
})
export class BackgroundChecksComponent implements OnInit {
  checks = signal<BackgroundCheckItem[]>([]);
  isLoading = signal(false);
  isSubmitting = signal(false);
  searchFilter = '';
  activeTab = signal<CheckrTabFilter>('ALL');
  mainSection = signal<CheckrSectionTab>('SCREENINGS');

  selectedCheck = signal<BackgroundCheckItem | null>(null);
  overrideStatusVal = 'clear';
  overrideReason = '';
  showOverrideModal = signal(false);

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalCount = computed(() => this.checks().length);
  clearCount = computed(() => this.checks().filter(c => c.status === 'clear').length);
  pendingCount = computed(() => this.checks().filter(c => c.status === 'pending').length);
  considerCount = computed(() => this.checks().filter(c => c.status === 'consider' || c.status === 'suspended').length);
  overrideCount = computed(() => this.checks().filter(c => c.manualOverride).length);

  filteredChecks = computed(() => {
    let list = this.checks();
    const tab = this.activeTab();
    if (tab === 'CLEAR') list = list.filter(c => c.status === 'clear');
    else if (tab === 'PENDING') list = list.filter(c => c.status === 'pending');
    else if (tab === 'CONSIDER') list = list.filter(c => c.status === 'consider' || c.status === 'suspended');
    else if (tab === 'OVERRIDDEN') list = list.filter(c => c.manualOverride);

    const q = this.searchFilter.toLowerCase().trim();
    if (q) {
      list = list.filter(c =>
        c.userName.toLowerCase().includes(q) ||
        c.userEmail.toLowerCase().includes(q) ||
        c.checkrCandidateId.toLowerCase().includes(q) ||
        c.checkrReportId.toLowerCase().includes(q)
      );
    }
    return list;
  });

  constructor(private bgUC: BackgroundCheckUseCase) {}

  ngOnInit() {
    this.loadChecks();
  }

  loadChecks() {
    this.isLoading.set(true);
    this.errorMessage.set(null);
    this.bgUC.listBackgroundChecks().subscribe({
      next: (items) => {
        this.checks.set(items);
        this.isLoading.set(false);
      },
      error: (err: any) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load background screening reports');
      }
    });
  }

  openOverride(item: BackgroundCheckItem) {
    this.selectedCheck.set(item);
    this.overrideStatusVal = item.status === 'clear' ? 'consider' : 'clear';
    this.overrideReason = 'Manual compliance review completed against certified identity documentation';
    this.showOverrideModal.set(true);
  }

  closeOverride() {
    this.showOverrideModal.set(false);
    this.selectedCheck.set(null);
  }

  submitOverride() {
    const item = this.selectedCheck();
    if (!item) return;

    if (!this.overrideReason.trim()) {
      this.errorMessage.set('Please provide a valid compliance override justification');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.bgUC.overrideBackgroundCheck(item.id, this.overrideStatusVal, this.overrideReason).subscribe({
      next: () => {
        this.isSubmitting.set(false);
        this.showOverrideModal.set(false);
        this.successMessage.set(`Background check status for ${item.userName} updated to ${this.overrideStatusVal.toUpperCase()}.`);
        this.loadChecks();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err: any) => {
        this.isSubmitting.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to override background screening');
      }
    });
  }

  exportCSV(): void {
    const data = this.checks();
    if (!data.length) return;

    const headers = ['ID', 'Provider Name', 'Provider Email', 'Candidate ID', 'Report ID', 'Package', 'Status', 'Manual Override', 'Completed At'];
    const rows = data.map(c => [
      `"${c.id}"`,
      `"${c.userName || ''}"`,
      `"${c.userEmail || ''}"`,
      `"${c.checkrCandidateId || ''}"`,
      `"${c.checkrReportId || ''}"`,
      `"${c.package || 'tasker_standard'}"`,
      `"${c.status}"`,
      c.manualOverride ? 'Yes' : 'No',
      `"${c.completedAt || ''}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `background_checks_export_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }
}

