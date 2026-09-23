import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { SupportRepository } from '../repositories/support.repository';
import { ContactMessage, ResolutionReport } from '../domain/entities/support.model';

@Injectable({
  providedIn: 'root'
})
export class SupportUseCase {
  constructor(private supportRepo: SupportRepository) {}

  listMessages(params?: { is_resolved?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ContactMessage[]; count: number }> {
    return this.supportRepo.listMessages(params);
  }

  toggleMessageResolved(id: string, is_resolved: boolean): Observable<{ status: string }> {
    return this.supportRepo.toggleMessageResolved(id, is_resolved);
  }

  deleteMessage(id: string): Observable<{ status: string }> {
    return this.supportRepo.deleteMessage(id);
  }

  listResolutions(params?: { is_reviewed?: boolean; search?: string; page?: number; pageSize?: number }): Observable<{ results: ResolutionReport[]; count: number }> {
    return this.supportRepo.listResolutions(params);
  }

  toggleResolutionReviewed(id: string, is_reviewed: boolean): Observable<{ status: string }> {
    return this.supportRepo.toggleResolutionReviewed(id, is_reviewed);
  }

  deleteResolution(id: string): Observable<{ status: string }> {
    return this.supportRepo.deleteResolution(id);
  }
}
