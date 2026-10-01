import { Observable } from 'rxjs';
import { DispatchIncidentSummary, DispatchOverrideRequest, ProviderFleetTelemetry } from '../domain/entities/dispatch.model';

export abstract class DispatchRepository {
  abstract listIncidents(status?: string): Observable<DispatchIncidentSummary[]>;
  abstract overrideIncident(id: string, req: DispatchOverrideRequest): Observable<{ status: string; message: string }>;
  abstract getProviderFleet(): Observable<ProviderFleetTelemetry[]>;
}
