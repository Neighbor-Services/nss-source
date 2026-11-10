import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { StripePaymentRepository } from '../../core/repositories/stripe-payment.repository';
import {
  StripePaymentListResponse,
  StripePaymentDetail,
  IssueRefundInput
} from '../../core/domain/entities/stripe-payment.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class StripePaymentRepositoryImpl implements StripePaymentRepository {
  constructor(private http: HttpClient) {}

  listPayments(
    status?: string,
    paymentType?: string,
    search?: string,
    page: number = 1,
    pageSize: number = 20
  ): Observable<StripePaymentListResponse> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payments}?page=${page}&page_size=${pageSize}`;
    if (status && status !== 'ALL') {
      url += `&status=${encodeURIComponent(status)}`;
    }
    if (paymentType && paymentType !== 'ALL') {
      url += `&payment_type=${encodeURIComponent(paymentType)}`;
    }
    if (search && search.trim()) {
      url += `&search=${encodeURIComponent(search.trim())}`;
    }
    return this.http.get<StripePaymentListResponse>(url);
  }

  getPaymentDetail(id: string): Observable<StripePaymentDetail> {
    return this.http.get<StripePaymentDetail>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payments}/${id}`
    );
  }

  refundPayment(
    id: string,
    input: IssueRefundInput
  ): Observable<{ status: string; message: string; detail: StripePaymentDetail }> {
    return this.http.post<{ status: string; message: string; detail: StripePaymentDetail }>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payments}/${id}/refund`,
      input
    );
  }

  syncPayments(): Observable<{ status: string; message: string; synced: number }> {
    return this.http.post<{ status: string; message: string; synced: number }>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payments}/sync`,
      {}
    );
  }

  exportPaymentsCSV(status?: string, paymentType?: string): Observable<Blob> {
    let url = `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.exportPayments}?`;
    if (status && status !== 'ALL') {
      url += `status=${encodeURIComponent(status)}&`;
    }
    if (paymentType && paymentType !== 'ALL') {
      url += `payment_type=${encodeURIComponent(paymentType)}&`;
    }
    return this.http.get(url, { responseType: 'blob' });
  }
}

