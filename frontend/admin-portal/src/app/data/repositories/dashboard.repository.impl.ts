import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { DashboardRepository } from '../../core/repositories/dashboard.repository';
import { DashboardStats } from '../../core/domain/entities/dashboard.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class DashboardRepositoryImpl implements DashboardRepository {
  constructor(private http: HttpClient) {}

  getDashboardStats(): Observable<DashboardStats> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.dashboardStats}`).pipe(
      map(res => ({
        totalUsers: res.total_users ?? res.totalUsers ?? 0,
        activeProviders: res.total_providers ?? res.active_providers ?? res.activeProviders ?? 0,
        activeSeekers: res.total_seekers ?? res.active_seekers ?? res.activeSeekers ?? 0,
        totalStaff: res.total_staff ?? res.totalStaff ?? 0,
        totalVerified: res.total_verified ?? res.totalVerified ?? 0,
        totalSuspended: res.total_suspended ?? res.totalSuspended ?? 0,
        openDisputes: res.open_disputes_count ?? res.open_disputes ?? res.openDisputes ?? 0,
        pendingVerifications: res.pending_verifications ?? res.pendingVerifications ?? 0,
        pendingBackgroundChecks: res.pending_background_checks ?? res.pendingBackground ?? 0,
        totalGMV: res.total_wallet_balance ?? res.total_gmv ?? res.totalGMV ?? 0,
        platformRevenue: res.platform_revenue ?? res.platformRevenue ?? 0,
        activeJobsCount: res.appointments_pending ?? res.active_jobs ?? res.activeJobsCount ?? 0,
        completedJobsCount: res.appointments_completed ?? res.completed_jobs ?? res.completedJobsCount ?? 0,
        activeSubscriptions: res.active_subscriptions ?? 0,
        pendingPayoutsCount: res.pending_payouts_count ?? 0,
        pendingPayoutsAmount: res.pending_payouts_amount ?? 0,
        totalWalletBalance: res.total_wallet_balance ?? 0,
        recentAuditLogs: res.recent_audit_logs ?? [],
        signupsLast30Days: res.signups_last_30_days ?? {}
      }))
    );
  }
}
