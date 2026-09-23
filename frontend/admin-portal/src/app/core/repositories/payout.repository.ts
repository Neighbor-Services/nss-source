import { Observable } from 'rxjs';
import { PayoutRequestItem, WalletItem, BatchPayoutResult } from '../domain/entities/payout.model';

export abstract class PayoutRepository {
  abstract listPayouts(status?: string): Observable<PayoutRequestItem[]>;
  abstract approvePayout(id: string): Observable<{ success: boolean }>;
  abstract rejectPayout(id: string, reason: string): Observable<{ success: boolean }>;
  abstract listWallets(): Observable<WalletItem[]>;
  abstract adjustWallet(walletId: string, amount: number, reason: string): Observable<{ success: boolean }>;
  abstract batchApprovePayouts(payoutIds: string[]): Observable<BatchPayoutResult>;
  abstract batchRejectPayouts(payoutIds: string[], reason?: string): Observable<BatchPayoutResult>;
}
