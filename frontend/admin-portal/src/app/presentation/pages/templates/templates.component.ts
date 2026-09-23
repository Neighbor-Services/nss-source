import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { TemplateUseCase } from '../../../core/usecases/template.usecase';
import { NotificationTemplate } from '../../../core/domain/entities/template.model';

export type TemplateSectionTab = 'EDITOR' | 'SMTP_GATEWAY' | 'FCM_PUSH';

@Component({
  selector: 'app-templates',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './templates.component.html',
  styleUrl: './templates.component.css'
})
export class TemplatesComponent implements OnInit {
  templates = signal<NotificationTemplate[]>([]);
  selectedTemplate = signal<NotificationTemplate | null>(null);
  activeTab = signal<'code' | 'preview'>('code');
  mainSection = signal<TemplateSectionTab>('EDITOR');

  loading = signal<boolean>(false);
  saving = signal<boolean>(false);
  sendingTest = signal<boolean>(false);

  testRecipientEmail = '';
  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);

  // Computed KPIs
  totalTemplates = computed(() => this.templates().length);
  emailTemplatesCount = computed(() => this.templates().filter(t => t.channel === 'EMAIL').length);
  pushTemplatesCount = computed(() => this.templates().filter(t => t.channel === 'PUSH').length);

  constructor(
    private templateUseCase: TemplateUseCase,
    private sanitizer: DomSanitizer
  ) {}

  getSafePreviewHtml(): SafeHtml {
    const t = this.selectedTemplate();
    if (!t) return '';
    let preview = t.bodyHtml || t.bodyText || '';
    // Substitute common preview tokens
    preview = preview
      .replace(/\{\{code\}\}/g, '<strong style="color: #6366f1; font-size: 1.2em;">849201</strong>')
      .replace(/\{\{email\}\}/g, 'user@example.com')
      .replace(/\{\{amount\}\}/g, '240.00')
      .replace(/\{\{date\}\}/g, new Date().toLocaleDateString())
      .replace(/\{\{provider_name\}\}/g, 'Alex Rivera')
      .replace(/\{\{customer_name\}\}/g, 'Sarah Jenkins')
      .replace(/\{\{service_title\}\}/g, 'HVAC System Maintenance')
      .replace(/\{\{appointment_date\}\}/g, 'Tomorrow at 10:00 AM')
      .replace(/\{\{address\}\}/g, '742 Evergreen Terrace, Springfield')
      .replace(/\{\{dispute_reason\}\}/g, 'Scope completion disagreement')
      .replace(/\{\{location\}\}/g, 'San Francisco, CA');

    return this.sanitizer.bypassSecurityTrustHtml(preview);
  }

  insertToken(token: string): void {
    const t = this.selectedTemplate();
    if (!t) return;
    const tokenStr = `{{${token}}}`;
    t.bodyHtml = (t.bodyHtml || '') + ` ${tokenStr} `;
    this.successMessage.set(`Inserted placeholder ${tokenStr}`);
    setTimeout(() => this.successMessage.set(null), 2500);
  }

  ngOnInit(): void {
    this.fetchTemplates();
  }

  fetchTemplates(): void {
    this.loading.set(true);
    this.errorMessage.set(null);
    this.templateUseCase.listNotificationTemplates().subscribe({
      next: (data) => {
        this.templates.set(data);
        if (data.length > 0 && !this.selectedTemplate()) {
          this.selectedTemplate.set({ ...data[0] });
        }
        this.loading.set(false);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to load email & notification templates');
        this.loading.set(false);
      }
    });
  }

  selectTemplate(template: NotificationTemplate): void {
    this.selectedTemplate.set({ ...template });
    this.errorMessage.set(null);
    this.successMessage.set(null);
  }

  saveSelectedTemplate(): void {
    const t = this.selectedTemplate();
    if (!t) return;

    this.saving.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.templateUseCase.updateNotificationTemplate(t.key, {
      subject: t.subject,
      bodyHtml: t.bodyHtml,
      bodyText: t.bodyText
    }).subscribe({
      next: (updated) => {
        this.saving.set(false);
        this.selectedTemplate.set({ ...updated });
        this.successMessage.set(`Template "${t.name}" updated successfully!`);
        this.fetchTemplates();
        setTimeout(() => this.successMessage.set(null), 4000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to save template changes');
        this.saving.set(false);
      }
    });
  }

  sendTestEmail(): void {
    const t = this.selectedTemplate();
    if (!t) return;

    if (!this.testRecipientEmail.trim()) {
      this.errorMessage.set('Please enter a valid recipient email address');
      return;
    }

    this.sendingTest.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    this.templateUseCase.testSendEmailTemplate(t.key, this.testRecipientEmail.trim()).subscribe({
      next: () => {
        this.sendingTest.set(false);
        this.successMessage.set(`Test email for "${t.name}" successfully dispatched to ${this.testRecipientEmail}!`);
        setTimeout(() => this.successMessage.set(null), 5000);
      },
      error: (err) => {
        this.errorMessage.set(err.error?.message || 'Failed to send test email');
        this.sendingTest.set(false);
      }
    });
  }
}
