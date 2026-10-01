import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { DispatchRepository } from '../repositories/dispatch.repository';
import { DispatchIncidentSummary, DispatchOverrideRequest, ProviderFleetTelemetry } from '../domain/entities/dispatch.model';

@Injectable({
  providedIn: 'root'
})
export class DispatchUseCase {
  constructor(private repo: DispatchRepository) {}

  listIncidents(status?: string): Observable<DispatchIncidentSummary[]> {
    return this.repo.listIncidents(status);
  }

  overrideIncident(id: string, req: DispatchOverrideRequest): Observable<{ status: string; message: string }> {
    return this.repo.overrideIncident(id, req);
  }

  getProviderFleet(): Observable<ProviderFleetTelemetry[]> {
    return this.repo.getProviderFleet();
  }
}
