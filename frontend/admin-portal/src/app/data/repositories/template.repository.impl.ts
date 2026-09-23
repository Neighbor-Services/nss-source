import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { TemplateRepository } from '../../core/repositories/template.repository';
import { NotificationTemplate } from '../../core/domain/entities/template.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class TemplateRepositoryImpl implements TemplateRepository {
  constructor(private http: HttpClient) {}

  listNotificationTemplates(): Observable<NotificationTemplate[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.templatesEmails}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((t: any) => ({
          id: t.id?.toString() || '',
          key: t.key || '',
          name: t.name || '',
          channel: t.channel || 'EMAIL',
          subject: t.subject || '',
          bodyHtml: t.body_html || '',
          bodyText: t.body_text || '',
          variables: Array.isArray(t.variables) ? t.variables : (typeof t.variables === 'string' ? JSON.parse(t.variables) : []),
          createdAt: t.created_at || new Date().toISOString()
        }));
      })
    );
  }

  updateNotificationTemplate(key: string, data: { subject?: string; bodyHtml?: string; bodyText?: string }): Observable<NotificationTemplate> {
    return this.http.put<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.templatesEmails}/${key}`, {
      subject: data.subject,
      body_html: data.bodyHtml,
      body_text: data.bodyText
    }).pipe(
      map(t => ({
        id: t.id?.toString() || '',
        key: t.key || key,
        name: t.name || '',
        channel: t.channel || 'EMAIL',
        subject: t.subject || data.subject || '',
        bodyHtml: t.body_html || data.bodyHtml || '',
        bodyText: t.body_text || data.bodyText || '',
        variables: Array.isArray(t.variables) ? t.variables : [],
        createdAt: t.created_at || new Date().toISOString()
      }))
    );
  }

  testSendEmailTemplate(key: string, email: string): Observable<{ success: boolean }> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.templatesTestSend}`, {
      key,
      email
    }).pipe(
      map(() => ({ success: true }))
    );
  }
}
