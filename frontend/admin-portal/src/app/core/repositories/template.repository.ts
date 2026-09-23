import { Observable } from 'rxjs';
import { NotificationTemplate } from '../domain/entities/template.model';

export abstract class TemplateRepository {
  abstract listNotificationTemplates(): Observable<NotificationTemplate[]>;
  abstract updateNotificationTemplate(key: string, data: { subject?: string; bodyHtml?: string; bodyText?: string }): Observable<NotificationTemplate>;
  abstract testSendEmailTemplate(key: string, email: string): Observable<{ success: boolean }>;
}
