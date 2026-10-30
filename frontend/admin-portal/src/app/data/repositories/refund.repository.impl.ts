import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { RefundRepository } from '../../core/repositories/refund.repository';
import { RefundRequestItem } from '../../core/domain/entities/refund.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class RefundRepositoryImpl implements RefundRepository {
  constructor(private http: HttpClient) {}

  listRefundRequests(status?: string, page: number = 1, pageSize: number = 20): Observable<{ results: RefundRequestItem[]; count: number }> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.refunds}?page=${page}&page_size=${pageSize}`;
    if (status && status !== 'ALL') {
      url += `&status=${encodeURIComponent(status)}`;
    }

    return this.http.get<any>(url).pipe(
      map(res => {
        const results = (res.results || []).map((r: any) => ({
          id: r.id,
          user_id: r.user_id,
          appointment_id: r.appointment_id,
          payment_intent_id: r.payment_intent_id,
          amount: r.amount || 0,
          currency: r.currency || 'USD',
          reason: r.reason || '',
          status: (r.status || 'PENDING').toUpperCase(),
          stripe_refund_id: r.stripe_refund_id,
          admin_id: r.admin_id,
          admin_notes: r.admin_notes,
          created_at: r.created_at || new Date().toISOString(),
          updated_at: r.updated_at,
          user: r.user,
          appointment: r.appointment
        }));
        return {
          results,
          count: res.count ?? results.length
        };
      })
    );
  }

  getRefundRequest(id: string): Observable<RefundRequestItem> {
    return this.http.get<RefundRequestItem>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.refunds}/${id}`);
  }

  processRefund(id: string, amount?: number, notes?: string): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.refunds}/${id}/process`, {
      amount,
      notes
    });
  }

  rejectRefund(id: string, reason: string): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.refunds}/${id}/reject`, {
      reason
    });
  }

  processDirectRefund(paymentIntentId: string, amount: number, reason?: string, notes?: string, appointmentId?: string): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.refunds}/direct`, {
      payment_intent_id: paymentIntentId,
      amount,
      reason,
      notes,
      appointment_id: appointmentId
    });
  }

  exportRefundsCSV(status?: string): Observable<Blob> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.exportRefunds}`;
    if (status && status !== 'ALL') {
      url += `?status=${encodeURIComponent(status)}`;
    }
    return this.http.get(url, { responseType: 'blob' });
  }
}

