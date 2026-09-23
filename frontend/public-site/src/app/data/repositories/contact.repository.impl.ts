import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ContactRepository } from '../../core/repositories/contact.repository';
import { ContactMessage, ContactResponse } from '../../core/domain/entities/contact.model';
import { API_CONFIG } from '../datasources/api.config';

@Injectable({
  providedIn: 'root'
})
export class ContactRepositoryImpl implements ContactRepository {
  constructor(private http: HttpClient) {}

  submitContactMessage(message: ContactMessage): Observable<ContactResponse> {
    return this.http.post<any>(`${API_CONFIG.baseUrl}${API_CONFIG.endpoints.contact}`, {
      first_name: message.firstName,
      last_name: message.lastName || '',
      email: message.email,
      inquiry_type: message.inquiryType,
      message: message.message
    }).pipe(
      map(res => ({
        status: 'success' as const,
        message: res?.message || 'Your message has been sent successfully. We will get back to you shortly.'
      }))
    );
  }
}
