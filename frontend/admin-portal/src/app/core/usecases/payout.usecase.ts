import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { PayoutRepository } from '../repositories/payout.repository';
import { PayoutRequestItem, WalletItem } from '../domain/entities/payout.model';

@Injectable({
  providedIn: 'root'
})
export class PayoutUseCase {
  constructor(private payoutRepo: PayoutRepository) {}

  listPayouts(status?: string): Observable<PayoutRequestItem[]> {
    return this.payoutRepo.listPayouts(status);
  }

  approvePayout(id: string): Observable<{ success: boolean }> {
    return this.payoutRepo.approvePayout(id);
  }

  rejectPayout(id: string, reason: string): Observable<{ success: boolean }> {
    return this.payoutRepo.rejectPayout(id, reason);
  }

  listWallets(): Observable<WalletItem[]> {
    return this.payoutRepo.listWallets();
  }

  adjustWallet(walletId: string, amount: number, reason: string): Observable<{ success: boolean }> {
    return this.payoutRepo.adjustWallet(walletId, amount, reason);
  }

  batchApprovePayouts(payoutIds: string[]): Observable<any> {
    return this.payoutRepo.batchApprovePayouts(payoutIds);
  }

  batchRejectPayouts(payoutIds: string[], reason?: string): Observable<any> {
    return this.payoutRepo.batchRejectPayouts(payoutIds, reason);
  }
}
