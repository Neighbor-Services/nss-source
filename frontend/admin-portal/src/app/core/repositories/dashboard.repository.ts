import { Observable } from 'rxjs';
import { DashboardStats } from '../domain/entities/dashboard.model';

export abstract class DashboardRepository {
  abstract getDashboardStats(): Observable<DashboardStats>;
  abstract getEscrowSummary(): Observable<any>;
  abstract getSubscriptionCohortStats(): Observable<any>;
  abstract getGeospatialHeatmap(): Observable<any>;
}

