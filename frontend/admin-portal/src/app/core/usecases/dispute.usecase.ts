import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { DisputeRepository } from '../repositories/dispute.repository';
import { DisputeItem } from '../domain/entities/dispute.model';

@Injectable({
  providedIn: 'root'
})
export class DisputeUseCase {
  constructor(private disputeRepo: DisputeRepository) {}

  listDisputes(status?: string): Observable<DisputeItem[]> {
    return this.disputeRepo.listDisputes(status);
  }

  resolveDispute(id: string, resolution: string, refundAmount?: number): Observable<{ success: boolean }> {
    return this.disputeRepo.resolveDispute(id, resolution, refundAmount);
  }

  rejectDispute(id: string, reason: string): Observable<{ success: boolean }> {
    return this.disputeRepo.rejectDispute(id, reason);
  }
}
