import { Observable } from 'rxjs';
import { VerificationItem } from '../domain/entities/verification.model';

export abstract class VerificationRepository {
  abstract listVerifications(status?: string): Observable<VerificationItem[]>;
  abstract approveVerification(id: string): Observable<{ success: boolean }>;
  abstract rejectVerification(id: string, reason: string): Observable<{ success: boolean }>;
  abstract batchVerifications(ids: string[], action: string, notes?: string): Observable<{ success: boolean; count: number }>;
}
