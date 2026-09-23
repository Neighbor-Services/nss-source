import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { VerificationRepository } from '../repositories/verification.repository';
import { VerificationItem } from '../domain/entities/verification.model';

@Injectable({
  providedIn: 'root'
})
export class VerificationUseCase {
  constructor(private verificationRepo: VerificationRepository) {}

  listVerifications(status?: string): Observable<VerificationItem[]> {
    return this.verificationRepo.listVerifications(status);
  }

  approveVerification(id: string): Observable<{ success: boolean }> {
    return this.verificationRepo.approveVerification(id);
  }

  rejectVerification(id: string, reason: string): Observable<{ success: boolean }> {
    return this.verificationRepo.rejectVerification(id, reason);
  }

  batchVerifications(ids: string[], action: string, notes?: string): Observable<{ success: boolean; count: number }> {
    return this.verificationRepo.batchVerifications(ids, action, notes);
  }
}
