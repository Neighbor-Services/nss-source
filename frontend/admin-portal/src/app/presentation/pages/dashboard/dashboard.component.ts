import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { DashboardUseCase } from '../../../core/usecases/dashboard.usecase';
import { DashboardStats } from '../../../core/domain/entities/dashboard.model';

export type DashboardMainTab = 'ANALYTICS' | 'LIFECYCLE' | 'CLUSTER_TELEMETRY';

export interface RecentActivityItem {
  id: string;
  type: 'BOOKING' | 'VERIFICATION' | 'PAYOUT' | 'DISPUTE' | 'SUBSCRIPTION';
  title: string;
  description: string;
  timestamp: string;
  status: string;
  statusType: 'success' | 'warning' | 'primary' | 'danger' | 'purple';
  route: string;
}

export interface CategoryBreakdown {
  name: string;
  percentage: number;
  jobsCount: number;
  revenue: number;
  color: string;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  stats = signal<DashboardStats | null>(null);
  isLoading = signal(false);
  errorMessage = signal<string | null>(null);
  activeTab = signal<DashboardMainTab>('ANALYTICS');

  recentActivities = signal<RecentActivityItem[]>([
    {
      id: 'ACT-1092',
      type: 'BOOKING',
      title: 'New Service Appointment Created',
      description: 'Seeker booked "Master Plumbing Inspection" with David A.',
      timestamp: '2 mins ago',
      status: 'MATCHED & ESCROWED',
      statusType: 'primary',
      route: '/admin/appointments'
    },
    {
      id: 'ACT-1091',
      type: 'VERIFICATION',
      title: 'Provider ID Verification Submitted',
      description: 'Sophie M. submitted Government ID & Facial Biometrics for review',
      timestamp: '14 mins ago',
      status: 'PENDING REVIEW',
      statusType: 'warning',
      route: '/admin/verifications'
    },
    {
      id: 'ACT-1090',
      type: 'SUBSCRIPTION',
      title: 'Provider Subscribed to Gold Tier',
      description: 'Marcus L. upgraded subscription to 3 service categories ($39.99/mo)',
      timestamp: '42 mins ago',
      status: 'ACTIVE',
      statusType: 'success',
      route: '/admin/subscriptions'
    },
    {
      id: 'ACT-1089',
      type: 'PAYOUT',
      title: 'Escrow Payout Disbursed',
      description: 'Released $180.00 to Elena R. after Security Code verification',
      timestamp: '1 hour ago',
      status: 'COMPLETED',
      statusType: 'success',
      route: '/admin/payouts'
    }
  ]);

  categoryBreakdowns = signal<CategoryBreakdown[]>([
    { name: 'Plumbing & Water', percentage: 38, jobsCount: 42, revenue: 6420, color: '#3a57e8' },
    { name: 'Electrical & Lighting', percentage: 24, jobsCount: 28, revenue: 4180, color: '#08B1BA' },
    { name: 'House Cleaning & Maid', percentage: 18, jobsCount: 22, revenue: 2950, color: '#1aa053' },
    { name: 'Handyman & Repairs', percentage: 12, jobsCount: 15, revenue: 1840, color: '#f59e0b' },
    { name: 'Landscaping & Lawn Care', percentage: 8, jobsCount: 11, revenue: 1220, color: '#7c3aed' }
  ]);

  constructor(private dashboardUC: DashboardUseCase) {}

  ngOnInit() {
    this.loadStats();
  }

  loadStats() {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.dashboardUC.getDashboardStats().subscribe({
      next: (s) => {
        this.stats.set(s);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.error || err?.message || 'Failed to load dashboard metrics.');
      }
    });
  }
}
