import { Observable } from 'rxjs';
import { ResolutionReport, ResolutionResponse } from '../domain/entities/resolution.model';

export abstract class ResolutionRepository {
  abstract submitResolutionReport(report: ResolutionReport): Observable<ResolutionResponse>;
}
