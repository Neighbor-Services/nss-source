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
        return raw.map((v: any) => {
          const providerUser = v.provider_details || v.user;
          const profile = providerUser?.profile;
          let userName = v.user_name || '';
          if (!userName && profile) {
            userName = `${profile.first_name || ''} ${profile.last_name || ''}`.trim();
          }
          if (!userName && providerUser) {
            userName = `${providerUser.first_name || ''} ${providerUser.last_name || ''}`.trim();
          }
          if (!userName) {
            userName = 'Provider User';
          }

          const userEmail = v.user_email || providerUser?.email || '';

          return {
            id: v.id?.toString() || '',
            userId: v.provider || v.provider_id || v.user_id || v.userId || '',
            userName,
            userEmail,
            documentType: v.document_type || 'Driver\'s License',
            documentUrl: v.document_front_url || v.document_front || v.document_url || v.id_front || '',
            documentBackUrl: v.document_back_url || v.document_back || '',
            selfieUrl: v.selfie_url || v.selfie || '',
            tradeLicenseUrl: v.trade_license_url || v.trade_license || '',
            licenseNumber: v.license_number || '',
            status: (v.status || 'pending').toLowerCase() as 'pending' | 'approved' | 'rejected',
            submittedAt: v.created_at || new Date().toISOString(),
            rejectionReason: v.reviewer_notes || v.rejection_reason || ''
          };
        });
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
