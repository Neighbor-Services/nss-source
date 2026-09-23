import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { BackgroundCheckRepository } from '../../core/repositories/background-check.repository';
import { BackgroundCheckItem } from '../../core/domain/entities/background-check.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class BackgroundCheckRepositoryImpl implements BackgroundCheckRepository {
  constructor(private http: HttpClient) {}

  listBackgroundChecks(): Observable<BackgroundCheckItem[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.backgroundChecks}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((b: any) => ({
          id: b.id?.toString() || '',
          userId: b.user_id || b.user?.id || '',
          userName: b.user_name || (b.user ? `${b.user.first_name || ''} ${b.user.last_name || ''}`.trim() : 'Provider User'),
          userEmail: b.user_email || b.user?.email || '',
          checkrCandidateId: b.candidate_id || b.checkr_candidate_id || 'N/A',
          checkrReportId: b.report_id || b.checkr_report_id || 'N/A',
          status: b.status || 'pending',
          package: b.package_name || b.package || 'Comprehensive Checkr Screening',
          turnaroundTime: b.turnaround || b.turnaround_time || 'Pending',
          completedAt: b.completed_at || '',
          manualOverride: b.manual_override || false
        }));
      })
    );
  }

  overrideBackgroundCheck(id: string, newStatus: string, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.backgroundChecks}/${id}/override`, {
      status: newStatus,
      notes: reason
    }).pipe(
      map(() => ({ success: true }))
    );
  }
}
