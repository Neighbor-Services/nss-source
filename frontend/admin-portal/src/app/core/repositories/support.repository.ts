import { Observable } from 'rxjs';
import { ContactMessage, ResolutionReport } from '../domain/entities/support.model';

export abstract class SupportRepository {
  abstract listMessages(params?: { is_resolved?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ContactMessage[]; count: number }>;
  abstract toggleMessageResolved(id: string, is_resolved: boolean): Observable<{ status: string }>;
  abstract deleteMessage(id: string): Observable<{ status: string }>;

  abstract listResolutions(params?: { is_reviewed?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ResolutionReport[]; count: number }>;
  abstract toggleResolutionReviewed(id: string, is_reviewed: boolean): Observable<{ status: string }>;
  abstract deleteResolution(id: string): Observable<{ status: string }>;
}
