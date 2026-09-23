import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AppointmentRepository } from '../../core/repositories/appointment.repository';
import { AdminAppointment } from '../../core/domain/entities/appointment.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class AppointmentRepositoryImpl implements AppointmentRepository {
  private readonly baseUrl = ADMIN_API_CONFIG.baseUrl + ADMIN_API_CONFIG.endpoints.appointments;

  constructor(private http: HttpClient) {}

  listAppointments(params?: { status?: string; search?: string; page?: number; pageSize?: number }): Observable<{ results: AdminAppointment[]; count: number }> {
    let httpParams = new HttpParams();
    if (params?.status) httpParams = httpParams.set('status', params.status);
    if (params?.search) httpParams = httpParams.set('search', params.search);
    if (params?.page) httpParams = httpParams.set('page', params.page.toString());
    if (params?.pageSize) httpParams = httpParams.set('page_size', params.pageSize.toString());

    return this.http.get<{ results: AdminAppointment[]; count: number }>(this.baseUrl, { params: httpParams });
  }

  getAppointmentById(id: string): Observable<AdminAppointment> {
    return this.http.get<AdminAppointment>(`${this.baseUrl}/${id}`);
  }

  updateAppointmentStatus(id: string, status: string): Observable<{ status: string }> {
    return this.http.post<{ status: string }>(`${this.baseUrl}/${id}/status`, { status });
  }
}
