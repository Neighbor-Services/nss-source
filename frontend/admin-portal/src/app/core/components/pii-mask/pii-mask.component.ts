import { Component, Input, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuditService } from '../../services/audit.service';

export type PiiType = 'PHONE' | 'EMAIL' | 'GOV_ID' | 'BANK_ACCOUNT' | 'GENERIC';

@Component({
  selector: 'app-pii-mask',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span class="pii-container">
      <span class="pii-text" [class.is-revealed]="isRevealed()">
        {{ isRevealed() ? value : maskedValue }}
      </span>
      <button
        type="button"
        class="pii-toggle-btn"
        (click)="toggleReveal($event)"
        [title]="isRevealed() ? 'Hide sensitive data' : 'Click to reveal (Audit logged)'"
      >
        @if (isRevealed()) {
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
            <line x1="1" y1="1" x2="23" y2="23"/>
          </svg>
        } @else {
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
        }
      </button>
    </span>
  `,
  styles: [`
    .pii-container {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-family: inherit;
    }
    .pii-text {
      color: var(--text-primary, #0f172a);
      letter-spacing: 0.3px;
    }
    .pii-text:not(.is-revealed) {
      font-family: monospace;
      color: var(--text-muted, #64748b);
    }
    .pii-toggle-btn {
      background: rgba(58, 87, 232, 0.08);
      border: 1px solid rgba(58, 87, 232, 0.2);
      border-radius: 4px;
      padding: 2px 5px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      color: #3a57e8;
      cursor: pointer;
      transition: all 0.2s ease;
    }
    .pii-toggle-btn:hover {
      background: rgba(58, 87, 232, 0.18);
      color: #1d39c4;
    }
  `]
})
export class PiiMaskComponent {
  @Input() value: string = '';
  @Input() type: PiiType = 'GENERIC';
  @Input() entityId?: string;

  isRevealed = signal(false);
  private audit = inject(AuditService);

  get maskedValue(): string {
    if (!this.value) return '••••••';
    const val = this.value.trim();

    switch (this.type) {
      case 'PHONE':
        if (val.length < 5) return '•••••';
        return val.slice(0, 3) + ' ••• ••• ' + val.slice(-2);

      case 'EMAIL':
        const parts = val.split('@');
        if (parts.length === 2) {
          const name = parts[0];
          const maskedName = name.length > 2 ? name[0] + '••••' + name[name.length - 1] : '•••';
          return `${maskedName}@${parts[1]}`;
        }
        return '•••••@••••.com';

      case 'BANK_ACCOUNT':
      case 'GOV_ID':
        if (val.length < 5) return '••••••••';
        return '•••••••• ' + val.slice(-4);

      case 'GENERIC':
      default:
        if (val.length <= 4) return '••••';
        return val.slice(0, 2) + '••••' + val.slice(-2);
    }
  }

  toggleReveal(e: Event) {
    e.stopPropagation();
    const nextState = !this.isRevealed();
    this.isRevealed.set(nextState);

    if (nextState) {
      this.audit.logAction({
        actionType: 'REVEAL_PII',
        targetEntity: 'USER_PII',
        targetId: this.entityId || '',
        details: `Admin revealed ${this.type} data: ${this.maskedValue}`
      });
    }
  }
}
