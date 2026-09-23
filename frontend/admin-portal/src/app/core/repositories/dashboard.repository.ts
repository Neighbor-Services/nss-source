import { Observable } from 'rxjs';
import { DashboardStats } from '../domain/entities/dashboard.model';

export abstract class DashboardRepository {
  abstract getDashboardStats(): Observable<DashboardStats>;
}
