import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { StripePaymentRepository } from '../repositories/stripe-payment.repository';
import {
  StripePaymentListResponse,
  StripePaymentDetail,
  IssueRefundInput
} from '../domain/entities/stripe-payment.model';

@Injectable({
  providedIn: 'root'
})
export class StripePaymentUseCase {
  constructor(private stripePaymentRepo: StripePaymentRepository) {}

  listPayments(
    status?: string,
    paymentType?: string,
    search?: string,
    page: number = 1,
    pageSize: number = 20
  ): Observable<StripePaymentListResponse> {
    return this.stripePaymentRepo.listPayments(status, paymentType, search, page, pageSize);
  }

  getPaymentDetail(id: string): Observable<StripePaymentDetail> {
    return this.stripePaymentRepo.getPaymentDetail(id);
  }

  refundPayment(
    id: string,
    input: IssueRefundInput
  ): Observable<{ status: string; message: string; detail: StripePaymentDetail }> {
    return this.stripePaymentRepo.refundPayment(id, input);
  }

  syncPayments(): Observable<{ status: string; message: string; synced: number }> {
    return this.stripePaymentRepo.syncPayments();
  }

  exportPaymentsCSV(status?: string, paymentType?: string): Observable<Blob> {
    return this.stripePaymentRepo.exportPaymentsCSV(status, paymentType);
  }
}

