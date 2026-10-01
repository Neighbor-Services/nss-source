import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DispatchRepository } from '../../core/repositories/dispatch.repository';
import { DispatchIncidentSummary, DispatchOverrideRequest, ProviderFleetTelemetry } from '../../core/domain/entities/dispatch.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class DispatchRepositoryImpl implements DispatchRepository {
  constructor(private http: HttpClient) {}

  listIncidents(status?: string): Observable<DispatchIncidentSummary[]> {
    let params = new HttpParams();
    if (status && status !== 'ALL') {
      params = params.set('status', status);
    }
    return this.http.get<DispatchIncidentSummary[]>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.dispatchIncidents}`,
      { params }
    );
  }

  overrideIncident(id: string, req: DispatchOverrideRequest): Observable<{ status: string; message: string }> {
    return this.http.post<{ status: string; message: string }>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.dispatchOverride}/${id}/override`,
      req
    );
  }

  getProviderFleet(): Observable<ProviderFleetTelemetry[]> {
    return this.http.get<ProviderFleetTelemetry[]>(
      `${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.telemetryProviders}`
    );
  }
}
