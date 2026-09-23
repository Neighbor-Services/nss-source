import { Observable } from 'rxjs';
import { BackgroundCheckItem } from '../domain/entities/background-check.model';

export abstract class BackgroundCheckRepository {
  abstract listBackgroundChecks(): Observable<BackgroundCheckItem[]>;
  abstract overrideBackgroundCheck(id: string, newStatus: string, reason: string): Observable<{ success: boolean }>;
}
