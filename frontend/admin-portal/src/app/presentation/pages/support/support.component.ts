import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { SupportUseCase } from '../../../core/usecases/support.usecase';
import { ContactMessage, ResolutionReport } from '../../../core/domain/entities/support.model';

@Component({
  selector: 'app-support',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './support.component.html',
  styleUrl: './support.component.css'
})
export class SupportComponent implements OnInit {
  activeTab = signal<'MESSAGES' | 'RESOLUTIONS'>('MESSAGES');

  // Contact Messages State
  messages = signal<ContactMessage[]>([]);
  totalMessages = signal(0);
  messageSearch = signal('');
  messageResolvedFilter = signal<'ALL' | 'RESOLVED' | 'UNRESOLVED'>('ALL');
  selectedMessage = signal<ContactMessage | null>(null);

  // Resolution Reports State
  resolutions = signal<ResolutionReport[]>([]);
  totalResolutions = signal(0);
  resolutionSearch = signal('');
  resolutionReviewedFilter = signal<'ALL' | 'REVIEWED' | 'UNREVIEWED'>('ALL');
  selectedResolution = signal<ResolutionReport | null>(null);

  // Pagination & Loading
  currentPage = signal(1);
  pageSize = signal(15);
  isLoading = signal(false);
  isActioning = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  pendingMessagesCount = computed(() => this.messages().filter(m => !m.is_resolved).length);
  openResolutionsCount = computed(() => this.resolutions().filter(r => !r.is_reviewed).length);

  constructor(
    private supportUC: SupportUseCase,
    private route: ActivatedRoute,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      const tabParam = params['tab'];
      if (tabParam === 'RESOLUTIONS' || tabParam === 'resolutions') {
        this.activeTab.set('RESOLUTIONS');
      } else if (tabParam === 'MESSAGES' || tabParam === 'messages') {
        this.activeTab.set('MESSAGES');
      }
      this.loadData();
    });
  }

  setTab(tab: 'MESSAGES' | 'RESOLUTIONS'): void {
    this.activeTab.set(tab);
    this.currentPage.set(1);
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { tab },
      queryParamsHandling: 'merge'
    });
    this.loadData();
  }

  loadData(): void {
    if (this.activeTab() === 'MESSAGES') {
      this.loadMessages();
    } else {
      this.loadResolutions();
    }
  }

  loadMessages(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    let isResolved: boolean | undefined = undefined;
    if (this.messageResolvedFilter() === 'RESOLVED') isResolved = true;
    if (this.messageResolvedFilter() === 'UNRESOLVED') isResolved = false;

    this.supportUC.listMessages({
      is_resolved: isResolved,
      search: this.messageSearch() || undefined,
      page: this.currentPage(),
      pageSize: this.pageSize()
    }).subscribe({
      next: (res) => {
        this.messages.set(res.results || []);
        this.totalMessages.set(res.count || 0);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load support messages.');
      }
    });
  }

  loadResolutions(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    let isReviewed: boolean | undefined = undefined;
    if (this.resolutionReviewedFilter() === 'REVIEWED') isReviewed = true;
    if (this.resolutionReviewedFilter() === 'UNREVIEWED') isReviewed = false;

    this.supportUC.listResolutions({
      is_reviewed: isReviewed,
      search: this.resolutionSearch() || undefined,
      page: this.currentPage(),
      pageSize: this.pageSize()
    }).subscribe({
      next: (res) => {
        this.resolutions.set(res.results || []);
        this.totalResolutions.set(res.count || 0);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load incident resolution reports.');
      }
    });
  }

  // Contact Message Actions
  viewMessage(msg: ContactMessage): void {
    this.selectedMessage.set(msg);
  }

  closeMessageModal(): void {
    this.selectedMessage.set(null);
  }

  toggleMessageResolved(msg: ContactMessage): void {
    const newStatus = !msg.is_resolved;
    this.isActioning.set(true);

    this.supportUC.toggleMessageResolved(msg.id, newStatus).subscribe({
      next: () => {
        this.isActioning.set(false);
        msg.is_resolved = newStatus;
        if (this.selectedMessage()?.id === msg.id) {
          this.selectedMessage.set({ ...msg, is_resolved: newStatus });
        }
        this.successMessage.set(`Message status updated to ${newStatus ? 'Resolved' : 'Pending'}.`);
        setTimeout(() => this.successMessage.set(null), 3500);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update message status.');
      }
    });
  }

  deleteMessage(msg: ContactMessage): void {
    if (!confirm(`Delete message from ${msg.first_name} ${msg.last_name}?`)) return;

    this.supportUC.deleteMessage(msg.id).subscribe({
      next: () => {
        if (this.selectedMessage()?.id === msg.id) {
          this.selectedMessage.set(null);
        }
        this.successMessage.set('Contact inquiry deleted.');
        setTimeout(() => this.successMessage.set(null), 3500);
        this.loadMessages();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete message.');
      }
    });
  }

  // Resolution Report Actions
  viewResolution(res: ResolutionReport): void {
    this.selectedResolution.set(res);
  }

  closeResolutionModal(): void {
    this.selectedResolution.set(null);
  }

  toggleResolutionReviewed(report: ResolutionReport): void {
    const newStatus = !report.is_reviewed;
    this.isActioning.set(true);

    this.supportUC.toggleResolutionReviewed(report.id, newStatus).subscribe({
      next: () => {
        this.isActioning.set(false);
        report.is_reviewed = newStatus;
        if (this.selectedResolution()?.id === report.id) {
          this.selectedResolution.set({ ...report, is_reviewed: newStatus });
        }
        this.successMessage.set(`Resolution report marked as ${newStatus ? 'Reviewed' : 'Under Review'}.`);
        setTimeout(() => this.successMessage.set(null), 3500);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update report status.');
      }
    });
  }

  deleteResolution(report: ResolutionReport): void {
    if (!confirm(`Delete resolution report for booking ${report.booking_ref || report.id}?`)) return;

    this.supportUC.deleteResolution(report.id).subscribe({
      next: () => {
        if (this.selectedResolution()?.id === report.id) {
          this.selectedResolution.set(null);
        }
        this.successMessage.set('Incident resolution report deleted.');
        setTimeout(() => this.successMessage.set(null), 3500);
        this.loadResolutions();
      },
      error: (err) => {
        this.errorMessage.set(err?.error?.message || 'Failed to delete resolution report.');
      }
    });
  }

  getInquiryBadgeClass(type?: string): string {
    const t = (type || '').toLowerCase();
    if (t.includes('bill') || t.includes('pay')) return 'badge-purple';
    if (t.includes('account') || t.includes('auth')) return 'badge-primary';
    if (t.includes('safety') || t.includes('urgent') || t.includes('damage')) return 'badge-danger';
    if (t.includes('provider') || t.includes('service')) return 'badge-success';
    if (t.includes('feed') || t.includes('suggest')) return 'badge-warning';
    return 'badge-info';
  }

  getAvatarInitials(first?: string, last?: string): string {
    const f = (first || 'U').trim().charAt(0).toUpperCase();
    const l = (last || '').trim().charAt(0).toUpperCase();
    return l ? `${f}${l}` : f;
  }
}
