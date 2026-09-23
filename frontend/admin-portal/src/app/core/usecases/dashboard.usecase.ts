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
}
