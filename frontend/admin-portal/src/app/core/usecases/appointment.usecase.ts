import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { AppointmentRepository } from '../repositories/appointment.repository';
import { AdminAppointment } from '../domain/entities/appointment.model';

@Injectable({
  providedIn: 'root'
})
export class AppointmentUseCase {
  constructor(private appointmentRepo: AppointmentRepository) {}

  listAppointments(params?: { status?: string; search?: string; page?: number; pageSize?: number }): Observable<{ results: AdminAppointment[]; count: number }> {
    return this.appointmentRepo.listAppointments(params);
  }

  getAppointmentById(id: string): Observable<AdminAppointment> {
    return this.appointmentRepo.getAppointmentById(id);
  }

  updateAppointmentStatus(id: string, status: string): Observable<{ status: string }> {
    return this.appointmentRepo.updateAppointmentStatus(id, status);
  }
}
