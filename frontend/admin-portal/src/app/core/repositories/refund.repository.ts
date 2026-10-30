import { Observable } from 'rxjs';
import { RefundRequestItem } from '../domain/entities/refund.model';

export abstract class RefundRepository {
  abstract listRefundRequests(status?: string, page?: number, pageSize?: number): Observable<{ results: RefundRequestItem[]; count: number }>;
  abstract getRefundRequest(id: string): Observable<RefundRequestItem>;
  abstract processRefund(id: string, amount?: number, notes?: string): Observable<any>;
  abstract rejectRefund(id: string, reason: string): Observable<any>;
  abstract processDirectRefund(paymentIntentId: string, amount: number, reason?: string, notes?: string, appointmentId?: string): Observable<any>;
  abstract exportRefundsCSV(status?: string): Observable<Blob>;
}
