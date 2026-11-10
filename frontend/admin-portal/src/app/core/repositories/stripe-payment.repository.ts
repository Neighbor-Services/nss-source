import { Observable } from 'rxjs';
import {
  StripePaymentListResponse,
  StripePaymentDetail,
  IssueRefundInput
} from '../domain/entities/stripe-payment.model';

export abstract class StripePaymentRepository {
  abstract listPayments(
    status?: string,
    paymentType?: string,
    search?: string,
    page?: number,
    pageSize?: number
  ): Observable<StripePaymentListResponse>;

  abstract getPaymentDetail(id: string): Observable<StripePaymentDetail>;

  abstract refundPayment(
    id: string,
    input: IssueRefundInput
  ): Observable<{ status: string; message: string; detail: StripePaymentDetail }>;

  abstract syncPayments(): Observable<{ status: string; message: string; synced: number }>;

  abstract exportPaymentsCSV(status?: string, paymentType?: string): Observable<Blob>;
}

