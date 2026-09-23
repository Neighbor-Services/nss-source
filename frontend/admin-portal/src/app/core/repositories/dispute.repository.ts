import { Observable } from 'rxjs';
import { DisputeItem } from '../domain/entities/dispute.model';

export abstract class DisputeRepository {
  abstract listDisputes(status?: string): Observable<DisputeItem[]>;
  abstract resolveDispute(id: string, resolution: string, refundAmount?: number): Observable<{ success: boolean }>;
  abstract rejectDispute(id: string, reason: string): Observable<{ success: boolean }>;
}
