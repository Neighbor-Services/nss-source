import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { RefundRepository } from '../repositories/refund.repository';
import { RefundRequestItem } from '../domain/entities/refund.model';

@Injectable({
  providedIn: 'root'
})
export class RefundUseCase {
  constructor(private refundRepo: RefundRepository) {}

  listRefundRequests(status?: string, page?: number, pageSize?: number): Observable<{ results: RefundRequestItem[]; count: number }> {
    return this.refundRepo.listRefundRequests(status, page, pageSize);
  }

  getRefundRequest(id: string): Observable<RefundRequestItem> {
    return this.refundRepo.getRefundRequest(id);
  }

  processRefund(id: string, amount?: number, notes?: string): Observable<any> {
    return this.refundRepo.processRefund(id, amount, notes);
  }

  rejectRefund(id: string, reason: string): Observable<any> {
    return this.refundRepo.rejectRefund(id, reason);
  }

  processDirectRefund(paymentIntentId: string, amount: number, reason?: string, notes?: string, appointmentId?: string): Observable<any> {
    return this.refundRepo.processDirectRefund(paymentIntentId, amount, reason, notes, appointmentId);
  }

  exportRefundsCSV(status?: string): Observable<Blob> {
    return this.refundRepo.exportRefundsCSV(status);
  }
}

