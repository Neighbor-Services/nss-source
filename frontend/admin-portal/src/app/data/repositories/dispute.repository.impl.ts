import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { DisputeRepository } from '../../core/repositories/dispute.repository';
import { DisputeItem } from '../../core/domain/entities/dispute.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class DisputeRepositoryImpl implements DisputeRepository {
  constructor(private http: HttpClient) {}

  listDisputes(status?: string): Observable<DisputeItem[]> {
    const url = status
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}?status=${status}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((d: any) => ({
          id: d.id?.toString() || '',
          appointmentId: d.appointment_id || d.appointment?.id || 'APT-REF',
          seekerName: d.seeker_name || (d.seeker ? `${d.seeker.first_name || ''} ${d.seeker.last_name || ''}`.trim() : 'Seeker'),
          providerName: d.provider_name || (d.provider ? `${d.provider.first_name || ''} ${d.provider.last_name || ''}`.trim() : 'Provider'),
          reason: d.reason || d.issue_type || 'Dispute',
          description: d.description || '',
          amount: d.amount || d.disputed_amount || 0,
          status: d.status || 'open',
          createdAt: d.created_at || new Date().toISOString(),
          evidenceUrls: d.evidence_urls || [],
          resolutionNotes: d.resolution_notes || ''
        }));
      })
    );
  }

  resolveDispute(id: string, resolution: string, refundAmount?: number): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}/${id}/resolve`, {
      resolution,
      refund_amount: refundAmount,
      notes: resolution
    }).pipe(
      map(() => ({ success: true }))
    );
  }

  rejectDispute(id: string, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}/${id}/reject`, {
      reason,
      notes: reason
    }).pipe(
      map(() => ({ success: true }))
    );
  }
}
