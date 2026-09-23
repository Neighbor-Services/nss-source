import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ResolutionRepository } from '../../core/repositories/resolution.repository';
import { ResolutionReport, ResolutionResponse } from '../../core/domain/entities/resolution.model';
import { API_CONFIG } from '../datasources/api.config';

@Injectable({
  providedIn: 'root'
})
export class ResolutionRepositoryImpl implements ResolutionRepository {
  constructor(private http: HttpClient) {}

  submitResolutionReport(report: ResolutionReport): Observable<ResolutionResponse> {
    return this.http.post<any>(`${API_CONFIG.baseUrl}${API_CONFIG.endpoints.resolution}`, {
      role: report.role,
      issue_type: report.issueType,
      booking_ref: report.bookingRef || '',
      other_neighbor: report.otherNeighbor || '',
      description: report.description,
      expected_outcome: report.expectedOutcome || ''
    }).pipe(
      map(res => ({
        status: 'success' as const,
        message: res?.message || 'Your resolution report has been submitted. Case specialist assigned.',
        caseId: res?.case_id || res?.id || ''
      }))
    );
  }
}
