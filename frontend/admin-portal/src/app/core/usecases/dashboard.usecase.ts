import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { DashboardRepository } from '../repositories/dashboard.repository';
import { DashboardStats } from '../domain/entities/dashboard.model';

@Injectable({
  providedIn: 'root'
})
export class DashboardUseCase {
  constructor(private dashboardRepo: DashboardRepository) {}

  getDashboardStats(): Observable<DashboardStats> {
    return this.dashboardRepo.getDashboardStats();
  }

  getEscrowSummary(): Observable<any> {
    return this.dashboardRepo.getEscrowSummary();
  }

  getSubscriptionCohortStats(): Observable<any> {
    return this.dashboardRepo.getSubscriptionCohortStats();
  }

  getGeospatialHeatmap(): Observable<any> {
    return this.dashboardRepo.getGeospatialHeatmap();
  }
}

