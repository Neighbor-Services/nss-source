import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { PayoutRepository } from '../../core/repositories/payout.repository';
import { PayoutRequestItem, WalletItem } from '../../core/domain/entities/payout.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class PayoutRepositoryImpl implements PayoutRepository {
  constructor(private http: HttpClient) {}

  listPayouts(status?: string): Observable<PayoutRequestItem[]> {
    const url = status
      ? `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payouts}?status=${status}`
      : `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payouts}`;

    return this.http.get<any>(url).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((p: any) => ({
          id: p.id?.toString() || '',
          providerId: p.provider_id || p.provider?.id || '',
          providerName: p.provider_name || (p.provider ? `${p.provider.first_name || ''} ${p.provider.last_name || ''}`.trim() : 'Provider'),
          amount: p.amount || 0,
          status: p.status || 'pending',
          requestedAt: p.created_at || new Date().toISOString(),
          bankAccountLast4: p.bank_last4 || p.bank_account_last4 || 'Direct Bank'
        }));
      })
    );
  }

  approvePayout(id: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payouts}/${id}/approve`, {}).pipe(
      map(() => ({ success: true }))
    );
  }

  rejectPayout(id: string, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payouts}/${id}/reject`, {
      reason,
      notes: reason
    }).pipe(
      map(() => ({ success: true }))
    );
  }

  listWallets(): Observable<WalletItem[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.wallets}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((w: any) => ({
          id: w.id?.toString() || '',
          userId: w.user_id || w.user?.id || '',
          userName: w.user_name || (w.user ? `${w.user.first_name || ''} ${w.user.last_name || ''}`.trim() : 'User'),
          balance: w.balance || 0,
          pendingBalance: w.pending_balance || 0,
          currency: w.currency || 'USD',
          updatedAt: w.updated_at || new Date().toISOString()
        }));
      })
    );
  }

  adjustWallet(walletId: string, amount: number, reason: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.wallets}/${walletId}/adjust`, {
      amount,
      reason
    }).pipe(
      map(() => ({ success: true }))
    );
  }

  batchApprovePayouts(payoutIds: string[]): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payoutsBatchApprove}`, {
      payout_ids: payoutIds
    });
  }

  batchRejectPayouts(payoutIds: string[], reason?: string): Observable<any> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.payoutsBatchReject}`, {
      payout_ids: payoutIds,
      reason: reason || 'Batch rejected by supervisor'
    });
  }
}
