import { Observable } from 'rxjs';
import { AdminAppointment } from '../domain/entities/appointment.model';

export abstract class AppointmentRepository {
  abstract listAppointments(params?: { status?: string; search?: string; page?: number; pageSize?: number }): Observable<{ results: AdminAppointment[]; count: number }>;
  abstract getAppointmentById(id: string): Observable<AdminAppointment>;
  abstract updateAppointmentStatus(id: string, status: string): Observable<{ status: string }>;
}
