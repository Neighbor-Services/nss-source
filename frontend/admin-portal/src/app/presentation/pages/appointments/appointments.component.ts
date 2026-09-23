import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { AppointmentUseCase } from '../../../core/usecases/appointment.usecase';
import { AdminAppointment } from '../../../core/domain/entities/appointment.model';

export type AppointmentStatusFilter = 'ALL' | 'SCHEDULED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED' | 'DISPUTED';

@Component({
  selector: 'app-appointments',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './appointments.component.html',
  styleUrl: './appointments.component.css'
})
export class AppointmentsComponent implements OnInit {
  appointments = signal<AdminAppointment[]>([]);
  totalCount = signal(0);
  
  // Search & Filter state
  searchQuery = '';
  selectedPaymentMode = '';
  activeStatus = signal<AppointmentStatusFilter>('ALL');
  
  // Pagination
  currentPage = signal(1);
  pageSize = signal(15);
  
  // Modal & Selection state
  selectedAppointment = signal<AdminAppointment | null>(null);
  overrideStatusValue = '';
  adminOverrideNotes = '';
  
  // Async states
  isLoading = signal(false);
  isActioning = signal(false);
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalAppointments = computed(() => this.totalCount() || this.appointments().length);
  grossVolume = computed(() => {
    return this.appointments().reduce((acc, a) => acc + (a.total_price || 0), 0);
  });
  scheduledCount = computed(() => this.appointments().filter(a => a.status === 'SCHEDULED').length);
  inProgressCount = computed(() => this.appointments().filter(a => a.status === 'IN_PROGRESS').length);
  completedCount = computed(() => this.appointments().filter(a => a.status === 'COMPLETED').length);
  activeCount = computed(() => this.appointments().filter(a => a.status === 'SCHEDULED' || a.status === 'IN_PROGRESS').length);
  disputedCount = computed(() => this.appointments().filter(a => a.status === 'DISPUTED').length);
  cancelledCount = computed(() => this.appointments().filter(a => a.status === 'CANCELLED').length);
  totalPages = computed(() => Math.max(1, Math.ceil(this.totalCount() / this.pageSize())));

  // Client-side filtered list if payment mode is selected or status refined
  filteredAppointments = computed(() => {
    let list = this.appointments();
    if (this.activeStatus() !== 'ALL') {
      list = list.filter(a => a.status === this.activeStatus());
    }
    if (this.selectedPaymentMode) {
      list = list.filter(a => a.payment_mode === this.selectedPaymentMode);
    }
    return list;
  });

  constructor(private appointmentUC: AppointmentUseCase) {}

  ngOnInit(): void {
    this.loadAppointments();
  }

  loadAppointments(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.appointmentUC.listAppointments({
      status: this.activeStatus() === 'ALL' ? undefined : this.activeStatus(),
      search: this.searchQuery.trim() || undefined,
      page: this.currentPage(),
      pageSize: this.pageSize()
    }).subscribe({
      next: (res) => {
        this.appointments.set(res.results || []);
        this.totalCount.set(res.count || (res.results ? res.results.length : 0));
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to load appointments ledger.');
      }
    });
  }

  setStatusFilter(status: AppointmentStatusFilter): void {
    this.activeStatus.set(status);
    this.currentPage.set(1);
    this.loadAppointments();
  }

  resetFilters(): void {
    this.searchQuery = '';
    this.selectedPaymentMode = '';
    this.activeStatus.set('ALL');
    this.currentPage.set(1);
    this.loadAppointments();
  }

  openDetails(apt: AdminAppointment): void {
    this.selectedAppointment.set(apt);
    this.overrideStatusValue = apt.status;
    this.adminOverrideNotes = '';
  }

  closeDetails(): void {
    this.selectedAppointment.set(null);
  }

  saveStatusOverride(): void {
    const apt = this.selectedAppointment();
    if (!apt || !this.overrideStatusValue) return;

    this.isActioning.set(true);
    this.appointmentUC.updateAppointmentStatus(apt.id, this.overrideStatusValue).subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Appointment #${apt.id.slice(0, 8)} status updated to ${this.overrideStatusValue}.`);
        this.closeDetails();
        this.loadAppointments();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to update appointment status.');
      }
    });
  }

  forceCancel(apt: AdminAppointment): void {
    if (!confirm(`Are you sure you want to force-cancel appointment #${apt.id.slice(0, 8)}?`)) return;

    this.isActioning.set(true);
    this.appointmentUC.updateAppointmentStatus(apt.id, 'CANCELLED').subscribe({
      next: () => {
        this.isActioning.set(false);
        this.successMessage.set(`Appointment #${apt.id.slice(0, 8)} has been cancelled.`);
        this.loadAppointments();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.isActioning.set(false);
        this.errorMessage.set(err?.error?.message || 'Failed to cancel appointment.');
      }
    });
  }

  exportCSV(): void {
    const data = this.appointments();
    if (!data.length) return;

    const headers = ['ID', 'Title', 'Status', 'Total Price', 'Payment Mode', 'Seeker Email', 'Provider Email', 'Scheduled At', 'Created At'];
    const rows = data.map(a => [
      `"${a.id}"`,
      `"${a.title || ''}"`,
      `"${a.status}"`,
      a.total_price || 0,
      `"${a.payment_mode}"`,
      `"${a.seeker_details?.email || a.seeker}"`,
      `"${a.provider_details?.email || a.provider}"`,
      `"${a.appointment_date || ''}"`,
      `"${a.created_at}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `appointments_export_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }

  changePage(newPage: number): void {
    if (newPage >= 1 && newPage <= this.totalPages()) {
      this.currentPage.set(newPage);
      this.loadAppointments();
    }
  }

  changePageSize(newSize: number): void {
    this.pageSize.set(newSize);
    this.currentPage.set(1);
    this.loadAppointments();
  }
}
