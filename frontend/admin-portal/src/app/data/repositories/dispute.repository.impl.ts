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
    const url = status && status !== 'ALL'
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}?status=${status}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((d: any) => {
          const raisedProfile = d.raised_by_user?.profile || d.raised_by_details || {};
          const raisedUser = d.raised_by_user || {};
          const seekerName = (
            raisedProfile.display_name ||
            `${raisedProfile.first_name || raisedUser.first_name || ''} ${raisedProfile.last_name || raisedUser.last_name || ''}`.trim() ||
            d.seeker_name ||
            raisedUser.email ||
            'Seeker Client'
          );

          const defProfile = d.defendant_user?.profile || d.defendant_details || {};
          const defUser = d.defendant_user || {};
          const providerName = (
            defProfile.display_name ||
            `${defProfile.first_name || defUser.first_name || ''} ${defProfile.last_name || defUser.last_name || ''}`.trim() ||
            d.provider_name ||
            defUser.email ||
            'Service Provider'
          );

          const appt = d.appointment_details || d.appointment || {};
          const rawApptId = appt.id || d.appointment || d.appointment_id || '';
          const apptDisplayId = rawApptId ? (rawApptId.length > 8 ? rawApptId.substring(0, 8).toUpperCase() : rawApptId) : (d.id ? d.id.substring(0, 8).toUpperCase() : 'APT-REF');

          const apptPrice = typeof appt.price === 'number' ? appt.price : (typeof appt === 'object' && appt.price ? Number(appt.price) : 0);
          const amount = d.amount || d.disputed_amount || apptPrice || 0;

          const evidenceList: string[] = [];
          if (d.evidence) evidenceList.push(d.evidence);
          if (d.evidence_url) evidenceList.push(d.evidence_url);
          if (Array.isArray(d.evidence_urls)) evidenceList.push(...d.evidence_urls);

          return {
            id: d.id?.toString() || '',
            appointmentId: apptDisplayId,
            appointmentTitle: appt.title || appt.service_name || 'Service Booking',
            seekerId: d.raised_by || raisedUser.id,
            seekerName,
            seekerEmail: raisedUser.email || '',
            seekerPhone: raisedProfile.phone || raisedUser.phone || '',
            seekerAvatar: raisedProfile.avatar || '',
            providerId: d.defendant || defUser.id,
            providerName,
            providerEmail: defUser.email || '',
            providerPhone: defProfile.phone || defUser.phone || '',
            providerAvatar: defProfile.avatar || '',
            reason: d.reason || d.issue_type || 'General Dispute',
            description: d.description || '',
            amount,
            status: (d.status || 'OPEN').toUpperCase(),
            createdAt: d.created_at || new Date().toISOString(),
            updatedAt: d.updated_at || '',
            evidence: d.evidence || d.evidence_url || '',
            evidenceUrls: Array.from(new Set(evidenceList.filter(Boolean))),
            resolutionNotes: d.resolution_notes || ''
          };
        });
      })
    );
  }

  resolveDispute(id: string, resolution: string, refundAmount?: number): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}/${id}/resolve`, {
      notes: resolution,
      resolution,
      refund_amount: refundAmount
    }).pipe(
      map(() => ({ success: true }))
    );
  }

  rejectDispute(id: string, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.disputes}/${id}/reject`, {
      notes: reason,
      reason
    }).pipe(
      map(() => ({ success: true }))
    );
  }
}
