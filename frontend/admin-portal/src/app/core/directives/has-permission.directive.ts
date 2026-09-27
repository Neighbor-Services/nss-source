import { Directive, Input, TemplateRef, ViewContainerRef, inject, effect } from '@angular/core';
import { AuthService } from '../../data/datasources/auth.service';

export type AdminRole = 'SUPER_ADMIN' | 'SUPPORT_AGENT' | 'KYC_REVIEWER' | 'FINANCE_OFFICER';

export type AdminPermission =
  | 'users:read'
  | 'users:write'
  | 'users:ban'
  | 'verifications:read'
  | 'verifications:write'
  | 'payouts:read'
  | 'payouts:write'
  | 'disputes:read'
  | 'disputes:write'
  | 'emergency:manage'
  | 'system:settings'
  | 'campaigns:send'
  | 'pii:reveal';

const ROLE_PERMISSIONS: Record<AdminRole, AdminPermission[]> = {
  SUPER_ADMIN: [
    'users:read',
    'users:write',
    'users:ban',
    'verifications:read',
    'verifications:write',
    'payouts:read',
    'payouts:write',
    'disputes:read',
    'disputes:write',
    'emergency:manage',
    'system:settings',
    'campaigns:send',
    'pii:reveal'
  ],
  FINANCE_OFFICER: [
    'users:read',
    'payouts:read',
    'payouts:write',
    'disputes:read',
    'disputes:write'
  ],
  KYC_REVIEWER: [
    'users:read',
    'verifications:read',
    'verifications:write',
    'pii:reveal'
  ],
  SUPPORT_AGENT: [
    'users:read',
    'disputes:read',
    'disputes:write',
    'emergency:manage',
    'campaigns:send'
  ]
};

@Directive({
  selector: '[hasPermission]',
  standalone: true
})
export class HasPermissionDirective {
  private authService = inject(AuthService);
  private templateRef = inject(TemplateRef<any>);
  private viewContainer = inject(ViewContainerRef);

  private requiredPermission?: AdminPermission | AdminPermission[];
  private isRendered = false;

  @Input() set hasPermission(permission: AdminPermission | AdminPermission[]) {
    this.requiredPermission = permission;
    this.updateView();
  }

  constructor() {
    effect(() => {
      // Re-evaluate when current user changes
      this.authService.currentUser();
      this.updateView();
    });
  }

  private updateView() {
    const user = this.authService.currentUser();
    if (!user) {
      this.clearView();
      return;
    }

    if (user.isSuperuser || user.is_superuser) {
      this.showView();
      return;
    }

    const role: AdminRole = (user.role as AdminRole) || 'SUPER_ADMIN';
    const permissions = ROLE_PERMISSIONS[role] || ROLE_PERMISSIONS.SUPER_ADMIN;

    if (!this.requiredPermission) {
      this.showView();
      return;
    }

    const required = Array.isArray(this.requiredPermission)
      ? this.requiredPermission
      : [this.requiredPermission];

    const hasAccess = required.some((p) => permissions.includes(p));

    if (hasAccess) {
      this.showView();
    } else {
      this.clearView();
    }
  }

  private showView() {
    if (!this.isRendered) {
      this.viewContainer.createEmbeddedView(this.templateRef);
      this.isRendered = true;
    }
  }

  private clearView() {
    if (this.isRendered) {
      this.viewContainer.clear();
      this.isRendered = false;
    }
  }
}
