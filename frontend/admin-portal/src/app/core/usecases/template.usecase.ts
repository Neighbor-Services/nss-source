import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { TemplateRepository } from '../repositories/template.repository';
import { NotificationTemplate } from '../domain/entities/template.model';

@Injectable({
  providedIn: 'root'
})
export class TemplateUseCase {
  constructor(private templateRepo: TemplateRepository) {}

  listNotificationTemplates(): Observable<NotificationTemplate[]> {
    return this.templateRepo.listNotificationTemplates();
  }

  updateNotificationTemplate(key: string, data: { subject?: string; bodyHtml?: string; bodyText?: string }): Observable<NotificationTemplate> {
    return this.templateRepo.updateNotificationTemplate(key, data);
  }

  testSendEmailTemplate(key: string, email: string): Observable<{ success: boolean }> {
    return this.templateRepo.testSendEmailTemplate(key, email);
  }
}
