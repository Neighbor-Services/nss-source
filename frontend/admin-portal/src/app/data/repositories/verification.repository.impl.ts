import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { VerificationRepository } from '../../core/repositories/verification.repository';
import { VerificationItem } from '../../core/domain/entities/verification.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class VerificationRepositoryImpl implements VerificationRepository {
  constructor(private http: HttpClient) {}

  listVerifications(status?: string): Observable<VerificationItem[]> {
    const url = status
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.verifications}?status=${status}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.verifications}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((v: any) => ({
          id: v.id?.toString() || '',
          userId: v.user_id || v.userId || '',
          userName: v.user_name || (v.user ? `${v.user.first_name || ''} ${v.user.last_name || ''}`.trim() : 'Provider User'),
          userEmail: v.user_email || v.user?.email || '',
          documentType: v.document_type || 'Government ID',
          documentUrl: v.document_url || v.id_front || '',
          selfieUrl: v.selfie_url || v.selfie || '',
          status: v.status || 'pending',
          submittedAt: v.created_at || new Date().toISOString(),
          rejectionReason: v.rejection_reason || ''
        }));
      })
    );
  }

  approveVerification(id: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.verifications}/${id}/approve`, {}).pipe(
      map(() => ({ success: true }))
    );
  }

  rejectVerification(id: string, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.verifications}/${id}/reject`, { reason }).pipe(
      map(() => ({ success: true }))
    );
  }

  batchVerifications(ids: string[], action: string, notes?: string): Observable<{ success: boolean; count: number }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.verificationsBatch}`, {
      ids,
      action,
      notes
    }).pipe(
      map(res => ({ success: true, count: res.count ?? ids.length }))
    );
  }
}
